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

// Middleware is a specification for a Middleware resource.
type Middleware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareSpec holds the Middleware configuration.
type MiddlewareSpec struct {
	AddPrefix         *dynamic.AddPrefix             `json:"addPrefix,omitempty"`
	StripPrefix       *dynamic.StripPrefix           `json:"stripPrefix,omitempty"`
	StripPrefixRegex  *dynamic.StripPrefixRegex      `json:"stripPrefixRegex,omitempty"`
	ReplacePath       *dynamic.ReplacePath           `json:"replacePath,omitempty"`
	ReplacePathRegex  *dynamic.ReplacePathRegex      `json:"replacePathRegex,omitempty"`
	Chain             *Chain                         `json:"chain,omitempty"`
	IPWhiteList       *dynamic.IPWhiteList           `json:"ipWhiteList,omitempty"`
	Headers           *dynamic.Headers               `json:"headers,omitempty"`
	Errors            *ErrorPage                     `json:"errors,omitempty"`
	RateLimit         *RateLimit                     `json:"rateLimit,omitempty"`
	RedirectRegex     *dynamic.RedirectRegex         `json:"redirectRegex,omitempty"`
	RedirectScheme    *dynamic.RedirectScheme        `json:"redirectScheme,omitempty"`
	BasicAuth         *BasicAuth                     `json:"basicAuth,omitempty"`
	DigestAuth        *DigestAuth                    `json:"digestAuth,omitempty"`
	ForwardAuth       *ForwardAuth                   `json:"forwardAuth,omitempty"`
	InFlightReq       *dynamic.InFlightReq           `json:"inFlightReq,omitempty"`
	Buffering         *dynamic.Buffering             `json:"buffering,omitempty"`
	CircuitBreaker    *CircuitBreaker                `json:"circuitBreaker,omitempty"`
	Compress          *dynamic.Compress              `json:"compress,omitempty"`
	PassTLSClientCert *dynamic.PassTLSClientCert     `json:"passTLSClientCert,omitempty"`
	Retry             *Retry                         `json:"retry,omitempty"`
	ContentType       *dynamic.ContentType           `json:"contentType,omitempty"`
	Plugin            map[string]apiextensionv1.JSON `json:"plugin,omitempty"`
}

// +k8s:deepcopy-gen=true

// ErrorPage holds the custom error page configuration.
type ErrorPage struct {
	Status  []string `json:"status,omitempty"`
	Service Service  `json:"service,omitempty"`
	Query   string   `json:"query,omitempty"`
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

// Chain holds a chain of middlewares.
type Chain struct {
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// +k8s:deepcopy-gen=true

// BasicAuth holds the HTTP basic authentication configuration.
type BasicAuth struct {
	Secret       string `json:"secret,omitempty"`
	Realm        string `json:"realm,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	HeaderField  string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// DigestAuth holds the Digest HTTP authentication configuration.
type DigestAuth struct {
	Secret       string `json:"secret,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	Realm        string `json:"realm,omitempty"`
	HeaderField  string `json:"headerField,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardAuth holds the http forward authentication configuration.
type ForwardAuth struct {
	Address                  string     `json:"address,omitempty"`
	TrustForwardHeader       bool       `json:"trustForwardHeader,omitempty"`
	AuthResponseHeaders      []string   `json:"authResponseHeaders,omitempty"`
	AuthResponseHeadersRegex string     `json:"authResponseHeadersRegex,omitempty"`
	AuthRequestHeaders       []string   `json:"authRequestHeaders,omitempty"`
	TLS                      *ClientTLS `json:"tls,omitempty"`
}

// ClientTLS holds TLS specific configurations as client.
type ClientTLS struct {
	CASecret           string `json:"caSecret,omitempty"`
	CAOptional         bool   `json:"caOptional,omitempty"`
	CertSecret         string `json:"certSecret,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareList is a list of Middleware resources.
type MiddlewareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Middleware `json:"items"`
}

// +k8s:deepcopy-gen=true

// RateLimit holds the rate limiting configuration for a given router.
type RateLimit struct {
	Average         int64                    `json:"average,omitempty"`
	Period          *intstr.IntOrString      `json:"period,omitempty"`
	Burst           *int64                   `json:"burst,omitempty"`
	SourceCriterion *dynamic.SourceCriterion `json:"sourceCriterion,omitempty"`
}

// +k8s:deepcopy-gen=true

// Retry holds the retry configuration.
type Retry struct {
	Attempts        int                `json:"attempts,omitempty"`
	InitialInterval intstr.IntOrString `json:"initialInterval,omitempty"`
}
