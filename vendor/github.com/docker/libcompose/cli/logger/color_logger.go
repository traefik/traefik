package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/docker/libcompose/logger"
	"golang.org/x/crypto/ssh/terminal"
)

// ColorLoggerFactory implements logger.Factory interface using ColorLogger.
type ColorLoggerFactory struct {
	maxLength int
	tty       bool
}

// ColorLogger implements logger.Logger interface with color support.
type ColorLogger struct {
	name        string
	colorPrefix string
	factory     *ColorLoggerFactory
}

// NewColorLoggerFactory creates a new ColorLoggerFactory.
func NewColorLoggerFactory() *ColorLoggerFactory {
	return &ColorLoggerFactory{
		tty: terminal.IsTerminal(int(os.Stdout.Fd())),
	}
}

// CreateContainerLogger implements logger.Factory.CreateContainerLogger.
func (c *ColorLoggerFactory) CreateContainerLogger(name string) logger.Logger {
	return c.create(name)
}

// CreateBuildLogger implements logger.Factory.CreateBuildLogger.
func (c *ColorLoggerFactory) CreateBuildLogger(name string) logger.Logger {
	return &logger.RawLogger{}
}

// CreatePullLogger implements logger.Factory.CreatePullLogger.
func (c *ColorLoggerFactory) CreatePullLogger(name string) logger.Logger {
	return &logger.NullLogger{}
}

// CreateBuildLogger implements logger.Factory.CreateContainerLogger.
func (c *ColorLoggerFactory) create(name string) logger.Logger {
	if c.maxLength < len(name) {
		c.maxLength = len(name)
	}

	return &ColorLogger{
		name:        name,
		factory:     c,
		colorPrefix: <-colorPrefix,
	}
}

// Out implements logger.Logger.Out.
func (c *ColorLogger) Out(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	logFmt, name := c.getLogFmt()
	message := fmt.Sprintf(logFmt, name, string(bytes))
	fmt.Print(message)
}

// Err implements logger.Logger.Err.
func (c *ColorLogger) Err(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	logFmt, name := c.getLogFmt()
	message := fmt.Sprintf(logFmt, name, string(bytes))
	fmt.Fprint(os.Stderr, message)
}

// OutWriter returns the base writer
func (c *ColorLogger) OutWriter() io.Writer {
	return os.Stdout
}

// ErrWriter returns the base writer
func (c *ColorLogger) ErrWriter() io.Writer {
	return os.Stderr
}

func (c *ColorLogger) getLogFmt() (string, string) {
	pad := c.factory.maxLength

	logFmt := "%s | %s"
	if c.factory.tty {
		logFmt = c.colorPrefix + " %s"
	}

	name := fmt.Sprintf("%-"+strconv.Itoa(pad)+"s", c.name)

	return logFmt, name
}
