package static

import (
	"fmt"
	"math"
	"strings"

	"github.com/traefik/traefik/v2/pkg/types"
)

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address          string                `description:"Entry point address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Transport        *EntryPointsTransport `description:"Configures communication between clients and Traefik." json:"transport,omitempty" toml:"transport,omitempty" yaml:"transport,omitempty"`
	ProxyProtocol    *ProxyProtocol        `description:"Proxy-Protocol configuration." json:"proxyProtocol,omitempty" toml:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty" label:"allowEmpty" file:"allowEmpty"`
	ForwardedHeaders *ForwardedHeaders     `description:"Trust client forwarding headers." json:"forwardedHeaders,omitempty" toml:"forwardedHeaders,omitempty" yaml:"forwardedHeaders,omitempty"`
	HTTP             HTTPConfig            `description:"HTTP configuration." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty"`
}

// GetAddress strips any potential protocol part of the address field of the
// entry point, in order to return the actual address.
func (ep EntryPoint) GetAddress() string {
	splitN := strings.SplitN(ep.Address, "/", 2)
	return splitN[0]
}

// GetProtocol returns the protocol part of the address field of the entry point.
// If none is specified, it defaults to "tcp".
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
func (ep *EntryPoint) SetDefaults() {
	ep.Transport = &EntryPointsTransport{}
	ep.Transport.SetDefaults()
	ep.ForwardedHeaders = &ForwardedHeaders{}
}

// HTTPConfig is the HTTP configuration of an entry point.
type HTTPConfig struct {
	Redirections *Redirections `description:"Set of redirection" json:"redirections,omitempty" toml:"redirections,omitempty" yaml:"redirections,omitempty"`
	Middlewares  []string      `description:"Default middlewares for the routers linked to the entry point." json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty"`
	TLS          *TLSConfig    `description:"Default TLS configuration for the routers linked to the entry point." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

// Redirections is a set of redirection for an entry point.
type Redirections struct {
	EntryPoint *RedirectEntryPoint `description:"Set of redirection for an entry point." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty"`
}

// RedirectEntryPoint is the definition of an entry point redirection.
type RedirectEntryPoint struct {
	To        string `description:"Targeted entry point of the redirection." json:"to,omitempty" toml:"to,omitempty" yaml:"to,omitempty"`
	Scheme    string `description:"Scheme used for the redirection." json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty"`
	Permanent bool   `description:"Applies a permanent redirection." json:"permanent,omitempty" toml:"permanent,omitempty" yaml:"permanent,omitempty"`
	Priority  int    `description:"Priority of the generated router." json:"priority,omitempty" toml:"priority,omitempty" yaml:"priority,omitempty"`
}

// SetDefaults sets the default values.
func (r *RedirectEntryPoint) SetDefaults() {
	r.Scheme = "https"
	r.Permanent = true
	r.Priority = math.MaxInt32
}

// TLSConfig is the default TLS configuration for all the routers associated to the concerned entry point.
type TLSConfig struct {
	Options      string         `description:"Default TLS options for the routers linked to the entry point." json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty"`
	CertResolver string         `description:"Default certificate resolver for the routers linked to the entry point." json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty"`
	Domains      []types.Domain `description:"Default TLS domains for the routers linked to the entry point." json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty"`
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
