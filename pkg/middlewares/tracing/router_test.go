package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/tracing"
)

func TestNewForwarder(t *testing.T) {
	type expected struct {
		tags     map[string]interface{}
		spanName string
	}

	testCases := []struct {
		desc     string
		tracer   tracing.Tracer
		service  string
		router   string
		expected expected
	}{
		{
			desc:    "base",
			tracer:  &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			service: "some-service.domain.tld",
			router:  "some-service.domain.tld",
			expected: expected{
				tags: map[string]interface{}{
					"span.kind":                 "internal",
					"http.request.method":       "GET",
					"http.response.status_code": int64(404),
					"network.protocol.version":  "1.1",
					"server.address":            "www.test.com",
					"server.port":               int64(80),
					"url.full":                  "http://www.test.com/traces?p=OpenTelemetry",
					"url.scheme":                "http",
					"traefik.service.name":      "some-service.domain.tld",
					"traefik.router.name":       "some-service.domain.tld",
					"user_agent.original":       "forwarder-test",
				},
				spanName: "router",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/traces?p=OpenTelemetry", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "forwarder-test")
			rw := httptest.NewRecorder()

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			handler := newRouter(context.Background(), test.tracer, test.router, test.service, next)
			handler.ServeHTTP(rw, req)

			span := test.tracer.(*MockTracer).Span

			tags := span.Tags
			assert.Equal(t, test.expected.tags, tags)
			assert.Equal(t, test.expected.spanName, span.SpanName)
		})
	}
}
