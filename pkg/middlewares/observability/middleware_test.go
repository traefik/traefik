package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestMiddlewareTracing_injectsTraceparent(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(prevPropagator) })

	tracer := tracing.NewTracer(&mockTracer{}, nil, nil, nil)
	ctx, span := tracer.Start(t.Context(), "entry", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	ctx = WithObservability(ctx, Observability{TracingEnabled: true, DetailedTracingEnabled: true})

	req := httptest.NewRequest(http.MethodGet, "http://www.test.com/", nil).WithContext(ctx)
	rw := httptest.NewRecorder()

	var gotTraceparent string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
	})

	NewMiddleware(next, "my-middleware", "MyType").ServeHTTP(rw, req)

	// mockSpan returns TraceID{1}/SpanID{1}, which the W3C TraceContext
	// propagator serializes as: 00-<traceid hex>-<spanid hex>-<flags>.
	assert.Equal(t, "00-01000000000000000000000000000000-0100000000000000-00", gotTraceparent)
}

func TestMiddlewareTracing_skipsWhenDetailedTracingDisabled(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(prevPropagator) })

	tracer := tracing.NewTracer(&mockTracer{}, nil, nil, nil)
	ctx, span := tracer.Start(t.Context(), "entry", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	ctx = WithObservability(ctx, Observability{TracingEnabled: true, DetailedTracingEnabled: false})

	req := httptest.NewRequest(http.MethodGet, "http://www.test.com/", nil).WithContext(ctx)
	rw := httptest.NewRecorder()

	var gotTraceparent string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
	})

	NewMiddleware(next, "my-middleware", "MyType").ServeHTTP(rw, req)

	assert.Empty(t, gotTraceparent)
}
