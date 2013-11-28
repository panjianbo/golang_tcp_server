package common

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"fmt"
	"runtime"
)

var (
	ConfFile string
)

func init() {
	flag.StringVar(&ConfFile, "c", "./config/server.conf", " set server config file path")
}

type Config struct {
	TcpAddr             string `json:"tcp_addr"`
	MonitorAddr         string `json:"monitor_addr"`
	TcpTimeout          int    `json:"tcp_timeout"`
	LogPath             string `json:"log_path"`
	RedisAddr           string `json:"redis_addr"`
}

func IsWindows() bool {
	return (runtime.GOOS == "windows")
}
func InitConfig(file string) (*Config, error) {
	c, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("ioutil.ReadFile(\"%s\") failed (%s)\n", file, err.Error())
		return nil, err
	}

	cf := &Config{
		TcpAddr:             "localhost:8080",
		MonitorAddr:         "localhost:8081",
		TcpTimeout:          300,
		RedisAddr:           "localhost:6379",
	}
	
	if ok := IsWindows(); ok {
		cf.LogPath = "d:\\appdatas\\tcp_server\\logs"
	}else{
		cf.LogPath = "/data/appdatas/tcp_server/logs"
	}
	
	if err = json.Unmarshal(c, cf); err != nil {
		fmt.Printf("json.Unmarshal(\"%s\", cf) failed (%s)\n", string(c), err.Error())
		return nil, err
	}

	return cf, nil
}