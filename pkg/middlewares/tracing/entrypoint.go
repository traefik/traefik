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

type entryPointTracing struct {
	tracer     trace.Tracer
	entryPoint string
	next       http.Handler
}

// WrapEntryPointHandler Wraps tracing to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, tracer trace.Tracer, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newEntryPoint(ctx, tracer, entryPointName, next), nil
	}
}

// newEntryPoint creates a new tracing middleware for incoming requests.
func newEntryPoint(ctx context.Context, tracer trace.Tracer, entryPointName string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", entryPointTypeName).Debug().Msg("Creating middleware")

	return &entryPointTracing{
		entryPoint: entryPointName,
		tracer:     tracer,
		next:       next,
	}
}

func (e *entryPointTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tracingCtx := tracing.ExtractCarrierIntoContext(req.Context(), req.Header)
	tracingCtx, span := e.tracer.Start(tracingCtx, "EntryPoint", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	req = req.WithContext(tracingCtx)

	span.SetAttributes(attribute.String("entry_point", e.entryPoint))

	tracing.LogServerRequest(span, req)

	recorder := newStatusCodeRecorder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status(), trace.SpanKindServer)
}
