package plugin

import (
	"github.com/Sirupsen/logrus"
	"github.com/containous/traefik/log"
	hclog "github.com/hashicorp/go-hclog"
	golog "log"
)

var _ hclog.Logger = (*LoggerAdapter)(nil)

type LoggerAdapter struct {
	logger *logrus.Entry
}

func (l *LoggerAdapter) Trace(msg string, args ...interface{}) {
	log.Debugf(msg, args)
}

// Emit a message and key/value pairs at the DEBUG level
func (l *LoggerAdapter) Debug(msg string, args ...interface{}) {
	log.Debugf(msg, args)
}

// Emit a message and key/value pairs at the INFO level
func (l *LoggerAdapter) Info(msg string, args ...interface{}) {
	log.Infof(msg, args)
}

// Emit a message and key/value pairs at the WARN level
func (l *LoggerAdapter) Warn(msg string, args ...interface{}) {
	log.Warnf(msg, args)
}

// Emit a message and key/value pairs at the ERROR level
func (l *LoggerAdapter) Error(msg string, args ...interface{}) {
	log.Errorf(msg, args)
}

// Indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (l *LoggerAdapter) IsTrace() bool {
	return logrus.DebugLevel == log.GetLevel()
}

// Indicate if DEBUG logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsDebug() bool {
	return logrus.DebugLevel == log.GetLevel()
}

// Indicate if INFO logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsInfo() bool {
	return logrus.InfoLevel == log.GetLevel()
}

// Indicate if WARN logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsWarn() bool {
	return logrus.WarnLevel == log.GetLevel()
}

// Indicate if ERROR logs would be emitted. This and the other Is* guards
func (l *LoggerAdapter) IsError() bool {
	return logrus.ErrorLevel == log.GetLevel()
}

// Creates a sublogger that will always have the given key/value pairs
func (l *LoggerAdapter) With(args ...interface{}) hclog.Logger {
	return l
}

// Create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (l *LoggerAdapter) Named(name string) hclog.Logger {
	return l
}

// Create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (l *LoggerAdapter) ResetNamed(name string) hclog.Logger {
	return l
}

// Return a value that conforms to the stdlib log.Logger interface
func (l *LoggerAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *golog.Logger {
	return golog.New(l.logger.Writer(), "", 0)
}
