package opentelemetry

import (
	"time"

	"github.com/traefik/traefik/v2/pkg/types"
)

// Config provides configuration settings for an open-telemetry tracer.
type Config struct {
	GRPC *ConfigGRPC `description:"GRPC specific configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" export:"true" label:"allowEmpty" file:"allowEmpty"`

	Compress bool              `description:"Enable compression on the sent data." json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" export:"true"`
	Endpoint string            `description:"Address of the collector endpoint." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	Retry    *retry            `description:"The retry policy for transient errors that may occurs when exporting traces." json:"retry,omitempty" toml:"retry,omitempty" yaml:"retry,omitempty" export:"true"`
	Timeout  time.Duration     `description:"The max waiting time for the backend to process each spans batch." json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	TLS      *types.ClientTLS  `description:"Enable TLS support" json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true" `
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.Endpoint = "https://localhost:4318/v1/traces"
	c.Retry = &retry{}
	c.Retry.SetDefaults()
	c.Timeout = 10 * time.Second
}

// ConfigGRPC provides configuration settings for an open-telemetry tracer.
type ConfigGRPC struct {
	Insecure           bool          `description:"Connect to endpoint using HTTP." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	ReconnectionPeriod time.Duration `description:"The minimum amount of time between connection attempts to the target endpoint." json:"reconnectionPeriod,omitempty" toml:"reconnectionPeriod,omitempty" yaml:"reconnectionPeriod,omitempty" export:"true"`
	ServiceConfig      string        `description:"Defines the default gRPC service config used." json:"serviceConfig,omitempty" toml:"serviceConfig,omitempty" yaml:"serviceConfig,omitempty" export:"true"`
}

type retry struct {
	InitialInterval time.Duration `description:"The time to wait after the first failure before retrying." json:"initialInterval,omitempty" toml:"initialInterval,omitempty" yaml:"initialInterval,omitempty" export:"true"`
	MaxElapsedTime  time.Duration `description:"The maximum amount of time (including retries) spent trying to send a request/batch." json:"maxElapsedTime,omitempty" toml:"maxElapsedTime,omitempty" yaml:"maxElapsedTime,omitempty" export:"true"`
	MaxInterval     time.Duration `description:"The upper bound on backoff interval." json:"maxInterval,omitempty" toml:"maxInterval,omitempty" yaml:"maxInterval,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (r *retry) SetDefaults() {
	r.InitialInterval = 5 * time.Second
	r.MaxElapsedTime = time.Minute
	r.MaxInterval = 30 * time.Second
}
