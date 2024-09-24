package logs

import "github.com/rs/zerolog"

type DatadogLogger struct {
	logger zerolog.Logger
}

func NewDatadogLogger(logger zerolog.Logger) *DatadogLogger {
	return &DatadogLogger{logger: logger}
}

func (d DatadogLogger) Log(msg string) {
	d.logger.Debug().CallerSkipFrame(1).Msg(msg)
}
