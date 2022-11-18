package logs

import (
	"github.com/rs/zerolog"
)

type HaystackLogger struct {
	logger zerolog.Logger
}

func NewHaystackLogger(logger zerolog.Logger) *HaystackLogger {
	return &HaystackLogger{logger: logger}
}

// Error prints the error message.
func (l HaystackLogger) Error(format string, v ...interface{}) {
	l.logger.Error().CallerSkipFrame(1).Msgf(format, v...)
}

// Info prints the info message.
func (l HaystackLogger) Info(format string, v ...interface{}) {
	l.logger.Info().CallerSkipFrame(1).Msgf(format, v...)
}

// Debug prints the info message.
func (l HaystackLogger) Debug(format string, v ...interface{}) {
	l.logger.Debug().CallerSkipFrame(1).Msgf(format, v...)
}
