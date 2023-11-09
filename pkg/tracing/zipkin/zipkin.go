package zipkin

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Name sets the name of this tracer.
const Name = "zipkin"

// Config provides configuration settings for a zipkin tracer.
type Config struct {
	HTTPEndpoint string            `description:"Sets the HTTP Endpoint to report traces to." json:"httpEndpoint,omitempty" toml:"httpEndpoint,omitempty" yaml:"httpEndpoint,omitempty"`
	SampleRate   float64           `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
	Attributes   map[string]string `description:"Defines additional attributes to be sent with the payloads." json:"attributes,omitempty" toml:"attributes,omitempty" yaml:"attributes,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.HTTPEndpoint = "http://localhost:9411/api/v2/spans"
	c.SampleRate = 1.0
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (trace.Tracer, io.Closer, error) {
	exporter, err := zipkin.New(c.HTTPEndpoint)
	if err != nil {
		return nil, nil, err
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)

	attr := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(version.Version),
	}

	for k, v := range c.Attributes {
		attr = append(attr, attribute.String(k, v))
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(attr...),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.SampleRate)),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSpanProcessor(batcher),
	)
	otel.SetTracerProvider(tracerProvider)

	log.Debug().Msgf("Zipkin tracer configured")

	return tracerProvider.Tracer(serviceName, trace.WithInstrumentationVersion(version.Version)), tpCloser{provider: tracerProvider}, nil
}

// tpCloser converts a TraceProvider into an io.Closer.
type tpCloser struct {
	provider *sdktrace.TracerProvider
}

func (t tpCloser) Close() error {
	return t.provider.Shutdown(context.Background())
}
