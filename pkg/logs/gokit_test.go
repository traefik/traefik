package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewGoKitWrapper(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewGoKitWrapper(zerolog.New(out).With().Caller().Logger())

	_ = logger.Log("foo")
}
