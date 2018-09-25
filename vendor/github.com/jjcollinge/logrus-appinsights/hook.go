package logrus_appinsights

import (
	"encoding/json"
	"fmt"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"github.com/sirupsen/logrus"
)

var defaultLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
	logrus.InfoLevel,
}

var levelMap = map[logrus.Level]appinsights.SeverityLevel{
	logrus.PanicLevel: appinsights.Critical,
	logrus.FatalLevel: appinsights.Critical,
	logrus.ErrorLevel: appinsights.Error,
	logrus.WarnLevel:  appinsights.Warning,
	logrus.InfoLevel:  appinsights.Information,
}

// AppInsightsHook is a logrus hook for Application Insights
type AppInsightsHook struct {
	client appinsights.TelemetryClient

	async        bool
	levels       []logrus.Level
	ignoreFields map[string]struct{}
	filters      map[string]func(interface{}) interface{}
}

// New returns an initialised logrus hook for Application Insights
func New(name string, conf Config) (*AppInsightsHook, error) {
	if conf.InstrumentationKey == "" {
		return nil, fmt.Errorf("InstrumentationKey is required and missing from configuration")
	}
	telemetryConf := appinsights.NewTelemetryConfiguration(conf.InstrumentationKey)
	if conf.MaxBatchSize != 0 {
		telemetryConf.MaxBatchSize = conf.MaxBatchSize
	}
	if conf.MaxBatchInterval != 0 {
		telemetryConf.MaxBatchInterval = conf.MaxBatchInterval
	}
	if conf.EndpointUrl != "" {
		telemetryConf.EndpointUrl = conf.EndpointUrl
	}
	telemetryClient := appinsights.NewTelemetryClientFromConfig(telemetryConf)
	if name != "" {
		telemetryClient.Context().Cloud().SetRoleName(name)
	}
	return &AppInsightsHook{
		client:       telemetryClient,
		levels:       defaultLevels,
		ignoreFields: make(map[string]struct{}),
		filters:      make(map[string]func(interface{}) interface{}),
	}, nil
}

// NewWithAppInsightsConfig returns an initialised logrus hook for Application Insights
func NewWithAppInsightsConfig(name string, conf *appinsights.TelemetryConfiguration) (*AppInsightsHook, error) {
	if conf == nil {
		return nil, fmt.Errorf("Nil configuration provided")
	}
	if conf.InstrumentationKey == "" {
		return nil, fmt.Errorf("InstrumentationKey is required in configuration")
	}
	telemetryClient := appinsights.NewTelemetryClientFromConfig(conf)
	if name != "" {
		telemetryClient.Context().Cloud().SetRoleName(name)
	}
	return &AppInsightsHook{
		client:       telemetryClient,
		levels:       defaultLevels,
		ignoreFields: make(map[string]struct{}),
		filters:      make(map[string]func(interface{}) interface{}),
	}, nil
}

// Levels returns logging level to fire this hook.
func (hook *AppInsightsHook) Levels() []logrus.Level {
	return hook.levels
}

// SetLevels sets logging level to fire this hook.
func (hook *AppInsightsHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}

// SetAsync sets async flag for sending logs asynchronously.
// If use this true, Fire() does not return error.
func (hook *AppInsightsHook) SetAsync(async bool) {
	hook.async = async
}

// AddIgnore adds field name to ignore.
func (hook *AppInsightsHook) AddIgnore(name string) {
	hook.ignoreFields[name] = struct{}{}
}

// AddFilter adds a custom filter function.
func (hook *AppInsightsHook) AddFilter(name string, fn func(interface{}) interface{}) {
	hook.filters[name] = fn
}

// Fire is invoked by logrus and sends log data to Application Insights.
func (hook *AppInsightsHook) Fire(entry *logrus.Entry) error {
	if !hook.async {
		return hook.fire(entry)
	}
	// async - fire and forget
	go hook.fire(entry)
	return nil
}

func (hook *AppInsightsHook) fire(entry *logrus.Entry) error {
	trace, err := hook.buildTrace(entry)
	if err != nil {
		return err
	}
	hook.client.TrackTraceTelemetry(trace)
	return nil
}

func (hook *AppInsightsHook) buildTrace(entry *logrus.Entry) (*appinsights.TraceTelemetry, error) {
	// Add the message as a field if it isn't already
	if _, ok := entry.Data["message"]; !ok {
		entry.Data["message"] = entry.Message
	}

	level := levelMap[entry.Level]
	trace := appinsights.NewTraceTelemetry(entry.Message, level)
	if trace == nil {
		return nil, fmt.Errorf("Could not create telemetry trace with entry %+v", entry)
	}
	for k, v := range entry.Data {
		if _, ok := hook.ignoreFields[k]; ok {
			continue
		}
		if fn, ok := hook.filters[k]; ok {
			v = fn(v) // apply custom filter
		} else {
			v = formatData(v) // use default formatter
		}
		vStr := fmt.Sprintf("%v", v)
		trace.SetProperty(k, vStr)
	}
	trace.SetProperty("source_level", entry.Level.String())
	trace.SetProperty("source_timestamp", entry.Time.String())
	return trace, nil
}

// formatData returns value as a suitable format.
func formatData(value interface{}) (formatted interface{}) {
	switch value := value.(type) {
	case json.Marshaler:
		return value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return value
	}
}

func stringPtr(str string) *string {
	return &str
}
