package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/traefik/traefik/v3/pkg/types"
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
		tracer = tracing.NewTracer(noop.Tracer{}, nil, nil, nil, "")
	}

	return &entryPointTracing{
		entryPoint: entryPointName,
		tracer:     tracer,
		next:       next,
	}
}

func (e *entryPointTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tracingCtx := tracing.ExtractCarrierIntoContext(req.Context(), req.Header)
	start := time.Now()
	// Initialise the span with the entry point name.
	spanName := e.getSpanName(req, e.entryPoint)
	tracingCtx, span := e.tracer.Start(tracingCtx, spanName, trace.WithSpanKind(trace.SpanKindServer), trace.WithTimestamp(start))

	// Associate the request context with the logger.
	logger := log.Ctx(tracingCtx).With().Ctx(tracingCtx).Logger()
	loggerCtx := logger.WithContext(tracingCtx)

	req = req.WithContext(loggerCtx)

	span.SetAttributes(attribute.String("entry_point", e.entryPoint))

	e.tracer.CaptureServerRequest(span, req)

	recorder := newRecorder(rw, http.StatusOK, spanName)
	e.next.ServeHTTP(recorder, req)

	e.tracer.CaptureResponse(span, recorder.Header(), recorder.Status(), trace.SpanKindServer)

	// Retrieve the last set SpanName from the recorder.
	spanName = recorder.SpanName()
	span.SetName(spanName)

	end := time.Now()
	span.End(trace.WithTimestamp(end))
}

// Translates the configured span name to it's dynamic value.
// Initially we use the entry point name, more detail is added further down the stack.
// The default value is "EntryPoint".
func (e *entryPointTracing) getSpanName(req *http.Request, entryPoint string) string {
	switch e.tracer.TraceNameFormat {
	case types.Static:
		return "EntryPoint " + entryPoint
	case types.Default:
		return "EntryPoint"
	case types.HostName:
		return req.Host
	case types.MethodAndRoute:
		return req.Method + " " + entryPoint
	default:
		return "EntryPoint"
	}
}
