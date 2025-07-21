package types

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otelsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	// AccessLogKeep is the keep string value.
	AccessLogKeep = "keep"
	// AccessLogDrop is the drop string value.
	AccessLogDrop = "drop"
	// AccessLogRedact is the redact string value.
	AccessLogRedact = "redact"
)

const (
	// CommonFormat is the common logging format (CLF).
	CommonFormat string = "common"
)

const OTelTraefikServiceName = "traefik"

// TraefikLog holds the configuration settings for the traefik logger.
type TraefikLog struct {
	Level   string `description:"Log level set to traefik logs." json:"level,omitempty" toml:"level,omitempty" yaml:"level,omitempty" export:"true"`
	Format  string `description:"Traefik log format: json | common" json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty" export:"true"`
	NoColor bool   `description:"When using the 'common' format, disables the colorized output." json:"noColor,omitempty" toml:"noColor,omitempty" yaml:"noColor,omitempty" export:"true"`

	FilePath   string `description:"Traefik log file path. Stdout is used when omitted or empty." json:"filePath,omitempty" toml:"filePath,omitempty" yaml:"filePath,omitempty"`
	MaxSize    int    `description:"Maximum size in megabytes of the log file before it gets rotated." json:"maxSize,omitempty" toml:"maxSize,omitempty" yaml:"maxSize,omitempty" export:"true"`
	MaxAge     int    `description:"Maximum number of days to retain old log files based on the timestamp encoded in their filename." json:"maxAge,omitempty" toml:"maxAge,omitempty" yaml:"maxAge,omitempty" export:"true"`
	MaxBackups int    `description:"Maximum number of old log files to retain." json:"maxBackups,omitempty" toml:"maxBackups,omitempty" yaml:"maxBackups,omitempty" export:"true"`
	Compress   bool   `description:"Determines if the rotated log files should be compressed using gzip." json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" export:"true"`

	OTLP *OTelLog `description:"Settings for OpenTelemetry." json:"otlp,omitempty" toml:"otlp,omitempty" yaml:"otlp,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values.
func (l *TraefikLog) SetDefaults() {
	l.Format = CommonFormat
	l.Level = "ERROR"
}

// AccessLog holds the configuration settings for the access logger (middlewares/accesslog).
type AccessLog struct {
	FilePath      string            `description:"Access log file path. Stdout is used when omitted or empty." json:"filePath,omitempty" toml:"filePath,omitempty" yaml:"filePath,omitempty"`
	Format        string            `description:"Access log format: json | common" json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty" export:"true"`
	Filters       *AccessLogFilters `description:"Access log filters, used to keep only specific access logs." json:"filters,omitempty" toml:"filters,omitempty" yaml:"filters,omitempty" export:"true"`
	Fields        *AccessLogFields  `description:"AccessLogFields." json:"fields,omitempty" toml:"fields,omitempty" yaml:"fields,omitempty" export:"true"`
	BufferingSize int64             `description:"Number of access log lines to process in a buffered way." json:"bufferingSize,omitempty" toml:"bufferingSize,omitempty" yaml:"bufferingSize,omitempty" export:"true"`
	AddInternals  bool              `description:"Enables access log for internal services (ping, dashboard, etc...)." json:"addInternals,omitempty" toml:"addInternals,omitempty" yaml:"addInternals,omitempty" export:"true"`

	OTLP *OTelLog `description:"Settings for OpenTelemetry." json:"otlp,omitempty" toml:"otlp,omitempty" yaml:"otlp,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values.
func (l *AccessLog) SetDefaults() {
	l.Format = CommonFormat
	l.FilePath = ""
	l.Filters = &AccessLogFilters{}
	l.Fields = &AccessLogFields{}
	l.Fields.SetDefaults()
}

// AccessLogFilters holds filters configuration.
type AccessLogFilters struct {
	StatusCodes   []string       `description:"Keep access logs with status codes in the specified range." json:"statusCodes,omitempty" toml:"statusCodes,omitempty" yaml:"statusCodes,omitempty" export:"true"`
	RetryAttempts bool           `description:"Keep access logs when at least one retry happened." json:"retryAttempts,omitempty" toml:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty" export:"true"`
	MinDuration   types.Duration `description:"Keep access logs when request took longer than the specified duration." json:"minDuration,omitempty" toml:"minDuration,omitempty" yaml:"minDuration,omitempty" export:"true"`
}

// FieldHeaders holds configuration for access log headers.
type FieldHeaders struct {
	DefaultMode string            `description:"Default mode for fields: keep | drop | redact" json:"defaultMode,omitempty" toml:"defaultMode,omitempty" yaml:"defaultMode,omitempty" export:"true"`
	Names       map[string]string `description:"Override mode for headers" json:"names,omitempty" toml:"names,omitempty" yaml:"names,omitempty" export:"true"`
}

