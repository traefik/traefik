package proxyprotocol

// Logger interface
type Logger interface {
	Printf(format string, v ...interface{})
}

// LoggerFunc wrap Printf-like function into proxyptocol.Logger
type LoggerFunc func(format string, v ...interface{})

// Printf call inner Printf-link function
func (logf LoggerFunc) Printf(format string, v ...interface{}) {
	logf(format, v...)
}

// FallbackLogger wrap Logger or nil
type FallbackLogger struct {
	Logger
}

// Printf call Printf on inner logger if it not nil
func (wrapper FallbackLogger) Printf(format string, v ...interface{}) {
	if nil == wrapper.Logger {
		return
	}
	wrapper.Logger.Printf(format, v...)
}
