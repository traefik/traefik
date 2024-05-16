package logs

import (
	"fmt"

	"github.com/rs/zerolog"
)

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
