package static

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address          string                `description:"Entry point address."`
	Transport        *EntryPointsTransport `description:"Configures communication between clients and Traefik."`
	ProxyProtocol    *ProxyProtocol        `description:"Proxy-Protocol configuration." label:"allowEmpty"`
	ForwardedHeaders *ForwardedHeaders     `description:"Trust client forwarding headers."`
}

// SetDefaults sets the default values.
func (e *EntryPoint) SetDefaults() {
	e.Transport = &EntryPointsTransport{}
	e.Transport.SetDefaults()
	e.ForwardedHeaders = &ForwardedHeaders{}
}

// ForwardedHeaders Trust client forwarding headers.
type ForwardedHeaders struct {
	Insecure   bool     `description:"Trust all forwarded headers." export:"true"`
	TrustedIPs []string `description:"Trust only forwarded headers from selected IPs."`
}

// ProxyProtocol contains Proxy-Protocol configuration.
type ProxyProtocol struct {
	Insecure   bool     `description:"Trust all." export:"true"`
	TrustedIPs []string `description:"Trust only selected IPs."`
}

// EntryPoints holds the HTTP entry point list.
type EntryPoints map[string]*EntryPoint

// EntryPointsTransport configures communication between clients and Traefik.
type EntryPointsTransport struct {
	LifeCycle          *LifeCycle          `description:"Timeouts influencing the server life cycle." export:"true"`
	RespondingTimeouts *RespondingTimeouts `description:"Timeouts for incoming requests to the Traefik instance." export:"true"`
}

// SetDefaults sets the default values.
func (t *EntryPointsTransport) SetDefaults() {
	t.LifeCycle = &LifeCycle{}
	t.LifeCycle.SetDefaults()
	t.RespondingTimeouts = &RespondingTimeouts{}
	t.RespondingTimeouts.SetDefaults()
}
