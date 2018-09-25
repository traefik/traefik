package appinsights

import "time"

type TelemetryClient interface {
	Context() TelemetryContext
	InstrumentationKey() string
	Channel() TelemetryChannel
	IsEnabled() bool
	SetIsEnabled(bool)
	Track(Telemetry)
	TrackEvent(string)
	TrackEventTelemetry(*EventTelemetry)
	TrackMetric(string, float32)
	TrackMetricTelemetry(*MetricTelemetry)
	TrackTrace(string)
	TrackTraceTelemetry(*TraceTelemetry)
	TrackRequest(string, string, string, time.Time, time.Duration, string, bool)
	TrackRequestTelemetry(*RequestTelemetry)
}

type telemetryClient struct {
	TelemetryConfiguration *TelemetryConfiguration
	channel                TelemetryChannel
	context                TelemetryContext
	isEnabled              bool
}

func NewTelemetryClient(iKey string) TelemetryClient {
	return NewTelemetryClientFromConfig(NewTelemetryConfiguration(iKey))
}

func NewTelemetryClientFromConfig(config *TelemetryConfiguration) TelemetryClient {
	channel := NewInMemoryChannel(config)
	context := NewClientTelemetryContext()
	return &telemetryClient{
		TelemetryConfiguration: config,
		channel:                channel,
		context:                context,
		isEnabled:              true,
	}
}

func (tc *telemetryClient) Context() TelemetryContext {
	return tc.context
}

func (tc *telemetryClient) Channel() TelemetryChannel {
	return tc.channel
}

func (tc *telemetryClient) InstrumentationKey() string {
	return tc.TelemetryConfiguration.InstrumentationKey
}

func (tc *telemetryClient) IsEnabled() bool {
	return tc.isEnabled
}

func (tc *telemetryClient) SetIsEnabled(isEnabled bool) {
	tc.isEnabled = isEnabled
}

func (tc *telemetryClient) Track(item Telemetry) {
	if tc.isEnabled {
		iKey := tc.context.InstrumentationKey()
		if len(iKey) == 0 {
			iKey = tc.TelemetryConfiguration.InstrumentationKey
		}

		itemContext := item.Context().(*telemetryContext)
		itemContext.iKey = iKey

		clientContext := tc.context.(*telemetryContext)

		for tagkey, tagval := range clientContext.tags {
			if itemContext.tags[tagkey] == "" {
				itemContext.tags[tagkey] = tagval
			}
		}

		tc.channel.Send(item)
	}
}

func (tc *telemetryClient) TrackEvent(name string) {
	item := NewEventTelemetry(name)
	tc.TrackEventTelemetry(item)
}

func (tc *telemetryClient) TrackEventTelemetry(event *EventTelemetry) {
	var item Telemetry
	item = event

	tc.Track(item)
}

func (tc *telemetryClient) TrackMetric(name string, value float32) {
	item := NewMetricTelemetry(name, value)
	tc.TrackMetricTelemetry(item)
}

func (tc *telemetryClient) TrackMetricTelemetry(metric *MetricTelemetry) {
	var item Telemetry
	item = metric

	tc.Track(item)
}

func (tc *telemetryClient) TrackTrace(message string) {
	item := NewTraceTelemetry(message, Information)
	tc.TrackTraceTelemetry(item)
}

func (tc *telemetryClient) TrackTraceTelemetry(trace *TraceTelemetry) {
	var item Telemetry
	item = trace

	tc.Track(item)
}

func (tc *telemetryClient) TrackRequest(name, method, url string, timestamp time.Time, duration time.Duration, responseCode string, success bool) {
	item := NewRequestTelemetry(name, method, url, timestamp, duration, responseCode, success)
	tc.TrackRequestTelemetry(item)
}

func (tc *telemetryClient) TrackRequestTelemetry(request *RequestTelemetry) {
	var item Telemetry
	item = request

	tc.Track(item)
}
