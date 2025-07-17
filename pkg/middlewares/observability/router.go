package observability

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	routerTypeName = "TracingRouter"
)

type routerTracing struct {
	router     string
	routerRule string
	service    string
	next       http.Handler
}

// WrapRouterHandler Wraps tracing to alice.Constructor.
func WrapRouterHandler(ctx context.Context, router, routerRule, service string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newRouter(ctx, router, routerRule, service, next), nil
	}
}

// newRouter creates a new tracing middleware that traces the internal requests.
func newRouter(ctx context.Context, router, routerRule, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", routerTypeName).
		Debug().Str(logs.RouterName, router).Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &routerTracing{
		router:     router,
		routerRule: routerRule,
		service:    service,
		next:       next,
	}
}

func (f *routerTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil && DetailedTracingEnabled(req.Context()) {
		tracingCtx, span := tracer.Start(req.Context(), "Router", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		req = req.WithContext(tracingCtx)

		span.SetAttributes(attribute.String("traefik.service.name", f.service))
		span.SetAttributes(attribute.String("traefik.router.name", f.router))
		span.SetAttributes(semconv.HTTPRoute(f.routerRule))
	}

	f.next.ServeHTTP(rw, req)
}
