package v1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// Middleware is the CRD implementation of a Traefik Middleware.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/overview/
type Middleware struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareSpec defines the desired state of a Middleware.
type MiddlewareSpec struct {
	AddPrefix         *dynamic.AddPrefix         `json:"addPrefix,omitempty"`
	StripPrefix       *StripPrefix               `json:"stripPrefix,omitempty"`
	StripPrefixRegex  *dynamic.StripPrefixRegex  `json:"stripPrefixRegex,omitempty"`
	ReplacePath       *dynamic.ReplacePath       `json:"replacePath,omitempty"`
	ReplacePathRegex  *dynamic.ReplacePathRegex  `json:"replacePathRegex,omitempty"`
	Chain             *Chain                     `json:"chain,omitempty"`
	IPAllowList       *dynamic.IPAllowList       `json:"ipAllowList,omitempty"`
	Headers           *Headers                   `json:"headers,omitempty"`
	Errors            *ErrorPage                 `json:"errors,omitempty"`
	RateLimit         *RateLimit                 `json:"rateLimit,omitempty"`
	RedirectRegex     *dynamic.RedirectRegex     `json:"redirectRegex,omitempty"`
	RedirectScheme    *dynamic.RedirectScheme    `json:"redirectScheme,omitempty"`
	BasicAuth         *BasicAuth                 `json:"basicAuth,omitempty"`
	DigestAuth        *DigestAuth                `json:"digestAuth,omitempty"`
	ForwardAuth       *ForwardAuth               `json:"forwardAuth,omitempty"`
	InFlightReq       *dynamic.InFlightReq       `json:"inFlightReq,omitempty"`
	Buffering         *dynamic.Buffering         `json:"buffering,omitempty"`
	CircuitBreaker    *CircuitBreaker            `json:"circuitBreaker,omitempty"`
	Compress          *dynamic.Compress          `json:"compress,omitempty"`
	PassTLSClientCert *dynamic.PassTLSClientCert `json:"passTLSClientCert,omitempty"`
	Retry             *Retry                     `json:"retry,omitempty"`
	ContentType       *ContentType               `json:"contentType,omitempty"`
	GrpcWeb           *dynamic.GrpcWeb           `json:"grpcWeb,omitempty"`
	// Plugin defines the middleware plugin configuration.
	// More info: https://doc.traefik.io/traefik/plugins/
	Plugin map[string]apiextensionv1.JSON `json:"plugin,omitempty"`

	// Deprecated: only here to hold v1alpha1 option.
	IPWhiteList *dynamic.IPWhiteList `json:"-"`
}

// +k8s:deepcopy-gen=true

// ContentType holds the content-type middleware configuration.
// This middleware sets the `Content-Type` header value to the media type detected from the response content,
// when it is not set by the backend.
type ContentType struct {
	// Deprecated: only here to hold v1alpha1 option.
	AutoDetect *bool `json:"-"`
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

	// Deprecated: only here to hold v1alpha1 option.
	SSLForceHost *bool `json:"-"`
	// Deprecated: only here to hold v1alpha1 option.
	SSLRedirect *bool `json:"-"`
	// Deprecated: only here to hold v1alpha1 option.
	SSLTemporaryRedirect *bool `json:"-"`
	// Deprecated: only here to hold v1alpha1 option.
	SSLHost *string `json:"-"`
	// Deprecated: only here to hold v1alpha1 option.
	FeaturePolicy *string `json:"-"`
}

// +k8s:deepcopy-gen=true

// StripPrefix holds the strip prefix middleware configuration.
// This middleware removes the specified prefixes from the URL path.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/stripprefix/
type StripPrefix struct {
	// Prefixes defines the prefixes to strip from the request URL.
	Prefixes []string `json:"prefixes,omitempty" toml:"prefixes,omitempty" yaml:"prefixes,omitempty" export:"true"`

	// Deprecated: only here to hold v1alpha1 option.
	ForceSlash *bool `json:"-"`
}

// +k8s:deepcopy-gen=true

// ErrorPage holds the custom error middleware configuration.
// This middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/errorpages/
type ErrorPage struct {
	// Status defines which status or range of statuses should result in an error page.
	// It can be either a status code as a number (500),
	// as multiple comma-separated numbers (500,502),
	// as ranges by separating two codes with a dash (500-599),
	// or a combination of the two (404,418,500-599).
	Status []string `json:"status,omitempty"`
	// Service defines the reference to a Kubernetes Service that will serve the error page.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/errorpages/#service
	Service Service `json:"service,omitempty"`
	// Query defines the URL for the error page (hosted by service).
	// The {status} variable can be used in order to insert the status code in the URL.
	Query string `json:"query,omitempty"`
}

// +k8s:deepcopy-gen=true

