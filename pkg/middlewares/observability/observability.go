package observability

import (
	"context"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
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
	Metadata               *dynamic.ObservabilityMetadata
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

// GetObservabilityMetadata returns the observability metadata.
func GetObservabilityMetadata(ctx context.Context) *dynamic.ObservabilityMetadata {
	obs, ok := ctx.Value(observabilityKey).(Observability)
	if ok {
		return obs.Metadata
	}
	return nil
}

type serviceObservabilityKey struct{}

// ServiceObservabilityState carries the service-level observability metadata
// produced by the service handler during request processing. The pointer is
// seeded once upstream (at the entrypoint chain) and mutated by the leaf
// service handler when the load balancer dispatches; access-log middleware
// reads the final state at log emission. Mirrors how *LogData is shared.
type ServiceObservabilityState struct {
	Metadata *dynamic.ServiceObservabilityMetadata
}

// WithServiceObservabilityState seeds an empty state container in the context
// so downstream handlers and the access-log middleware share the same pointer.
func WithServiceObservabilityState(ctx context.Context) context.Context {
	return context.WithValue(ctx, serviceObservabilityKey{}, &ServiceObservabilityState{})
}

// GetServiceObservabilityState returns the per-request state container, or nil
// if none has been seeded upstream.
func GetServiceObservabilityState(ctx context.Context) *ServiceObservabilityState {
	s, _ := ctx.Value(serviceObservabilityKey{}).(*ServiceObservabilityState)
	return s
}

// NewServiceMetadataHandler wraps the service handler so that when the leaf is
// reached at request time, its metadata is published into the per-request state
// container. A no-op when the service has no observability metadata attached.
func NewServiceMetadataHandler(cfg *dynamic.ServiceObservabilityConfig, next http.Handler) http.Handler {
	if cfg == nil || cfg.Metadata == nil {
		return next
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if state := GetServiceObservabilityState(req.Context()); state != nil {
			state.Metadata = cfg.Metadata
		}
		next.ServeHTTP(rw, req)
	})
}

// SetStatusErrorf flags the span as in error and log an event.
func SetStatusErrorf(ctx context.Context, format string, args ...any) {
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
