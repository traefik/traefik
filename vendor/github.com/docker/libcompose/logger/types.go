package logger

import (
	"io"
)

// Factory defines methods a factory should implement, to create a Logger
// based on the specified container, image or service name.
type Factory interface {
	CreateContainerLogger(name string) Logger
	CreateBuildLogger(name string) Logger
	CreatePullLogger(name string) Logger
}

// Logger defines methods to implement for being a logger.
type Logger interface {
	Out(bytes []byte)
	Err(bytes []byte)
	OutWriter() io.Writer
	ErrWriter() io.Writer
}

// Wrapper is a wrapper around Logger that implements the Writer interface,
// mainly use by docker/pkg/stdcopy functions.
type Wrapper struct {
	Err    bool
	Logger Logger
}

func (l *Wrapper) Write(bytes []byte) (int, error) {
	if l.Err {
		l.Logger.Err(bytes)
	} else {
		l.Logger.Out(bytes)
	}
	return len(bytes), nil
}
