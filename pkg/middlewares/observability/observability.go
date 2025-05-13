package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type contextKey int

// DisableMetricsKey is a context key used to disable the metrics.
const DisableMetricsKey contextKey = iota

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
