package logs

import (
	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
)

func NewAWSWrapper(logger zerolog.Logger) logging.LoggerFunc {
	if logger.GetLevel() > zerolog.DebugLevel {
		return func(classification logging.Classification, format string, args ...any) {}
	}

	return func(classification logging.Classification, format string, args ...any) {
		logger.Debug().CallerSkipFrame(2).MsgFunc(msgFunc(args...))
	}
}
