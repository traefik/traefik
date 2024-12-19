package logs

import (
	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
)

func NewAWSWrapper(logger zerolog.Logger) logging.LoggerFunc {
	if logger.GetLevel() > zerolog.DebugLevel {
		return logging.Nop{}.Logf
	}

	return func(classification logging.Classification, format string, v ...interface{}) {
		logger.Debug().CallerSkipFrame(2).Msgf(format, v...)
	}
}