// CircuitBreaker holds the circuit breaker configuration.
type CircuitBreaker struct {
	// Expression is the condition that triggers the tripped state.
	Expression string `json:"expression,omitempty" toml:"expression,omitempty" yaml:"expression,omitempty" export:"true"`
	// CheckPeriod is the interval between successive checks of the circuit breaker condition (when in standby state).
	CheckPeriod *intstr.IntOrString `json:"checkPeriod,omitempty" toml:"checkPeriod,omitempty" yaml:"checkPeriod,omitempty" export:"true"`
	// FallbackDuration is the duration for which the circuit breaker will wait before trying to recover (from a tripped state).
	FallbackDuration *intstr.IntOrString `json:"fallbackDuration,omitempty" toml:"fallbackDuration,omitempty" yaml:"fallbackDuration,omitempty" export:"true"`
	// RecoveryDuration is the duration for which the circuit breaker will try to recover (as soon as it is in recovering state).
	RecoveryDuration *intstr.IntOrString `json:"recoveryDuration,omitempty" toml:"recoveryDuration,omitempty" yaml:"recoveryDuration,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Chain holds the configuration of the chain middleware.
// This middleware enables to define reusable combinations of other pieces of middleware.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/chain/
type Chain struct {
	// Middlewares is the list of MiddlewareRef which composes the chain.
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// +k8s:deepcopy-gen=true

// BasicAuth holds the basic auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/
type BasicAuth struct {
	// Secret is the name of the referenced Kubernetes Secret containing user credentials.
	Secret string `json:"secret,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty"`
	// RemoveHeader sets the removeHeader option to true to remove the authorization header before forwarding the request to your service.
	// Default: false.
	RemoveHeader bool `json:"removeHeader,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// DigestAuth holds the digest auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/digestauth/
type DigestAuth struct {
	// Secret is the name of the referenced Kubernetes Secret containing user credentials.
	Secret string `json:"secret,omitempty"`
	// RemoveHeader defines whether to remove the authorization header before forwarding the request to the backend.
	RemoveHeader bool `json:"removeHeader,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardAuth holds the forward auth middleware configuration.
// This middleware delegates the request authentication to a Service.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/forwardauth/
type ForwardAuth struct {
	// Address defines the authentication server address.
	Address string `json:"address,omitempty"`
	// TrustForwardHeader defines whether to trust (ie: forward) all X-Forwarded-* headers.
	TrustForwardHeader bool `json:"trustForwardHeader,omitempty"`
	// AuthResponseHeaders defines the list of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers.
	AuthResponseHeaders []string `json:"authResponseHeaders,omitempty"`
	// AuthResponseHeadersRegex defines the regex to match headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/forwardauth/#authresponseheadersregex
	AuthResponseHeadersRegex string `json:"authResponseHeadersRegex,omitempty"`
	// AuthRequestHeaders defines the list of the headers to copy from the request to the authentication server.
	// If not set or empty then all request headers are passed.
	AuthRequestHeaders []string `json:"authRequestHeaders,omitempty"`
	// TLS defines the configuration used to secure the connection to the authentication server.
	TLS *ClientTLS `json:"tls,omitempty"`
}

// ClientTLS holds the client TLS configuration.
type ClientTLS struct {
	// CASecret is the name of the referenced Kubernetes Secret containing the CA to validate the server certificate.
	// The CA certificate is extracted from key `tls.ca` or `ca.crt`.
	CASecret string `json:"caSecret,omitempty"`
	// CertSecret is the name of the referenced Kubernetes Secret containing the client certificate.
	// The client certificate is extracted from the keys `tls.crt` and `tls.key`.
	CertSecret string `json:"certSecret,omitempty"`
	// InsecureSkipVerify defines whether the server certificates should be validated.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// Deprecated: only here to hold v1alpha1 option.
	CAOptional *bool `json:"-"`
}

// +k8s:deepcopy-gen=true

// RateLimit holds the rate limit configuration.
// This middleware ensures that services will receive a fair amount of requests, and allows one to define what fair is.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/ratelimit/
type RateLimit struct {
	// Average is the maximum rate, by default in requests/s, allowed for the given source.
	// It defaults to 0, which means no rate limiting.
	// The rate is actually defined by dividing Average by Period. So for a rate below 1req/s,
	// one needs to define a Period larger than a second.
	Average int64 `json:"average,omitempty"`
	// Period, in combination with Average, defines the actual maximum rate, such as:
	// r = Average / Period. It defaults to a second.
	Period *intstr.IntOrString `json:"period,omitempty"`
	// Burst is the maximum number of requests allowed to arrive in the same arbitrarily small period of time.
	// It defaults to 1.
	Burst *int64 `json:"burst,omitempty"`
	// SourceCriterion defines what criterion is used to group requests as originating from a common source.
	// If several strategies are defined at the same time, an error will be raised.
	// If none are set, the default is to use the request's remote address field (as an ipStrategy).
	SourceCriterion *dynamic.SourceCriterion `json:"sourceCriterion,omitempty"`
}

// +k8s:deepcopy-gen=true

// Retry holds the retry middleware configuration.
// This middleware reissues requests a given number of times to a backend server if that server does not reply.
// As soon as the server answers, the middleware stops retrying, regardless of the response status.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/http/retry/
type Retry struct {
	// Attempts defines how many times the request should be retried.
	Attempts int `json:"attempts,omitempty"`
	// InitialInterval defines the first wait time in the exponential backoff series.
	// The maximum interval is calculated as twice the initialInterval.
	// If unspecified, requests will be retried immediately.
	// The value of initialInterval should be provided in seconds or as a valid duration format,
	// see https://pkg.go.dev/time#ParseDuration.
	InitialInterval intstr.IntOrString `json:"initialInterval,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareList is a collection of Middleware resources.
type MiddlewareList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of Middleware.
	Items []Middleware `json:"items"`
}
