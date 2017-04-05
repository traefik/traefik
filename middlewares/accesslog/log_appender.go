package accesslog

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// default format for time presentation
const commonLogTimeFormat = "02/Jan/2006:15:04:05 -0700"

// Flusher types have I/O buffers that can be flushed through.
type Flusher interface {
	Flush() error
}

// LogFormatter types can write log data to some appender terminal.
type LogFormatter interface {
	// Write formats a log event and writes the formatted text to the writer provided.
	Write(w io.Writer, event *LogData) error
}

// LogAppender holds the settings and current state of the back-end part of the logger,
// i.e. the part that handles writing to a file, socket etc.
type LogAppender struct {
	settings  *types.AccessLog
	formatter LogFormatter
	file      io.Writer
	buf       io.Writer
}

// NewLogAppender creates a new instance of LogAppender using the settings provided.
// The returned appended will have determined the required output formatter and opened
// the relevant file (or etc) for output.
func NewLogAppender(settings *types.AccessLog) (LogAppender, error) {
	settings = applyDefaults(settings)

	var formatter LogFormatter
	switch settings.Format {
	case "json":
		formatter = newJSONLogFormatter(settings)
	default:
		formatter = newCommonLogFormatter(settings)
		settings.Format = "clf"
	}

	la := LogAppender{settings: settings, formatter: formatter}
	return la, la.Open()
}

func applyDefaults(settings *types.AccessLog) *types.AccessLog {
	if settings.GzipLevel == 0 {
		settings.GzipLevel = gzip.DefaultCompression
	}

	if settings.TimeFormat == "" {
		settings.TimeFormat = commonLogTimeFormat
	}
	return settings
}

// IsOpen determines whether the appender is currently open.
func (l LogAppender) IsOpen() bool {
	return l.file != nil
}

func (l *LogAppender) openLogFileAndBuffer() error {
	bufferSize, _, err := types.AsSI(l.settings.BufferSize)
	if err != nil && l.settings.BufferSize != "" {
		return err
	}

	log.Debugf("Opening access log %s (%s)", l.settings.File, l.settings.Format)
	useGzip := strings.HasSuffix(l.settings.File, ".gz")
	l.file, err = os.OpenFile(l.settings.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.buf = l.file

	if useGzip {
		// benchmarks show there is no point adding more buffering when gzipping
		l.buf, err = gzip.NewWriterLevel(l.file, l.settings.GzipLevel)
		if err != nil {
			return err
		}
	} else if bufferSize > 0 {
		l.buf = bufio.NewWriterSize(l.file, int(bufferSize))
	}
	return nil
}

// Open opens the output file used by the appender.
func (l *LogAppender) Open() error {
	if len(l.settings.File) > 0 && !l.IsOpen() {
		err := l.openLogFileAndBuffer()
		if err != nil {
			return err
		}

		// this prevents write races and is a performance optimisation over
		// simply using the goroutine with an unbuffered channel
		l.buf = LinearWriter(l.buf)
	}
	return nil
}

func (l *LogAppender) flushBufferAndCloseFile() error {
	log.Debugf("Closing access log %s (%s)", l.settings.File, l.settings.Format)
	var err error
	if f, ok := l.buf.(Flusher); ok {
		err = f.Flush()
		if err != nil {
			return err
		}
	}
	l.buf = nil

	if c, ok := l.file.(io.Closer); ok {
		err = c.Close()
	}
	l.file = nil
	return err
}

// Close flushed the remaining log messages through and then closes the output file.
// If the appender was using a separate goroutine, this will be terminated.
func (l *LogAppender) Close() error {
	if l.IsOpen() {
		err := l.flushBufferAndCloseFile()
		if err != nil {
			return err
		}
	}
	return nil
}

// Write writes a log event to the current appender.
func (l LogAppender) Write(logDataTable *LogData) error {
	return l.formatter.Write(l.buf, logDataTable)
}

//-------------------------------------------------------------------------------------------------

type commonLogFormatter struct {
	timeFormat string
}

func newCommonLogFormatter(settings *types.AccessLog) commonLogFormatter {
	return commonLogFormatter{settings.TimeFormat}
}

func (l commonLogFormatter) Write(w io.Writer, logDataTable *LogData) error {
	timestamp := logDataTable.Core[StartUTC].(time.Time).Format(l.timeFormat)
	elapsedMillis := logDataTable.Core[Duration].(time.Duration).Nanoseconds() / 1000000

	_, err := fmt.Fprintf(w, "%s - %s [%s] \"%s %s %s\" %d %d \"%s\" %s %d %s %s %dms\n",
		logDataTable.Core[ClientHost],
		logDataTable.Core[ClientUsername],
		timestamp,
		logDataTable.Core[RequestMethod],
		logDataTable.Core[RequestPath],
		logDataTable.Core[RequestProtocol],
		logDataTable.Core[OriginStatus],
		logDataTable.Core[OriginContentSize],
		logDataTable.Request.Get("Referer"),
		quoted(logDataTable.Request.Get("User-Agent")),
		logDataTable.Core[RequestCount],
		quoted(logDataTable.Core[FrontendName]),
		quoted(logDataTable.Core[BackendURL]),
		elapsedMillis)
	return err
}

func quoted(v interface{}) string {
	if v == nil {
		return "-"
	}

	switch s := v.(type) {
	case string:
		if s == "" {
			return "-"
		}
		return `"` + s + `"`

	case fmt.Stringer:
		return `"` + s.String() + `"`
	}

	return "-"
}
