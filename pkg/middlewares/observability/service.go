package observability

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceTypeName = "TracingService"
)

type serviceNameKey struct{}

// WithServiceName stores the logical service name in the context.
func WithServiceName(ctx context.Context, service string) context.Context {
	return context.WithValue(ctx, serviceNameKey{}, service)
}

// ServiceNameFromContext retrieves the logical service name from the context.
func ServiceNameFromContext(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(serviceNameKey{}).(string)
	return name, ok && name != ""
}

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
	// Store service name in context for use by the RoundTripper client span.
	req = req.WithContext(WithServiceName(req.Context(), t.service))

	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil && DetailedTracingEnabled(req.Context()) {
		tracingCtx, span := tracer.Start(req.Context(), "Service", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		req = req.WithContext(tracingCtx)

		span.SetAttributes(attribute.String("traefik.service.name", t.service))
	}

	t.next.ServeHTTP(rw, req)
}
