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
	opNameFunc := generateEntryPointSpanName

	ctx, _ := e.Extract(opentracing.HTTPHeaders, HTTPHeadersCarrier(r.Header))
	span := e.StartSpan(opNameFunc(r, e.entryPoint, e.SpanNameLimit), ext.RPCServerOption(ctx))
	ext.Component.Set(span, e.ServiceName)
	LogRequest(span, r)
	ext.SpanKindRPCServer.Set(span)

	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))

	recorder := newStatusCodeRecoder(w, 200)
	next(recorder, r)

	LogResponseCode(span, recorder.Status())
	span.Finish()
}

// generateEntryPointSpanName will return a Span name of an appropriate lenth based on the 'spanLimit' argument.  If needed, it will be truncated, but will not be less than 24 characters.
func generateEntryPointSpanName(r *http.Request, entryPoint string, spanLimit int) string {
	name := fmt.Sprintf("Entrypoint %s %s", entryPoint, r.Host)

	if spanLimit > 0 && len(name) > spanLimit {
		if spanLimit < EntryPointMaxLengthNumber {
			log.Warnf("SpanNameLimit is set to be less than required static number of characters, defaulting to %d + 3", EntryPointMaxLengthNumber)
			spanLimit = EntryPointMaxLengthNumber + 3
		}
		hash := computeHash(name)
		limit := (spanLimit - EntryPointMaxLengthNumber) / 2
		name = fmt.Sprintf("Entrypoint %s %s %s", truncateString(entryPoint, limit), truncateString(r.Host, limit), hash)
	}

	return name
}
