package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
)

func TestEntryPointMiddleware(t *testing.T) {
	type expected struct {
		Tags     map[string]interface{}
		SpanName string
	}

	testCases := []struct {
		desc       string
		entryPoint string
		tracing    tracing.Tracer
		expected   expected
	}{
		{
			desc:       "basic test",
			entryPoint: "test",
			tracing:    &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			expected: expected{
				Tags: map[string]interface{}{
					"span.kind":                 trace.SpanKindServer.String(),
					"http.request.method":       http.MethodGet,
					"http.response.status_code": int64(http.StatusNotFound),
					"client.port":               int64(1234),
					"client.socket.address":     "",
					"http.request.body.size":    int64(0),
					"network.protocol.version":  "1.1",
					"client.address":            "10.0.0.1",
					"server.address":            "www.test.com",
					"url.path":                  "/search",
					"url.query":                 "q=Opentelemetry",
					"url.scheme":                "http",
					"user_agent.original":       "MyUserAgent",
					"component":                 "",
					"entry_point":               "test",
				},
				SpanName: "entry_point",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "MyUserAgent")
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			handler := newEntryPoint(context.Background(), test.tracing, test.entryPoint, next)
			handler.ServeHTTP(rw, req)

			span := test.tracing.(*MockTracer).Span

			tags := span.Tags
			assert.Equal(t, test.expected.Tags, tags)
			assert.Equal(t, test.expected.SpanName, span.SpanName)
		})
	}
}
