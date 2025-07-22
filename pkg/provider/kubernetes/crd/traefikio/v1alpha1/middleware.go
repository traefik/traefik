package v1alpha1

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
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/overview/
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
	AddPrefix        *dynamic.AddPrefix        `json:"addPrefix,omitempty"`
	StripPrefix      *dynamic.StripPrefix      `json:"stripPrefix,omitempty"`
	StripPrefixRegex *dynamic.StripPrefixRegex `json:"stripPrefixRegex,omitempty"`
	ReplacePath      *dynamic.ReplacePath      `json:"replacePath,omitempty"`
	ReplacePathRegex *dynamic.ReplacePathRegex `json:"replacePathRegex,omitempty"`
	Chain            *Chain                    `json:"chain,omitempty"`
	// Deprecated: please use IPAllowList instead.
	IPWhiteList       *dynamic.IPWhiteList       `json:"ipWhiteList,omitempty"`
	IPAllowList       *dynamic.IPAllowList       `json:"ipAllowList,omitempty"`
	Headers           *dynamic.Headers           `json:"headers,omitempty"`
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
	Compress          *Compress                  `json:"compress,omitempty"`
	PassTLSClientCert *dynamic.PassTLSClientCert `json:"passTLSClientCert,omitempty"`
	Retry             *Retry                     `json:"retry,omitempty"`
	ContentType       *dynamic.ContentType       `json:"contentType,omitempty"`
	GrpcWeb           *dynamic.GrpcWeb           `json:"grpcWeb,omitempty"`
	// Plugin defines the middleware plugin configuration.
	// More info: https://doc.traefik.io/traefik/plugins/
	Plugin map[string]apiextensionv1.JSON `json:"plugin,omitempty"`
}

// +k8s:deepcopy-gen=true

// ErrorPage holds the custom error middleware configuration.
// This middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/errorpages/
type ErrorPage struct {
	// Status defines which status or range of statuses should result in an error page.
	// It can be either a status code as a number (500),
	// as multiple comma-separated numbers (500,502),
	// as ranges by separating two codes with a dash (500-599),
	// or a combination of the two (404,418,500-599).
	// +kubebuilder:validation:items:Pattern=`^([1-5][0-9]{2}[,-]?)+$`
	Status []string `json:"status,omitempty"`
	// StatusRewrites defines a mapping of status codes that should be returned instead of the original error status codes.
	// For example: "418": 404 or "410-418": 404
	StatusRewrites map[string]int `json:"statusRewrites,omitempty"`
	// Service defines the reference to a Kubernetes Service that will serve the error page.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/errorpages/#service
	Service Service `json:"service,omitempty"`
	// Query defines the URL for the error page (hosted by service).
	// The {status} variable can be used in order to insert the status code in the URL.
	// The {originalStatus} variable can be used in order to insert the upstream status code in the URL.
	// The {url} variable can be used in order to insert the escaped request URL.
	Query string `json:"query,omitempty"`
}

// +k8s:deepcopy-gen=true

