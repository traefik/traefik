package logs

import "github.com/rs/zerolog"

type ElasticLogger struct {
	logger zerolog.Logger
}

func NewElasticLogger(logger zerolog.Logger) *ElasticLogger {
	return &ElasticLogger{logger: logger}
}

func (l ElasticLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l ElasticLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}
