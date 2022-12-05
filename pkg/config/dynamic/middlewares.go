package dynamic

import (
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/types"
)

// +k8s:deepcopy-gen=true

// Middleware holds the Middleware configuration.
type Middleware struct {
	AddPrefix         *AddPrefix         `json:"addPrefix,omitempty" toml:"addPrefix,omitempty" yaml:"addPrefix,omitempty" export:"true"`
	StripPrefix       *StripPrefix       `json:"stripPrefix,omitempty" toml:"stripPrefix,omitempty" yaml:"stripPrefix,omitempty" export:"true"`
	StripPrefixRegex  *StripPrefixRegex  `json:"stripPrefixRegex,omitempty" toml:"stripPrefixRegex,omitempty" yaml:"stripPrefixRegex,omitempty" export:"true"`
	ReplacePath       *ReplacePath       `json:"replacePath,omitempty" toml:"replacePath,omitempty" yaml:"replacePath,omitempty" export:"true"`
	ReplacePathRegex  *ReplacePathRegex  `json:"replacePathRegex,omitempty" toml:"replacePathRegex,omitempty" yaml:"replacePathRegex,omitempty" export:"true"`
	Chain             *Chain             `json:"chain,omitempty" toml:"chain,omitempty" yaml:"chain,omitempty" export:"true"`
	IPAllowList       *IPAllowList       `json:"ipAllowList,omitempty" toml:"ipAllowList,omitempty" yaml:"ipAllowList,omitempty" export:"true"`
	Headers           *Headers           `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	Errors            *ErrorPage         `json:"errors,omitempty" toml:"errors,omitempty" yaml:"errors,omitempty" export:"true"`
	RateLimit         *RateLimit         `json:"rateLimit,omitempty" toml:"rateLimit,omitempty" yaml:"rateLimit,omitempty" export:"true"`
	RedirectRegex     *RedirectRegex     `json:"redirectRegex,omitempty" toml:"redirectRegex,omitempty" yaml:"redirectRegex,omitempty" export:"true"`
	RedirectScheme    *RedirectScheme    `json:"redirectScheme,omitempty" toml:"redirectScheme,omitempty" yaml:"redirectScheme,omitempty" export:"true"`
	BasicAuth         *BasicAuth         `json:"basicAuth,omitempty" toml:"basicAuth,omitempty" yaml:"basicAuth,omitempty" export:"true"`
	DigestAuth        *DigestAuth        `json:"digestAuth,omitempty" toml:"digestAuth,omitempty" yaml:"digestAuth,omitempty" export:"true"`
	ForwardAuth       *ForwardAuth       `json:"forwardAuth,omitempty" toml:"forwardAuth,omitempty" yaml:"forwardAuth,omitempty" export:"true"`
	InFlightReq       *InFlightReq       `json:"inFlightReq,omitempty" toml:"inFlightReq,omitempty" yaml:"inFlightReq,omitempty" export:"true"`
	Buffering         *Buffering         `json:"buffering,omitempty" toml:"buffering,omitempty" yaml:"buffering,omitempty" export:"true"`
	CircuitBreaker    *CircuitBreaker    `json:"circuitBreaker,omitempty" toml:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty" export:"true"`
	Compress          *Compress          `json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	PassTLSClientCert *PassTLSClientCert `json:"passTLSClientCert,omitempty" toml:"passTLSClientCert,omitempty" yaml:"passTLSClientCert,omitempty" export:"true"`
	Retry             *Retry             `json:"retry,omitempty" toml:"retry,omitempty" yaml:"retry,omitempty" export:"true"`
	ContentType       *ContentType       `json:"contentType,omitempty" toml:"contentType,omitempty" yaml:"contentType,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	GrpcWeb           *GrpcWeb           `json:"grpcWeb,omitempty" toml:"grpcWeb,omitempty" yaml:"grpcWeb,omitempty" export:"true"`

	Plugin map[string]PluginConf `json:"plugin,omitempty" toml:"plugin,omitempty" yaml:"plugin,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// GrpcWeb holds the gRPC web middleware configuration.
// This middleware converts a gRPC web request to an HTTP/2 gRPC request.
type GrpcWeb struct {
	// AllowOrigins is a list of allowable origins.
	// Can also be a wildcard origin "*".
	AllowOrigins []string `json:"allowOrigins,omitempty" toml:"allowOrigins,omitempty" yaml:"allowOrigins,omitempty"`
}

// +k8s:deepcopy-gen=true

// ContentType holds the content-type middleware configuration.
// This middleware sets the `Content-Type` header value to the media type detected from the response content,
// when it is not set by the backend.
type ContentType struct{}

// +k8s:deepcopy-gen=true

// AddPrefix holds the add prefix middleware configuration.
// This middleware updates the path of a request before forwarding it.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/addprefix/
type AddPrefix struct {
	// Prefix is the string to add before the current path in the requested URL.
	// It should include a leading slash (/).
	Prefix string `json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// BasicAuth holds the basic auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/
type BasicAuth struct {
	// Users is an array of authorized users.
	// Each user must be declared using the name:hashed-password format.
	// Tip: Use htpasswd to generate the passwords.
	Users Users `json:"users,omitempty" toml:"users,omitempty" yaml:"users,omitempty" loggable:"false"`
	// UsersFile is the path to an external file that contains the authorized users.
	UsersFile string `json:"usersFile,omitempty" toml:"usersFile,omitempty" yaml:"usersFile,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty" toml:"realm,omitempty" yaml:"realm,omitempty"`
	// RemoveHeader sets the removeHeader option to true to remove the authorization header before forwarding the request to your service.
	// Default: false.
	RemoveHeader bool `json:"removeHeader,omitempty" toml:"removeHeader,omitempty" yaml:"removeHeader,omitempty" export:"true"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty" toml:"headerField,omitempty" yaml:"headerField,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Buffering holds the buffering middleware configuration.
// This middleware retries or limits the size of requests that can be forwarded to backends.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/buffering/#maxrequestbodybytes
type Buffering struct {
	// MaxRequestBodyBytes defines the maximum allowed body size for the request (in bytes).
	// If the request exceeds the allowed size, it is not forwarded to the service, and the client gets a 413 (Request Entity Too Large) response.
	// Default: 0 (no maximum).
	MaxRequestBodyBytes int64 `json:"maxRequestBodyBytes,omitempty" toml:"maxRequestBodyBytes,omitempty" yaml:"maxRequestBodyBytes,omitempty" export:"true"`
	// MemRequestBodyBytes defines the threshold (in bytes) from which the request will be buffered on disk instead of in memory.
	// Default: 1048576 (1Mi).
	MemRequestBodyBytes int64 `json:"memRequestBodyBytes,omitempty" toml:"memRequestBodyBytes,omitempty" yaml:"memRequestBodyBytes,omitempty" export:"true"`
	// MaxResponseBodyBytes defines the maximum allowed response size from the service (in bytes).
	// If the response exceeds the allowed size, it is not forwarded to the client. The client gets a 500 (Internal Server Error) response instead.
	// Default: 0 (no maximum).
	MaxResponseBodyBytes int64 `json:"maxResponseBodyBytes,omitempty" toml:"maxResponseBodyBytes,omitempty" yaml:"maxResponseBodyBytes,omitempty" export:"true"`
	// MemResponseBodyBytes defines the threshold (in bytes) from which the response will be buffered on disk instead of in memory.
	// Default: 1048576 (1Mi).
	MemResponseBodyBytes int64 `json:"memResponseBodyBytes,omitempty" toml:"memResponseBodyBytes,omitempty" yaml:"memResponseBodyBytes,omitempty" export:"true"`
	// RetryExpression defines the retry conditions.
	// It is a logical combination of functions with operators AND (&&) and OR (||).
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/buffering/#retryexpression
	RetryExpression string `json:"retryExpression,omitempty" toml:"retryExpression,omitempty" yaml:"retryExpression,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Chain holds the chain middleware configuration.
// This middleware enables to define reusable combinations of other pieces of middleware.
type Chain struct {
	// Middlewares is the list of middleware names which composes the chain.
	Middlewares []string `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// CircuitBreaker holds the circuit breaker middleware configuration.
// This middleware protects the system from stacking requests to unhealthy services, resulting in cascading failures.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/circuitbreaker/
type CircuitBreaker struct {
	// Expression defines the expression that, once matched, opens the circuit breaker and applies the fallback mechanism instead of calling the services.
	Expression string `json:"expression,omitempty" toml:"expression,omitempty" yaml:"expression,omitempty" export:"true"`
	// CheckPeriod is the interval between successive checks of the circuit breaker condition (when in standby state).
	CheckPeriod ptypes.Duration `json:"checkPeriod,omitempty" toml:"checkPeriod,omitempty" yaml:"checkPeriod,omitempty" export:"true"`
	// FallbackDuration is the duration for which the circuit breaker will wait before trying to recover (from a tripped state).
	FallbackDuration ptypes.Duration `json:"fallbackDuration,omitempty" toml:"fallbackDuration,omitempty" yaml:"fallbackDuration,omitempty" export:"true"`
	// RecoveryDuration is the duration for which the circuit breaker will try to recover (as soon as it is in recovering state).
	RecoveryDuration ptypes.Duration `json:"recoveryDuration,omitempty" toml:"recoveryDuration,omitempty" yaml:"recoveryDuration,omitempty" export:"true"`
}

// SetDefaults sets the default values on a RateLimit.
func (c *CircuitBreaker) SetDefaults() {
	c.CheckPeriod = ptypes.Duration(100 * time.Millisecond)
	c.FallbackDuration = ptypes.Duration(10 * time.Second)
	c.RecoveryDuration = ptypes.Duration(10 * time.Second)
}

// +k8s:deepcopy-gen=true

// Compress holds the compress middleware configuration.
// This middleware compresses responses before sending them to the client, using gzip compression.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/compress/
type Compress struct {
	// ExcludedContentTypes defines the list of content types to compare the Content-Type header of the incoming requests and responses before compressing.
	// `application/grpc` is always excluded.
	ExcludedContentTypes []string `json:"excludedContentTypes,omitempty" toml:"excludedContentTypes,omitempty" yaml:"excludedContentTypes,omitempty" export:"true"`
	// MinResponseBodyBytes defines the minimum amount of bytes a response body must have to be compressed.
	// Default: 1024.
	MinResponseBodyBytes int `json:"minResponseBodyBytes,omitempty" toml:"minResponseBodyBytes,omitempty" yaml:"minResponseBodyBytes,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// DigestAuth holds the digest auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/digestauth/
type DigestAuth struct {
	// Users defines the authorized users.
	// Each user should be declared using the name:realm:encoded-password format.
	Users Users `json:"users,omitempty" toml:"users,omitempty" yaml:"users,omitempty" loggable:"false"`
	// UsersFile is the path to an external file that contains the authorized users for the middleware.
	UsersFile string `json:"usersFile,omitempty" toml:"usersFile,omitempty" yaml:"usersFile,omitempty"`
	// RemoveHeader defines whether to remove the authorization header before forwarding the request to the backend.
	RemoveHeader bool `json:"removeHeader,omitempty" toml:"removeHeader,omitempty" yaml:"removeHeader,omitempty" export:"true"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty" toml:"realm,omitempty" yaml:"realm,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty" toml:"headerField,omitempty" yaml:"headerField,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ErrorPage holds the custom error middleware configuration.
// This middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.
type ErrorPage struct {
	// Status defines which status or range of statuses should result in an error page.
	// It can be either a status code as a number (500),
	// as multiple comma-separated numbers (500,502),
	// as ranges by separating two codes with a dash (500-599),
	// or a combination of the two (404,418,500-599).
	Status []string `json:"status,omitempty" toml:"status,omitempty" yaml:"status,omitempty" export:"true"`
	// Service defines the name of the service that will serve the error page.
	Service string `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	// Query defines the URL for the error page (hosted by service).
	// The {status} variable can be used in order to insert the status code in the URL.
	Query string `json:"query,omitempty" toml:"query,omitempty" yaml:"query,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ForwardAuth holds the forward auth middleware configuration.
// This middleware delegates the request authentication to a Service.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/forwardauth/
type ForwardAuth struct {
	// Address defines the authentication server address.
	Address string `json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	// TLS defines the configuration used to secure the connection to the authentication server.
	TLS *types.ClientTLS `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	// TrustForwardHeader defines whether to trust (ie: forward) all X-Forwarded-* headers.
	TrustForwardHeader bool `json:"trustForwardHeader,omitempty" toml:"trustForwardHeader,omitempty" yaml:"trustForwardHeader,omitempty" export:"true"`
	// AuthResponseHeaders defines the list of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers.
	AuthResponseHeaders []string `json:"authResponseHeaders,omitempty" toml:"authResponseHeaders,omitempty" yaml:"authResponseHeaders,omitempty" export:"true"`
	// AuthResponseHeadersRegex defines the regex to match headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/forwardauth/#authresponseheadersregex
	AuthResponseHeadersRegex string `json:"authResponseHeadersRegex,omitempty" toml:"authResponseHeadersRegex,omitempty" yaml:"authResponseHeadersRegex,omitempty" export:"true"`
	// AuthRequestHeaders defines the list of the headers to copy from the request to the authentication server.
	// If not set or empty then all request headers are passed.
	AuthRequestHeaders []string `json:"authRequestHeaders,omitempty" toml:"authRequestHeaders,omitempty" yaml:"authRequestHeaders,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Headers holds the headers middleware configuration.
// This middleware manages the requests and responses headers.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/headers/#customrequestheaders
type Headers struct {
	// CustomRequestHeaders defines the header names and values to apply to the request.
	CustomRequestHeaders map[string]string `json:"customRequestHeaders,omitempty" toml:"customRequestHeaders,omitempty" yaml:"customRequestHeaders,omitempty" export:"true"`
	// CustomResponseHeaders defines the header names and values to apply to the response.
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty" toml:"customResponseHeaders,omitempty" yaml:"customResponseHeaders,omitempty" export:"true"`

	// AccessControlAllowCredentials defines whether the request can include user credentials.
	AccessControlAllowCredentials bool `json:"accessControlAllowCredentials,omitempty" toml:"accessControlAllowCredentials,omitempty" yaml:"accessControlAllowCredentials,omitempty" export:"true"`
	// AccessControlAllowHeaders defines the Access-Control-Request-Headers values sent in preflight response.
	AccessControlAllowHeaders []string `json:"accessControlAllowHeaders,omitempty" toml:"accessControlAllowHeaders,omitempty" yaml:"accessControlAllowHeaders,omitempty" export:"true"`
	// AccessControlAllowMethods defines the Access-Control-Request-Method values sent in preflight response.
	AccessControlAllowMethods []string `json:"accessControlAllowMethods,omitempty" toml:"accessControlAllowMethods,omitempty" yaml:"accessControlAllowMethods,omitempty" export:"true"`
	// AccessControlAllowOriginList is a list of allowable origins. Can also be a wildcard origin "*".
	AccessControlAllowOriginList []string `json:"accessControlAllowOriginList,omitempty" toml:"accessControlAllowOriginList,omitempty" yaml:"accessControlAllowOriginList,omitempty"`
	// AccessControlAllowOriginListRegex is a list of allowable origins written following the Regular Expression syntax (https://golang.org/pkg/regexp/).
	AccessControlAllowOriginListRegex []string `json:"accessControlAllowOriginListRegex,omitempty" toml:"accessControlAllowOriginListRegex,omitempty" yaml:"accessControlAllowOriginListRegex,omitempty"`
	// AccessControlExposeHeaders defines the Access-Control-Expose-Headers values sent in preflight response.
	AccessControlExposeHeaders []string `json:"accessControlExposeHeaders,omitempty" toml:"accessControlExposeHeaders,omitempty" yaml:"accessControlExposeHeaders,omitempty" export:"true"`
	// AccessControlMaxAge defines the time that a preflight request may be cached.
	AccessControlMaxAge int64 `json:"accessControlMaxAge,omitempty" toml:"accessControlMaxAge,omitempty" yaml:"accessControlMaxAge,omitempty" export:"true"`
	// AddVaryHeader defines whether the Vary header is automatically added/updated when the AccessControlAllowOriginList is set.
	AddVaryHeader bool `json:"addVaryHeader,omitempty" toml:"addVaryHeader,omitempty" yaml:"addVaryHeader,omitempty" export:"true"`
	// AllowedHosts defines the fully qualified list of allowed domain names.
	AllowedHosts []string `json:"allowedHosts,omitempty" toml:"allowedHosts,omitempty" yaml:"allowedHosts,omitempty"`
	// HostsProxyHeaders defines the header keys that may hold a proxied hostname value for the request.
	HostsProxyHeaders []string `json:"hostsProxyHeaders,omitempty" toml:"hostsProxyHeaders,omitempty" yaml:"hostsProxyHeaders,omitempty" export:"true"`
	// SSLProxyHeaders defines the header keys with associated values that would indicate a valid HTTPS request.
	// It can be useful when using other proxies (example: "X-Forwarded-Proto": "https").
	SSLProxyHeaders map[string]string `json:"sslProxyHeaders,omitempty" toml:"sslProxyHeaders,omitempty" yaml:"sslProxyHeaders,omitempty"`
	// STSSeconds defines the max-age of the Strict-Transport-Security header.
	// If set to 0, the header is not set.
	STSSeconds int64 `json:"stsSeconds,omitempty" toml:"stsSeconds,omitempty" yaml:"stsSeconds,omitempty" export:"true"`
	// STSIncludeSubdomains defines whether the includeSubDomains directive is appended to the Strict-Transport-Security header.
	STSIncludeSubdomains bool `json:"stsIncludeSubdomains,omitempty" toml:"stsIncludeSubdomains,omitempty" yaml:"stsIncludeSubdomains,omitempty" export:"true"`
	// STSPreload defines whether the preload flag is appended to the Strict-Transport-Security header.
	STSPreload bool `json:"stsPreload,omitempty" toml:"stsPreload,omitempty" yaml:"stsPreload,omitempty" export:"true"`
	// ForceSTSHeader defines whether to add the STS header even when the connection is HTTP.
	ForceSTSHeader bool `json:"forceSTSHeader,omitempty" toml:"forceSTSHeader,omitempty" yaml:"forceSTSHeader,omitempty" export:"true"`
	// FrameDeny defines whether to add the X-Frame-Options header with the DENY value.
	FrameDeny bool `json:"frameDeny,omitempty" toml:"frameDeny,omitempty" yaml:"frameDeny,omitempty" export:"true"`
	// CustomFrameOptionsValue defines the X-Frame-Options header value.
	// This overrides the FrameDeny option.
	CustomFrameOptionsValue string `json:"customFrameOptionsValue,omitempty" toml:"customFrameOptionsValue,omitempty" yaml:"customFrameOptionsValue,omitempty"`
	// ContentTypeNosniff defines whether to add the X-Content-Type-Options header with the nosniff value.
	ContentTypeNosniff bool `json:"contentTypeNosniff,omitempty" toml:"contentTypeNosniff,omitempty" yaml:"contentTypeNosniff,omitempty" export:"true"`
	// BrowserXSSFilter defines whether to add the X-XSS-Protection header with the value 1; mode=block.
	BrowserXSSFilter bool `json:"browserXssFilter,omitempty" toml:"browserXssFilter,omitempty" yaml:"browserXssFilter,omitempty" export:"true"`
	// CustomBrowserXSSValue defines the X-XSS-Protection header value.
	// This overrides the BrowserXssFilter option.
	CustomBrowserXSSValue string `json:"customBrowserXSSValue,omitempty" toml:"customBrowserXSSValue,omitempty" yaml:"customBrowserXSSValue,omitempty"`
	// ContentSecurityPolicy defines the Content-Security-Policy header value.
	ContentSecurityPolicy string `json:"contentSecurityPolicy,omitempty" toml:"contentSecurityPolicy,omitempty" yaml:"contentSecurityPolicy,omitempty"`
	// PublicKey is the public key that implements HPKP to prevent MITM attacks with forged certificates.
	PublicKey string `json:"publicKey,omitempty" toml:"publicKey,omitempty" yaml:"publicKey,omitempty"`
	// ReferrerPolicy defines the Referrer-Policy header value.
	// This allows sites to control whether browsers forward the Referer header to other sites.
	ReferrerPolicy string `json:"referrerPolicy,omitempty" toml:"referrerPolicy,omitempty" yaml:"referrerPolicy,omitempty" export:"true"`
	// PermissionsPolicy defines the Permissions-Policy header value.
	// This allows sites to control browser features.
	PermissionsPolicy string `json:"permissionsPolicy,omitempty" toml:"permissionsPolicy,omitempty" yaml:"permissionsPolicy,omitempty" export:"true"`
	// IsDevelopment defines whether to mitigate the unwanted effects of the AllowedHosts, SSL, and STS options when developing.
	// Usually testing takes place using HTTP, not HTTPS, and on localhost, not your production domain.
	// If you would like your development environment to mimic production with complete Host blocking, SSL redirects,
	// and STS headers, leave this as false.
	IsDevelopment bool `json:"isDevelopment,omitempty" toml:"isDevelopment,omitempty" yaml:"isDevelopment,omitempty" export:"true"`
}

// HasCustomHeadersDefined checks to see if any of the custom header elements have been set.
func (h *Headers) HasCustomHeadersDefined() bool {
	return h != nil && (len(h.CustomResponseHeaders) != 0 ||
		len(h.CustomRequestHeaders) != 0)
}

// HasCorsHeadersDefined checks to see if any of the cors header elements have been set.
func (h *Headers) HasCorsHeadersDefined() bool {
	return h != nil && (h.AccessControlAllowCredentials ||
		len(h.AccessControlAllowHeaders) != 0 ||
		len(h.AccessControlAllowMethods) != 0 ||
		len(h.AccessControlAllowOriginList) != 0 ||
		len(h.AccessControlAllowOriginListRegex) != 0 ||
		len(h.AccessControlExposeHeaders) != 0 ||
		h.AccessControlMaxAge != 0 ||
		h.AddVaryHeader)
}

// HasSecureHeadersDefined checks to see if any of the secure header elements have been set.
func (h *Headers) HasSecureHeadersDefined() bool {
	return h != nil && (len(h.AllowedHosts) != 0 ||
		len(h.HostsProxyHeaders) != 0 ||
		len(h.SSLProxyHeaders) != 0 ||
		h.STSSeconds != 0 ||
		h.STSIncludeSubdomains ||
		h.STSPreload ||
		h.ForceSTSHeader ||
		h.FrameDeny ||
		h.CustomFrameOptionsValue != "" ||
		h.ContentTypeNosniff ||
		h.BrowserXSSFilter ||
		h.CustomBrowserXSSValue != "" ||
		h.ContentSecurityPolicy != "" ||
		h.PublicKey != "" ||
		h.ReferrerPolicy != "" ||
		h.PermissionsPolicy != "" ||
		h.IsDevelopment)
}

// +k8s:deepcopy-gen=true

// IPStrategy holds the IP strategy configuration used by Traefik to determine the client IP.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/ipallowlist/#ipstrategy
type IPStrategy struct {
	// Depth tells Traefik to use the X-Forwarded-For header and take the IP located at the depth position (starting from the right).
	Depth int `json:"depth,omitempty" toml:"depth,omitempty" yaml:"depth,omitempty" export:"true"`
	// ExcludedIPs configures Traefik to scan the X-Forwarded-For header and select the first IP not in the list.
	ExcludedIPs []string `json:"excludedIPs,omitempty" toml:"excludedIPs,omitempty" yaml:"excludedIPs,omitempty"`
	// TODO(mpl): I think we should make RemoteAddr an explicit field. For one thing, it would yield better documentation.
}

// Get an IP selection strategy.
// If nil return the RemoteAddr strategy
// else return a strategy based on the configuration using the X-Forwarded-For Header.
// Depth override the ExcludedIPs.
func (s *IPStrategy) Get() (ip.Strategy, error) {
	if s == nil {
		return &ip.RemoteAddrStrategy{}, nil
	}

	if s.Depth > 0 {
		return &ip.DepthStrategy{
			Depth: s.Depth,
		}, nil
	}

	if len(s.ExcludedIPs) > 0 {
		checker, err := ip.NewChecker(s.ExcludedIPs)
		if err != nil {
			return nil, err
		}
		return &ip.PoolStrategy{
			Checker: checker,
		}, nil
	}

	return &ip.RemoteAddrStrategy{}, nil
}

// +k8s:deepcopy-gen=true

// IPAllowList holds the IP allowlist middleware configuration.
// This middleware accepts / refuses requests based on the client IP.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/ipallowlist/
type IPAllowList struct {
	// SourceRange defines the set of allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string    `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty"  label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// InFlightReq holds the in-flight request middleware configuration.
// This middleware limits the number of requests being processed and served concurrently.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/inflightreq/
type InFlightReq struct {
	// Amount defines the maximum amount of allowed simultaneous in-flight request.
	// The middleware responds with HTTP 429 Too Many Requests if there are already amount requests in progress (based on the same sourceCriterion strategy).
	Amount int64 `json:"amount,omitempty" toml:"amount,omitempty" yaml:"amount,omitempty" export:"true"`
	// SourceCriterion defines what criterion is used to group requests as originating from a common source.
	// If several strategies are defined at the same time, an error will be raised.
	// If none are set, the default is to use the requestHost.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/inflightreq/#sourcecriterion
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty" toml:"sourceCriterion,omitempty" yaml:"sourceCriterion,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// PassTLSClientCert holds the pass TLS client cert middleware configuration.
// This middleware adds the selected data from the passed client TLS certificate to a header.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/passtlsclientcert/
type PassTLSClientCert struct {
	// PEM sets the X-Forwarded-Tls-Client-Cert header with the certificate.
	PEM bool `json:"pem,omitempty" toml:"pem,omitempty" yaml:"pem,omitempty" export:"true"`
	// Info selects the specific client certificate details you want to add to the X-Forwarded-Tls-Client-Cert-Info header.
	Info *TLSClientCertificateInfo `json:"info,omitempty" toml:"info,omitempty" yaml:"info,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// SourceCriterion defines what criterion is used to group requests as originating from a common source.
// If none are set, the default is to use the request's remote address field.
// All fields are mutually exclusive.
type SourceCriterion struct {
	IPStrategy *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty" export:"true"`
	// RequestHeaderName defines the name of the header used to group incoming requests.
	RequestHeaderName string `json:"requestHeaderName,omitempty" toml:"requestHeaderName,omitempty" yaml:"requestHeaderName,omitempty" export:"true"`
	// RequestHost defines whether to consider the request Host as the source.
	RequestHost bool `json:"requestHost,omitempty" toml:"requestHost,omitempty" yaml:"requestHost,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// RateLimit holds the rate limit configuration.
// This middleware ensures that services will receive a fair amount of requests, and allows one to define what fair is.
type RateLimit struct {
	// Average is the maximum rate, by default in requests/s, allowed for the given source.
	// It defaults to 0, which means no rate limiting.
	// The rate is actually defined by dividing Average by Period. So for a rate below 1req/s,
	// one needs to define a Period larger than a second.
	Average int64 `json:"average,omitempty" toml:"average,omitempty" yaml:"average,omitempty" export:"true"`

	// Period, in combination with Average, defines the actual maximum rate, such as:
	// r = Average / Period. It defaults to a second.
	Period ptypes.Duration `json:"period,omitempty" toml:"period,omitempty" yaml:"period,omitempty" export:"true"`

	// Burst is the maximum number of requests allowed to arrive in the same arbitrarily small period of time.
	// It defaults to 1.
	Burst int64 `json:"burst,omitempty" toml:"burst,omitempty" yaml:"burst,omitempty" export:"true"`

	// SourceCriterion defines what criterion is used to group requests as originating from a common source.
	// If several strategies are defined at the same time, an error will be raised.
	// If none are set, the default is to use the request's remote address field (as an ipStrategy).
	SourceCriterion *SourceCriterion `json:"sourceCriterion,omitempty" toml:"sourceCriterion,omitempty" yaml:"sourceCriterion,omitempty" export:"true"`
}

// SetDefaults sets the default values on a RateLimit.
func (r *RateLimit) SetDefaults() {
	r.Burst = 1
	r.Period = ptypes.Duration(time.Second)
}

// +k8s:deepcopy-gen=true

// RedirectRegex holds the redirect regex middleware configuration.
// This middleware redirects a request using regex matching and replacement.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/redirectregex/#regex
type RedirectRegex struct {
	// Regex defines the regex used to match and capture elements from the request URL.
	Regex string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	// Replacement defines how to modify the URL to have the new target URL.
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
	// Permanent defines whether the redirection is permanent (301).
	Permanent bool `json:"permanent,omitempty" toml:"permanent,omitempty" yaml:"permanent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// RedirectScheme holds the redirect scheme middleware configuration.
// This middleware redirects requests from a scheme/port to another.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/redirectscheme/
type RedirectScheme struct {
	// Scheme defines the scheme of the new URL.
	Scheme string `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty" export:"true"`
	// Port defines the port of the new URL.
	Port string `json:"port,omitempty" toml:"port,omitempty" yaml:"port,omitempty" export:"true"`
	// Permanent defines whether the redirection is permanent (301).
	Permanent bool `json:"permanent,omitempty" toml:"permanent,omitempty" yaml:"permanent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ReplacePath holds the replace path middleware configuration.
// This middleware replaces the path of the request URL and store the original path in an X-Replaced-Path header.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/replacepath/
type ReplacePath struct {
	// Path defines the path to use as replacement in the request URL.
	Path string `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ReplacePathRegex holds the replace path regex middleware configuration.
// This middleware replaces the path of a URL using regex matching and replacement.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/replacepathregex/
type ReplacePathRegex struct {
	// Regex defines the regular expression used to match and capture the path from the request URL.
	Regex string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty" export:"true"`
	// Replacement defines the replacement path format, which can include captured variables.
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Retry holds the retry middleware configuration.
// This middleware reissues requests a given number of times to a backend server if that server does not reply.
// As soon as the server answers, the middleware stops retrying, regardless of the response status.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/retry/
type Retry struct {
	// Attempts defines how many times the request should be retried.
	Attempts int `json:"attempts,omitempty" toml:"attempts,omitempty" yaml:"attempts,omitempty" export:"true"`
	// InitialInterval defines the first wait time in the exponential backoff series.
	// The maximum interval is calculated as twice the initialInterval.
	// If unspecified, requests will be retried immediately.
	// The value of initialInterval should be provided in seconds or as a valid duration format,
	// see https://pkg.go.dev/time#ParseDuration.
	InitialInterval ptypes.Duration `json:"initialInterval,omitempty" toml:"initialInterval,omitempty" yaml:"initialInterval,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// StripPrefix holds the strip prefix middleware configuration.
// This middleware removes the specified prefixes from the URL path.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/stripprefix/
type StripPrefix struct {
	// Prefixes defines the prefixes to strip from the request URL.
	Prefixes []string `json:"prefixes,omitempty" toml:"prefixes,omitempty" yaml:"prefixes,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// StripPrefixRegex holds the strip prefix regex middleware configuration.
// This middleware removes the matching prefixes from the URL path.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/stripprefixregex/
type StripPrefixRegex struct {
	// Regex defines the regular expression to match the path prefix from the request URL.
	Regex []string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TLSClientCertificateInfo holds the client TLS certificate info configuration.
type TLSClientCertificateInfo struct {
	// NotAfter defines whether to add the Not After information from the Validity part.
	NotAfter bool `json:"notAfter,omitempty" toml:"notAfter,omitempty" yaml:"notAfter,omitempty" export:"true"`
	// NotBefore defines whether to add the Not Before information from the Validity part.
	NotBefore bool `json:"notBefore,omitempty" toml:"notBefore,omitempty" yaml:"notBefore,omitempty" export:"true"`
	// Sans defines whether to add the Subject Alternative Name information from the Subject Alternative Name part.
	Sans bool `json:"sans,omitempty" toml:"sans,omitempty" yaml:"sans,omitempty" export:"true"`
	// SerialNumber defines whether to add the client serialNumber information.
	SerialNumber bool `json:"serialNumber,omitempty" toml:"serialNumber,omitempty" yaml:"serialNumber,omitempty" export:"true"`
	// Subject defines the client certificate subject details to add to the X-Forwarded-Tls-Client-Cert-Info header.
	Subject *TLSClientCertificateSubjectDNInfo `json:"subject,omitempty" toml:"subject,omitempty" yaml:"subject,omitempty" export:"true"`
	// Issuer defines the client certificate issuer details to add to the X-Forwarded-Tls-Client-Cert-Info header.
	Issuer *TLSClientCertificateIssuerDNInfo `json:"issuer,omitempty" toml:"issuer,omitempty" yaml:"issuer,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TLSClientCertificateIssuerDNInfo holds the client TLS certificate distinguished name info configuration.
// cf https://tools.ietf.org/html/rfc3739
type TLSClientCertificateIssuerDNInfo struct {
	// Country defines whether to add the country information into the issuer.
	Country bool `json:"country,omitempty" toml:"country,omitempty" yaml:"country,omitempty" export:"true"`
	// Province defines whether to add the province information into the issuer.
	Province bool `json:"province,omitempty" toml:"province,omitempty" yaml:"province,omitempty" export:"true"`
	// Locality defines whether to add the locality information into the issuer.
	Locality bool `json:"locality,omitempty" toml:"locality,omitempty" yaml:"locality,omitempty" export:"true"`
	// Organization defines whether to add the organization information into the issuer.
	Organization bool `json:"organization,omitempty" toml:"organization,omitempty" yaml:"organization,omitempty" export:"true"`
	// CommonName defines whether to add the organizationalUnit information into the issuer.
	CommonName bool `json:"commonName,omitempty" toml:"commonName,omitempty" yaml:"commonName,omitempty" export:"true"`
	// SerialNumber defines whether to add the serialNumber information into the issuer.
	SerialNumber bool `json:"serialNumber,omitempty" toml:"serialNumber,omitempty" yaml:"serialNumber,omitempty" export:"true"`
	// DomainComponent defines whether to add the domainComponent information into the issuer.
	DomainComponent bool `json:"domainComponent,omitempty" toml:"domainComponent,omitempty" yaml:"domainComponent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// TLSClientCertificateSubjectDNInfo holds the client TLS certificate distinguished name info configuration.
// cf https://tools.ietf.org/html/rfc3739
type TLSClientCertificateSubjectDNInfo struct {
	// Country defines whether to add the country information into the subject.
	Country bool `json:"country,omitempty" toml:"country,omitempty" yaml:"country,omitempty" export:"true"`
	// Province defines whether to add the province information into the subject.
	Province bool `json:"province,omitempty" toml:"province,omitempty" yaml:"province,omitempty" export:"true"`
	// Locality defines whether to add the locality information into the subject.
	Locality bool `json:"locality,omitempty" toml:"locality,omitempty" yaml:"locality,omitempty" export:"true"`
	// Organization defines whether to add the organization information into the subject.
	Organization bool `json:"organization,omitempty" toml:"organization,omitempty" yaml:"organization,omitempty" export:"true"`
	// OrganizationalUnit defines whether to add the organizationalUnit information into the subject.
	OrganizationalUnit bool `json:"organizationalUnit,omitempty" toml:"organizationalUnit,omitempty" yaml:"organizationalUnit,omitempty" export:"true"`
	// CommonName defines whether to add the organizationalUnit information into the subject.
	CommonName bool `json:"commonName,omitempty" toml:"commonName,omitempty" yaml:"commonName,omitempty" export:"true"`
	// SerialNumber defines whether to add the serialNumber information into the subject.
	SerialNumber bool `json:"serialNumber,omitempty" toml:"serialNumber,omitempty" yaml:"serialNumber,omitempty" export:"true"`
	// DomainComponent defines whether to add the domainComponent information into the subject.
	DomainComponent bool `json:"domainComponent,omitempty" toml:"domainComponent,omitempty" yaml:"domainComponent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Users holds a list of users.
type Users []string
