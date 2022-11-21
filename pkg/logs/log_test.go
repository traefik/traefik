package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNoLevel(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NoLevel(zerolog.New(out).With().Caller().Logger(), zerolog.DebugLevel)

	logger.Info().Msg("foo")
}
