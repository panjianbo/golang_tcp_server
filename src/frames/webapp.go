package frames

import (
	"encoding/base64"
	"net/http"
	"io"
	"time"
	"common"
	"fmt"
	"strconv"
)

type handlerMonitor struct {
	listener *controledListener
	errLogger *common.ErrorLogger
}

func (self *handlerMonitor) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	
	if req.URL.Path == "/status/" {
		
		io.WriteString(resp, "<html><body>")
		io.WriteString(resp, "<p align=\"center\">连接总数："+ strconv.FormatInt(self.listener.GetCount(), 10))
		io.WriteString(resp, "；当前连接数：" + strconv.Itoa(self.listener.clientDict.Len()) + "</p>")
		if detail := req.FormValue("detail"); detail == "1"{
			io.WriteString(resp, "<table border=\"1\" align=\"center\"><tr><td>客户端地址</td><td>首次连接时间</td><td>最后收发时间</td><td>操作</td></tr>")
			items := self.listener.clientDict.Items()
			for key, value := range items {
				io.WriteString(resp, "<tr><td>" + key.(string) + "</td>")
				connect := value.(*workerConn)
				io.WriteString(resp, "<td>" + connect.timefirst.Format("2006-01-02 15:04:05") + "</td>")
				io.WriteString(resp, "<td>" + connect.timelast.Format("2006-01-02 15:04:05") + "</td>")
				io.WriteString(resp, "<td><button onclick=alert('"+key.(string)+"')>关闭</button></td>")
			}
			io.WriteString(resp, "</table>")
		}
		io.WriteString(resp, "</body></html>")
	} else if req.URL.Path == "/command/" {
		addr := req.FormValue("addr")
		if addr == ""{
			fmt.Fprintln(resp, "parameter addr not found!")
			return
		}
		
		connect := self.listener.clientDict.Get(addr)
		if connect == nil{
			fmt.Fprintln(resp, "parameter addr not connected!")
			return
		}
		
		commandType := req.FormValue("ct")
		if commandType == "close" {
			connect.(*workerConn).conn.Close()
			fmt.Fprintln(resp, "command close execute ok!")
		}else if commandType == "send"{
			data := req.FormValue("data")
			if data == ""{
				fmt.Fprintln(resp, "command parameter data not found!")
				return
			}
			connect.(*workerConn).conn.Write([]byte(data))
			fmt.Fprintln(resp, "command send execute ok!")
		}else if commandType == "send_base64"{
			data := req.FormValue("data")
			if data == ""{
				fmt.Fprintln(resp, "command parameter data not found!")
				return
			}
			decodeData, err := base64.StdEncoding.DecodeString(data)
			if err != nil{
				fmt.Fprintln(resp, "command parameter data decode error!")
				return
			}
			connect.(*workerConn).conn.Write(decodeData)
			fmt.Fprintln(resp, "command send execute ok!")
		}else{
			fmt.Fprintln(resp, "command not allowed")
		}
	} else {
		fmt.Fprintln(resp, "not allow")
	}
}

func StartHttpMonitor(monitor *controledListener, listener *controledListener, errLogger *common.ErrorLogger){
 	handler := new(handlerMonitor)
 	handler.listener = listener
 	handler.errLogger = errLogger
 	
	serve := &http.Server{
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	serve.Serve(monitor)
}

