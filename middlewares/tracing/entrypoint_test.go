package tracing

import (
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_entryPointMiddleware_ServeHTTP(t *testing.T) {
	type fields struct {
		entryPoint string
		Tracing    *Tracing
	}
	type args struct {
		w    http.ResponseWriter
		r    *http.Request
		next http.HandlerFunc
	}
	tests := []struct {
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
					assert.Equal(t, got, want, "ServeHTTP() = %+v want %+v", got, want)
					assert.Equal(t, defaultMockSpan.OpName, "Entrypoint te... ww... 39b97e58")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			e := &entryPointMiddleware{
				entryPoint: tt.fields.entryPoint,
				Tracing:    tt.fields.Tracing,
			}
			e.ServeHTTP(tt.args.w, tt.args.r, tt.args.next)
		})
	}
}
