package types

// HostResolverConfig contain configuration for CNAME Flattening.
type HostResolverConfig struct {
	CnameFlattening bool   `description:"A flag to enable/disable CNAME flattening" export:"true"`
	ResolvConfig    string `description:"resolv.conf used for DNS resolving" export:"true"`
	ResolvDepth     int    `description:"The maximal depth of DNS recursive resolving" export:"true"`
}

// SetDefaults sets the default values.
func (h *HostResolverConfig) SetDefaults() {
	h.CnameFlattening = false
	h.ResolvConfig = "/etc/resolv.conf"
	h.ResolvDepth = 5
}
