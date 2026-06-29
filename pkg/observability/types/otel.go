package types

import (
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// OTelGRPC provides configuration settings for the gRPC open-telemetry.
type OTelGRPC struct {
	Endpoint string            `description:"Sets the gRPC endpoint (host:port) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Insecure bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	TLS      *types.ClientTLS  `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (o *OTelGRPC) SetDefaults() {
	o.Endpoint = "localhost:4317"
}

// OTelHTTP provides configuration settings for the HTTP open-telemetry.
type OTelHTTP struct {
	Endpoint string            `description:"Sets the HTTP endpoint (scheme://host:port/path) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	TLS      *types.ClientTLS  `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (o *OTelHTTP) SetDefaults() {
	o.Endpoint = "https://localhost:4318"
}

// ServiceResourceAttributes returns service resource attributes shared by all OpenTelemetry signals.
func ServiceResourceAttributes(serviceName, serviceNamespace string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(version.Version),
	}
	if serviceNamespace != "" {
		attrs = append(attrs, semconv.ServiceNamespace(serviceNamespace))
	}

	return attrs
}
