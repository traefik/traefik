package tracing

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	forwarderTypeName = "TracingForwarder"
)

type forwarderMiddleware struct {
	router  string
	service string
	next    http.Handler
}

// NewForwarder creates a new forwarder middleware that traces the outgoing request.
func NewForwarder(ctx context.Context, router, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", forwarderTypeName).
		Debug().Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &forwarderMiddleware{
		router:  router,
		service: service,
		next:    next,
	}
}

func (f *forwarderMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tr, err := tracing.FromContext(req.Context())
	if err != nil {
		f.next.ServeHTTP(rw, req)
		return
	}

	opParts := []string{f.service, f.router}

	span, req, finish := tr.StartSpanf(req, trace.SpanKindClient, "forward", opParts, "/", trace.WithSpanKind(trace.SpanKindClient))
	defer finish()

	span.SetAttributes(attribute.String("traefik.service.name", f.service))
	span.SetAttributes(attribute.String("traefik.router.name", f.router))
	span.SetAttributes(semconv.HTTPMethod(req.Method))
	span.SetAttributes(attribute.String("http.host", req.Host))
	span.SetAttributes(semconv.HTTPURL(req.URL.String()))

	tracing.InjectRequestHeaders(req)

	recorder := newStatusCodeRecorder(rw, 200)

	f.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status())
}
