package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func TestEntryPointMiddlewareServeHTTP(t *testing.T) {
	type fields struct {
		entryPoint string
		Tracing    *Tracing
	}
	type args struct {
		w    http.ResponseWriter
		r    *http.Request
		next http.HandlerFunc
	}

	testCases := []struct {
		desc   string
		fields fields
		args   args
	}{
		{
			desc: "basic test",
			fields: fields{
				entryPoint: "test",
				Tracing: &Tracing{
					SpanNameLimit: 25,
					tracer:        defaultMockTracer,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "http://www.test.com", nil),
				next: func(http.ResponseWriter, *http.Request) {
					// Asserts go here...
					want := make(map[string]interface{})
					want["span.kind"] = ext.SpanKindRPCServerEnum
					want["http.method"] = "GET"
					want["component"] = ""
					want["http.url"] = "http://www.test.com"
					want["http.host"] = "www.test.com"

					got := defaultMockSpan.Tags
					assert.Equal(t, want, got, "ServeHTTP() = %+v want %+v", got, want)
					assert.Equal(t, "Entrypoint te... ww... 39b97e58", defaultMockSpan.OpName)
				},
			},
		},
		{
			desc: "no truncation test",
			fields: fields{
				entryPoint: "test",
				Tracing: &Tracing{
					SpanNameLimit: 0,
					tracer:        defaultMockTracer,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "http://www.test.com", nil),
				next: func(http.ResponseWriter, *http.Request) {
					// Asserts go here...
					want := make(map[string]interface{})
					want["span.kind"] = ext.SpanKindRPCServerEnum
					want["http.method"] = "GET"
					want["component"] = ""
					want["http.url"] = "http://www.test.com"
					want["http.host"] = "www.test.com"

					got := defaultMockSpan.Tags
					assert.Equal(t, want, got, "ServeHTTP() = %+v want %+v", got, want)
					assert.Equal(t, "Entrypoint test www.test.com", defaultMockSpan.OpName)
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defaultMockSpan.Reset()
			e := &entryPointMiddleware{
				entryPoint: test.fields.entryPoint,
				Tracing:    test.fields.Tracing,
			}

			e.ServeHTTP(test.args.w, test.args.r, test.args.next)
		})
	}
}
