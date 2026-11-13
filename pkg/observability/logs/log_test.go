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

func TestNewConsoleWriter_NoTimestamp(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cw := NewConsoleWriter(buf, true, true)

	logger := zerolog.New(cw).With().Logger()
	logger.Info().Str("field1", "val1").Msg("test message")

	output := buf.String()
	assert.Equal(t, "INF test message field1=val1\n", output)
}

func TestNewConsoleWriter_WithTimestamp(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cw := NewConsoleWriter(buf, true, false)

	logger := zerolog.New(cw).With().Timestamp().Logger()
	logger.Info().Str("field1", "val1").Msg("test message")

	output := buf.String()
	assert.Contains(t, output, "INF test message field1=val1\n")
	assert.NotEqual(t, "INF test message field1=val1\n", output)
}

func TestNewLogger_JSON_NoTimestamp(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := NewLogger(buf, "json", false, true)

	logger.Info().Str("key", "value").Msg("json message")

	output := buf.String()
	assert.NotContains(t, output, `"time"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"json message"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestNewLogger_JSON_WithTimestamp(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := NewLogger(buf, "json", false, false)

	logger.Info().Str("key", "value").Msg("json message")

	output := buf.String()
	assert.Contains(t, output, `"time"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"json message"`)
	assert.Contains(t, output, `"key":"value"`)
}

