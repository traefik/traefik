package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	entryPointTypeName = "TracingEntryPoint"
)

type entryPointTracing struct {
	tracer *tracing.Tracer

	entryPoint string
	next       http.Handler
}

// EntryPointHandler Wraps tracing to alice.Constructor.
func EntryPointHandler(ctx context.Context, tracer *tracing.Tracer, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newEntryPoint(ctx, tracer, entryPointName, next), nil
	}
}

// newEntryPoint creates a new tracing middleware for incoming requests.
func newEntryPoint(ctx context.Context, tracer *tracing.Tracer, entryPointName string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", entryPointTypeName).Debug().Msg("Creating middleware")

	if tracer == nil {
		tracer = tracing.NewTracer(noop.Tracer{}, nil, nil, nil)
	}

	return &entryPointTracing{
		entryPoint: entryPointName,
		tracer:     tracer,
		next:       next,
	}
}

func (e *entryPointTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if e.tracer == nil || !TracingEnabled(req.Context()) {
		e.next.ServeHTTP(rw, req)
		return
	}

	tracingCtx := tracing.ExtractCarrierIntoContext(req.Context(), req.Header)
	start := time.Now()

	// Follow semantic conventions defined by OTEL: https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name
	// The span name initially uses only {method} (e.g. "GET") because the route
	// is not yet known at the entrypoint level. When a router matches the
	// request, the router tracing middleware enriches this server span with
	// the http.route attribute and updates the span name to "{method} {route}"
	// (e.g. "GET /api/v1/ml-scribe"). If no router matches (404), the span
	// name remains "{method}".
	tracingCtx, span := e.tracer.Start(tracingCtx, req.Method, trace.WithSpanKind(trace.SpanKindServer), trace.WithTimestamp(start))

	// Associate the request context with the logger.
	// This allows the logger to be aware of the tracing context and log accordingly (TraceID, SpanID, etc.).
	logger := log.Ctx(tracingCtx).With().Ctx(tracingCtx).Logger()
	loggerCtx := logger.WithContext(tracingCtx)

	req = req.WithContext(loggerCtx)

	span.SetAttributes(attribute.String("entry_point", e.entryPoint))

	e.tracer.CaptureServerRequest(span, req)

	recorder := newStatusCodeRecorder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	e.tracer.CaptureResponse(span, recorder.Header(), recorder.Status(), trace.SpanKindServer)

	end := time.Now()
	span.End(trace.WithTimestamp(end))
}
