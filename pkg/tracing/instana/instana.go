package instana

import (
	"context"
	"fmt"
	"io"

	instana "github.com/instana/go-otel-exporter"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Name sets the name of this tracer.
const Name = "instana"

// Config provides configuration settings for an instana tracer.
type Config struct {
	GlobalTags map[string]string `description:"Sets a list of key:value tags on all spans." json:"globalTags,omitempty" toml:"globalTags,omitempty" yaml:"globalTags,omitempty" export:"true"`
	SampleRate float64           `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.SampleRate = 1.0
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (trace.Tracer, io.Closer, error) {
	// this takes env variables INSTANA_ENDPOINT_URL and INSTANA_AGENT_KEY
	exporter := instana.New()

	attr := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(version.Version),
	}

	for k, v := range c.GlobalTags {
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
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Tracer("github.com/traefik/traefik", trace.WithInstrumentationVersion(version.Version)), tpCloser{provider: tracerProvider}, nil
}

// tpCloser converts a TraceProvider into an io.Closer.
type tpCloser struct {
	provider *sdktrace.TracerProvider
}

func (t tpCloser) Close() error {
	return t.provider.Shutdown(context.Background())
}
