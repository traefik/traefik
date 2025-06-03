package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNewRouter(t *testing.T) {
	type expected struct {
		attributes []attribute.KeyValue
		name       string
	}

	testCases := []struct {
		desc            string
		service         string
		router          string
		routerRule      string
		traceNameFormat string
		expected        []expected
	}{
		{
			desc:            "base",
			service:         "myService",
			router:          "myRouter",
			routerRule:      "Path(`/`)",
			traceNameFormat: "default",
			expected: []expected{
				{
					name: "EntryPoint",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "server"),
					},
				},
				{
					name: "Router",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "internal"),
						attribute.String("traefik.service.name", "myService"),
						attribute.String("traefik.router.name", "myRouter"),
						attribute.String("http.route", "Path(`/`)"),
					},
				},
			},
		},
		{
			desc:            "static span name",
			service:         "myService",
			router:          "myRouter",
			routerRule:      "Path(`/`)",
			traceNameFormat: "static",
			expected: []expected{
				{
					name: "Router myRouter",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "server"),
					},
				},
				{
					name: "Router",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "internal"),
						attribute.String("traefik.service.name", "myService"),
						attribute.String("traefik.router.name", "myRouter"),
						attribute.String("http.route", "Path(`/`)"),
					},
				},
			},
		},
		{
			desc:            "Method and route span name",
			service:         "myService",
			router:          "myRouter",
			routerRule:      "Path(`/`)",
			traceNameFormat: "methodAndRoute",
			expected: []expected{
				{
					name: "GET Path(`/`)",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "server"),
					},
				},
				{
					name: "Router",
					attributes: []attribute.KeyValue{
						attribute.String("span.kind", "internal"),
						attribute.String("traefik.service.name", "myService"),
						attribute.String("traefik.router.name", "myRouter"),
						attribute.String("http.route", "Path(`/`)"),
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/traces?p=OpenTelemetry", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "router-test")

			tracer := &mockTracer{}
			tracer.TraceNameFormat = test.traceNameFormat
			tracingCtx, entryPointSpan := tracer.Start(req.Context(), "EntryPoint", trace.WithSpanKind(trace.SpanKindServer))
			defer entryPointSpan.End()

			req = req.WithContext(tracingCtx)

			rw := newRecorder(httptest.NewRecorder(), http.StatusOK, "EntryPoint")
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			handler := newRouter(t.Context(), test.router, test.routerRule, test.service, next)
			handler.ServeHTTP(rw, req)

			entryPointSpan.SetName(rw.SpanName())

			for i, span := range tracer.spans {
				assert.Equal(t, test.expected[i].name, span.name)
				assert.Equal(t, test.expected[i].attributes, span.attributes)
			}
			assert.Len(t, tracer.spans, len(test.expected), "Expected number of spans does not match actual number of spans")
		})
	}
}
