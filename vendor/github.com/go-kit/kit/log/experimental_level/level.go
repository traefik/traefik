package level

import (
	"github.com/go-kit/kit/log"
)

var (
	levelKey        = "level"
	errorLevelValue = "error"
	warnLevelValue  = "warn"
	infoLevelValue  = "info"
	debugLevelValue = "debug"
)

// AllowAll is an alias for AllowDebugAndAbove.
func AllowAll() []string {
	return AllowDebugAndAbove()
}

// AllowDebugAndAbove allows all of the four default log levels.
// Its return value may be provided as the Allowed parameter in the Config.
func AllowDebugAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue, infoLevelValue, debugLevelValue}
}

// AllowInfoAndAbove allows the default info, warn, and error log levels.
// Its return value may be provided as the Allowed parameter in the Config.
func AllowInfoAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue, infoLevelValue}
}

// AllowWarnAndAbove allows the default warn and error log levels.
// Its return value may be provided as the Allowed parameter in the Config.
func AllowWarnAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue}
}

// AllowErrorOnly allows only the default error log level.
// Its return value may be provided as the Allowed parameter in the Config.
func AllowErrorOnly() []string {
	return []string{errorLevelValue}
}

// AllowNone allows none of the default log levels.
// Its return value may be provided as the Allowed parameter in the Config.
func AllowNone() []string {
	return []string{}
}

// Error returns a logger with the level key set to ErrorLevelValue.
func Error(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(levelKey, errorLevelValue)
}

// Warn returns a logger with the level key set to WarnLevelValue.
func Warn(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(levelKey, warnLevelValue)
}

// Info returns a logger with the level key set to InfoLevelValue.
func Info(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(levelKey, infoLevelValue)
}

// Debug returns a logger with the level key set to DebugLevelValue.
func Debug(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(levelKey, debugLevelValue)
}

// Config parameterizes the leveled logger.
type Config struct {
	// Allowed enumerates the accepted log levels. If a log event is encountered
	// with a level key set to a value that isn't explicitly allowed, the event
	// will be squelched, and ErrNotAllowed returned.
	Allowed []string

	// ErrNotAllowed is returned to the caller when Log is invoked with a level
	// key that hasn't been explicitly allowed. By default, ErrNotAllowed is
	// nil; in this case, the log event is squelched with no error.
	ErrNotAllowed error

	// SquelchNoLevel will squelch log events with no level key, so that they
	// don't proceed through to the wrapped logger. If SquelchNoLevel is set to
	// true and a log event is squelched in this way, ErrNoLevel is returned to
	// the caller.
	SquelchNoLevel bool

	// ErrNoLevel is returned to the caller when SquelchNoLevel is true, and Log
	// is invoked without a level key. By default, ErrNoLevel is nil; in this
	// case, the log event is squelched with no error.
	ErrNoLevel error
}

// New wraps the logger and implements level checking. See the commentary on the
// Config object for a detailed description of how to configure levels.
func New(next log.Logger, config Config) log.Logger {
	return &logger{
		next:           next,
		allowed:        makeSet(config.Allowed),
		errNotAllowed:  config.ErrNotAllowed,
		squelchNoLevel: config.SquelchNoLevel,
		errNoLevel:     config.ErrNoLevel,
	}
}

type logger struct {
	next           log.Logger
	allowed        map[string]struct{}
	errNotAllowed  error
	squelchNoLevel bool
	errNoLevel     error
}

func (l *logger) Log(keyvals ...interface{}) error {
	var hasLevel, levelAllowed bool
	for i := 0; i < len(keyvals); i += 2 {
		if k, ok := keyvals[i].(string); !ok || k != levelKey {
			continue
		}
		hasLevel = true
		if i >= len(keyvals) {
			continue
		}
		v, ok := keyvals[i+1].(string)
		if !ok {
			continue
		}
		_, levelAllowed = l.allowed[v]
		break
	}
	if !hasLevel && l.squelchNoLevel {
		return l.errNoLevel
	}
	if hasLevel && !levelAllowed {
		return l.errNotAllowed
	}
	return l.next.Log(keyvals...)
}

func makeSet(a []string) map[string]struct{} {
	m := make(map[string]struct{}, len(a))
	for _, s := range a {
		m[s] = struct{}{}
	}
	return m
}
