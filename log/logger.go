package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Logger allows overriding the logrus logger behavior
type Logger interface {
	logrus.FieldLogger
	WriterLevel(logrus.Level) *io.PipeWriter
}

var (
	logger      Logger
	logFilePath string
	logFile     *os.File
)

func init() {
	logger = logrus.StandardLogger().WithFields(logrus.Fields{})
	logrus.SetOutput(os.Stdout)
}

// Context sets the Context of the logger
func Context(context interface{}) *logrus.Entry {
	return logger.WithField("context", context)
}

// SetOutput sets the standard logger output.
func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}

// SetFormatter sets the standard logger formatter.
func SetFormatter(formatter logrus.Formatter) {
	logrus.SetFormatter(formatter)
}

// SetLevel sets the standard logger level.
func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

// SetLogger sets the logger.
func SetLogger(l Logger) {
	logger = l
}

// GetLevel returns the standard logger level.
func GetLevel() logrus.Level {
	return logrus.GetLevel()
}

// AddHook adds a hook to the standard logger hooks.
func AddHook(hook logrus.Hook) {
	logrus.AddHook(hook)
}

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return logger.WithError(err)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	logger.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	logger.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	logger.Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	logger.Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	logger.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

// OpenFile opens the log file using the specified path
func OpenFile(path string) error {
	logFilePath = path
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err == nil {
		SetOutput(logFile)
	}

	return err
}

// CloseFile closes the log and sets the Output to stdout
func CloseFile() error {
	logrus.SetOutput(os.Stdout)

	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// RotateFile closes and reopens the log file to allow for rotation
// by an external source.  If the log isn't backed by a file then
// it does nothing.
func RotateFile() error {
	if logFile == nil && logFilePath == "" {
		Debug("Traefik log is not writing to a file, ignoring rotate request")
		return nil
	}

	if logFile != nil {
		defer func(f *os.File) {
			f.Close()
		}(logFile)
	}

	if err := OpenFile(logFilePath); err != nil {
		return fmt.Errorf("error opening log file: %s", err)
	}

	return nil
}

// Writer logs writer (Level Info)
func Writer() *io.PipeWriter {
	return WriterLevel(logrus.InfoLevel)
}

// WriterLevel logs writer for a specific level.
func WriterLevel(level logrus.Level) *io.PipeWriter {
	return logger.WriterLevel(level)
}

// CustomWriterLevel logs writer for a specific level. (with a custom scanner buffer size.)
// adapted from github.com/Sirupsen/logrus/writer.go
func CustomWriterLevel(level logrus.Level, maxScanTokenSize int) *io.PipeWriter {
	reader, writer := io.Pipe()

	var printFunc func(args ...interface{})

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
		printFunc = Print
	}

	go writerScanner(reader, maxScanTokenSize, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

// extract from github.com/Sirupsen/logrus/writer.go
// Hack the buffer size
func writerScanner(reader io.ReadCloser, scanTokenSize int, printFunc func(args ...interface{})) {
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
