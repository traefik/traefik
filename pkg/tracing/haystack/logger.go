package haystack

import (
	"github.com/traefik/traefik/v2/pkg/log"
)

type haystackLogger struct {
	logger log.Logger
}

// Error prints the error message.
func (l haystackLogger) Error(format string, v ...any) {
	l.logger.Errorf(format, v...)
}

// Info prints the info message.
func (l haystackLogger) Info(format string, v ...any) {
	l.logger.Infof(format, v...)
}

// Debug prints the info message.
func (l haystackLogger) Debug(format string, v ...any) {
	l.logger.Debugf(format, v...)
}
