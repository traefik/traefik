package static

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/types"
)

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address          string                `description:"Entry point address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	AllowACMEByPass  bool                  `description:"Enables handling of ACME TLS and HTTP challenges with custom routers." json:"allowACMEByPass,omitempty" toml:"allowACMEByPass,omitempty" yaml:"allowACMEByPass,omitempty"`
	ReusePort        bool                  `description:"Enables EntryPoints from the same or different processes listening on the same TCP/UDP port." json:"reusePort,omitempty" toml:"reusePort,omitempty" yaml:"reusePort,omitempty"`
	AsDefault        bool                  `description:"Adds this EntryPoint to the list of default EntryPoints to be used on routers that don't have any Entrypoint defined." json:"asDefault,omitempty" toml:"asDefault,omitempty" yaml:"asDefault,omitempty"`
	Transport        *EntryPointsTransport `description:"Configures communication between clients and Traefik." json:"transport,omitempty" toml:"transport,omitempty" yaml:"transport,omitempty" export:"true"`
	ProxyProtocol    *ProxyProtocol        `description:"Proxy-Protocol configuration." json:"proxyProtocol,omitempty" toml:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	ForwardedHeaders *ForwardedHeaders     `description:"Trust client forwarding headers." json:"forwardedHeaders,omitempty" toml:"forwardedHeaders,omitempty" yaml:"forwardedHeaders,omitempty" export:"true"`
	HTTP             HTTPConfig            `description:"HTTP configuration." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
	HTTP2            *HTTP2Config          `description:"HTTP/2 configuration." json:"http2,omitempty" toml:"http2,omitempty" yaml:"http2,omitempty" export:"true"`
	HTTP3            *HTTP3Config          `description:"HTTP/3 configuration." json:"http3,omitempty" toml:"http3,omitempty" yaml:"http3,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	UDP              *UDPConfig            `description:"UDP configuration." json:"udp,omitempty" toml:"udp,omitempty" yaml:"udp,omitempty"`
	Observability    *ObservabilityConfig  `description:"Observability configuration." json:"observability,omitempty" toml:"observability,omitempty" yaml:"observability,omitempty" export:"true"`
}

// GetAddress strips any potential protocol part of the address field of the
// entry point, in order to return the actual address.
func (ep *EntryPoint) GetAddress() string {
	splitN := strings.SplitN(ep.Address, "/", 2)
	return splitN[0]
}

// GetProtocol returns the protocol part of the address field of the entry point.
// If none is specified, it defaults to "tcp".
func (ep *EntryPoint) GetProtocol() (string, error) {
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
	ep.UDP = &UDPConfig{}
	ep.UDP.SetDefaults()
	ep.HTTP = HTTPConfig{}
	ep.HTTP.SetDefaults()
	ep.HTTP2 = &HTTP2Config{}
	ep.HTTP2.SetDefaults()
}

// HTTPConfig is the HTTP configuration of an entry point.
type HTTPConfig struct {
	Redirections          *Redirections `description:"Set of redirection" json:"redirections,omitempty" toml:"redirections,omitempty" yaml:"redirections,omitempty" export:"true"`
	Middlewares           []string      `description:"Default middlewares for the routers linked to the entry point." json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	TLS                   *TLSConfig    `description:"Default TLS configuration for the routers linked to the entry point." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	EncodeQuerySemicolons bool          `description:"Defines whether request query semicolons should be URLEncoded." json:"encodeQuerySemicolons,omitempty" toml:"encodeQuerySemicolons,omitempty" yaml:"encodeQuerySemicolons,omitempty"`
	SanitizePath          *bool         `description:"Defines whether to enable request path sanitization (removal of /./, /../ and multiple slash sequences)." json:"sanitizePath,omitempty" toml:"sanitizePath,omitempty" yaml:"sanitizePath,omitempty" export:"true"`
	MaxHeaderBytes        int           `description:"Maximum size of request headers in bytes." json:"maxHeaderBytes,omitempty" toml:"maxHeaderBytes,omitempty" yaml:"maxHeaderBytes,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *HTTPConfig) SetDefaults() {
	sanitizePath := true
	c.SanitizePath = &sanitizePath
	c.MaxHeaderBytes = http.DefaultMaxHeaderBytes
}

// HTTP2Config is the HTTP2 configuration of an entry point.
type HTTP2Config struct {
	MaxConcurrentStreams int32 `description:"Specifies the number of concurrent streams per connection that each client is allowed to initiate." json:"maxConcurrentStreams,omitempty" toml:"maxConcurrentStreams,omitempty" yaml:"maxConcurrentStreams,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *HTTP2Config) SetDefaults() {
	c.MaxConcurrentStreams = 250 // https://cs.opensource.google/go/x/net/+/cd36cc07:http2/server.go;l=58
}

