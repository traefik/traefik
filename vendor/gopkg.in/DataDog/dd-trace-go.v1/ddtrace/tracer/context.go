package tracer

import (
	"context"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/internal"
)

type contextKey struct{}

var activeSpanKey = contextKey{}

// ContextWithSpan returns a copy of the given context which includes the span s.
func ContextWithSpan(ctx context.Context, s Span) context.Context {
	return context.WithValue(ctx, activeSpanKey, s)
}

// SpanFromContext returns the span contained in the given context. A second return
// value indicates if a span was found in the context. If no span is found, a no-op
// span is returned.
func SpanFromContext(ctx context.Context) (Span, bool) {
	if ctx == nil {
		return &internal.NoopSpan{}, false
	}
	v := ctx.Value(activeSpanKey)
	if s, ok := v.(ddtrace.Span); ok {
		return s, true
	}
	return &internal.NoopSpan{}, false
}

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span. If the ChildOf
// option is passed, the span from context will take precedence over it as the parent span.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	if s, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ChildOf(s.Context()))
	}
	s := StartSpan(operationName, opts...)
	return s, ContextWithSpan(ctx, s)
}
