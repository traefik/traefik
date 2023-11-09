package opentelemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

// Config provides configuration settings for the open-telemetry tracer.
type Config struct {
	// NOTE: as no gRPC option is implemented yet, the type is empty and is used as a boolean for upward compatibility purposes.
	GRPC *struct{} `description:"gRPC specific configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	Address    string            `description:"Sets the address (host:port) of the collector endpoint." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Path       string            `description:"Sets the URL path of the collector endpoint." json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	Insecure   bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	Headers    map[string]string `description:"Defines additional connection headers to be sent with the payloads." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	GlobalTags map[string]string `description:"Sets a list of key:value tags on all spans." json:"globalTags,omitempty" toml:"globalTags,omitempty" yaml:"globalTags,omitempty" export:"true"`
	SampleRate float64           `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
	TLS        *types.ClientTLS  `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.Address = "localhost:4318"
	c.SampleRate = 1.0
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (trace.Tracer, io.Closer, error) {
	var (
		err      error
		exporter *otlptrace.Exporter
	)
	if c.GRPC != nil {
		exporter, err = c.setupGRPCExporter()
	} else {
		exporter, err = c.setupHTTPExporter()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("setting up exporter: %w", err)
	}

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
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)

	log.Debug().Msg("OpenTelemetry tracer configured")

	return tracerProvider.Tracer("github.com/traefik/traefik", trace.WithInstrumentationVersion(version.Version)), tpCloser{provider: tracerProvider}, err
}

func (c *Config) setupHTTPExporter() (*otlptrace.Exporter, error) {
	host, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid collector address %q: %w", c.Address, err)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
		otlptracehttp.WithHeaders(c.Headers),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	}

	if c.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if c.Path != "" {
		opts = append(opts, otlptracehttp.WithURLPath(c.Path))
	}

	if c.TLS != nil {
		tlsConfig, err := c.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConfig))
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func (c *Config) setupGRPCExporter() (*otlptrace.Exporter, error) {
	host, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid collector address %q: %w", c.Address, err)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
		otlptracegrpc.WithHeaders(c.Headers),
		otlptracegrpc.WithCompressor(gzip.Name),
	}

	if c.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if c.TLS != nil {
		tlsConfig, err := c.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}

// tpCloser converts a TraceProvider into an io.Closer.
type tpCloser struct {
	provider *sdktrace.TracerProvider
}

func (t tpCloser) Close() error {
	return t.provider.Shutdown(context.Background())
}
