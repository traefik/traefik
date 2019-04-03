package zipkintracer

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

// ErrMissingValue adds a Missing Value Error when the Logging Parameters are
// not even in number
var ErrMissingValue = errors.New("(MISSING)")

// Logger is the fundamental interface for all log operations. Log creates a
// log event from keyvals, a variadic sequence of alternating keys and values.
// The signature is compatible with the Go kit log package.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// NewNopLogger provides a Logger that discards all Log data sent to it.
func NewNopLogger() Logger {
	return &nopLogger{}
}

// LogWrapper wraps a standard library logger into a Logger compatible with this
// package.
func LogWrapper(l *log.Logger) Logger {
	return &wrappedLogger{l: l}
}

// wrappedLogger implements Logger
type wrappedLogger struct {
	l *log.Logger
}

// Log implements Logger
func (l *wrappedLogger) Log(k ...interface{}) error {
	if len(k)%2 == 1 {
		k = append(k, ErrMissingValue)
	}
	o := make([]string, len(k)/2)
	for i := 0; i < len(k); i += 2 {
		o[i/2] = fmt.Sprintf("%s=%q", k[i], k[i+1])
	}
	l.l.Println(strings.Join(o, " "))
	return nil
}

// nopLogger implements Logger
type nopLogger struct{}

// Log implements Logger
func (*nopLogger) Log(_ ...interface{}) error { return nil }

// LoggerFunc is an adapter to allow use of ordinary functions as Loggers. If
// f is a function with the appropriate signature, LoggerFunc(f) is a Logger
// object that calls f.
type LoggerFunc func(...interface{}) error

// Log implements Logger by calling f(keyvals...).
func (f LoggerFunc) Log(keyvals ...interface{}) error {
	return f(keyvals...)
}
