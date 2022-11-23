package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewRetryableHTTPLogger(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewRetryableHTTPLogger(zerolog.New(out).With().Caller().Logger())

	logger.Info("foo")
}
