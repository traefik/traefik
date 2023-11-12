package datadog

import "github.com/traefik/traefik/v2/pkg/log"

type DatadogLogger struct {
	logger log.Logger
}

func newDatadogLogger(logger log.Logger) *DatadogLogger {
	return &DatadogLogger{logger: logger}
}

func (d DatadogLogger) Log(msg string) {
	d.logger.Debug(msg)
}
