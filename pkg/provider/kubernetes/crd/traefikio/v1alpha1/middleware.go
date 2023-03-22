package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// Middleware is the CRD implementation of a Traefik Middleware.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/overview/
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
	StripPrefix       *dynamic.StripPrefix       `json:"stripPrefix,omitempty"`
	StripPrefixRegex  *dynamic.StripPrefixRegex  `json:"stripPrefixRegex,omitempty"`
	ReplacePath       *dynamic.ReplacePath       `json:"replacePath,omitempty"`
	ReplacePathRegex  *dynamic.ReplacePathRegex  `json:"replacePathRegex,omitempty"`
	Chain             *Chain                     `json:"chain,omitempty"`
	IPWhiteList       *dynamic.IPWhiteList       `json:"ipWhiteList,omitempty"`
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
	Compress          *dynamic.Compress          `json:"compress,omitempty"`
	PassTLSClientCert *dynamic.PassTLSClientCert `json:"passTLSClientCert,omitempty"`
	Retry             *Retry                     `json:"retry,omitempty"`
	ContentType       *dynamic.ContentType       `json:"contentType,omitempty"`
	// Plugin defines the middleware plugin configuration.
	// More info: https://doc.traefik.io/traefik/plugins/
	Plugin map[string]apiextensionv1.JSON `json:"plugin,omitempty"`
}

// +k8s:deepcopy-gen=true

// ErrorPage holds the custom error middleware configuration.
// This middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/errorpages/
type ErrorPage struct {
	// Status defines which status or range of statuses should result in an error page.
	// It can be either a status code as a number (500),
	// as multiple comma-separated numbers (500,502),
	// as ranges by separating two codes with a dash (500-599),
	// or a combination of the two (404,418,500-599).
	Status []string `json:"status,omitempty"`
	// Service defines the reference to a Kubernetes Service that will serve the error page.
	// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/errorpages/#service
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
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/chain/
type Chain struct {
	// Middlewares is the list of MiddlewareRef which composes the chain.
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// +k8s:deepcopy-gen=true

// BasicAuth holds the basic auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/basicauth/
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
	// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// DigestAuth holds the digest auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/digestauth/
type DigestAuth struct {
	// Secret is the name of the referenced Kubernetes Secret containing user credentials.
	Secret string `json:"secret,omitempty"`
	// RemoveHeader defines whether to remove the authorization header before forwarding the request to the backend.
	RemoveHeader bool `json:"removeHeader,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardAuth holds the forward auth middleware configuration.
// This middleware delegates the request authentication to a Service.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/forwardauth/
type ForwardAuth struct {
	// Address defines the authentication server address.
	Address string `json:"address,omitempty"`
	// TrustForwardHeader defines whether to trust (ie: forward) all X-Forwarded-* headers.
	TrustForwardHeader bool `json:"trustForwardHeader,omitempty"`
	// AuthResponseHeaders defines the list of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers.
	AuthResponseHeaders []string `json:"authResponseHeaders,omitempty"`
	// AuthResponseHeadersRegex defines the regex to match headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.
	// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/forwardauth/#authresponseheadersregex
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
	CAOptional         bool `json:"caOptional,omitempty"`
}

// +k8s:deepcopy-gen=true

// RateLimit holds the rate limit configuration.
// This middleware ensures that services will receive a fair amount of requests, and allows one to define what fair is.
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/ratelimit/
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
// More info: https://doc.traefik.io/traefik/v2.10/middlewares/http/retry/
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
