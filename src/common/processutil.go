package common

import (
	"os"
)

func KillCurrentProcess() error {
	pid := os.Getpid()
	if pid == 1 { // init provided sockets, for example systemd
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
