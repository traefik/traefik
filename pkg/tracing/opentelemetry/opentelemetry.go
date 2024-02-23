package opentelemetry

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

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
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

// Config provides configuration settings for the open-telemetry tracer.
type Config struct {
	GRPC *types.OtelGRPC `description:"gRPC configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" export:"true"`
	HTTP *types.OtelHTTP `description:"HTTP configuration for the OpenTelemetry collector." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.HTTP = &types.OtelHTTP{}
	c.HTTP.SetDefaults()
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string, sampleRate float64, globalAttributes map[string]string) (trace.Tracer, io.Closer, error) {
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

	for k, v := range globalAttributes {
		attr = append(attr, attribute.String(k, v))
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(attr...),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOSType(),
		resource.WithProcessCommandArgs(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building resource: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)

	log.Debug().Msg("OpenTelemetry tracer configured")

	return tracerProvider.Tracer("github.com/traefik/traefik"), &tpCloser{provider: tracerProvider}, err
}

func (c *Config) setupHTTPExporter() (*otlptrace.Exporter, error) {
	endpoint, err := url.Parse(c.HTTP.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", c.HTTP.Endpoint, err)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint.Host),
		otlptracehttp.WithHeaders(c.HTTP.Headers),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	}

	if endpoint.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if endpoint.Path != "" {
		opts = append(opts, otlptracehttp.WithURLPath(endpoint.Path))
	}

	if c.HTTP.TLS != nil {
		tlsConfig, err := c.HTTP.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConfig))
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func (c *Config) setupGRPCExporter() (*otlptrace.Exporter, error) {
	host, port, err := net.SplitHostPort(c.GRPC.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", c.GRPC.Endpoint, err)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
		otlptracegrpc.WithHeaders(c.GRPC.Headers),
		otlptracegrpc.WithCompressor(gzip.Name),
	}

	if c.GRPC.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if c.GRPC.TLS != nil {
		tlsConfig, err := c.GRPC.TLS.CreateTLSConfig(context.Background())
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

func (t *tpCloser) Close() error {
	if t == nil {
		return nil
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	return t.provider.Shutdown(ctx)
}
