package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

func TestNewForwarder(t *testing.T) {
	type expected struct {
		Tags          map[string]interface{}
		OperationName string
	}

	testCases := []struct {
		desc          string
		spanNameLimit int
		tracing       *trackingBackenMock
		service       string
		router        string
		expected      expected
	}{
		{
			desc:          "Simple Forward Tracer without truncation and hashing",
			spanNameLimit: 101,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			service: "some-service.domain.tld",
			router:  "some-service.domain.tld",
			expected: expected{
				Tags: map[string]interface{}{
					"http.host":    "www.test.com",
					"http.method":  "GET",
					"http.url":     "http://www.test.com/toto",
					"service.name": "some-service.domain.tld",
					"router.name":  "some-service.domain.tld",
					"span.kind":    ext.SpanKindRPCClientEnum,
				},
				OperationName: "forward some-service.domain.tld/some-service.domain.tld",
			},
		},
		{
			desc:          "Simple Forward Tracer with truncation and hashing",
			spanNameLimit: 101,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			service: "some-service-100.slug.namespace.environment.domain.tld",
			router:  "some-service-100.slug.namespace.environment.domain.tld",
			expected: expected{
				Tags: map[string]interface{}{
					"http.host":    "www.test.com",
					"http.method":  "GET",
					"http.url":     "http://www.test.com/toto",
					"service.name": "some-service-100.slug.namespace.environment.domain.tld",
					"router.name":  "some-service-100.slug.namespace.environment.domain.tld",
					"span.kind":    ext.SpanKindRPCClientEnum,
				},
				OperationName: "forward some-service-100.slug.namespace.enviro.../some-service-100.slug.namespace.enviro.../bc4a0d48",
			},
		},
		{
			desc:          "Exactly 101 chars",
			spanNameLimit: 101,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			service: "some-service1.namespace.environment.domain.tld",
			router:  "some-service1.namespace.environment.domain.tld",
			expected: expected{
				Tags: map[string]interface{}{
					"http.host":    "www.test.com",
					"http.method":  "GET",
					"http.url":     "http://www.test.com/toto",
					"service.name": "some-service1.namespace.environment.domain.tld",
					"router.name":  "some-service1.namespace.environment.domain.tld",
					"span.kind":    ext.SpanKindRPCClientEnum,
				},
				OperationName: "forward some-service1.namespace.environment.domain.tld/some-service1.namespace.environment.domain.tld",
			},
		},
		{
			desc:          "More than 101 chars",
			spanNameLimit: 101,
			tracing: &trackingBackenMock{
				tracer: &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			service: "some-service1.frontend.namespace.environment.domain.tld",
			router:  "some-service1.backend.namespace.environment.domain.tld",
			expected: expected{
				Tags: map[string]interface{}{
					"http.host":    "www.test.com",
					"http.method":  "GET",
					"http.url":     "http://www.test.com/toto",
					"service.name": "some-service1.frontend.namespace.environment.domain.tld",
					"router.name":  "some-service1.backend.namespace.environment.domain.tld",
					"span.kind":    ext.SpanKindRPCClientEnum,
				},
				OperationName: "forward some-service1.frontend.namespace.envir.../some-service1.backend.namespace.enviro.../fa49dd23",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			newTracing, err := tracing.NewTracing("", test.spanNameLimit, test.tracing)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/toto", nil)
			req = req.WithContext(tracing.WithTracing(req.Context(), newTracing))

			rw := httptest.NewRecorder()

			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				span := test.tracing.tracer.(*MockTracer).Span

				tags := span.Tags
				assert.Equal(t, test.expected.Tags, tags)
				assert.True(t, len(test.expected.OperationName) <= test.spanNameLimit,
					"the len of the operation name %q [len: %d] doesn't respect limit %d",
					test.expected.OperationName, len(test.expected.OperationName), test.spanNameLimit)
				assert.Equal(t, test.expected.OperationName, span.OpName)
			})

			handler := NewForwarder(context.Background(), test.router, test.service, next)
			handler.ServeHTTP(rw, req)
		})
	}
}
