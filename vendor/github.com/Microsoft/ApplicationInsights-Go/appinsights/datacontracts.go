package appinsights

import (
	"fmt"
	"time"
)

type Telemetry interface {
	Timestamp() time.Time
	Context() TelemetryContext
	baseTypeName() string
	baseData() Domain
	SetProperty(string, string)
}

type BaseTelemetry struct {
	timestamp time.Time
	context   TelemetryContext
}

type TraceTelemetry struct {
	BaseTelemetry
	data *messageData
}

func NewTraceTelemetry(message string, severityLevel SeverityLevel) *TraceTelemetry {
	now := time.Now()
	data := &messageData{
		Message:       message,
		SeverityLevel: severityLevel,
	}

	data.Ver = 2

	item := &TraceTelemetry{
		data: data,
	}

	item.timestamp = now
	item.context = NewItemTelemetryContext()

	return item
}

func (item *TraceTelemetry) Timestamp() time.Time {
	return item.timestamp
}

func (item *TraceTelemetry) Context() TelemetryContext {
	return item.context
}

func (item *TraceTelemetry) baseTypeName() string {
	return "Message"
}

func (item *TraceTelemetry) baseData() Domain {
	return item.data
}

func (item *TraceTelemetry) SetProperty(key, value string) {
	if item.data.Properties == nil {
		item.data.Properties = make(map[string]string)
	}
	item.data.Properties[key] = value
}

type EventTelemetry struct {
	BaseTelemetry
	data *eventData
}

func NewEventTelemetry(name string) *EventTelemetry {
	now := time.Now()
	data := &eventData{
		Name: name,
	}

	data.Ver = 2

	item := &EventTelemetry{
		data: data,
	}

	item.timestamp = now
	item.context = NewItemTelemetryContext()

	return item
}

func (item *EventTelemetry) Timestamp() time.Time {
	return item.timestamp
}

func (item *EventTelemetry) Context() TelemetryContext {
	return item.context
}

func (item *EventTelemetry) baseTypeName() string {
	return "Event"
}

func (item *EventTelemetry) baseData() Domain {
	return item.data
}

func (item *EventTelemetry) SetProperty(key, value string) {
	if item.data.Properties == nil {
		item.data.Properties = make(map[string]string)
	}
	item.data.Properties[key] = value
}

type MetricTelemetry struct {
	BaseTelemetry
	data *metricData
}

func NewMetricTelemetry(name string, value float32) *MetricTelemetry {
	now := time.Now()
	metric := &DataPoint{
		Name:  name,
		Value: value,
		Count: 1,
	}

	data := &metricData{
		Metrics: make([]*DataPoint, 1),
	}

	data.Ver = 2
	data.Metrics[0] = metric

	item := &MetricTelemetry{
		data: data,
	}

	item.timestamp = now
	item.context = NewItemTelemetryContext()

	return item
}

func (item *MetricTelemetry) Timestamp() time.Time {
	return item.timestamp
}

func (item *MetricTelemetry) Context() TelemetryContext {
	return item.context
}

func (item *MetricTelemetry) baseTypeName() string {
	return "Metric"
}

func (item *MetricTelemetry) baseData() Domain {
	return item.data
}

func (item *MetricTelemetry) SetProperty(key, value string) {
	if item.data.Properties == nil {
		item.data.Properties = make(map[string]string)
	}
	item.data.Properties[key] = value
}

type RequestTelemetry struct {
	BaseTelemetry
	data *requestData
}

func NewRequestTelemetry(name, httpMethod, url string, timestamp time.Time, duration time.Duration, responseCode string, success bool) *RequestTelemetry {
	now := time.Now()
	data := &requestData{
		Name:         name,
		StartTime:    timestamp.Format(time.RFC3339Nano),
		Duration:     formatDuration(duration),
		ResponseCode: responseCode,
		Success:      success,
		HttpMethod:   httpMethod,
		Url:          url,
		Id:           randomId(),
	}

	data.Ver = 2

	item := &RequestTelemetry{
		data: data,
	}

	item.timestamp = now
	item.context = NewItemTelemetryContext()

	return item
}

func (item *RequestTelemetry) Timestamp() time.Time {
	return item.timestamp
}

func (item *RequestTelemetry) Context() TelemetryContext {
	return item.context
}

func (item *RequestTelemetry) baseTypeName() string {
	return "Request"
}

func (item *RequestTelemetry) baseData() Domain {
	return item.data
}

func (item *RequestTelemetry) SetProperty(key, value string) {
	if item.data.Properties == nil {
		item.data.Properties = make(map[string]string)
	}
	item.data.Properties[key] = value
}

func formatDuration(d time.Duration) string {
	ticks := int64(d/(time.Nanosecond*100)) % 10000000
	seconds := int64(d/time.Second) % 60
	minutes := int64(d/time.Minute) % 60
	hours := int64(d/time.Hour) % 24
	days := int64(d / (time.Hour * 24))

	return fmt.Sprintf("%d.%02d:%02d:%02d.%07d", days, hours, minutes, seconds, ticks)
}
