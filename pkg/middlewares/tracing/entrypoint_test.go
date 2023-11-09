package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
)

func TestEntryPointMiddleware(t *testing.T) {
	type expected struct {
		Tags          map[string]interface{}
		OperationName string
	}

	testCases := []struct {
		desc          string
		entryPoint    string
		spanNameLimit int
		tracing       *trackingBackenMock
		expected      expected
	}{
		{
			desc:          "no truncation test",
			entryPoint:    "test",
			spanNameLimit: 0,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			expected: expected{
				Tags: map[string]interface{}{
					"span.kind":                   trace.SpanKindServer.String(),
					"http.method":                 http.MethodGet,
					"component":                   "",
					"http.url":                    "http://www.test.com",
					"http.host":                   "www.test.com",
					"http.client_ip":              "10.0.0.1",
					"http.request_content_length": int64(0),
					"http.user_agent":             "entrypoint-test",
				},
				OperationName: "EntryPoint test www.test.com",
			},
		},
		{
			desc:          "basic test",
			entryPoint:    "test",
			spanNameLimit: 25,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			expected: expected{
				Tags: map[string]interface{}{
					"span.kind":                   trace.SpanKindServer.String(),
					"http.method":                 http.MethodGet,
					"component":                   "",
					"http.url":                    "http://www.test.com",
					"http.host":                   "www.test.com",
					"http.client_ip":              "10.0.0.1",
					"http.request_content_length": int64(0),
					"http.user_agent":             "entrypoint-test",
				},
				OperationName: "EntryPoint te... ww... 0c15301b",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			newTracing, err := tracing.NewTracing("", test.spanNameLimit, test.tracing)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				span := test.tracing.tracer.(*MockTracer).Span

				tags := span.Tags
				assert.Equal(t, test.expected.Tags, tags)
				assert.Equal(t, test.expected.OperationName, span.OpName)
			})

			handler := NewEntryPoint(context.Background(), newTracing, test.entryPoint, next)
			handler.ServeHTTP(rw, req)
		})
	}
}
