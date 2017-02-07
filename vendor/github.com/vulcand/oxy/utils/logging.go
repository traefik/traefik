package utils

import (
	"io"
	"log"
)

var NullLogger Logger = &NOPLogger{}

// Logger defines a simple logging interface
type Logger interface {
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type FileLogger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewFileLogger(w io.Writer, lvl LogLevel) *FileLogger {
	l := &FileLogger{}
	flag := log.Ldate | log.Ltime | log.Lmicroseconds
	if lvl <= INFO {
		l.info = log.New(w, "INFO: ", flag)
	}
	if lvl <= WARN {
		l.warn = log.New(w, "WARN: ", flag)
	}
	if lvl <= ERROR {
		l.error = log.New(w, "ERR: ", flag)
	}
	return l
}

func (f *FileLogger) Infof(format string, args ...interface{}) {
	if f.info == nil {
		return
	}
	f.info.Printf(format, args...)
}

func (f *FileLogger) Warningf(format string, args ...interface{}) {
	if f.warn == nil {
		return
	}
	f.warn.Printf(format, args...)
}

func (f *FileLogger) Errorf(format string, args ...interface{}) {
	if f.error == nil {
		return
	}
	f.error.Printf(format, args...)
}

type NOPLogger struct {
}

func (*NOPLogger) Infof(format string, args ...interface{}) {

}
func (*NOPLogger) Warningf(format string, args ...interface{}) {
}

func (*NOPLogger) Errorf(format string, args ...interface{}) {
}

func (*NOPLogger) Info(string) {

}
func (*NOPLogger) Warning(string) {
}

func (*NOPLogger) Error(string) {
}

type LogLevel int

const (
	INFO = iota
	WARN
	ERROR
)
