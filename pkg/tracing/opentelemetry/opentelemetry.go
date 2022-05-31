package opentelemetry

import (
	"context"
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
	traefikversion "github.com/traefik/traefik/v2/pkg/version"
	"go.opentelemetry.io/otel"
	oteltracer "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

// Setup sets up the tracer.
func (c *Config) Setup(componentName string) (opentracing.Tracer, io.Closer, error) {
	// Tracer
	bt := oteltracer.NewBridgeTracer()

	// TODO add schema URL
	bt.SetOpenTelemetryTracer(otel.Tracer(componentName, trace.WithInstrumentationVersion(traefikversion.Version)))
	opentracing.SetGlobalTracer(bt)

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
		return nil, nil, fmt.Errorf("setup exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tracerProvider)

	log.WithoutContext().Debug("OpenTelemetry tracer configured")

	return bt, tpCloser{provider: tracerProvider}, err
}

func (c *Config) setupHTTPExporter() (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithHeaders(c.Headers),
	}

	if c.Compress {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	}

	if c.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(c.Endpoint))
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
			return nil, fmt.Errorf("create TLS client config: %w", err)
		}

		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConfig))
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func (c *Config) setupGRPCExporter() (*otlptrace.Exporter, error) {
	// TODO: handle DialOption
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithHeaders(c.Headers),
	}

	if c.Compress {
		opts = append(opts, otlptracegrpc.WithCompressor(gzip.Name))
	}

	if c.Endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(c.Endpoint))
	}

	if c.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if c.TLS != nil {
		tlsConfig, err := c.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("create TLS client config: %w", err)
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
