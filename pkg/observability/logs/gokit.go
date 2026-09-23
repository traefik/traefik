package logs

import (
	kitlog "github.com/go-kit/log"
	"github.com/rs/zerolog"
)

func NewGoKitWrapper(logger zerolog.Logger) kitlog.LoggerFunc {
	if logger.GetLevel() > zerolog.DebugLevel {
		return func(args ...any) error { return nil }
	}

	return func(args ...any) error {
		logger.Debug().CallerSkipFrame(2).MsgFunc(msgFunc(args...))
		return nil
	}
}
