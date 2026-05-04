package ingressnginx

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
	netv1 "k8s.io/api/networking/v1"
)

// model is a complete, self-contained snapshot of all ingress resources resolved into a Traefik-agnostic intermediate form.
type model struct {
	// Backends holds all resolved upstream services, keyed by backend.Name.
	Backends map[string]*backend

	// Servers holds one entry per distinct hostname across all ingresses.
	Servers map[string]*server

	// PassthroughBackends holds ssl-passthrough entries.
	PassthroughBackends []*sslPassthroughBackend

	// DefaultBackend is the catch-all backend (nil if not configured). It is
	// either provider-level (DefaultBackendService) or sourced from an ingress
	// with spec.defaultBackend and no rules.
	DefaultBackend *backend

	// DefaultBackendLocation, when non-nil, holds metadata for an ingress-level
	// spec.defaultBackend on an ingress with no rules. The translator uses it
	// to attach retry middleware, ServersTransport, and ingress observability
	// metadata to the global catch-all routers.
	DefaultBackendLocation *location

	// Certs holds all TLS certificates resolved from Kubernetes Secrets across all ingresses.
	// The map key is the certificate PEM, the value is the matching private key PEM.
	// Using the cert PEM as the key naturally deduplicates certificates that appear in multiple ingresses.
	Certs map[string]string
}

// backend represents a resolved upstream service.
// Name format follows ingress-nginx: "namespace-serviceName-port".
type backend struct {
	// Name is the unique key for this backend.
	Name string

	// Namespace is the Kubernetes namespace of the service.
	Namespace string

	// ServiceName is the original Kubernetes service name (without namespace or port).
	ServiceName string

	// Endpoints holds the resolved pod addresses.
	Endpoints []endpoint
}

// endpoint is a single resolved pod address.
type endpoint struct {
	// Address is the host:port string ready for use in a Traefik server URL.
	Address string

	// Fenced is true when the pod is terminating (Serving=true, Terminating=true).
	Fenced bool
}

// server represents a virtual host, one per distinct hostname across all ingresses.
// Mirrors ingress-nginx's ingress.Server.
type server struct {
	// Hostname is the virtual host name (e.g. "example.com").
	Hostname string

	// Locations holds the per-path routing units for this hostname.
	Locations []*location
}

// location is the per-path routing unit. It is the central type of the metamodel:
// it holds the backend reference, all resolved annotation data, pre-fetched
// k8s resources (secrets, configmaps), and pre-built middleware configurations,
// so the translator needs no k8s access and performs no annotation interpretation.
//
// Mirrors ingress-nginx's ingress.Location, with Traefik-specific extensions.
type location struct {
	// Path is the HTTP path.
	Path string

	// PathType is the Kubernetes PathType for this location.
	PathType *netv1.PathType

	// Aliases holds the additional hostnames (server-alias) that should be
	// included in this location's host rule. Builder resolves conflicts
	// (already-defined hosts, claims by other ingresses) before populating it.
	Aliases []string

	// UseRegex is true when this location's path must be compiled as a regex
	// (resolved by the builder from use-regex / rewrite-target annotations
	// across all ingresses sharing the host).
	UseRegex bool

	// BackendName is the key into Configuration.Backends for the primary backend.
	BackendName string

	// ServersTransportName is the unique name of the per-ingress transport.
	ServersTransportName string

	// ServersTransport holds the resolved per-ingress transport config (nil for
	// locations that do not need a custom transport, e.g. ingress default backend).
	// The translator registers it once per unique ServersTransportName.
	ServersTransport *dynamic.ServersTransport

	// Config holds all parsed annotation values for this location.
	Config IngressConfig

	// Canary holds the canary routing config (nil if no canary applies).
	Canary *canaryConfig

	// BasicAuth holds the resolved basic-auth middleware (nil unless auth-type=basic).
	BasicAuth *dynamic.BasicAuth

	// DigestAuth holds the resolved digest-auth middleware (nil unless auth-type=digest).
	DigestAuth *dynamic.DigestAuth

	// ResolvedCustomHeaders holds the pre-fetched custom response headers.
	// Populated in Phase 1 from the custom-headers ConfigMap.
	ResolvedCustomHeaders map[string]string

	// ResolvedHTTPErrorBackendName is the key into Configuration.Backends for the
	// custom-http-errors default backend service. Empty if no custom errors configured.
	ResolvedHTTPErrorBackendName string

	// TLSOptionName is set to the TLS option name when auth-tls-secret is configured.
	TLSOptionName string

	// TLSOption holds the resolved client-auth TLS option for this Location
	// (nil when auth-tls-secret is not configured). The translator registers it
	// once per unique TLSOptionName in conf.TLS.Options.
	TLSOption *tls.Options

	// HasTLS is true when the parent ingress has a TLS section covering this
	// location's host. The certs themselves live on the parent Server.
	HasTLS bool

	// ServerSnippet is the resolved server-snippet for this location's hostname
	// (may come from a different ingress on the same host).
	ServerSnippet string

	// LocationIndex is the per-ingress-rule path index (pi) set by the builder.
	// Used by the translator to build stable router key names that are independent
	// of how many other ingresses share the same virtual host.
	LocationIndex int

	// RuleIndex is the spec.rules index (ri) this location was generated from.
	RuleIndex int

	// Observability metadata.
	Namespace   string
	IngressName string
	ServiceName string
	ServicePort string

	// SSLRedirectOnly is true when the non-TLS router should only perform an
	// HTTPS redirect. All other middlewares are suppressed for that router.
	SSLRedirectOnly bool

	// AccessLog, if non-nil, overrides the router-level access log setting.
	AccessLog *bool

	// AppRoot, if non-nil, is the path to redirect bare "/" requests to.
	AppRoot *string

	// UpstreamVhost, if non-nil, overrides the Host header forwarded to the backend.
	UpstreamVhost *dynamic.UpstreamVHost

	// CustomHTTPErrors, if non-nil, configures custom error-page routing.
	CustomHTTPErrors *middlewareCustomHTTPErrors

	// FromToWwwRedirect, if non-nil, describes the extra www↔non-www redirect router.
	FromToWwwRedirect *middlewareFromToWwwRedirect

	// Redirect, if non-nil, configures a permanent or temporary URL redirect.
	Redirect *dynamic.RedirectRegex

	// Buffering, if non-nil, configures request/response buffering limits.
	Buffering *dynamic.Buffering

	// IPAllowList, if non-nil, restricts access to the given CIDR ranges.
	IPAllowList *dynamic.IPAllowList

	// CORS, if non-nil, configures Cross-Origin Resource Sharing headers.
	CORS *dynamic.Headers

	// RewriteTarget, if non-nil, rewrites the request path before forwarding.
	RewriteTarget *dynamic.RewriteTarget

	// RateLimitRPM, if non-nil, applies a per-minute request rate limit.
	RateLimitRPM *dynamic.RateLimit

	// RateLimitRPS, if non-nil, applies a per-second request rate limit.
	RateLimitRPS *dynamic.RateLimit

	// LimitConnections, if non-nil, caps concurrent in-flight requests per source IP.
	LimitConnections *dynamic.InFlightReq

	// AuthTLSPassCert, if non-nil, forwards the client TLS certificate to the backend.
	AuthTLSPassCert *dynamic.AuthTLSPassCertificateToUpstream

	// SnippetAuth, if non-nil, configures server/configuration snippets and forward auth.
	SnippetAuth *dynamic.Snippet

	// Retry, if non-nil, configures request retry behavior.
	Retry *dynamic.Retry

	// If the loc is in an error state.
	Error bool

	// IsIngressDefaultBackend is true when this location represents the
	// ingress-level spec.defaultBackend fallback (host-only rule, no path,
	// no serversTransport).
	IsIngressDefaultBackend bool
}

