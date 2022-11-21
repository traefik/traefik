package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewLogrusStdWrapper(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewLogrusWrapper(zerolog.New(out).With().Caller().Logger())

	logger.Println("foo")
}
