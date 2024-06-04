package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

type mockTracerProvider struct {
	embedded.TracerProvider
}

var _ trace.TracerProvider = mockTracerProvider{}

func (p mockTracerProvider) Tracer(string, ...trace.TracerOption) trace.Tracer {
	return &mockTracer{}
}

type mockTracer struct {
	embedded.Tracer

	spans []*mockSpan
}

var _ trace.Tracer = &mockTracer{}

func (t *mockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	config := trace.NewSpanStartConfig(opts...)
	span := &mockSpan{}
	span.SetName(name)
	span.SetAttributes(attribute.String("span.kind", config.SpanKind().String()))
	span.SetAttributes(config.Attributes()...)
	t.spans = append(t.spans, span)
	return trace.ContextWithSpan(ctx, span), span
}

// mockSpan is an implementation of Span that preforms no operations.
type mockSpan struct {
	embedded.Span

	name       string
	attributes []attribute.KeyValue
}

var _ trace.Span = &mockSpan{}

func (*mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}})
}
func (*mockSpan) IsRecording() bool                  { return false }
func (s *mockSpan) SetStatus(_ codes.Code, _ string) {}
func (s *mockSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.attributes = append(s.attributes, kv...)
}
func (s *mockSpan) End(...trace.SpanEndOption)                  {}
func (s *mockSpan) RecordError(_ error, _ ...trace.EventOption) {}
func (s *mockSpan) AddEvent(_ string, _ ...trace.EventOption)   {}
func (s *mockSpan) AddLink(_ trace.Link)                        {}

func (s *mockSpan) SetName(name string) { s.name = name }

func (s *mockSpan) TracerProvider() trace.TracerProvider {
	return nil
}
