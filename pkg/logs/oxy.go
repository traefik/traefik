package logs

import "github.com/rs/zerolog"

type OxyWrapper struct {
	logger zerolog.Logger
}

func NewOxyWrapper(logger zerolog.Logger) *OxyWrapper {
	return &OxyWrapper{logger: logger}
}

func (l OxyWrapper) Debug(s string, i ...interface{}) {
	l.logger.Debug().Msgf(s, i...)
}

func (l OxyWrapper) Info(s string, i ...interface{}) {
	l.logger.Info().Msgf(s, i...)
}

func (l OxyWrapper) Warn(s string, i ...interface{}) {
	l.logger.Warn().Msgf(s, i...)
}

func (l OxyWrapper) Error(s string, i ...interface{}) {
	l.logger.Error().Msgf(s, i...)
}
