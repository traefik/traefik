package logs

import "github.com/rs/zerolog"

type OxyWrapper struct {
	logger zerolog.Logger
}

func NewOxyWrapper(logger zerolog.Logger) *OxyWrapper {
	return &OxyWrapper{logger: logger}
}

func (l OxyWrapper) Debugf(s string, i ...interface{}) {
	l.logger.Debug().Msgf(s, i...)
}

func (l OxyWrapper) Infof(s string, i ...interface{}) {
	l.logger.Info().Msgf(s, i...)
}

func (l OxyWrapper) Warnf(s string, i ...interface{}) {
	l.logger.Warn().Msgf(s, i...)
}

func (l OxyWrapper) Errorf(s string, i ...interface{}) {
	l.logger.Error().Msgf(s, i...)
}

func (l OxyWrapper) Fatalf(s string, i ...interface{}) {
	l.logger.Fatal().Msgf(s, i...)
}
