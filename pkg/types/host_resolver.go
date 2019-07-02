package types

// HostResolverConfig contain configuration for CNAME Flattening.
type HostResolverConfig struct {
	CnameFlattening bool   `description:"A flag to enable/disable CNAME flattening" json:"cnameFlattening,omitempty" toml:"cnameFlattening,omitempty" yaml:"cnameFlattening,omitempty" export:"true"`
	ResolvConfig    string `description:"resolv.conf used for DNS resolving" json:"resolvConfig,omitempty" toml:"resolvConfig,omitempty" yaml:"resolvConfig,omitempty" export:"true"`
	ResolvDepth     int    `description:"The maximal depth of DNS recursive resolving" json:"resolvDepth,omitempty" toml:"resolvDepth,omitempty" yaml:"resolvDepth,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (h *HostResolverConfig) SetDefaults() {
	h.CnameFlattening = false
	h.ResolvConfig = "/etc/resolv.conf"
	h.ResolvDepth = 5
}
