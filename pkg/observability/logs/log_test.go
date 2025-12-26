package logs

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNoLevel(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cwb := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true}

	out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, cwb)

	logger := NoLevel(zerolog.New(out).With().Caller().Logger(), zerolog.DebugLevel)

	logger.Info().Msg("foo")

	assert.Equal(t, "<nil> INF log_test.go:21 > foo\n", buf.String())
}
