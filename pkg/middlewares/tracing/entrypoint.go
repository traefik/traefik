package tracing

import (
	"context"
	"fmt"
	"net/http"
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
)

const (
	entryPointTypeName = "TracingEntryPoint"
)

type entryPointTracing struct {
	tracer trace.Tracer

	entryPoint            string
	next                  http.Handler
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
}

// WrapEntryPointHandler Wraps tracing to alice.Constructor.
func WrapEntryPointHandler(ctx context.Context, tracer trace.Tracer, semConvMetricRegistry *metrics.SemConvMetricsRegistry, entryPointName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newEntryPoint(ctx, tracer, semConvMetricRegistry, entryPointName, next), nil
	}
}

// newEntryPoint creates a new tracing middleware for incoming requests.
func newEntryPoint(ctx context.Context, tracer trace.Tracer, semConvMetricRegistry *metrics.SemConvMetricsRegistry, entryPointName string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", entryPointTypeName).Debug().Msg("Creating middleware")

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

	tracing.LogServerRequest(span, req)

	recorder := newStatusCodeRecorder(rw, http.StatusOK)
	e.next.ServeHTTP(recorder, req)

	tracing.LogResponseCode(span, recorder.Status(), trace.SpanKindServer)

	end := time.Now()
	span.End(trace.WithTimestamp(end))

	if e.semConvMetricRegistry != nil && e.semConvMetricRegistry.HttpServerRequestDuration() != nil {
		var attrs []attribute.KeyValue

		if recorder.Status() < 100 || recorder.Status() >= 600 {
			attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code ; %d", recorder.Status())))
		} else if recorder.Status() >= 400 {
			attrs = append(attrs, attribute.Key("error.type").Int(recorder.Status()))
		}

		attrs = append(attrs, semconv.HTTPRequestMethodKey.String(req.Method))
		attrs = append(attrs, semconv.HTTPResponseStatusCode(recorder.Status()))
		attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
		attrs = append(attrs, semconv.NetworkProtocolVersion(proto(req.Proto)))
		attrs = append(attrs, semconv.ServerAddress(req.Host))
		attrs = append(attrs, semconv.URLScheme(req.Header.Get("X-Forwarded-Proto")))

		e.semConvMetricRegistry.HttpServerRequestDuration().Record(req.Context(), end.Sub(start).Seconds(), metric.WithAttributes(attrs...))
	}
}

func proto(proto string) string {
	switch proto {
	case "HTTP/1.0":
		return "1.0"
	case "HTTP/1.1":
		return "1.1"
	case "HTTP/2":
		return "2"
	case "HTTP/3":
		return "3"
	default:
		return proto
	}
}
