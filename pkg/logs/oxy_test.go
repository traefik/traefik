package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewOxyWrapper(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewOxyWrapper(zerolog.New(out).With().Caller().Logger())

	logger.Info("foo")
}
