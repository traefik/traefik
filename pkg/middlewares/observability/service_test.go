package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNewService(t *testing.T) {
	type expected struct {
		attributes []attribute.KeyValue
		name       string
	}

	testCases := []struct {
		desc     string
		service  string
		expected []expected
	}{
		{
			desc:    "base",
			service: "myService",
			expected: []expected{
				{
					name: "EntryPoint",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "server"),
					},
				},
				{
					name: "Service",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "internal"),
						attribute.String("http.request.method", "GET"),
						attribute.Int64("http.response.status_code", int64(404)),
						attribute.String("network.protocol.version", "1.1"),
						attribute.String("server.address", "www.test.com"),
						attribute.Int64("server.port", int64(80)),
						attribute.String("url.full", "http://www.test.com/traces?p=OpenTelemetry"),
						attribute.String("url.scheme", "http"),
						attribute.String("traefik.service.name", "myService"),
						attribute.String("user_agent.original", "service-test"),
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/traces?p=OpenTelemetry", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "service-test")

			tracer := &mockTracer{}
			tracingCtx, entryPointSpan := tracer.Start(req.Context(), "EntryPoint", trace.WithSpanKind(trace.SpanKindServer))
			defer entryPointSpan.End()

			req = req.WithContext(tracingCtx)

			rw := httptest.NewRecorder()
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			handler := NewService(context.Background(), test.service, next)
			handler.ServeHTTP(rw, req)

			for i, span := range tracer.spans {
				assert.Equal(t, test.expected[i].name, span.name)
				assert.Equal(t, test.expected[i].attributes, span.attributes)
			}
		})
	}
}
