package log

type DatadogLogger struct {
	logger Logger
}

func NewDatadogLogger(logger Logger) *DatadogLogger {
	return &DatadogLogger{logger: logger}
}

func (d DatadogLogger) Log(msg string) {
	d.logger.Error(msg)
}
