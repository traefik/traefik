package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type MockTracer struct {
	embedded.Tracer
	Span *MockSpan
}

// Start belongs to the Tracer interface.
func (n MockTracer) Start(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	c := trace.NewSpanStartConfig(opts...)
	n.Span.SetAttributes(c.Attributes()...)
	n.Span.SetAttributes(attribute.String("span.kind", c.SpanKind().String()))
	n.Span.SpanName = operationName
	return ctx, n.Span
}

// StartSpan belongs to the Tracer interface.
func (n MockTracer) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	n.Span.SpanName = operationName
	return ctx, n.Span
}

// GetServiceName returns the service name.
func (n MockTracer) GetServiceName() string {
	return ""
}

// Close mocks of Close.
func (n MockTracer) Close() {}

// MockSpan a span mock.
type MockSpan struct {
	embedded.Span
	SpanName string
	Tags     map[string]interface{}
	Conf     trace.SpanContextConfig
	Provider trace.TracerProvider
}

func (n MockSpan) AddEvent(event string, foo ...trace.EventOption)     {}
func (n MockSpan) End(options ...trace.SpanEndOption)                  {}
func (n MockSpan) IsRecording() bool                                   { return false }
func (n MockSpan) RecordError(err error, options ...trace.EventOption) {}
func (n MockSpan) SpanContext() trace.SpanContext                      { return trace.NewSpanContext(n.Conf) }
func (n MockSpan) SetStatus(code codes.Code, description string)       {}
func (n MockSpan) SetName(name string)                                 {}
func (n MockSpan) SetAttributes(kv ...attribute.KeyValue) {
	for _, v := range kv {
		n.Tags[string(v.Key)] = v.Value.AsInterface()
	}
}
func (n MockSpan) TracerProvider() trace.TracerProvider { return n.Provider }
