package tracing

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	entryPointTypeName = "TracingEntryPoint"
)

// NewEntryPoint creates a new middleware that the incoming request.
func NewEntryPoint(ctx context.Context, t *tracing.Tracing, entryPointName string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", entryPointTypeName).Debug().Msg("Creating middleware")

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
	span, req, finish := e.StartSpanf(req, trace.SpanKindServer, "EntryPoint", []string{e.entryPoint, req.Host}, " ", trace.WithSpanKind(trace.SpanKindServer))
	defer finish()

	span.SetAttributes(attribute.String("component", e.ServiceName))
	tracing.LogRequest(span, req)

	req = req.WithContext(tracing.WithTracing(req.Context(), e.Tracing))

	recorder := newStatusCodeRecorder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status())
}

// WrapEntryPointHandler Wraps tracing to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, tracer *tracing.Tracing, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewEntryPoint(ctx, tracer, entryPointName, next), nil
	}
}
