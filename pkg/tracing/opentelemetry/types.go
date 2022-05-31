package opentelemetry

import (
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/types"
)

// Config provides configuration settings for an open-telemetry tracer.
type Config struct {
	GRPC *GRPC `description:"GRPC specific configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" export:"true" label:"allowEmpty" file:"allowEmpty"`

	Compress bool              `description:"Enable compression on the sent data." json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" export:"true"`
	Endpoint string            `description:"Address of the collector endpoint." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	Retry    *Retry            `description:"The retry policy for transient errors that may occurs when exporting traces." json:"retry,omitempty" toml:"retry,omitempty" yaml:"retry,omitempty" export:"true"`
	Timeout  ptypes.Duration   `description:"The max waiting time for the backend to process each spans batch." json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	TLS      *types.ClientTLS  `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true" `
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.Endpoint = "https://localhost:4318/v1/traces"
	c.Retry = &Retry{}
	c.Retry.SetDefaults()
	c.Timeout = ptypes.Duration(10 * time.Second)
}

// GRPC provides gRPC configuration settings for the open-telemetry tracer.
type GRPC struct {
	Insecure           bool            `description:"Connect to endpoint using HTTP." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	ReconnectionPeriod ptypes.Duration `description:"The minimum amount of time between connection attempts to the target endpoint." json:"reconnectionPeriod,omitempty" toml:"reconnectionPeriod,omitempty" yaml:"reconnectionPeriod,omitempty" export:"true"`
	ServiceConfig      string          `description:"Defines the default gRPC service config used." json:"serviceConfig,omitempty" toml:"serviceConfig,omitempty" yaml:"serviceConfig,omitempty" export:"true"`
}

// Retry provides retry configuration settings for the open-telemetry tracer.
type Retry struct {
	InitialInterval ptypes.Duration `description:"The time to wait after the first failure before retrying." json:"initialInterval,omitempty" toml:"initialInterval,omitempty" yaml:"initialInterval,omitempty" export:"true"`
	MaxElapsedTime  ptypes.Duration `description:"The maximum amount of time (including retries) spent trying to send a request/batch." json:"maxElapsedTime,omitempty" toml:"maxElapsedTime,omitempty" yaml:"maxElapsedTime,omitempty" export:"true"`
	MaxInterval     ptypes.Duration `description:"The upper bound on backoff interval." json:"maxInterval,omitempty" toml:"maxInterval,omitempty" yaml:"maxInterval,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (r *Retry) SetDefaults() {
	r.InitialInterval = ptypes.Duration(5 * time.Second)
	r.MaxElapsedTime = ptypes.Duration(time.Minute)
	r.MaxInterval = ptypes.Duration(30 * time.Second)
}
