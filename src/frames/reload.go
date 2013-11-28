package frames

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"sync"
	"common"
)

const (
	FDKey = "TCP_SERVER_LISTEN_FD"
)

// Allows for us to notice when the connection is closed.
type conn struct {
	net.Conn
	closed bool
	wg     *sync.WaitGroup
	lock   sync.Mutex
}

type controledListener struct {
	net.Listener
	count   int64
	clientDict *common.RWLockDict
	wg      sync.WaitGroup
}

func initListener(tcpaddr *net.TCPAddr) (listener net.Listener, err error) {
	countStr := os.Getenv(FDKey)
	if countStr == "" {
		return net.ListenTCP("tcp", tcpaddr)
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(count), "listen socket")
	listener, err = net.FileListener(f)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func GetControledListener(tcpAddr *net.TCPAddr) (*controledListener, error) {
	listener, err := initListener(tcpAddr)
	if err != nil {
		return nil, err
	}
	return &controledListener{Listener: listener, clientDict: common.NewRWLockDict()}, nil
}

func ReceiveSignupSignal(controled *controledListener, errLogger *common.ErrorLogger) {
	err := restart(controled.Listener, errLogger)
	if err == nil {
		controled.Listener.Close()
	}
}

func (controled *controledListener) GetCount() int64 {
	return controled.count
}

func (controled *controledListener) Accept() (c net.Conn, err error) {
	c, err = controled.Listener.Accept()
	if err != nil {
		return
	}

	controled.wg.Add(1)
	// Wrap the returned connection, so that we can observe when it is closed.
	c = conn{Conn: c, wg: &controled.wg}
	return
}

func (c conn) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.Conn.Close()
	if !c.closed && err == nil {
		c.wg.Done()
		c.closed = true
	}
	return err
}

// Re-exec this image without dropping the listener passed to this function.
func restart(listener net.Listener, errLogger *common.ErrorLogger) error {
	argv0, err := exec.LookPath(os.Args[0])
	if nil != err {
		return err
	}
	wd, err := os.Getwd()
	if nil != err {
		return err
	}
	v := reflect.ValueOf(listener).Elem().FieldByName("fd").Elem()
	fd := uintptr(v.FieldByName("sysfd").Int())
	allFiles := append([]*os.File{os.Stdin, os.Stdout, os.Stderr},
		os.NewFile(fd, string(v.FieldByName("sysfile").String())))

	p, err := os.StartProcess(argv0, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   append(os.Environ(), fmt.Sprintf("%s=%d", FDKey, fd)),
		Files: allFiles,
	})
	if nil != err {
		return err
	}
	errLogger.Printf("spawned child %d\n", p.Pid)
	return nil
}
