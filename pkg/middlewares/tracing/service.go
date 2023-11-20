package tracing

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
	tracer  tracing.Tracer
	next    http.Handler
}

// NewService creates a new tracing middleware that traces the outgoing requests.
func NewService(ctx context.Context, tracer tracing.Tracer, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", serviceTypeName).
		Debug().Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &serviceTracing{
		service: service,
		tracer:  tracer,
		next:    next,
	}
}

func (f *serviceTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tracingCtx := tracing.Propagator(req.Context(), req.Header)
	tracingCtx, span := f.tracer.Start(tracingCtx, "router", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	req = req.WithContext(tracingCtx)

	span.SetAttributes(attribute.String("traefik.service.name", f.service))

	tracing.LogRequest(span, req, trace.SpanKindClient)

	tracing.InjectRequestHeaders(req.Context(), req.Header)

	recorder := newStatusCodeRecorder(rw, 200)

	f.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status())
}
