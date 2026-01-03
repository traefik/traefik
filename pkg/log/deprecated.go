package log

import (
	"bufio"
	"io"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Debug logs a message at level Debug on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Debug(...) instead.
func Debug(args ...any) {
	mainLogger.Debug(args...)
}

// Debugf logs a message at level Debug on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Debugf(...) instead.
func Debugf(format string, args ...any) {
	mainLogger.Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Info(...) instead.
func Info(args ...any) {
	mainLogger.Info(args...)
}

// Infof logs a message at level Info on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Infof(...) instead.
func Infof(format string, args ...any) {
	mainLogger.Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Warn(...) instead.
func Warn(args ...any) {
	mainLogger.Warn(args...)
}

// Warnf logs a message at level Warn on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Warnf(...) instead.
func Warnf(format string, args ...any) {
	mainLogger.Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Error(...) instead.
func Error(args ...any) {
	mainLogger.Error(args...)
}

// Errorf logs a message at level Error on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Errorf(...) instead.
func Errorf(format string, args ...any) {
	mainLogger.Errorf(format, args...)
}

// Panic logs a message at level Panic on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Panic(...) instead.
func Panic(args ...any) {
	mainLogger.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Fatal(...) instead.
func Fatal(args ...any) {
	mainLogger.Fatal(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
//
// Deprecated: use log.FromContext(ctx).Fatalf(...) instead.
func Fatalf(format string, args ...any) {
	mainLogger.Fatalf(format, args...)
}

// AddHook adds a hook to the standard logger hooks.
func AddHook(hook logrus.Hook) {
	logrus.AddHook(hook)
}

// CustomWriterLevel logs writer for a specific level. (with a custom scanner buffer size.)
// adapted from github.com/Sirupsen/logrus/writer.go.
func CustomWriterLevel(level logrus.Level, maxScanTokenSize int) *io.PipeWriter {
	reader, writer := io.Pipe()

	var printFunc func(args ...any)

	switch level {
	case logrus.DebugLevel:
		printFunc = Debug
	case logrus.InfoLevel:
		printFunc = Info
	case logrus.WarnLevel:
		printFunc = Warn
	case logrus.ErrorLevel:
		printFunc = Error
	case logrus.FatalLevel:
		printFunc = Fatal
	case logrus.PanicLevel:
		printFunc = Panic
	default:
		printFunc = mainLogger.Print
	}

	go writerScanner(reader, maxScanTokenSize, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

// extract from github.com/Sirupsen/logrus/writer.go
// Hack the buffer size.
func writerScanner(reader io.ReadCloser, scanTokenSize int, printFunc func(args ...any)) {
	scanner := bufio.NewScanner(reader)

	if scanTokenSize > bufio.MaxScanTokenSize {
		buf := make([]byte, bufio.MaxScanTokenSize)
		scanner.Buffer(buf, scanTokenSize)
	}

	for scanner.Scan() {
		printFunc(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		Errorf("Error while reading from Writer: %s", err)
	}
	reader.Close()
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close()
}
