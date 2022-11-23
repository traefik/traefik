package logs

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewDatadogLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cwb := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true}

	out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, cwb)

	logger := NewDatadogLogger(zerolog.New(out).With().Caller().Logger())

	logger.Log("foo")

	assert.Equal(t, "<nil> DBG datadog_test.go:21 > foo\n", buf.String())
}
