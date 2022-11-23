package logs

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryableHTTPLogger(t *testing.T) {
	var out bytes.Buffer

	logger := NewRetryableHTTPLogger(zerolog.New(&out).With().Logger())

	logger.Info("foo")

	assert.Equal(t, "{\"level\":\"info\",\"message\":\"Foo\"}\n", out.String())
}