// middlewareCustomHTTPErrors configures error-page routing for specific HTTP status codes.
type middlewareCustomHTTPErrors struct {
	// Status is the list of HTTP status codes that trigger the error page.
	Status []string
	// ErrorServiceName is the Traefik service name for the provider-level default backend.
	// Set when using the global default backend; empty when using a per-ingress annotation.
	ErrorServiceName string
	// ErrorBackendName is the key into Configuration.Backends for a per-ingress annotation
	// default backend. When set, the translator creates a per-router Traefik service.
	// Exactly one of ErrorServiceName or ErrorBackendName is non-empty.
	ErrorBackendName string
	// The following fields are forwarded as headers to the error service so it
	// can identify the origin of the error.
	Namespace   string
	IngressName string
	ServiceName string
	ServicePort string
}

// middlewareFromToWwwRedirect describes the extra router needed for www↔non-www redirects.
type middlewareFromToWwwRedirect struct {
	// ExtraRouterRule is the Traefik rule expression for the redirect router
	// (e.g. `Host("www.example.com")`).
	ExtraRouterRule string
	// TargetHostname is the hostname to redirect to. The translator builds the
	// redirect URL as "$1://<TargetHostname>$2/$3".
	TargetHostname string
}

// canaryConfig holds the canary routing parameters for a Location.
// Mirrors ingress-nginx's canary.Config / TrafficShapingPolicy.
type canaryConfig struct {
	// BackendName is the key into Configuration.Backends for the canary backend.
	BackendName string

	Weight        int
	WeightTotal   int
	Header        string
	HeaderValue   string
	HeaderPattern string
	Cookie        string
}

// RequiresCanaryRouter returns true when a dedicated canary router is needed
// (cookie or header based routing).
func (c *canaryConfig) RequiresCanaryRouter() bool {
	return c.Cookie != "" || c.Header != ""
}

// RequiresNonCanaryRouter returns true when a dedicated non-canary router is needed.
func (c *canaryConfig) RequiresNonCanaryRouter() bool {
	return c.Weight > 0 && ((c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "") || c.Cookie != "")
}

// sslPassthroughBackend holds a TLS passthrough entry.
type sslPassthroughBackend struct {
	// BackendName is the key into Configuration.Backends.
	BackendName string

	// Hostname is the SNI hostname for the TCP router rule.
	Hostname string

	// RouterKey is the unique key used to name the TCP router.
	RouterKey string
}
