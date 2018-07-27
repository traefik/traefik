package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"testing"
)

type MockTracer struct {
}

type MockSpan struct {
	OpName string
	Tags   map[string]interface{}
}

type MockSpanContext struct {
}

var (
	defaultMockSpanContext = MockSpanContext{}
	defaultMockSpan        = MockSpan{
		Tags: make(map[string]interface{}),
	}
	defaultMockTracer = MockTracer{}
)

// MockSpanContext:
func (n MockSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

// MockSpan:
func (n MockSpan) Context() opentracing.SpanContext                { return defaultMockSpanContext }
func (n MockSpan) SetBaggageItem(key, val string) opentracing.Span { return defaultMockSpan }
func (n MockSpan) BaggageItem(key string) string                   { return "" }
func (n MockSpan) SetTag(key string, value interface{}) opentracing.Span {
	n.Tags[key] = value
	return n
}
func (n MockSpan) LogFields(fields ...log.Field)                          {}
func (n MockSpan) LogKV(keyVals ...interface{})                           {}
func (n MockSpan) Finish()                                                {}
func (n MockSpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (n MockSpan) SetOperationName(operationName string) opentracing.Span { return n }
func (n MockSpan) Tracer() opentracing.Tracer                             { return defaultMockTracer }
func (n MockSpan) LogEvent(event string)                                  {}
func (n MockSpan) LogEventWithPayload(event string, payload interface{})  {}
func (n MockSpan) Log(data opentracing.LogData)                           {}

// StartSpan belongs to the Tracer interface.
func (n MockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	defaultMockSpan.OpName = operationName
	return defaultMockSpan
}

// Inject belongs to the Tracer interface.
func (n MockTracer) Inject(sp opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}

// Extract belongs to the Tracer interface.
func (n MockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, opentracing.ErrSpanContextNotFound
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		desc  string
		text  string
		limit int
		want  string
	}{
		{
			desc:  "basic truncate with limit 10",
			text:  "some very long pice of text",
			limit: 10,
			want:  "some ve...",
		},
		{
			desc:  "short text less than limit 10",
			text:  "short",
			limit: 10,
			want:  "short",
		},
		{
			desc:  "truncate long FQDN to 39 chars",
			text:  "some-service-100.slug.namespace.environment.domain.tld",
			limit: 39,
			want:  "some-service-100.slug.namespace.envi...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := TruncateString(tt.text, tt.limit); got != tt.want || len(got) > tt.limit {
				t.Errorf("TruncateString() = %v, want %v - limit %v, length %v", got, tt.want, tt.limit, len(got))
			}
		})
	}
}

func TestComputeHash(t *testing.T) {
	tests := []struct {
		desc string
		text string
		want string
	}{
		{
			desc: "hashing",
			text: "some very long pice of text",
			want: "0258ea1c",
		},
		{
			desc: "short text less than limit 10",
			text: "short",
			want: "f9b0078b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := ComputeHash(tt.text); got != tt.want {
				t.Errorf("ComputeHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
