package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/tracing/opentelemetry"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Backend is an abstraction for tracking backend (OpenTelemetry, ...).
type Backend interface {
	Setup(serviceName string, sampleRate float64, globalAttributes map[string]string) (trace.Tracer, io.Closer, error)
}

// NewTracing Creates a Tracing.
func NewTracing(conf *static.Tracing) (trace.Tracer, io.Closer, error) {
	var backend Backend

	if conf.OTLP != nil {
		backend = conf.OTLP
	}

	if backend == nil {
		log.Debug().Msg("Could not initialize tracing, using OpenTelemetry by default")
		defaultBackend := &opentelemetry.Config{}
		backend = defaultBackend
	}

	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	return backend.Setup(conf.ServiceName, conf.SampleRate, conf.GlobalAttributes)
}

// TracerFromContext extracts the trace.Tracer from the given context.
func TracerFromContext(ctx context.Context) trace.Tracer {
	// Prevent picking trace.noopSpan tracer.
	if !trace.SpanContextFromContext(ctx).IsValid() {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if span != nil && span.TracerProvider() != nil {
		return span.TracerProvider().Tracer("github.com/traefik/traefik")
	}

	return nil
}

// ExtractCarrierIntoContext reads cross-cutting concerns from the carrier into a Context.
func ExtractCarrierIntoContext(ctx context.Context, headers http.Header) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// InjectContextIntoCarrier sets cross-cutting concerns from the request context into the request headers.
func InjectContextIntoCarrier(req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))
}
