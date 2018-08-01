package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestEntryPointMiddlewareServeHTTP(t *testing.T) {
	expectedTags := map[string]interface{}{
		"span.kind":   ext.SpanKindRPCServerEnum,
		"http.method": "GET",
		"component":   "",
		"http.url":    "http://www.test.com",
		"http.host":   "www.test.com",
	}

	testCases := []struct {
		desc         string
		entryPoint   string
		tracing      *Tracing
		expectedTags map[string]interface{}
		expectedName string
	}{
		{
			desc:       "no truncation test",
			entryPoint: "test",
			tracing: &Tracing{
				SpanNameLimit: 0,
				tracer:        &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			expectedTags: expectedTags,
			expectedName: "Entrypoint test www.test.com",
		}, {
			desc:       "basic test",
			entryPoint: "test",
			tracing: &Tracing{
				SpanNameLimit: 25,
				tracer:        &MockTracer{Span: &MockSpan{Tags: make(map[string]interface{})}},
			},
			expectedTags: expectedTags,
			expectedName: "Entrypoint te... ww... 39b97e58",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			e := &entryPointMiddleware{
				entryPoint: test.entryPoint,
				Tracing:    test.tracing,
			}

			next := func(http.ResponseWriter, *http.Request) {
				span := test.tracing.tracer.(*MockTracer).Span

				actual := span.Tags
				assert.Equal(t, test.expectedTags, actual)
				assert.Equal(t, test.expectedName, span.OpName)
			}

			e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://www.test.com", nil), next)
		})
	}
}
