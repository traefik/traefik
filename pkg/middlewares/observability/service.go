package observability

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceTypeName = "TracingService"
)

type serviceTracing struct {
	service string
	next    http.Handler
}

// NewService creates a new tracing middleware that traces the outgoing requests.
func NewService(ctx context.Context, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", serviceTypeName).
		Debug().Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &serviceTracing{
		service: service,
		next:    next,
	}
}

func (t *serviceTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil && DetailedTracingEnabled(req.Context()) {
		tracingCtx, span := tracer.Start(req.Context(), "Service", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		req = req.WithContext(tracingCtx)

		span.SetAttributes(attribute.String("traefik.service.name", t.service))
	}

	t.next.ServeHTTP(rw, req)
}
