package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewAWSWrapper(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewAWSWrapper(zerolog.New(out).With().Caller().Logger())

	logger.Log("foo")
}
