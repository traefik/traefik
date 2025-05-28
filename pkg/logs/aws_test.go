package logs

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewAWSWrapper(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	cwb := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true}

	out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, cwb)

	logger := NewAWSWrapper(zerolog.New(out).With().Caller().Logger())

	logger.Logf(logging.Debug, "%s", "foo")

	assert.Equal(t, "<nil> DBG aws_test.go:22 > foo\n", buf.String())
}
