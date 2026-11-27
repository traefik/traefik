package middlewares

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

func TestWithLogger_chained(t *testing.T) {
	var buf bytes.Buffer
	base := zerolog.New(&buf).With().Str(logs.EntryPointName, "web").Logger()

	ctx, _ := WithLogger(base.WithContext(t.Context()), "first", "typeA")
	ctx, _ = WithLogger(ctx, "second", "typeB")

	log.Ctx(ctx).Info().Msg("test")

	assertLogLine(t, buf.String(), map[string]string{
		logs.EntryPointName: "web",
		logs.MiddlewareName: "second",
		logs.MiddlewareType: "typeB",
	})
}

func TestGetLogger_afterWithLogger(t *testing.T) {
	var buf bytes.Buffer
	base := zerolog.New(&buf).With().Str(logs.EntryPointName, "web").Logger()

	ctx, _ := WithLogger(base.WithContext(t.Context()), "first", "typeA")

	GetLogger(ctx, "second", "typeB").Info().Msg("test")

	assertLogLine(t, buf.String(), map[string]string{
		logs.EntryPointName: "web",
		logs.MiddlewareName: "second",
		logs.MiddlewareType: "typeB",
	})
}

func TestWithLogger_contextLoggerReplaced(t *testing.T) {
	var buf bytes.Buffer
	base := zerolog.New(&buf).With().Str(logs.EntryPointName, "web").Logger()

	ctx, _ := WithLogger(base.WithContext(t.Context()), "first", "typeA")
	ctx = log.Ctx(ctx).With().Str("extra", "value").Logger().WithContext(ctx)
	_, logger := WithLogger(ctx, "second", "typeB")

	logger.Info().Msg("test")

	// The extra field must not be lost, at the cost of repeating the middleware fields.
	assert.Equal(t, 2, strings.Count(buf.String(), `"`+logs.MiddlewareName+`"`))
	assert.Contains(t, buf.String(), `"extra":"value"`)
}

func assertLogLine(t *testing.T, line string, expected map[string]string) {
	t.Helper()

	assert.Equal(t, 1, strings.Count(line, `"`+logs.MiddlewareName+`"`))
	assert.Equal(t, 1, strings.Count(line, `"`+logs.MiddlewareType+`"`))

	var fields map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &fields))

	for key, value := range expected {
		assert.Equal(t, value, fields[key])
	}
}
