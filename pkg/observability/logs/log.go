package logs

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

// NewConsoleWriter creates a zerolog.ConsoleWriter with timestamp and color settings.
func NewConsoleWriter(out io.Writer, noColor bool, noTimestamp bool) zerolog.ConsoleWriter {
	cw := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: time.RFC3339,
		NoColor:    noColor,
	}
	if noTimestamp {
		cw.PartsExclude = []string{zerolog.TimestampFieldName}
	}
	return cw
}

// NewLogger creates a new zerolog.Logger configured with format, color, and timestamp settings.
func NewLogger(w io.Writer, format string, noColor bool, noTimestamp bool) zerolog.Logger {
	var out io.Writer = w
	if format != "json" {
		out = NewConsoleWriter(w, noColor, noTimestamp)
	}

	logger := zerolog.New(out).With()
	if !noTimestamp {
		logger = logger.Timestamp()
	}

	return logger.Logger()
}

func NoLevel(logger zerolog.Logger, level zerolog.Level) zerolog.Logger {
	return logger.Hook(NewNoLevelHook(logger.GetLevel(), level))
}

type NoLevelHook struct {
	minLevel zerolog.Level
	level    zerolog.Level
}

func NewNoLevelHook(minLevel zerolog.Level, level zerolog.Level) *NoLevelHook {
	return &NoLevelHook{minLevel: minLevel, level: level}
}

func (n NoLevelHook) Run(e *zerolog.Event, level zerolog.Level, _ string) {
	if n.minLevel > n.level {
		e.Discard()
		return
	}

	if level == zerolog.NoLevel {
		e.Str("level", n.level.String())
	}
}

func msgFunc(i ...any) func() string {
	return func() string { return fmt.Sprint(i...) }
}
