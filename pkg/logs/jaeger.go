package logs

import (
	"github.com/rs/zerolog"
)

// JaegerLogger is an implementation of the Logger interface that delegates to traefik log.
type JaegerLogger struct {
	logger zerolog.Logger
}

func NewJaegerLogger(logger zerolog.Logger) *JaegerLogger {
	return &JaegerLogger{logger: logger}
}

func (l *JaegerLogger) Error(msg string) {
	l.logger.Error().CallerSkipFrame(1).Msg(msg)
}

// Infof logs a message at debug priority.
func (l *JaegerLogger) Infof(msg string, args ...interface{}) {
	l.logger.Debug().CallerSkipFrame(1).Msgf(msg, args...)
}
