package common

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type ErrorLogger struct {
	dirPath string
	name    string
	lock    sync.Mutex
}

func NewErrorLogger(dirPath string, name string) (*ErrorLogger, error) {
	this := ErrorLogger{}
	this.dirPath = fmt.Sprintf("%s/%s", dirPath, name)
	this.name = name
	EnsureDirExists(this.dirPath)
	return &this, nil
}

func (this *ErrorLogger) Writeln(line string) error {
	t := time.Now()
	fileName := fmt.Sprintf("%s_%d%02d%02d", this.name, t.Year(), t.Month(), t.Day())

	this.lock.Lock()
	defer this.lock.Unlock()

	file, err := os.OpenFile(fmt.Sprintf("%s/%s.log", this.dirPath, fileName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var outputLine string
	if strings.HasSuffix(line, "\n") {
		outputLine = fmt.Sprintf("[%02d:%02d:%02d] %s", t.Hour(), t.Minute(), t.Second(), line)
	} else {
		outputLine = fmt.Sprintf("[%02d:%02d:%02d] %s\n", t.Hour(), t.Minute(), t.Second(), line)
	}
	_, err = file.WriteString(outputLine)
	return err
}

func (this *ErrorLogger) Printf(format string, v ...interface{}) {
	this.Writeln(fmt.Sprintf(format, v...))
}

func (this *ErrorLogger) Println(v ...interface{}) {
	this.Writeln(fmt.Sprintln(v...))
}

func (this *ErrorLogger) Fatalf(format string, v ...interface{}) {
	this.Printf(format, v...)
	os.Exit(-1)
}

func (this *ErrorLogger) Fatalln(v ...interface{}) {
	this.Println(v...)
	os.Exit(-1)
}
