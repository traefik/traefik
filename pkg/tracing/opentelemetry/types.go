package opentelemetry

import "github.com/traefik/traefik/v2/pkg/types"

// Config provides configuration settings for the open-telemetry tracer.
type Config struct {
	GRPC *struct{} `description:"GRPC specific configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" export:"true" label:"allowEmpty" file:"allowEmpty"`

	Endpoint string            `description:"Sets the address of the collector endpoint." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Path     string            `description:"Sets the default URL path for sending traces." json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
	Insecure bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	Compress bool              `description:"Enables compression of the sent data." json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" export:"true"`
	Headers  map[string]string `description:"Defines additional headers to be sent with the payloads." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	TLS      *types.ClientTLS  `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	// FIXME keep??
	c.Endpoint = "https://localhost:4318/v1/traces"

}
