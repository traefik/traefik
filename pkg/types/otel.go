package types

// OTelGRPC provides configuration settings for the gRPC open-telemetry.
type OTelGRPC struct {
	Endpoint string            `description:"Sets the gRPC endpoint (host:port) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Insecure bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	TLS      *ClientTLS        `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (o *OTelGRPC) SetDefaults() {
	o.Endpoint = "localhost:4317"
}

// OTelHTTP provides configuration settings for the HTTP open-telemetry.
type OTelHTTP struct {
	Endpoint string            `description:"Sets the HTTP endpoint (scheme://host:port/path) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	TLS      *ClientTLS        `description:"Defines client transport security parameters." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (o *OTelHTTP) SetDefaults() {
	o.Endpoint = "https://localhost:4318"
}
