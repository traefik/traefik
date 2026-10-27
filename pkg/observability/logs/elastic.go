package logs

import "github.com/rs/zerolog"

type ElasticLogger struct {
	logger zerolog.Logger
}

func NewElasticLogger(logger zerolog.Logger) *ElasticLogger {
	return &ElasticLogger{logger: logger}
}

func (l ElasticLogger) Debugf(format string, args ...any) {
	l.logger.Debug().CallerSkipFrame(1).Msgf(format, args...)
}

func (l ElasticLogger) Errorf(format string, args ...any) {
	l.logger.Error().CallerSkipFrame(1).Msgf(format, args...)
}
