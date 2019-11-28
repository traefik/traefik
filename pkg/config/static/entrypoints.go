package static

import (
	"fmt"
	"strings"
)

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address          string                `description:"Entry point address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Transport        *EntryPointsTransport `description:"Configures communication between clients and Traefik." json:"transport,omitempty" toml:"transport,omitempty" yaml:"transport,omitempty"`
	ProxyProtocol    *ProxyProtocol        `description:"Proxy-Protocol configuration." json:"proxyProtocol,omitempty" toml:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty" label:"allowEmpty"`
	ForwardedHeaders *ForwardedHeaders     `description:"Trust client forwarding headers." json:"forwardedHeaders,omitempty" toml:"forwardedHeaders,omitempty" yaml:"forwardedHeaders,omitempty"`
}

func (ep EntryPoint) GetAddress() string {
	splitN := strings.SplitN(ep.Address, "/", 2)
	return splitN[0]
}

func (ep EntryPoint) GetProtocol() (string, error) {
	splitN := strings.SplitN(ep.Address, "/", 2)
	if len(splitN) < 2 {
		return "tcp", nil
	}

	protocol := strings.ToLower(splitN[1])
	if protocol == "tcp" || protocol == "udp" {
		return protocol, nil
	}

	return "", fmt.Errorf("invalid protocol: %s", splitN[1])
}

// SetDefaults sets the default values.
func (e *EntryPoint) SetDefaults() {
	e.Transport = &EntryPointsTransport{}
	e.Transport.SetDefaults()
	e.ForwardedHeaders = &ForwardedHeaders{}
}

// ForwardedHeaders Trust client forwarding headers.
type ForwardedHeaders struct {
	Insecure   bool     `description:"Trust all forwarded headers." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	TrustedIPs []string `description:"Trust only forwarded headers from selected IPs." json:"trustedIPs,omitempty" toml:"trustedIPs,omitempty" yaml:"trustedIPs,omitempty"`
}

// ProxyProtocol contains Proxy-Protocol configuration.
type ProxyProtocol struct {
	Insecure   bool     `description:"Trust all." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	TrustedIPs []string `description:"Trust only selected IPs." json:"trustedIPs,omitempty" toml:"trustedIPs,omitempty" yaml:"trustedIPs,omitempty"`
}

// EntryPoints holds the HTTP entry point list.
type EntryPoints map[string]*EntryPoint

// EntryPointsTransport configures communication between clients and Traefik.
type EntryPointsTransport struct {
	LifeCycle          *LifeCycle          `description:"Timeouts influencing the server life cycle." json:"lifeCycle,omitempty" toml:"lifeCycle,omitempty" yaml:"lifeCycle,omitempty" export:"true"`
	RespondingTimeouts *RespondingTimeouts `description:"Timeouts for incoming requests to the Traefik instance." json:"respondingTimeouts,omitempty" toml:"respondingTimeouts,omitempty" yaml:"respondingTimeouts,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (t *EntryPointsTransport) SetDefaults() {
	t.LifeCycle = &LifeCycle{}
	t.LifeCycle.SetDefaults()
	t.RespondingTimeouts = &RespondingTimeouts{}
	t.RespondingTimeouts.SetDefaults()
}
