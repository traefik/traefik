package logs

import (
	"github.com/rs/zerolog"
)

type InstanaLogger struct {
	logger zerolog.Logger
}

func NewInstanaLogger(logger zerolog.Logger) *InstanaLogger {
	return &InstanaLogger{logger: logger}
}

func (l InstanaLogger) Debug(args ...interface{}) {
	l.logger.Debug().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l InstanaLogger) Info(args ...interface{}) {
	l.logger.Info().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l InstanaLogger) Warn(args ...interface{}) {
	l.logger.Warn().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}

func (l InstanaLogger) Error(args ...interface{}) {
	l.logger.Error().CallerSkipFrame(1).MsgFunc(msgFunc(args...))
}
