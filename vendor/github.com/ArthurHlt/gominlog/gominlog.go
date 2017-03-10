package gominlog

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
)

type Level int

const (
	Loff     = Level(^uint(0) >> 1)
	Lsevere  = Level(1000)
	Lerror   = Level(900)
	Lwarning = Level(800)
	Linfo    = Level(700)
	Ldebug   = Level(600)
	Lall     = Level(-Loff - 1)
)

type MinLog struct {
	log         *log.Logger
	level       Level
	packageName string
	isColorized bool
}

func NewClassicMinLog() *MinLog {
	minLog := &MinLog{}
	logWriter := os.Stdout
	flags := log.Lshortfile | log.Ldate | log.Ltime
	minLog.log = log.New(logWriter, "", flags)
	minLog.isColorized = true
	minLog.packageName = ""
	minLog.level = Lall
	return minLog
}
func NewClassicMinLogWithPackageName(packageName string) *MinLog {
	minLog := NewClassicMinLog()
	minLog.SetPackageName(packageName)
	return minLog
}
func NewMinLog(appName string, level Level, withColor bool, flag int) *MinLog {
	minLog := &MinLog{}
	logWriter := os.Stdout
	minLog.log = log.New(logWriter, "", flag)
	minLog.isColorized = withColor
	minLog.packageName = appName
	minLog.level = level
	return minLog
}
func NewMinLogWithLogger(packageName string, level Level, withColor bool, logger *log.Logger) *MinLog {
	minLog := &MinLog{}
	minLog.log = logger
	minLog.isColorized = withColor
	minLog.packageName = packageName
	minLog.level = level
	return minLog
}
func (this *MinLog) GetLevel() Level {
	return Level(this.level)
}

func (this *MinLog) SetWriter(writer io.Writer) {
	this.log.SetOutput(writer)
}

func (this *MinLog) SetLevel(level Level) {
	this.level = level
}
func (this *MinLog) SetPackageName(newPackageName string) {
	this.packageName = newPackageName
}
func (this *MinLog) GetPackageName() string {
	return this.packageName
}
func (this *MinLog) SetLogger(l *log.Logger) {
	this.log = l
}
func (this *MinLog) WithColor(isColorized bool) {
	this.isColorized = isColorized
}
func (this *MinLog) IsColorized() bool {
	return this.isColorized
}
func (this *MinLog) GetLogger() *log.Logger {
	return this.log
}

func (this *MinLog) logMessage(typeLog string, colorFg ct.Color, colorBg ct.Color, args ...interface{}) {
	var text string
	msg := ""
	flags := this.log.Flags()
	if (log.Lshortfile | flags) == flags {
		msg += this.trace()
		this.log.SetFlags(flags - log.Lshortfile)
	}
	text, ok := args[0].(string)
	if !ok {
		panic("Firt argument should be a string")
	}
	if len(args) > 1 {
		newArgs := args[1:]
		msg += typeLog + ": " + fmt.Sprintf(text, newArgs...)
	} else {
		msg += typeLog + ": " + text
	}
	this.writeMsgInLogger(msg, colorFg, colorBg)
	this.log.SetFlags(flags)
}
func (this *MinLog) writeMsgInLogger(msg string, colorFg ct.Color, colorBg ct.Color) {
	if this.isColorized && colorFg > 0 {
		ct.Foreground(colorFg, false)
	}
	if this.isColorized && colorBg > 0 {
		ct.ChangeColor(colorFg, false, colorBg, false)
	}
	this.log.Print(msg)
	if this.isColorized {
		ct.ResetColor()
	}
}
func (this *MinLog) Error(args ...interface{}) {
	if this.level > Lerror {
		return
	}
	this.logMessage("ERROR", ct.Red, 0, args...)
}

func (this *MinLog) Severe(args ...interface{}) {
	if this.level > Lsevere {
		return
	}
	this.logMessage("SEVERE", ct.Red, ct.Yellow, args...)
}

func (this *MinLog) Debug(args ...interface{}) {
	if this.level > Ldebug {
		return
	}
	this.logMessage("DEBUG", ct.Blue, 0, args...)
}

func (this *MinLog) Info(args ...interface{}) {
	if this.level > Linfo {
		return
	}
	this.logMessage("INFO", ct.Cyan, 0, args...)
}

func (this *MinLog) Warning(args ...interface{}) {
	if this.level > Lwarning {
		return
	}
	this.logMessage("WARNING", ct.Yellow, 0, args...)
}
func (this *MinLog) trace() string {
	var shortFile string
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[2])
	file, line := f.FileLine(pc[2])
	if this.packageName == "" {
		execFileSplit := strings.Split(os.Args[0], "/")
		this.packageName = execFileSplit[len(execFileSplit)-1]
	}
	regex, err := regexp.Compile(regexp.QuoteMeta(this.packageName) + "/(.*)")
	if err != nil {
		panic(err)
	}
	subMatch := regex.FindStringSubmatch(file)
	if len(subMatch) < 2 {
		fileSplit := strings.Split(file, "/")
		shortFile = fileSplit[len(fileSplit)-1]
	} else {
		shortFile = subMatch[1]
	}

	return fmt.Sprintf("/%s/%s:%d ", this.packageName, shortFile, line)
}
