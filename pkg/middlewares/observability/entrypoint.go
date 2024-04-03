package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	entryPointTypeName = "TracingEntryPoint"
)

type entryPointTracing struct {
	tracer *tracing.Tracer

	entryPoint            string
	next                  http.Handler
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
}

// WrapEntryPointHandler Wraps tracing to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, tracer *tracing.Tracer, semConvMetricRegistry *metrics.SemConvMetricsRegistry, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		if tracer == nil {
			tracer = tracing.NewTracer(noop.Tracer{}, nil, nil)
		}

		return newEntryPoint(ctx, tracer, semConvMetricRegistry, entryPointName, next), nil
	}
}

// newEntryPoint creates a new tracing middleware for incoming requests.
func newEntryPoint(ctx context.Context, tracer *tracing.Tracer, semConvMetricRegistry *metrics.SemConvMetricsRegistry, entryPointName string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", entryPointTypeName).Debug().Msg("Creating middleware")

	if tracer == nil {
		tracer = tracing.NewTracer(noop.Tracer{}, nil, nil)
	}

	return &entryPointTracing{
		entryPoint:            entryPointName,
		tracer:                tracer,
		semConvMetricRegistry: semConvMetricRegistry,
		next:                  next,
	}
}

func (e *entryPointTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	tracingCtx := tracing.ExtractCarrierIntoContext(req.Context(), req.Header)
	start := time.Now()
	tracingCtx, span := e.tracer.Start(tracingCtx, "EntryPoint", trace.WithSpanKind(trace.SpanKindServer), trace.WithTimestamp(start))

	req = req.WithContext(tracingCtx)

	span.SetAttributes(attribute.String("entry_point", e.entryPoint))

	e.tracer.CaptureServerRequest(span, req)

	recorder := newStatusCodeRecorder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	e.tracer.CaptureResponse(span, recorder.Header(), recorder.Status(), trace.SpanKindServer)

	end := time.Now()
	span.End(trace.WithTimestamp(end))

	if e.semConvMetricRegistry != nil && e.semConvMetricRegistry.HTTPServerRequestDuration() != nil {
		var attrs []attribute.KeyValue

		if recorder.Status() < 100 || recorder.Status() >= 600 {
			attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code ; %d", recorder.Status())))
		} else if recorder.Status() >= 400 {
			attrs = append(attrs, attribute.Key("error.type").String(strconv.Itoa(recorder.Status())))
		}

		attrs = append(attrs, semconv.HTTPRequestMethodKey.String(req.Method))
		attrs = append(attrs, semconv.HTTPResponseStatusCode(recorder.Status()))
		attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
		attrs = append(attrs, semconv.NetworkProtocolVersion(Proto(req.Proto)))
		attrs = append(attrs, semconv.ServerAddress(req.Host))
		attrs = append(attrs, semconv.URLScheme(req.Header.Get("X-Forwarded-Proto")))

		e.semConvMetricRegistry.HTTPServerRequestDuration().Record(req.Context(), end.Sub(start).Seconds(), metric.WithAttributes(attrs...))
	}
}
