/*
 *  Copyright 2018 Expedia Group.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

package haystack

import (
	"encoding/json"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

/*_Span implements opentracing.Span*/
type _Span struct {
	tracer  *Tracer
	context *SpanContext

	operationName string

	// startTime is the timestamp indicating when the span began, with microseconds precision.
	startTime time.Time
	// duration returns duration of the span with microseconds precision.
	duration time.Duration

	tags []opentracing.Tag
	logs []opentracing.LogRecord
}

// SetOperationName sets or changes the operation name.
func (span *_Span) SetOperationName(operationName string) opentracing.Span {
	span.operationName = operationName
	return span
}

// SetTag implements SetTag() of opentracing.Span
func (span *_Span) SetTag(key string, value interface{}) opentracing.Span {
	span.tags = append(span.tags,
		opentracing.Tag{
			Key:   key,
			Value: value,
		})
	return span
}

// LogFields implements opentracing.Span API
func (span *_Span) LogFields(fields ...log.Field) {
	log := opentracing.LogRecord{
		Fields:    fields,
		Timestamp: time.Now(),
	}
	span.logs = append(span.logs, log)
}

// LogKV implements opentracing.Span API
func (span *_Span) LogKV(alternatingKeyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(alternatingKeyValues...)
	if err != nil {
		span.LogFields(log.Error(err), log.String("function", "LogKV"))
		return
	}
	span.LogFields(fields...)
}

// LogEvent implements opentracing.Span API
func (span *_Span) LogEvent(event string) {
	span.Log(opentracing.LogData{Event: event})
}

// LogEventWithPayload implements opentracing.Span API
func (span *_Span) LogEventWithPayload(event string, payload interface{}) {
	span.Log(opentracing.LogData{Event: event, Payload: payload})
}

// Log implements opentracing.Span API
func (span *_Span) Log(ld opentracing.LogData) {
	span.logs = append(span.logs, ld.ToLogRecord())
}

// SetBaggageItem implements SetBaggageItem() of opentracing.SpanContext
func (span *_Span) SetBaggageItem(key, value string) opentracing.Span {
	span.context = span.context.WithBaggageItem(key, value)
	span.LogFields(log.String("event", "baggage"), log.String("payload", key), log.String("payload", value))
	return span
}

// BaggageItem implements BaggageItem() of opentracing.SpanContext
func (span *_Span) BaggageItem(key string) string {
	return span.context.Baggage[key]
}

// Finish implements opentracing.Span API
func (span *_Span) Finish() {
	span.FinishWithOptions(opentracing.FinishOptions{})
}

// FinishWithOptions implements opentracing.Span API
func (span *_Span) FinishWithOptions(options opentracing.FinishOptions) {
	if options.FinishTime.IsZero() {
		options.FinishTime = span.tracer.timeNow()
	}
	span.duration = options.FinishTime.Sub(span.startTime)
	if options.LogRecords != nil {
		span.logs = append(span.logs, options.LogRecords...)
	}
	for _, ld := range options.BulkLogData {
		span.logs = append(span.logs, ld.ToLogRecord())
	}
	span.tracer.DispatchSpan(span)
}

// Context implements opentracing.Span API
func (span *_Span) Context() opentracing.SpanContext {
	return span.context
}

/*Tracer returns the tracer*/
func (span *_Span) Tracer() opentracing.Tracer {
	return span.tracer
}

/*OperationName allows retrieving current operation name*/
func (span *_Span) OperationName() string {
	return span.operationName
}

/*ServiceName returns the name of the service*/
func (span *_Span) ServiceName() string {
	return span.tracer.serviceName
}

func (span *_Span) String() string {
	data, err := json.Marshal(map[string]interface{}{
		"traceId":       span.context.TraceID,
		"spanId":        span.context.SpanID,
		"parentSpanId":  span.context.ParentID,
		"operationName": span.OperationName(),
		"serviceName":   span.ServiceName(),
		"tags":          span.Tags(),
		"logs":          span.logs,
	})
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (span *_Span) Tags() []opentracing.Tag {
	return span.tags
}
