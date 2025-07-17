package observability

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type contextKey int

const observabilityKey contextKey = iota

type Observability struct {
	AccessLogsEnabled      bool
	MetricsEnabled         bool
	SemConvMetricsEnabled  bool
	TracingEnabled         bool
	DetailedTracingEnabled bool
}

// WithObservabilityHandler sets the observability state in the context for the next handler.
// This is also used for testing purposes to control whether access logs are enabled or not.
func WithObservabilityHandler(next http.Handler, obs Observability) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(rw, req.WithContext(WithObservability(req.Context(), obs)))
	})
}

// WithObservability injects the observability state into the context.
func WithObservability(ctx context.Context, obs Observability) context.Context {
	return context.WithValue(ctx, observabilityKey, obs)
}

// AccessLogsEnabled returns whether access-logs are enabled.
func AccessLogsEnabled(ctx context.Context) bool {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	return ok && obs.AccessLogsEnabled
}

// MetricsEnabled returns whether metrics are enabled.
func MetricsEnabled(ctx context.Context) bool {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	return ok && obs.MetricsEnabled
}

// SemConvMetricsEnabled returns whether metrics are enabled.
func SemConvMetricsEnabled(ctx context.Context) bool {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	return ok && obs.SemConvMetricsEnabled
}

// TracingEnabled returns whether tracing is enabled.
func TracingEnabled(ctx context.Context) bool {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	return ok && obs.TracingEnabled
}

// DetailedTracingEnabled returns whether detailed tracing is enabled.
func DetailedTracingEnabled(ctx context.Context) bool {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	return ok && obs.DetailedTracingEnabled
}

// SetStatusErrorf flags the span as in error and log an event.
func SetStatusErrorf(ctx context.Context, format string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(format, args...))
	}
}

func Proto(proto string) string {
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
