package common

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	TruncateDay    int8 = 1
	TruncateHour   int8 = 2
	TruncateMinute int8 = 3
	flushLineCount int  = 200
)

type DataLogger struct {
	dirPath      string
	name         string
	writingName  string
	truncateMode int8
	lines        chan string
	quit         chan bool
	lock         sync.Mutex
}

func NewDataLogger(dirPath string, name string, truncateMode int8) (*DataLogger, error) {
	this := DataLogger{}
	this.dirPath = fmt.Sprintf("%s/%s", dirPath, name)
	this.name = name
	this.writingName = this.getFileName()
	this.truncateMode = truncateMode
	this.lines = make(chan string, 1000000)
	this.quit = make(chan bool, 1)
	EnsureDirExists(this.dirPath)
	this.autoFlush()
	return &this, nil
}

func (this DataLogger) getFileName() (fileName string) {
	t := time.Now()
	switch this.truncateMode {
	case TruncateDay:
		fileName = fmt.Sprintf("%s_%d%02d%02d0000", this.name, t.Year(), t.Month(), t.Day())
	case TruncateHour:
		fileName = fmt.Sprintf("%s_%d%02d%02d%02d00", this.name, t.Year(), t.Month(), t.Day(), t.Hour())
	case TruncateMinute:
		fileName = fmt.Sprintf("%s_%d%02d%02d%02d%02d", this.name, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	default:
		fileName = this.name
	}
	return
}

func (this *DataLogger) Writeln(line string) {
	if strings.HasSuffix(line, "\n") {
		this.lines <- line
	} else {
		this.lines <- fmt.Sprintf("%s\n", line)
	}
}

func (this *DataLogger) Flush() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	lineCount := len(this.lines)

	if lineCount == 0 {
		return nil
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.log", this.dirPath, this.getFileName()), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := bytes.Buffer{}
	for i := 1; i <= lineCount; i++ {
		buffer.WriteString(<-this.lines)
		if i%flushLineCount == 0 {
			file.WriteString(buffer.String())
			buffer = bytes.Buffer{}
		}
	}
	if buffer.Len() > 0 {
		file.WriteString(buffer.String())
	}
	return nil
}

func (this *DataLogger) Close() {
	this.Flush()
	this.quit <- true
}

func (this *DataLogger) autoFlush() {
	go func() {
		timer := time.NewTicker(time.Second)
		for {
			select {
			case <-timer.C:
				func() {
					lineCount := len(this.lines)
					if (lineCount >= flushLineCount) || (lineCount > 0 && this.writingName != this.getFileName()) {
						if err := this.Flush(); err != nil {
							panic("failed Flush")
						}
					}
				}()
			case <-this.quit:
				break
			}
		}
	}()
}
