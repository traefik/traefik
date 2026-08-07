package logs

import "github.com/rs/zerolog"

type OxyWrapper struct {
	logger zerolog.Logger
}

func NewOxyWrapper(logger zerolog.Logger) *OxyWrapper {
	return &OxyWrapper{logger: logger}
}

func (l OxyWrapper) Debug(s string, i ...any) {
	l.logger.Debug().CallerSkipFrame(1).Msgf(s, i...)
}

func (l OxyWrapper) Info(s string, i ...any) {
	l.logger.Info().CallerSkipFrame(1).Msgf(s, i...)
}

func (l OxyWrapper) Warn(s string, i ...any) {
	l.logger.Warn().CallerSkipFrame(1).Msgf(s, i...)
}

func (l OxyWrapper) Error(s string, i ...any) {
	l.logger.Error().CallerSkipFrame(1).Msgf(s, i...)
}
