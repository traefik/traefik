package plugin

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/containous/traefik/log"
	hclog "github.com/hashicorp/go-hclog"
	golog "log"
)

var _ hclog.Logger = (*LoggerAdapter)(nil)

// LoggerAdapter wraps the "*logrus.Entry" and provides an interface adapter for go-hclog's "hclog.Logger"
type LoggerAdapter struct {
	logger *logrus.Entry
}

// Trace emit a message and key/value pairs at the DEBUG level
func (l *LoggerAdapter) Trace(msg string, args ...interface{}) {
	log.Debug(msg + argsToString(args))
}

// Debug emit a message and key/value pairs at the DEBUG level
func (l *LoggerAdapter) Debug(msg string, args ...interface{}) {
	log.Debug(msg + argsToString(args))
}

// Info emit a message and key/value pairs at the INFO level
func (l *LoggerAdapter) Info(msg string, args ...interface{}) {
	log.Info(msg + argsToString(args))
}

// Warn emit a message and key/value pairs at the WARN level
func (l *LoggerAdapter) Warn(msg string, args ...interface{}) {
	log.Warn(msg + argsToString(args))
}

// Error emit a message and key/value pairs at the ERROR level
func (l *LoggerAdapter) Error(msg string, args ...interface{}) {
	log.Error(msg + argsToString(args))
}

// IsTrace indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (l *LoggerAdapter) IsTrace() bool {
	return logrus.DebugLevel == log.GetLevel()
}

// IsDebug indicate if DEBUG logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsDebug() bool {
	return logrus.DebugLevel == log.GetLevel()
}

// IsInfo indicate if INFO logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsInfo() bool {
	return logrus.InfoLevel == log.GetLevel()
}

// IsWarn indicate if WARN logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsWarn() bool {
	return logrus.WarnLevel == log.GetLevel()
}

// IsError indicate if ERROR logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsError() bool {
	return logrus.ErrorLevel == log.GetLevel()
}

// With creates a sublogger that will always have the given key/value pairs
func (l *LoggerAdapter) With(args ...interface{}) hclog.Logger {
	return l
}

// Named create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (l *LoggerAdapter) Named(name string) hclog.Logger {
	return l
}

// ResetNamed create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (l *LoggerAdapter) ResetNamed(name string) hclog.Logger {
	return l
}

// StandardLogger return a value that conforms to the stdlib log.Logger interface
func (l *LoggerAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *golog.Logger {
	return golog.New(l.logger.Writer(), "", 0)
}

func argsToString(args ...interface{}) string {
	list := args
	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			args, ok := args[0].([]interface{})
			if !ok {
				return ""
			}
			list = args
		}

		out := ":"
		if len(list)%2 == 0 {
			for i := 0; i < len(list); i = i + 2 {
				key := fmt.Sprintf("%v", list[i])
				val := fmt.Sprintf("%v", list[i+1])

				out = out + " " + key + "=" + val
			}
			return out
		}
	}
	return ""
}
