package tracing

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

type MockTracer struct {
	Span *MockSpan
}

type MockSpan struct {
	OpName string
	Tags   map[string]interface{}
}

type MockSpanContext struct {
}

// MockSpanContext:
func (n MockSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

// MockSpan:
func (n MockSpan) Context() opentracing.SpanContext { return MockSpanContext{} }
func (n MockSpan) SetBaggageItem(key, val string) opentracing.Span {
	return MockSpan{Tags: make(map[string]interface{})}
}
func (n MockSpan) BaggageItem(key string) string { return "" }
func (n MockSpan) SetTag(key string, value interface{}) opentracing.Span {
	n.Tags[key] = value
	return n
}
func (n MockSpan) LogFields(fields ...log.Field)                          {}
func (n MockSpan) LogKV(keyVals ...interface{})                           {}
func (n MockSpan) Finish()                                                {}
func (n MockSpan) FinishWithOptions(opts opentracing.FinishOptions)       {}
func (n MockSpan) SetOperationName(operationName string) opentracing.Span { return n }
func (n MockSpan) Tracer() opentracing.Tracer                             { return MockTracer{} }
func (n MockSpan) LogEvent(event string)                                  {}
func (n MockSpan) LogEventWithPayload(event string, payload interface{})  {}
func (n MockSpan) Log(data opentracing.LogData)                           {}
func (n MockSpan) Reset() {
	n.Tags = make(map[string]interface{})
}

// StartSpan belongs to the Tracer interface.
func (n MockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	n.Span.OpName = operationName
	return n.Span
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
	testCases := []struct {
		desc     string
		text     string
		limit    int
		expected string
	}{
		{
			desc:     "short text less than limit 10",
			text:     "short",
			limit:    10,
			expected: "short",
		},
		{
			desc:     "basic truncate with limit 10",
			text:     "some very long pice of text",
			limit:    10,
			expected: "some ve...",
		},
		{
			desc:     "truncate long FQDN to 39 chars",
			text:     "some-service-100.slug.namespace.environment.domain.tld",
			limit:    39,
			expected: "some-service-100.slug.namespace.envi...",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := truncateString(test.text, test.limit)

			assert.Equal(t, test.expected, actual)
			assert.True(t, len(actual) <= test.limit)
		})
	}
}

func TestComputeHash(t *testing.T) {
	testCases := []struct {
		desc     string
		text     string
		expected string
	}{
		{
			desc:     "hashing",
			text:     "some very long pice of text",
			expected: "0258ea1c",
		},
		{
			desc:     "short text less than limit 10",
			text:     "short",
			expected: "f9b0078b",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := computeHash(test.text)

			assert.Equal(t, test.expected, actual)
		})
	}
}
