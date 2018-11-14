package log

import "github.com/sirupsen/logrus"

// Debug logs a message at level Debug on the standard logger.
// Deprecated
func Debug(args ...interface{}) {
	mainLogger.Debug(args...)
}

// Debugf logs a message at level Debug on the standard logger.
// Deprecated
func Debugf(format string, args ...interface{}) {
	mainLogger.Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
// Deprecated
func Info(args ...interface{}) {
	mainLogger.Info(args...)
}

// Infof logs a message at level Info on the standard logger.
// Deprecated
func Infof(format string, args ...interface{}) {
	mainLogger.Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
// Deprecated
func Warn(args ...interface{}) {
	mainLogger.Warn(args...)
}

// Warnf logs a message at level Warn on the standard logger.
// Deprecated
func Warnf(format string, args ...interface{}) {
	mainLogger.Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
// Deprecated
func Error(args ...interface{}) {
	mainLogger.Error(args...)
}

// Errorf logs a message at level Error on the standard logger.
// Deprecated
func Errorf(format string, args ...interface{}) {
	mainLogger.Errorf(format, args...)
}

// Panic logs a message at level Panic on the standard logger.
// Deprecated
func Panic(args ...interface{}) {
	mainLogger.Panic(args...)
}

// Panicf logs a message at level Panic on the standard logger.
// Deprecated
func Panicf(format string, args ...interface{}) {
	mainLogger.Panicf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
// Deprecated
func Fatal(args ...interface{}) {
	mainLogger.Fatal(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
// Deprecated
func Fatalf(format string, args ...interface{}) {
	mainLogger.Fatalf(format, args...)
}

// AddHook adds a hook to the standard logger hooks.
func AddHook(hook logrus.Hook) {
	logrus.AddHook(hook)
}
