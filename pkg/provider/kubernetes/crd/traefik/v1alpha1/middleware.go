package v1alpha1

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Middleware is a specification for a Middleware resource.
type Middleware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// Middleware holds the Middleware configuration.
type MiddlewareSpec struct {
	AddPrefix         *dynamic.AddPrefix         `json:"addPrefix,omitempty" toml:"addPrefix,omitempty" yaml:"addPrefix,omitempty"`
	StripPrefix       *dynamic.StripPrefix       `json:"stripPrefix,omitempty" toml:"stripPrefix,omitempty" yaml:"stripPrefix,omitempty"`
	StripPrefixRegex  *dynamic.StripPrefixRegex  `json:"stripPrefixRegex,omitempty" toml:"stripPrefixRegex,omitempty" yaml:"stripPrefixRegex,omitempty"`
	ReplacePath       *dynamic.ReplacePath       `json:"replacePath,omitempty" toml:"replacePath,omitempty" yaml:"replacePath,omitempty"`
	ReplacePathRegex  *dynamic.ReplacePathRegex  `json:"replacePathRegex,omitempty" toml:"replacePathRegex,omitempty" yaml:"replacePathRegex,omitempty"`
	Chain             *Chain                     `json:"chain,omitempty" toml:"chain,omitempty" yaml:"chain,omitempty"`
	IPWhiteList       *dynamic.IPWhiteList       `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty"`
	Headers           *dynamic.Headers           `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	Errors            *dynamic.ErrorPage         `json:"errors,omitempty" toml:"errors,omitempty" yaml:"errors,omitempty"`
	RateLimit         *dynamic.RateLimit         `json:"rateLimit,omitempty" toml:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	RedirectRegex     *dynamic.RedirectRegex     `json:"redirectRegex,omitempty" toml:"redirectRegex,omitempty" yaml:"redirectRegex,omitempty"`
	RedirectScheme    *dynamic.RedirectScheme    `json:"redirectScheme,omitempty" toml:"redirectScheme,omitempty" yaml:"redirectScheme,omitempty"`
	BasicAuth         *dynamic.BasicAuth         `json:"basicAuth,omitempty" toml:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	DigestAuth        *dynamic.DigestAuth        `json:"digestAuth,omitempty" toml:"digestAuth,omitempty" yaml:"digestAuth,omitempty"`
	ForwardAuth       *dynamic.ForwardAuth       `json:"forwardAuth,omitempty" toml:"forwardAuth,omitempty" yaml:"forwardAuth,omitempty"`
	InFlightReq       *dynamic.InFlightReq       `json:"inFlightReq,omitempty" toml:"inFlightReq,omitempty" yaml:"inFlightReq,omitempty"`
	Buffering         *dynamic.Buffering         `json:"buffering,omitempty" toml:"buffering,omitempty" yaml:"buffering,omitempty"`
	CircuitBreaker    *dynamic.CircuitBreaker    `json:"circuitBreaker,omitempty" toml:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty"`
	Compress          *dynamic.Compress          `json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" label:"allowEmpty"`
	PassTLSClientCert *dynamic.PassTLSClientCert `json:"passTLSClientCert,omitempty" toml:"passTLSClientCert,omitempty" yaml:"passTLSClientCert,omitempty"`
	Retry             *dynamic.Retry             `json:"retry,omitempty" toml:"retry,omitempty" yaml:"retry,omitempty"`
}

// +k8s:deepcopy-gen=true

// Chain holds a chain of middlewares
type Chain struct {
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareList is a list of Middleware resources.
type MiddlewareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Middleware `json:"items"`
}
