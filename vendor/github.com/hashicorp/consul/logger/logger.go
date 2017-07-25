package logger

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hashicorp/go-syslog"
	"github.com/hashicorp/logutils"
	"github.com/mitchellh/cli"
)

// Config is used to set up logging.
type Config struct {
	// LogLevel is the minimum level to be logged.
	LogLevel string

	// EnableSyslog controls forwarding to syslog.
	EnableSyslog bool

	// SyslogFacility is the destination for syslog forwarding.
	SyslogFacility string
}

// Setup is used to perform setup of several logging objects:
//
// * A LevelFilter is used to perform filtering by log level.
// * A GatedWriter is used to buffer logs until startup UI operations are
//   complete. After this is flushed then logs flow directly to output
//   destinations.
// * A LogWriter provides a mean to temporarily hook logs, such as for running
//   a command like "consul monitor".
// * An io.Writer is provided as the sink for all logs to flow to.
//
// The provided ui object will get any log messages related to setting up
// logging itself, and will also be hooked up to the gated logger. The final bool
// parameter indicates if logging was set up successfully.
func Setup(config *Config, ui cli.Ui) (*logutils.LevelFilter, *GatedWriter, *LogWriter, io.Writer, bool) {
	// The gated writer buffers logs at startup and holds until it's flushed.
	logGate := &GatedWriter{
		Writer: &cli.UiWriter{Ui: ui},
	}

	// Set up the level filter.
	logFilter := LevelFilter()
	logFilter.MinLevel = logutils.LogLevel(strings.ToUpper(config.LogLevel))
	logFilter.Writer = logGate
	if !ValidateLevelFilter(logFilter.MinLevel, logFilter) {
		ui.Error(fmt.Sprintf(
			"Invalid log level: %s. Valid log levels are: %v",
			logFilter.MinLevel, logFilter.Levels))
		return nil, nil, nil, nil, false
	}

	// Set up syslog if it's enabled.
	var syslog io.Writer
	if config.EnableSyslog {
		retries := 12
		delay := 5 * time.Second
		for i := 0; i <= retries; i++ {
			l, err := gsyslog.NewLogger(gsyslog.LOG_NOTICE, config.SyslogFacility, "consul")
			if err != nil {
				ui.Error(fmt.Sprintf("Syslog setup error: %v", err))
				if i == retries {
					timeout := time.Duration(retries) * delay
					ui.Error(fmt.Sprintf("Syslog setup did not succeed within timeout (%s).", timeout.String()))
					return nil, nil, nil, nil, false
				} else {
					ui.Error(fmt.Sprintf("Retrying syslog setup in %s...", delay.String()))
					time.Sleep(delay)
				}
			} else {
				syslog = &SyslogWrapper{l, logFilter}
				break
			}
		}
	}

	// Create a log writer, and wrap a logOutput around it
	logWriter := NewLogWriter(512)
	var logOutput io.Writer
	if syslog != nil {
		logOutput = io.MultiWriter(logFilter, logWriter, syslog)
	} else {
		logOutput = io.MultiWriter(logFilter, logWriter)
	}
	return logFilter, logGate, logWriter, logOutput, true
}
