package tracing

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/trace"
)

// Traceable embeds tracing information.
type Traceable interface {
	GetTracingInformation() (name string, spanKind trace.SpanKind)
}

// WrapMiddleware adds traceability to an alice.Constructor.
func WrapMiddleware(ctx context.Context, constructor alice.Constructor, tracer tracing.Tracer) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if constructor == nil {
			return nil, nil
		}
		handler, err := constructor(next)
		if err != nil {
			return nil, err
		}

		if traceableHandler, ok := handler.(Traceable); ok {
			name, spanKind := traceableHandler.GetTracingInformation()
			log.Ctx(ctx).Debug().Str(logs.MiddlewareName, name).Msg("Adding tracing to middleware")
			return newMiddleware(handler, name, spanKind, tracer), nil
		}
		return handler, nil
	}
}

// newMiddleware returns a http.Handler struct.
func newMiddleware(next http.Handler, name string, spanKind trace.SpanKind, tracer tracing.Tracer) http.Handler {
	return &middlewareTracing{
		next:     next,
		name:     name,
		spanKind: spanKind,
		tracer:   tracer,
	}
}

// middlewareTracing is used to wrap http handler middleware.
type middlewareTracing struct {
	next     http.Handler
	name     string
	spanKind trace.SpanKind
	tracer   tracing.Tracer
}

func (w *middlewareTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctxTracing := tracing.Propagator(req.Context(), req.Header)
	ctxTracing, span := w.tracer.Start(ctxTracing, w.name, trace.WithSpanKind(w.spanKind))
	defer span.End()

	req = req.WithContext(ctxTracing)

	if w.next != nil {
		w.next.ServeHTTP(rw, req)
	}
}
