package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewDatadogLogger(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewDatadogLogger(zerolog.New(out).With().Caller().Logger())

	logger.Log("foo")
}
