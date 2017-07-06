package logger

import (
	"fmt"
	"io"
	"os"
)

// RawLogger is a logger.Logger and logger.Factory implementation that prints raw data with no formatting.
type RawLogger struct {
}

// Out is a no-op function.
func (r *RawLogger) Out(message []byte) {
	fmt.Print(string(message))

}

// Err is a no-op function.
func (r *RawLogger) Err(message []byte) {
	fmt.Fprint(os.Stderr, string(message))

}

// CreateContainerLogger allows RawLogger to implement logger.Factory.
func (r *RawLogger) CreateContainerLogger(_ string) Logger {
	return &RawLogger{}
}

// CreateBuildLogger allows RawLogger to implement logger.Factory.
func (r *RawLogger) CreateBuildLogger(_ string) Logger {
	return &RawLogger{}
}

// CreatePullLogger allows RawLogger to implement logger.Factory.
func (r *RawLogger) CreatePullLogger(_ string) Logger {
	return &RawLogger{}
}

// OutWriter returns the base writer
func (r *RawLogger) OutWriter() io.Writer {
	return os.Stdout
}

// ErrWriter returns the base writer
func (r *RawLogger) ErrWriter() io.Writer {
	return os.Stderr
}