// AccessLogFields holds configuration for access log fields.
type AccessLogFields struct {
	DefaultMode string            `description:"Default mode for fields: keep | drop" json:"defaultMode,omitempty" toml:"defaultMode,omitempty" yaml:"defaultMode,omitempty"  export:"true"`
	Names       map[string]string `description:"Override mode for fields" json:"names,omitempty" toml:"names,omitempty" yaml:"names,omitempty" export:"true"`
	Headers     *FieldHeaders     `description:"Headers to keep, drop or redact" json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (f *AccessLogFields) SetDefaults() {
	f.DefaultMode = AccessLogKeep
	f.Headers = &FieldHeaders{
		DefaultMode: AccessLogDrop,
	}
}

// Keep check if the field need to be kept or dropped.
func (f *AccessLogFields) Keep(field string) bool {
	defaultKeep := true
	if f != nil {
		defaultKeep = checkFieldValue(f.DefaultMode, defaultKeep)

		if v, ok := f.Names[field]; ok {
			return checkFieldValue(v, defaultKeep)
		}
	}
	return defaultKeep
}

// KeepHeader checks if the headers need to be kept, dropped or redacted and returns the status.
func (f *AccessLogFields) KeepHeader(header string) string {
	defaultValue := AccessLogKeep
	if f != nil && f.Headers != nil {
		defaultValue = checkFieldHeaderValue(f.Headers.DefaultMode, defaultValue)

		if v, ok := f.Headers.Names[header]; ok {
			return checkFieldHeaderValue(v, defaultValue)
		}
	}
	return defaultValue
}

func checkFieldValue(value string, defaultKeep bool) bool {
	switch value {
	case AccessLogKeep:
		return true
	case AccessLogDrop:
		return false
	default:
		return defaultKeep
	}
}

func checkFieldHeaderValue(value, defaultValue string) string {
	if value == AccessLogKeep || value == AccessLogDrop || value == AccessLogRedact {
		return value
	}
	return defaultValue
}

// OTelLog provides configuration settings for the open-telemetry logger.
type OTelLog struct {
	ServiceName        string            `description:"Defines the service name resource attribute." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
	ResourceAttributes map[string]string `description:"Defines additional resource attributes (key:value)." json:"resourceAttributes,omitempty" toml:"resourceAttributes,omitempty" yaml:"resourceAttributes,omitempty"`
	GRPC               *OTelGRPC         `description:"gRPC configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	HTTP               *OTelHTTP         `description:"HTTP configuration for the OpenTelemetry collector." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values.
func (o *OTelLog) SetDefaults() {
	o.ServiceName = OTelTraefikServiceName
	o.HTTP = &OTelHTTP{}
	o.HTTP.SetDefaults()
}

// NewLoggerProvider creates a new OpenTelemetry logger provider.
func (o *OTelLog) NewLoggerProvider(ctx context.Context) (*otelsdk.LoggerProvider, error) {
	var (
		err      error
		exporter otelsdk.Exporter
	)
	if o.GRPC != nil {
		exporter, err = o.buildGRPCExporter()
	} else {
		exporter, err = o.buildHTTPExporter()
	}
	if err != nil {
		return nil, fmt.Errorf("setting up exporter: %w", err)
	}

	var resAttrs []attribute.KeyValue
	for k, v := range o.ResourceAttributes {
		resAttrs = append(resAttrs, attribute.String(k, v))
	}

	res, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(K8sAttributesDetector{}),
		// The following order allows the user to override the service name and version,
		// as well as any other attributes set by the above detectors.
		resource.WithAttributes(
			semconv.ServiceName(o.ServiceName),
			semconv.ServiceVersion(version.Version),
		),
		resource.WithAttributes(resAttrs...),
		// Use the environment variables to allow overriding above resource attributes.
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("building resource: %w", err)
	}

	// Register the trace provider to allow the global logger to access it.
	bp := otelsdk.NewBatchProcessor(exporter)
	loggerProvider := otelsdk.NewLoggerProvider(
		otelsdk.WithResource(res),
		otelsdk.WithProcessor(bp),
	)

	return loggerProvider, nil
}

func (o *OTelLog) buildHTTPExporter() (*otlploghttp.Exporter, error) {
	endpoint, err := url.Parse(o.HTTP.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", o.HTTP.Endpoint, err)
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint.Host),
		otlploghttp.WithHeaders(o.HTTP.Headers),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	}

	if endpoint.Scheme == "http" {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	if endpoint.Path != "" {
		opts = append(opts, otlploghttp.WithURLPath(endpoint.Path))
	}

	if o.HTTP.TLS != nil {
		tlsConfig, err := o.HTTP.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlploghttp.WithTLSClientConfig(tlsConfig))
	}

	return otlploghttp.New(context.Background(), opts...)
}

func (o *OTelLog) buildGRPCExporter() (*otlploggrpc.Exporter, error) {
	host, port, err := net.SplitHostPort(o.GRPC.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid collector endpoint %q: %w", o.GRPC.Endpoint, err)
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(fmt.Sprintf("%s:%s", host, port)),
		otlploggrpc.WithHeaders(o.GRPC.Headers),
		otlploggrpc.WithCompressor(gzip.Name),
	}

	if o.GRPC.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	if o.GRPC.TLS != nil {
		tlsConfig, err := o.GRPC.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("creating TLS client config: %w", err)
		}

		opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	}

	return otlploggrpc.New(context.Background(), opts...)
}
