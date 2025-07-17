package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	semConvServerMetricsTypeName = "SemConvServerMetrics"
)

type semConvServerMetrics struct {
	next                  http.Handler
	semConvMetricRegistry *metrics.SemConvMetricsRegistry
}

// SemConvServerMetricsHandler return the alice.Constructor for semantic conventions servers metrics.
func SemConvServerMetricsHandler(ctx context.Context, semConvMetricRegistry *metrics.SemConvMetricsRegistry) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newServerMetricsSemConv(ctx, semConvMetricRegistry, next), nil
	}
}

// newServerMetricsSemConv creates a new semConv server metrics middleware for incoming requests.
func newServerMetricsSemConv(ctx context.Context, semConvMetricRegistry *metrics.SemConvMetricsRegistry, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", semConvServerMetricsTypeName).Debug().Msg("Creating middleware")

	return &semConvServerMetrics{
		semConvMetricRegistry: semConvMetricRegistry,
		next:                  next,
	}
}

func (e *semConvServerMetrics) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if e.semConvMetricRegistry == nil || e.semConvMetricRegistry.HTTPServerRequestDuration() == nil || !SemConvMetricsEnabled(req.Context()) {
		e.next.ServeHTTP(rw, req)
		return
	}

	start := time.Now()
	e.next.ServeHTTP(rw, req)
	end := time.Now()

	ctx := req.Context()
	capt, err := capture.FromContext(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str(logs.MiddlewareType, semConvServerMetricsTypeName).Msg("Could not get Capture")
		return
	}

	var attrs []attribute.KeyValue

	if capt.StatusCode() < 100 || capt.StatusCode() >= 600 {
		attrs = append(attrs, attribute.Key("error.type").String(fmt.Sprintf("Invalid HTTP status code ; %d", capt.StatusCode())))
	} else if capt.StatusCode() >= 400 {
		attrs = append(attrs, attribute.Key("error.type").String(strconv.Itoa(capt.StatusCode())))
	}

	attrs = append(attrs, semconv.HTTPRequestMethodKey.String(req.Method))
	attrs = append(attrs, semconv.HTTPResponseStatusCode(capt.StatusCode()))
	attrs = append(attrs, semconv.NetworkProtocolName(strings.ToLower(req.Proto)))
	attrs = append(attrs, semconv.NetworkProtocolVersion(Proto(req.Proto)))
	attrs = append(attrs, semconv.ServerAddress(req.Host))
	attrs = append(attrs, semconv.URLScheme(req.Header.Get("X-Forwarded-Proto")))

	e.semConvMetricRegistry.HTTPServerRequestDuration().Record(req.Context(), end.Sub(start).Seconds(), metric.WithAttributes(attrs...))
}
