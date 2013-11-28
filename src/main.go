package main 

import (
	"fmt"
	"flag"
	"common"
	"frames"
)

var(
	Conf *common.Config
)

func init(){
	
}

func main() {
	var err error
	flag.Parse()
	Conf, err = common.InitConfig(common.ConfFile)
	 if err != nil {
		fmt.Printf("initConfig(\"%s\") failed (%s)\n", common.ConfFile, err.Error())
		return
	}
	errLogger, _ := common.NewErrorLogger(Conf.LogPath, "error")
	
	frames.RunServer(Conf.TcpAddr, Conf.MonitorAddr, errLogger)
}