// HTTP3Config is the HTTP3 configuration of an entry point.
type HTTP3Config struct {
	AdvertisedPort int `description:"UDP port to advertise, on which HTTP/3 is available." json:"advertisedPort,omitempty" toml:"advertisedPort,omitempty" yaml:"advertisedPort,omitempty" export:"true"`
}

// Redirections is a set of redirection for an entry point.
type Redirections struct {
	EntryPoint *RedirectEntryPoint `description:"Set of redirection for an entry point." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
}

// RedirectEntryPoint is the definition of an entry point redirection.
type RedirectEntryPoint struct {
	To        string `description:"Targeted entry point of the redirection." json:"to,omitempty" toml:"to,omitempty" yaml:"to,omitempty" export:"true"`
	Scheme    string `description:"Scheme used for the redirection." json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	Permanent bool   `description:"Applies a permanent redirection." json:"permanent,omitempty" toml:"permanent,omitempty" yaml:"permanent,omitempty" export:"true"`
	Priority  int    `description:"Priority of the generated router." json:"priority,omitempty" toml:"priority,omitempty" yaml:"priority,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (r *RedirectEntryPoint) SetDefaults() {
	r.Scheme = "https"
	r.Permanent = true
	r.Priority = math.MaxInt - 1
}

// TLSConfig is the default TLS configuration for all the routers associated to the concerned entry point.
type TLSConfig struct {
	Options      string         `description:"Default TLS options for the routers linked to the entry point." json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string         `description:"Default certificate resolver for the routers linked to the entry point." json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []types.Domain `description:"Default TLS domains for the routers linked to the entry point." json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

// ForwardedHeaders Trust client forwarding headers.
type ForwardedHeaders struct {
	Insecure   bool     `description:"Trust all forwarded headers." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	TrustedIPs []string `description:"Trust only forwarded headers from selected IPs." json:"trustedIPs,omitempty" toml:"trustedIPs,omitempty" yaml:"trustedIPs,omitempty"`
	Connection []string `description:"List of Connection headers that are allowed to pass through the middleware chain before being removed." json:"connection,omitempty" toml:"connection,omitempty" yaml:"connection,omitempty"`
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
	LifeCycle            *LifeCycle          `description:"Timeouts influencing the server life cycle." json:"lifeCycle,omitempty" toml:"lifeCycle,omitempty" yaml:"lifeCycle,omitempty" export:"true"`
	RespondingTimeouts   *RespondingTimeouts `description:"Timeouts for incoming requests to the Traefik instance." json:"respondingTimeouts,omitempty" toml:"respondingTimeouts,omitempty" yaml:"respondingTimeouts,omitempty" export:"true"`
	KeepAliveMaxTime     ptypes.Duration     `description:"Maximum duration before closing a keep-alive connection." json:"keepAliveMaxTime,omitempty" toml:"keepAliveMaxTime,omitempty" yaml:"keepAliveMaxTime,omitempty" export:"true"`
	KeepAliveMaxRequests int                 `description:"Maximum number of requests before closing a keep-alive connection." json:"keepAliveMaxRequests,omitempty" toml:"keepAliveMaxRequests,omitempty" yaml:"keepAliveMaxRequests,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (t *EntryPointsTransport) SetDefaults() {
	t.LifeCycle = &LifeCycle{}
	t.LifeCycle.SetDefaults()
	t.RespondingTimeouts = &RespondingTimeouts{}
	t.RespondingTimeouts.SetDefaults()
}

// UDPConfig is the UDP configuration of an entry point.
type UDPConfig struct {
	Timeout ptypes.Duration `description:"Timeout defines how long to wait on an idle session before releasing the related resources." json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// SetDefaults sets the default values.
func (u *UDPConfig) SetDefaults() {
	u.Timeout = ptypes.Duration(DefaultUDPTimeout)
}

// ObservabilityConfig holds the observability configuration for an entry point.
type ObservabilityConfig struct {
	AccessLogs     *bool                  `description:"Enables access-logs for this entryPoint." json:"accessLogs,omitempty" toml:"accessLogs,omitempty" yaml:"accessLogs,omitempty" export:"true"`
	Metrics        *bool                  `description:"Enables metrics for this entryPoint." json:"metrics,omitempty" toml:"metrics,omitempty" yaml:"metrics,omitempty" export:"true"`
	Tracing        *bool                  `description:"Enables tracing for this entryPoint." json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty" export:"true"`
	TraceVerbosity types.TracingVerbosity `description:"Defines the tracing verbosity level for this entryPoint." json:"traceVerbosity,omitempty" toml:"traceVerbosity,omitempty" yaml:"traceVerbosity,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (o *ObservabilityConfig) SetDefaults() {
	defaultValue := true
	o.AccessLogs = &defaultValue
	o.Metrics = &defaultValue
	o.Tracing = &defaultValue
	o.TraceVerbosity = types.MinimalVerbosity
}
