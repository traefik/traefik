package logs

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNewElasticLogger(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logger := NewElasticLogger(zerolog.New(out).With().Caller().Logger())

	logger.Errorf("foo")
}
