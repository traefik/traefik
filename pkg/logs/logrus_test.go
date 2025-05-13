package logs

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewLogrusStdWrapper(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cwb := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true}

	out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, cwb)

	logger := NewLogrusWrapper(zerolog.New(out).With().Caller().Logger())

	logger.Println("foo")

	assert.Equal(t, "<nil> DBG logrus_test.go:21 > foo\n", buf.String())
}
