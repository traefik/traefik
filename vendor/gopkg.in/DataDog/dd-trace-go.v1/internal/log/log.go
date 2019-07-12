// Package log provides logging utilities for the tracer.
package log

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/version"
)

// Level specifies the logging level that the log package prints at.
type Level int

const (
	// LevelDebug represents debug level messages.
	LevelDebug Level = iota
	// LevelWarn represents warning and errors.
	LevelWarn
)

var prefixMsg = fmt.Sprintf("Datadog Tracer %s", version.Tag)

var (
	mu     sync.RWMutex   // guards below fields
	level                 = LevelWarn
	logger ddtrace.Logger = &defaultLogger{l: log.New(os.Stderr, "", log.LstdFlags)}
)

// UseLogger sets l as the active logger.
func UseLogger(l ddtrace.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logger = l
}

// SetLevel sets the given lvl for logging.
func SetLevel(lvl Level) {
	mu.Lock()
	defer mu.Unlock()
	level = lvl
}

// Debug prints the given message if the level is LevelDebug.
func Debug(fmt string, a ...interface{}) {
	mu.RLock()
	lvl := level
	mu.RUnlock()
	if lvl != LevelDebug {
		return
	}
	printMsg("DEBUG", fmt, a...)
}

// Warn prints a warning message.
func Warn(fmt string, a ...interface{}) {
	printMsg("WARN", fmt, a...)
}

var (
	errmu   sync.RWMutex                // guards below fields
	erragg  = map[string]*errorReport{} // aggregated errors
	errrate = time.Minute               // the rate at which errors are reported
	erron   bool                        // true if errors are being aggregated
)

func init() {
	if v := os.Getenv("DD_LOGGING_RATE"); v != "" {
		if sec, err := strconv.ParseUint(v, 10, 64); err != nil {
			Warn("Invalid value for DD_LOGGING_RATE: %v", err)
		} else {
			errrate = time.Duration(sec) * time.Second
		}
	}
}

type errorReport struct {
	first time.Time // time when first error occurred
	err   error
	count uint64
}

// Error reports an error. Errors get aggregated and logged periodically. The
// default is once per minute or once every DD_LOGGING_RATE number of seconds.
func Error(format string, a ...interface{}) {
	key := format // format should 99.9% of the time be constant
	if reachedLimit(key) {
		// avoid too much lock contention on spammy errors
		return
	}
	errmu.Lock()
	defer errmu.Unlock()
	report, ok := erragg[key]
	if !ok {
		erragg[key] = &errorReport{
			err:   fmt.Errorf(format, a...),
			first: time.Now(),
		}
		report = erragg[key]
	}
	report.count++
	if errrate == 0 {
		flushLocked()
		return
	}
	if !erron {
		erron = true
		time.AfterFunc(errrate, Flush)
	}
}

// defaultErrorLimit specifies the maximum number of errors gathered in a report.
const defaultErrorLimit = 200

// reachedLimit reports whether the maximum count has been reached for this key.
func reachedLimit(key string) bool {
	errmu.RLock()
	e, ok := erragg[key]
	confirm := ok && e.count > defaultErrorLimit
	errmu.RUnlock()
	return confirm
}

// Flush flushes and resets all aggregated errors to the logger.
func Flush() {
	errmu.Lock()
	defer errmu.Unlock()
	flushLocked()
}

func flushLocked() {
	for _, report := range erragg {
		msg := fmt.Sprintf("%v", report.err)
		if report.count > defaultErrorLimit {
			msg += fmt.Sprintf(", %d+ additional messages skipped (first occurrence: %s)", defaultErrorLimit, report.first.Format(time.RFC822))
		} else if report.count > 1 {
			msg += fmt.Sprintf(", %d additional messages skipped (first occurrence: %s)", report.count-1, report.first.Format(time.RFC822))
		} else {
			msg += fmt.Sprintf(" (occurred: %s)", report.first.Format(time.RFC822))
		}
		printMsg("ERROR", msg)
	}
	for k := range erragg {
		// compiler-optimized map-clearing post go1.11 (golang/go#20138)
		delete(erragg, k)
	}
	erron = false
}

func printMsg(lvl, format string, a ...interface{}) {
	msg := fmt.Sprintf("%s %s: %s", prefixMsg, lvl, fmt.Sprintf(format, a...))
	mu.RLock()
	logger.Log(msg)
	mu.RUnlock()
}

type defaultLogger struct{ l *log.Logger }

func (p *defaultLogger) Log(msg string) { p.l.Print(msg) }
