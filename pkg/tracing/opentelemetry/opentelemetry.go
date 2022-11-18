package opentelemetry

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/traefik/traefik/v2/pkg/version"
	"go.opentelemetry.io/otel"
	oteltracer "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

// Config provides configuration settings for the open-telemetry tracer.
type Config struct {
	// NOTE: as no gRPC option is implemented yet, the type is empty and is used as a boolean for upward compatibility purposes.
	GRPC *struct{} `description:"gRPC specific configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" export:"true" label:"allowEmpty" file:"allowEmpty"`

	Address  string            `description:"Sets the address of the collector endpoint." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Path     string            `description:"Sets the default URL path for sending traces." json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	Insecure bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	Headers  map[string]string `description:"Defines additional headers to be sent with the payloads." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	TLS      *types.ClientTLS  `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
}

// Setup sets up the tracer.
func (c *Config) Setup(componentName string) (opentracing.Tracer, io.Closer, error) {
	if c.Address == "" {
		return nil, nil, errors.New("address property is missing")
	}

	bt := oteltracer.NewBridgeTracer()
	bt.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	bt.SetOpenTelemetryTracer(otel.Tracer(componentName, trace.WithInstrumentationVersion(version.Version)))
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
		return nil, nil, fmt.Errorf("setting up exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)

	log.WithoutContext().Debug("OpenTelemetry tracer configured")

	return bt, tpCloser{provider: tracerProvider}, err
}

func (c *Config) setupHTTPExporter() (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(c.Address),
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
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(c.Address),
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
