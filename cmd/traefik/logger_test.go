package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

func TestSetupLogger_CommonFormat_NoTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	defaultLogWriter = buf
	t.Cleanup(func() {
		defaultLogWriter = os.Stdout
	})

	staticConfig := &static.Configuration{
		Log: &otypes.TraefikLog{
			Level:       "INFO",
			Format:      "common",
			NoColor:     true,
			NoTimestamp: true,
		},
	}

	err := setupLogger(context.Background(), staticConfig)
	require.NoError(t, err)

	log.Info().Str("component", "test").Msg("hello world")

	output := buf.String()
	assert.Contains(t, output, "INF hello world component=test")
	assert.NotRegexp(t, regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`), output)
}

func TestSetupLogger_CommonFormat_WithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	defaultLogWriter = buf
	t.Cleanup(func() {
		defaultLogWriter = os.Stdout
	})

	staticConfig := &static.Configuration{
		Log: &otypes.TraefikLog{
			Level:       "INFO",
			Format:      "common",
			NoColor:     true,
			NoTimestamp: false,
		},
	}

	err := setupLogger(context.Background(), staticConfig)
	require.NoError(t, err)

	log.Info().Str("component", "test").Msg("hello world")

	output := buf.String()
	assert.Contains(t, output, "INF hello world component=test")
	assert.Regexp(t, regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`), output)
}

func TestSetupLogger_JSONFormat_NoTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	defaultLogWriter = buf
	t.Cleanup(func() {
		defaultLogWriter = os.Stdout
	})

	staticConfig := &static.Configuration{
		Log: &otypes.TraefikLog{
			Level:       "INFO",
			Format:      "json",
			NoTimestamp: true,
		},
	}

	err := setupLogger(context.Background(), staticConfig)
	require.NoError(t, err)

	log.Info().Str("component", "test").Msg("hello json")

	var entry map[string]any
	err = json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.NotContains(t, entry, "time")
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "hello json", entry["message"])
	assert.Equal(t, "test", entry["component"])
}

func TestSetupLogger_JSONFormat_WithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	defaultLogWriter = buf
	t.Cleanup(func() {
		defaultLogWriter = os.Stdout
	})

	staticConfig := &static.Configuration{
		Log: &otypes.TraefikLog{
			Level:       "INFO",
			Format:      "json",
			NoTimestamp: false,
		},
	}

	err := setupLogger(context.Background(), staticConfig)
	require.NoError(t, err)

	log.Info().Str("component", "test").Msg("hello json")

	var entry map[string]any
	err = json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Contains(t, entry, "time")
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "hello json", entry["message"])
	assert.Equal(t, "test", entry["component"])
}
