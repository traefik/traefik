package logs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/rs/zerolog"
)

func NewAWSWrapper(logger zerolog.Logger) aws.LoggerFunc {
	if logger.GetLevel() > zerolog.DebugLevel {
		return func(args ...interface{}) {}
	}

	return func(args ...interface{}) {
		logger.Debug().CallerSkipFrame(2).MsgFunc(msgFunc(args...))
	}
}
