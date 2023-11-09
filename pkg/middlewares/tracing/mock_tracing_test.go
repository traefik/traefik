package tracing

import (
	"context"
	"io"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type MockTracer struct {
	Span *MockSpan
}

// Start belongs to the Tracer interface.
func (n MockTracer) Start(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	c := trace.NewSpanStartConfig(opts...)
	n.Span.SetAttributes(c.Attributes()...)
	n.Span.SetAttributes(attribute.String("span.kind", c.SpanKind().String()))
	n.Span.OpName = operationName
	return ctx, n.Span
}

// StartSpan belongs to the Tracer interface.
func (n MockTracer) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	n.Span.OpName = operationName
	return ctx, n.Span
}

// MockSpan a span mock.
type MockSpan struct {
	OpName   string
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

type trackingBackenMock struct {
	tracer trace.Tracer
}

func (t *trackingBackenMock) Setup(componentName string) (trace.Tracer, io.Closer, error) {
	return t.tracer, io.NopCloser(nil), nil
}
