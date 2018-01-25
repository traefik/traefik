package tracing

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/urfave/negroni"
)

type entryPointMiddleware struct {
	entryPoint string
	*Tracing
}

// NewEntryPoint creates a new middleware that the incoming request
func (t *Tracing) NewEntryPoint(name string) negroni.Handler {
	log.Debug("Added entrypoint tracing middleware")
	return &entryPointMiddleware{Tracing: t, entryPoint: name}
}

func (e *entryPointMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	opNameFunc := func(r *http.Request) string {
		return fmt.Sprintf("Entrypoint %s %s", e.entryPoint, r.Host)
	}

	ctx, _ := e.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := e.StartSpan(opNameFunc(r), ext.RPCServerOption(ctx))
	ext.Component.Set(span, e.ServiceName)
	LogRequest(span, r)
	ext.SpanKindRPCServer.Set(span)

	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

	recorder := newStatusCodeRecoder(w, 200)
	next(recorder, r)

	LogResponseCode(span, recorder.Status())
	span.Finish()
}
