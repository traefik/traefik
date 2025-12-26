package logs

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryableHTTPLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cwb := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true}

	out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, cwb)

	logger := NewRetryableHTTPLogger(zerolog.New(out).With().Caller().Logger())

	logger.Info("foo")

	assert.Equal(t, "<nil> INF hclog_test.go:21 > Foo\n", buf.String())
}
