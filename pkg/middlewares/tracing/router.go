package tracing

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	routerTypeName = "TracingRouter"
)

type routerTracing struct {
	router  string
	service string
	tracer  tracing.Tracer
	next    http.Handler
}

// newRouter creates a new tracing middleware that traces the internal requests.
func newRouter(ctx context.Context, tracer tracing.Tracer, router, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", routerTypeName).
		Debug().Str(logs.RouterName, router).Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &routerTracing{
		router:  router,
		service: service,
		tracer:  tracer,
		next:    next,
	}
}

func (f *routerTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tracingCtx := tracing.Propagator(req.Context(), req.Header)
	tracingCtx, span := f.tracer.Start(tracingCtx, "router", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	req = req.WithContext(tracingCtx)

	span.SetAttributes(attribute.String("traefik.service.name", f.service))
	span.SetAttributes(attribute.String("traefik.router.name", f.router))

	tracing.LogRequest(span, req, trace.SpanKindInternal)

	tracing.InjectRequestHeaders(req.Context(), req.Header)

	recorder := newStatusCodeRecorder(rw, 200)

	f.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status())
}

// WrapRouterHandler Wraps tracing to alice.Constructor.
func WrapRouterHandler(ctx context.Context, tracer tracing.Tracer, router, service string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newRouter(ctx, tracer, router, service, next), nil
	}
}
