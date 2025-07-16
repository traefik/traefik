package observability

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type contextKey int

const (
	accessLogsKey contextKey = iota
	metricsKey
	semConvMetricsKey
	minimalTracingKey
	detailedTracingKey
)

// AccessLogsEnabled returns whether metrics are enabled.
func AccessLogsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(accessLogsKey).(bool); ok {
		return enabled
	}

	return false
}

// WithAccessLogEnabledHandler sets the access log enabled state in the context for the next handler.
// This is used for testing purposes to control whether access logs are enabled or not.
func WithAccessLogEnabledHandler(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), accessLogsKey, enabled)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// MetricsEnabled returns whether metrics are enabled.
func MetricsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(metricsKey).(bool); ok {
		return enabled
	}

	return false
}

// WithMetricsEnabledHandler sets the metrics enabled state in the context for the next handler.
// This is used for testing purposes to control whether metrics are enabled or not.
func WithMetricsEnabledHandler(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), metricsKey, enabled)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// SemConvMetricsEnabled returns whether metrics are enabled.
func SemConvMetricsEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(semConvMetricsKey).(bool); ok {
		return enabled
	}

	return false
}

// WithSemConvMetricsEnabled sets the semantic conventions metrics enabled state in the context.
// This is used for testing purposes to control whether semantic conventions metrics are enabled or not.
func WithSemConvMetricsEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, semConvMetricsKey, enabled)
}

// WithSemConvMetricsEnabledHandler sets the semantic conventions metrics enabled state in the context for the next handler.
// This is used for testing purposes to control whether semantic conventions metrics are enabled or not.
func WithSemConvMetricsEnabledHandler(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), semConvMetricsKey, enabled)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// MinimalTraceEnabled returns whether minimal tracing is enabled.
func MinimalTraceEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(minimalTracingKey).(bool); ok {
		return enabled
	}

	return false
}

// WithMinimalTraceEnabledHandler sets the minimal tracing enabled state in the context for the next handler.
// This is used for testing purposes to control whether minimal tracing is enabled or not.
func WithMinimalTraceEnabledHandler(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), minimalTracingKey, enabled)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// DetailedTraceEnabled returns whether detailed tracing is enabled.
func DetailedTraceEnabled(ctx context.Context) bool {
	if enabled, ok := ctx.Value(detailedTracingKey).(bool); ok {
		return enabled
	}

	return false
}

// WithDetailedTraceEnabledHandler sets the detailed tracing enabled state in the context for the next handler.
// This is used for testing purposes to control whether detailed tracing is enabled or not.
func WithDetailedTraceEnabledHandler(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), detailedTracingKey, enabled)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
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
