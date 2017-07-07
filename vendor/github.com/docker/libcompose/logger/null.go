package logger

import (
	"io"
)

// NullLogger is a logger.Logger and logger.Factory implementation that does nothing.
type NullLogger struct {
}

// Out is a no-op function.
func (n *NullLogger) Out(_ []byte) {
}

// Err is a no-op function.
func (n *NullLogger) Err(_ []byte) {
}

// CreateContainerLogger allows NullLogger to implement logger.Factory.
func (n *NullLogger) CreateContainerLogger(_ string) Logger {
	return &NullLogger{}
}

// CreateBuildLogger allows NullLogger to implement logger.Factory.
func (n *NullLogger) CreateBuildLogger(_ string) Logger {
	return &NullLogger{}
}

// CreatePullLogger allows NullLogger to implement logger.Factory.
func (n *NullLogger) CreatePullLogger(_ string) Logger {
	return &NullLogger{}
}

// OutWriter returns the base writer
func (n *NullLogger) OutWriter() io.Writer {
	return nil
}

// ErrWriter returns the base writer
func (n *NullLogger) ErrWriter() io.Writer {
	return nil
}
