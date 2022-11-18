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
	l.logger.Debug().MsgFunc(MsgFunc(args...))
}

func (l InstanaLogger) Info(args ...interface{}) {
	l.logger.Info().MsgFunc(MsgFunc(args...))
}

func (l InstanaLogger) Warn(args ...interface{}) {
	l.logger.Warn().MsgFunc(MsgFunc(args...))
}

func (l InstanaLogger) Error(args ...interface{}) {
	l.logger.Error().MsgFunc(MsgFunc(args...))
}
