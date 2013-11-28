package frames

import (
	"business"
	"bytes"
	"common"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type workerConn struct {
	conn      net.Conn
	raddr     net.Addr
	buff      *bytes.Buffer
	timeout   time.Duration
	timefirst time.Time
	timelast  time.Time
	listener  *controledListener
	errLogger *common.ErrorLogger
}

type Services struct {
	wg *sync.WaitGroup
}

func newWorkConn(conn net.Conn, listener *controledListener, errLogger *common.ErrorLogger) (worker *workerConn, err error) {
	worker = new(workerConn)
	worker.conn = conn
	worker.raddr = conn.RemoteAddr()
	worker.buff = new(bytes.Buffer)
	worker.timeout = time.Second * 60 * 4
	worker.timefirst = time.Now()
	worker.timelast = time.Now()
	worker.errLogger = errLogger
	worker.listener = listener
	return worker, nil
}

func (worker *workerConn) RunWorker() {
	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			worker.errLogger.Println("panic serving:", worker.conn.RemoteAddr(), err, buf)
		}
		worker.conn.Close()
	}()

	for {
		if d := worker.timeout; d != 0 {
			worker.conn.SetDeadline(time.Now().Add(d))
		}

		buf := make([]byte, 2048)
		count, err := worker.conn.Read(buf)
		if err != nil {
			worker.errLogger.Println("read data error:", worker.conn.RemoteAddr(), err)
			if err == io.EOF {
				break
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				break
			}
			break
		}
		worker.timelast = time.Now()
		worker.buff.Write(buf[:count])

		response, err := business.AnalysisInputData(worker.buff)
		if err != nil {
			break
		}

		if response == nil {
			continue
		}
		for _, buff := range response {
			count, err = worker.conn.Write(buff)
			if err != nil {
				worker.errLogger.Println("write data error:", worker.conn.RemoteAddr(), err)
				break
		}
		}
		
	}
	worker.listener.clientDict.Del(worker.raddr.String())
	worker.errLogger.Println("client gone:", worker.raddr.String(), ", left count:", worker.listener.clientDict.Len())
}

func (services *Services) RunTcpServe(listener *controledListener, errLogger *common.ErrorLogger) {
	defer services.wg.Done()
	defer listener.Close()

	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				errLogger.Println("http: Accept error:", err, "retry:", tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			errLogger.Println("http: Accept error:;", err)
			break
		}
		tempDelay = 0

		worker, err := newWorkConn(conn, listener, errLogger)
		if err != nil {
			continue
		}

		listener.clientDict.Set(conn.RemoteAddr().String(), worker)
		listener.count += 1

		go worker.RunWorker()
	}
	//监听socket关闭后，在这里等所有连接结束
	listener.wg.Wait()
}

func (services *Services) RunHttpServe(monitor *controledListener, listener *controledListener, errLogger *common.ErrorLogger) {
	defer services.wg.Done()
	StartHttpMonitor(monitor, listener, errLogger)
}

func RunServer(addr_tcp string, addr_monitor string, errLogger *common.ErrorLogger) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	errLogger.Println("program start.")

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr_tcp)
	if err != nil {
		errLogger.Fatalln("ResolveTCPAddr:", err)
	}

	controledListener, err := GetControledListener(tcpAddr)
	if err != nil {
		errLogger.Fatalln("GetControledListener:", err)
	}
	
	tcpMonitor, err := net.ResolveTCPAddr("tcp", addr_monitor)
	if err != nil {
		errLogger.Fatalln("ResolveMonitorAddr:", err)
	}

	controledMonitor, err := GetControledListener(tcpMonitor)
	if err != nil {
		errLogger.Fatalln("GetControledMonitorListener:", err)
	}
	
	//强杀
	common.SignalWatchRegister(func() { os.Exit(-1) }, syscall.SIGTERM)
	//和平结束
	common.SignalWatchRegister(func() { controledListener.Close(); controledMonitor.Close()}, os.Interrupt)
	//重启
	common.SignalWatchRegister(func() { ReceiveSignupSignal(controledListener, errLogger) }, syscall.SIGHUP)

	common.SignalWatchRun()

	services := Services{wg: &sync.WaitGroup{}}

	//启动tcp服务
	services.wg.Add(1)
	go services.RunTcpServe(controledListener, errLogger)

	services.wg.Add(1)
	go services.RunHttpServe(controledMonitor, controledListener, errLogger) 
	//等等所有服务协程都结束
	services.wg.Wait()

	errLogger.Println("program stop normal.")
}
