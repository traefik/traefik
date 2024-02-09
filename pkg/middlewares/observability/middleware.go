package observability

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Traceable embeds tracing information.
type Traceable interface {
	GetTracingInformation() (name string, typeName string, spanKind trace.SpanKind)
}

// WrapMiddleware adds traceability to an alice.Constructor.
func WrapMiddleware(ctx context.Context, constructor alice.Constructor) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if constructor == nil {
			return nil, nil
		}
		handler, err := constructor(next)
		if err != nil {
			return nil, err
		}

		if traceableHandler, ok := handler.(Traceable); ok {
			name, typeName, spanKind := traceableHandler.GetTracingInformation()
			log.Ctx(ctx).Debug().Str(logs.MiddlewareName, name).Msg("Adding tracing to middleware")
			return NewMiddleware(handler, name, typeName, spanKind), nil
		}
		return handler, nil
	}
}

// NewMiddleware returns a http.Handler struct.
func NewMiddleware(next http.Handler, name string, typeName string, spanKind trace.SpanKind) http.Handler {
	return &middlewareTracing{
		next:     next,
		name:     name,
		typeName: typeName,
		spanKind: spanKind,
	}
}

// middlewareTracing is used to wrap http handler middleware.
type middlewareTracing struct {
	next     http.Handler
	name     string
	typeName string
	spanKind trace.SpanKind
}

func (w *middlewareTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil {
		tracingCtx, span := tracer.Start(req.Context(), w.typeName, trace.WithSpanKind(w.spanKind))
		defer span.End()

		req = req.WithContext(tracingCtx)

		span.SetAttributes(attribute.String("traefik.middleware.name", w.name))
	}

	if w.next != nil {
		w.next.ServeHTTP(rw, req)
	}
}
