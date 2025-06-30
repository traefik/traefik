package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func TestEntryPointMiddleware_tracing(t *testing.T) {
	type expected struct {
		name       string
		attributes []attribute.KeyValue
	}

	testCases := []struct {
		desc       string
		entryPoint string
		expected   expected
	}{
		{
			desc:       "basic test",
			entryPoint: "test",
			expected: expected{
				name: "EntryPoint",
				attributes: []attribute.KeyValue{
					attribute.String("span.kind", "server"),
					attribute.String("entry_point", "test"),
					attribute.String("http.request.method", "GET"),
					attribute.String("network.protocol.version", "1.1"),
					attribute.Int64("http.request.body.size", int64(0)),
					attribute.String("url.path", "/search"),
					attribute.String("url.query", "q=Opentelemetry&token=REDACTED"),
					attribute.String("url.scheme", "http"),
					attribute.String("user_agent.original", "entrypoint-test"),
					attribute.String("server.address", "www.test.com"),
					attribute.String("network.peer.address", "10.0.0.1"),
					attribute.String("client.address", "10.0.0.1"),
					attribute.Int64("client.port", int64(1234)),
					attribute.Int64("network.peer.port", int64(1234)),
					attribute.StringSlice("http.request.header.x-foo", []string{"foo", "bar"}),
					attribute.Int64("http.response.status_code", int64(404)),
					attribute.StringSlice("http.response.header.x-bar", []string{"foo", "bar"}),
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry&token=123", nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")
			req.Header.Set("X-Forwarded-Proto", "http")
			req.Header.Set("X-Foo", "foo")
			req.Header.Add("X-Foo", "bar")

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("X-Bar", "foo")
				rw.Header().Add("X-Bar", "bar")
				rw.WriteHeader(http.StatusNotFound)
			})

			tracer := &mockTracer{}

			handler := newEntryPoint(t.Context(), tracing.NewTracer(tracer, []string{"X-Foo"}, []string{"X-Bar"}, []string{"q"}), test.entryPoint, next)
			handler.ServeHTTP(rw, req)

			for _, span := range tracer.spans {
				assert.Equal(t, test.expected.name, span.name)
				assert.Equal(t, test.expected.attributes, span.attributes)
			}
		})
	}
}
