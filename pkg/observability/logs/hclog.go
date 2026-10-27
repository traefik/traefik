package logs

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rs/zerolog"
)

// RetryableHTTPLogger wraps our logger and implements retryablehttp.LeveledLogger.
// The retry library sends fields as pairs of keys and values as structured logging,
// so we need to adapt them to our logger.
type RetryableHTTPLogger struct {
	logger zerolog.Logger
}

// NewRetryableHTTPLogger creates an implementation of the retryablehttp.LeveledLogger.
func NewRetryableHTTPLogger(logger zerolog.Logger) *RetryableHTTPLogger {
	return &RetryableHTTPLogger{logger: logger}
}

// Error starts a new message with error level.
func (l RetryableHTTPLogger) Error(msg string, keysAndValues ...any) {
	logWithLevel(l.logger.Error().CallerSkipFrame(2), msg, keysAndValues...)
}

// Info starts a new message with info level.
func (l RetryableHTTPLogger) Info(msg string, keysAndValues ...any) {
	logWithLevel(l.logger.Info().CallerSkipFrame(2), msg, keysAndValues...)
}

// Debug starts a new message with debug level.
func (l RetryableHTTPLogger) Debug(msg string, keysAndValues ...any) {
	logWithLevel(l.logger.Debug().CallerSkipFrame(2), msg, keysAndValues...)
}

// Warn starts a new message with warn level.
func (l RetryableHTTPLogger) Warn(msg string, keysAndValues ...any) {
	logWithLevel(l.logger.Warn().CallerSkipFrame(2), msg, keysAndValues...)
}

func logWithLevel(ev *zerolog.Event, msg string, kvs ...any) {
	if len(kvs)%2 == 0 {
		for i := 0; i < len(kvs)-1; i += 2 {
			// The first item of the pair (the key) is supposed to be a string.
			key, ok := kvs[i].(string)
			if !ok {
				continue
			}

			val := kvs[i+1]

			var s fmt.Stringer
			if s, ok = val.(fmt.Stringer); ok {
				ev.Str(key, s.String())
			} else {
				ev.Interface(key, val)
			}
		}
	}

	// Capitalize first character.
	first := true
	msg = strings.Map(func(r rune) rune {
		if first {
			first = false
			return unicode.ToTitle(r)
		}

		return r
	}, msg)

	ev.Msg(msg)
}