// CircuitBreaker holds the circuit breaker configuration.
type CircuitBreaker struct {
	// Expression is the condition that triggers the tripped state.
	Expression string `json:"expression,omitempty" toml:"expression,omitempty" yaml:"expression,omitempty" export:"true"`
	// CheckPeriod is the interval between successive checks of the circuit breaker condition (when in standby state).
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	CheckPeriod *intstr.IntOrString `json:"checkPeriod,omitempty" toml:"checkPeriod,omitempty" yaml:"checkPeriod,omitempty" export:"true"`
	// FallbackDuration is the duration for which the circuit breaker will wait before trying to recover (from a tripped state).
	FallbackDuration *intstr.IntOrString `json:"fallbackDuration,omitempty" toml:"fallbackDuration,omitempty" yaml:"fallbackDuration,omitempty" export:"true"`
	// RecoveryDuration is the duration for which the circuit breaker will try to recover (as soon as it is in recovering state).
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	RecoveryDuration *intstr.IntOrString `json:"recoveryDuration,omitempty" toml:"recoveryDuration,omitempty" yaml:"recoveryDuration,omitempty" export:"true"`
	// ResponseCode is the status code that the circuit breaker will return while it is in the open state.
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=599
	ResponseCode int `json:"responseCode,omitempty" toml:"responseCode,omitempty" yaml:"responseCode,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Chain holds the configuration of the chain middleware.
// This middleware enables to define reusable combinations of other pieces of middleware.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/chain/
type Chain struct {
	// Middlewares is the list of MiddlewareRef which composes the chain.
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// +k8s:deepcopy-gen=true

// BasicAuth holds the basic auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/basicauth/
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
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// DigestAuth holds the digest auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/digestauth/
type DigestAuth struct {
	// Secret is the name of the referenced Kubernetes Secret containing user credentials.
	Secret string `json:"secret,omitempty"`
	// RemoveHeader defines whether to remove the authorization header before forwarding the request to the backend.
	RemoveHeader bool `json:"removeHeader,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardAuth holds the forward auth middleware configuration.
// This middleware delegates the request authentication to a Service.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/forwardauth/
type ForwardAuth struct {
	// Address defines the authentication server address.
	Address string `json:"address,omitempty"`
	// TrustForwardHeader defines whether to trust (ie: forward) all X-Forwarded-* headers.
	TrustForwardHeader bool `json:"trustForwardHeader,omitempty"`
	// AuthResponseHeaders defines the list of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers.
	AuthResponseHeaders []string `json:"authResponseHeaders,omitempty"`
	// AuthResponseHeadersRegex defines the regex to match headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/forwardauth/#authresponseheadersregex
	AuthResponseHeadersRegex string `json:"authResponseHeadersRegex,omitempty"`
	// AuthRequestHeaders defines the list of the headers to copy from the request to the authentication server.
	// If not set or empty then all request headers are passed.
	AuthRequestHeaders []string `json:"authRequestHeaders,omitempty"`
	// TLS defines the configuration used to secure the connection to the authentication server.
	TLS *ClientTLSWithCAOptional `json:"tls,omitempty"`
	// AddAuthCookiesToResponse defines the list of cookies to copy from the authentication server response to the response.
	AddAuthCookiesToResponse []string `json:"addAuthCookiesToResponse,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/forwardauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
	// ForwardBody defines whether to send the request body to the authentication server.
	ForwardBody bool `json:"forwardBody,omitempty"`
	// MaxBodySize defines the maximum body size in bytes allowed to be forwarded to the authentication server.
	MaxBodySize *int64 `json:"maxBodySize,omitempty"`
	// PreserveLocationHeader defines whether to forward the Location header to the client as is or prefix it with the domain name of the authentication server.
	PreserveLocationHeader bool `json:"preserveLocationHeader,omitempty"`
	// PreserveRequestMethod defines whether to preserve the original request method while forwarding the request to the authentication server.
	PreserveRequestMethod bool `json:"preserveRequestMethod,omitempty"`
}

// +k8s:deepcopy-gen=true

// ClientTLSWithCAOptional holds the client TLS configuration.
// TODO: This has to be removed once the CAOptional option is removed.
type ClientTLSWithCAOptional struct {
	ClientTLS `json:",inline"`

	// Deprecated: TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634).
	CAOptional *bool `json:"caOptional,omitempty"`
}

// +k8s:deepcopy-gen=true

// RateLimit holds the rate limit configuration.
// This middleware ensures that services will receive a fair amount of requests, and allows one to define what fair is.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/ratelimit/
type RateLimit struct {
	// Average is the maximum rate, by default in requests/s, allowed for the given source.
	// It defaults to 0, which means no rate limiting.
	// The rate is actually defined by dividing Average by Period. So for a rate below 1req/s,
	// one needs to define a Period larger than a second.
	// +kubebuilder:validation:Minimum=0
	Average *int64 `json:"average,omitempty"`
	// Period, in combination with Average, defines the actual maximum rate, such as:
	// r = Average / Period. It defaults to a second.
	// +kubebuilder:validation:XIntOrString
	Period *intstr.IntOrString `json:"period,omitempty"`
	// Burst is the maximum number of requests allowed to arrive in the same arbitrarily small period of time.
	// It defaults to 1.
	// +kubebuilder:validation:Minimum=0
	Burst *int64 `json:"burst,omitempty"`
	// SourceCriterion defines what criterion is used to group requests as originating from a common source.
	// If several strategies are defined at the same time, an error will be raised.
	// If none are set, the default is to use the request's remote address field (as an ipStrategy).
	SourceCriterion *dynamic.SourceCriterion `json:"sourceCriterion,omitempty"`
	// Redis hold the configs of Redis as bucket in rate limiter.
	Redis *Redis `json:"redis,omitempty"`
}

// +k8s:deepcopy-gen=true

// Redis contains the configuration for using Redis in middleware.
// In a Kubernetes setup, the username and password are stored in a Secret file within the same namespace as the middleware.
type Redis struct {
	// Endpoints contains either a single address or a seed list of host:port addresses.
	// Default value is ["localhost:6379"].
	Endpoints []string `json:"endpoints,omitempty"`
	// TLS defines TLS-specific configurations, including the CA, certificate, and key,
	// which can be provided as a file path or file content.
	TLS *ClientTLS `json:"tls,omitempty"`
	// Secret defines the name of the referenced Kubernetes Secret containing Redis credentials.
	Secret string `json:"secret,omitempty"`
	// DB defines the Redis database that will be selected after connecting to the server.
	DB int `json:"db,omitempty"`
	// PoolSize defines the initial number of socket connections.
	// If the pool runs out of available connections, additional ones will be created beyond PoolSize.
	// This can be limited using MaxActiveConns.
	// // Default value is 0, meaning 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize int `json:"poolSize,omitempty"`
	// MinIdleConns defines the minimum number of idle connections.
	// Default value is 0, and idle connections are not closed by default.
	MinIdleConns int `json:"minIdleConns,omitempty"`
	// MaxActiveConns defines the maximum number of connections allocated by the pool at a given time.
	// Default value is 0, meaning there is no limit.
	MaxActiveConns int `json:"maxActiveConns,omitempty"`
	// ReadTimeout defines the timeout for socket read operations.
	// Default value is 3 seconds.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	ReadTimeout *intstr.IntOrString `json:"readTimeout,omitempty"`
	// WriteTimeout defines the timeout for socket write operations.
	// Default value is 3 seconds.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	WriteTimeout *intstr.IntOrString `json:"writeTimeout,omitempty"`
	// DialTimeout sets the timeout for establishing new connections.
	// Default value is 5 seconds.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	DialTimeout *intstr.IntOrString `json:"dialTimeout,omitempty"`
}

// +k8s:deepcopy-gen=true

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
}

// +k8s:deepcopy-gen=true

// Compress holds the compress middleware configuration.
// This middleware compresses responses before sending them to the client, using gzip, brotli, or zstd compression.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/compress/
type Compress struct {
	// ExcludedContentTypes defines the list of content types to compare the Content-Type header of the incoming requests and responses before compressing.
	// `application/grpc` is always excluded.
	ExcludedContentTypes []string `json:"excludedContentTypes,omitempty"`
	// IncludedContentTypes defines the list of content types to compare the Content-Type header of the responses before compressing.
	IncludedContentTypes []string `json:"includedContentTypes,omitempty"`
	// MinResponseBodyBytes defines the minimum amount of bytes a response body must have to be compressed.
	// Default: 1024.
	// +kubebuilder:validation:Minimum=0
	MinResponseBodyBytes *int `json:"minResponseBodyBytes,omitempty"`
	// Encodings defines the list of supported compression algorithms.
	Encodings []string `json:"encodings,omitempty"`
	// DefaultEncoding specifies the default encoding if the `Accept-Encoding` header is not in the request or contains a wildcard (`*`).
	DefaultEncoding *string `json:"defaultEncoding,omitempty"`
}

// +k8s:deepcopy-gen=true

// Retry holds the retry middleware configuration.
// This middleware reissues requests a given number of times to a backend server if that server does not reply.
// As soon as the server answers, the middleware stops retrying, regardless of the response status.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/http/retry/
type Retry struct {
	// Attempts defines how many times the request should be retried.
	// +kubebuilder:validation:Minimum=0
	Attempts int `json:"attempts,omitempty"`
	// InitialInterval defines the first wait time in the exponential backoff series.
	// The maximum interval is calculated as twice the initialInterval.
	// If unspecified, requests will be retried immediately.
	// The value of initialInterval should be provided in seconds or as a valid duration format,
	// see https://pkg.go.dev/time#ParseDuration.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
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
