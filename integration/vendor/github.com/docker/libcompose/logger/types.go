package logger

// Factory defines methods a factory should implement, to create a Logger
// based on the specified name.
type Factory interface {
	Create(name string) Logger
}

// Logger defines methods to implement for being a logger.
type Logger interface {
	Out(bytes []byte)
	Err(bytes []byte)
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
