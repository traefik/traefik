package tracing

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	entryPointTypeName = "TracingEntryPoint"
)

// NewEntryPoint creates a new middleware that the incoming request.
func NewEntryPoint(ctx context.Context, t *tracing.Tracing, entryPointName string, next http.Handler) http.Handler {
	log.FromContext(middlewares.GetLoggerCtx(ctx, "tracing", entryPointTypeName)).Debug("Creating middleware")

	return &entryPointMiddleware{
		entryPoint: entryPointName,
		Tracing:    t,
		next:       next,
	}
}

type entryPointMiddleware struct {
	*tracing.Tracing
	entryPoint string
	next       http.Handler
}

func (e *entryPointMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	spanCtx, err := e.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(req.Context(), "tracing", entryPointTypeName)).
			Debugf("Failed to extract the context: %v", err)
	}

	span, req, finish := e.StartSpanf(req, ext.SpanKindRPCServerEnum, "EntryPoint", []string{e.entryPoint, req.Host}, " ", ext.RPCServerOption(spanCtx))
	defer finish()

	ext.Component.Set(span, e.ServiceName)
	tracing.LogRequest(span, req)

	req = req.WithContext(tracing.WithTracing(req.Context(), e.Tracing))

	recorder := newStatusCodeRecoder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status())
}

// WrapEntryPointHandler Wraps tracing to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, tracer *tracing.Tracing, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewEntryPoint(ctx, tracer, entryPointName, next), nil
	}
}
