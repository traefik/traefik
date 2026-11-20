package types

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	ttypes "github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

type TracingVerbosity string

const (
	MinimalVerbosity  TracingVerbosity = "minimal"
	DetailedVerbosity TracingVerbosity = "detailed"
)

func (v TracingVerbosity) Allows(verbosity TracingVerbosity) bool {
	switch v {
	case DetailedVerbosity:
		return verbosity == DetailedVerbosity || verbosity == MinimalVerbosity
	default:
		return verbosity == MinimalVerbosity
	}
}

// envSamplers mirrors the OTEL_TRACES_SAMPLER values that the OpenTelemetry Go SDK supports in sdk/trace/sampler_env.go, which it does not export.
var envSamplers = []string{
	"always_on",
	"always_off",
	"traceidratio",
	"parentbased_always_on",
	"parentbased_always_off",
	"parentbased_traceidratio",
}

// OTelTracing provides configuration settings for the open-telemetry tracer.
type OTelTracing struct {
	GRPC *OTelGRPC `description:"gRPC configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	HTTP *OTelHTTP `description:"HTTP configuration for the OpenTelemetry collector." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values.
func (c *OTelTracing) SetDefaults() {
	c.HTTP = &OTelHTTP{}
	c.HTTP.SetDefaults()
}

// Setup sets up the tracer.
func (c *OTelTracing) Setup(ctx context.Context, serviceName string, sampleRate float64, resourceAttributes map[string]string) (trace.Tracer, io.Closer, error) {
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

	var resAttrs []attribute.KeyValue
	for k, v := range resourceAttributes {
		resAttrs = append(resAttrs, attribute.String(k, v))
	}

	res, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(ttypes.K8sAttributesDetector{}),
		// The following order allows the user to override the service name and version,
		// as well as any other attributes set by the above detectors.
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version.Version),
		),
		resource.WithAttributes(resAttrs...),
		// Use the environment variables to allow overriding above resource attributes.
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building resource: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	}

	// The SDK applies the OTEL_TRACES_SAMPLER sampler before the options, so an explicit sampler would override it.
	useEnvSampler, envErr := checkEnvSampler()
	if envErr != nil {
		log.Warn().Err(envErr).Msg("Using sampleRate instead of the OTEL_TRACES_SAMPLER env variables")
	}

	if !useEnvSampler {
		opts = append(opts, sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))))
	}

	tracerProvider := sdktrace.NewTracerProvider(opts...)

	otel.SetTracerProvider(tracerProvider)

	log.Debug().Msg("OpenTelemetry tracer configured")

	return tracerProvider.Tracer("github.com/traefik/traefik"), &tpCloser{provider: tracerProvider}, err
}

// checkEnvSampler reports whether the OTEL_TRACES_SAMPLER env variables define a valid sampler.
// The SDK samples every root span when it rejects a value, so checkEnvSampler returns an error and sampleRate applies instead.
func checkEnvSampler() (bool, error) {
	value := os.Getenv("OTEL_TRACES_SAMPLER")
	sampler := strings.ToLower(strings.TrimSpace(value))
	if sampler == "" {
		return false, nil
	}

	if !slices.Contains(envSamplers, sampler) {
		return false, fmt.Errorf("unsupported OTEL_TRACES_SAMPLER value %q", value)
	}

	arg := strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER_ARG"))
	if arg == "" || !strings.HasSuffix(sampler, "traceidratio") {
		return true, nil
	}

	ratio, err := strconv.ParseFloat(arg, 64)
	// The negated comparison also rejects NaN, which the SDK accepts without a usable ratio.
	if err != nil || !(ratio >= 0 && ratio <= 1) {
		return false, fmt.Errorf("invalid OTEL_TRACES_SAMPLER_ARG value %q, it must be a ratio between 0 and 1", arg)
	}

	return true, nil
}

func (c *OTelTracing) setupHTTPExporter() (*otlptrace.Exporter, error) {
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

func (c *OTelTracing) setupGRPCExporter() (*otlptrace.Exporter, error) {
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
