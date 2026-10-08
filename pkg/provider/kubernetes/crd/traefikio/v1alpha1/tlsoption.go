package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSOption is the CRD implementation of a Traefik TLS Option, allowing to configure some parameters of the TLS connection.
// More info: https://doc.traefik.io/traefik/v3.7/reference/routing-configuration/http/tls/tls-certificates/#certificates-stores#tls-options
type TLSOption struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSOptionSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSOptionSpec defines the desired state of a TLSOption.
type TLSOptionSpec struct {
	// MinVersion defines the minimum TLS version that Traefik will accept.
	// Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13.
	// Default: VersionTLS10.
	MinVersion string `json:"minVersion,omitempty"`
	// MaxVersion defines the maximum TLS version that Traefik will accept.
	// Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13.
	// Default: None.
	MaxVersion string `json:"maxVersion,omitempty"`
	// CipherSuites defines the list of supported cipher suites for TLS versions up to TLS 1.2.
	// More info: https://doc.traefik.io/traefik/v3.7/reference/routing-configuration/http/tls/tls-certificates/#certificates-stores#cipher-suites
	CipherSuites []string `json:"cipherSuites,omitempty"`
	// CurvePreferences defines the preferred elliptic curves.
	// More info: https://doc.traefik.io/traefik/v3.7/reference/routing-configuration/http/tls/tls-certificates/#certificates-stores#curve-preferences
	CurvePreferences []string `json:"curvePreferences,omitempty"`
	// ClientAuth defines the server's policy for TLS Client Authentication.
	ClientAuth ClientAuth `json:"clientAuth,omitempty"`
	// SniStrict defines whether Traefik allows connections from clients connections that do not specify a server_name extension.
	SniStrict bool `json:"sniStrict,omitempty"`
	// ALPNProtocols defines the list of supported application level protocols for the TLS handshake, in order of preference.
	// More info: https://doc.traefik.io/traefik/v3.7/reference/routing-configuration/http/tls/tls-certificates/#certificates-stores#alpn-protocols
	ALPNProtocols []string `json:"alpnProtocols,omitempty"`
	// DisableSessionTickets disables TLS session resumption via session tickets.
	DisableSessionTickets bool `json:"disableSessionTickets,omitempty"`
	// PreferServerCipherSuites defines whether the server chooses a cipher suite among his own instead of among the client's.
	// It is enabled automatically when minVersion or maxVersion is set.
	//
	// Deprecated: https://github.com/golang/go/issues/45430
	PreferServerCipherSuites *bool `json:"preferServerCipherSuites,omitempty"`
}

// +k8s:deepcopy-gen=true

// ClientAuth holds the TLS client authentication configuration.
type ClientAuth struct {
	// SecretNames defines the names of the referenced Kubernetes Secret storing certificate details.
	SecretNames []string `json:"secretNames,omitempty"`
	// ClientAuthType defines the client authentication type to apply.
	// +kubebuilder:validation:Enum=NoClientCert;RequestClientCert;RequireAnyClientCert;VerifyClientCertIfGiven;RequireAndVerifyClientCert;RequireAndVerifyClientCertWithExpiry
	ClientAuthType string  `json:"clientAuthType,omitempty"`
	Expiry         *Expiry `json:"expiry,omitempty"`
}

// +k8s:deepcopy-gen=true

// CRLHTTP defines the parameters of the client authentication regarding client certificate expiry based on CRLs
// when loading CRLs from HTTP endpoints.
type CRLHTTPWhitelist struct {
	// Enables whitelist handling, setting it to false may lead to high memory usage
	// due to a high number of CRL being stored
	Enabled bool `json:"enabled,omitempty"`
	// CRL locations whitelist
	DistributionPoints []string `json:"distrivutionPoints,omitempty"`
}

// +k8s:deepcopy-gen=true

// CRLHTTP defines the parameters of the client authentication regarding client certificate expiry based on CRLs
// when loading CRLs from HTTP endpoints.
type CRLHTTP struct {
	// Defines how CRL expiry is handled.
	// Available values are: "Open", "FailedClosed"
	// +kubebuilder:validation:Enum=Open;FailedClosed
	// Default: Open
	ExpirationStrategy string `json:"expirationStrategy,omitempty" toml:"expirationStrategy,omitempty" yaml:"expirationStrategy,omitempty" export:"true"`
	// Timeout for HTTP requests used to fetch CRLs for this TLS Options.
	// If unset, a default timeout configured at the Traefik static level is used.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Timeout *intstr.IntOrString `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	// Error backoff duration for HTTP requests when endpoints are in error
	// If unset, a default duration is set by traefik admins
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	ErrorBackoff *intstr.IntOrString `json:"errorBackoff,omitempty" toml:"errorBackoff,omitempty" yaml:"errorBackoff,omitempty" export:"true"`
	// Max CRL file size exprimed in bytes.
	// Default: 1MB
	MaxCRLBytes *int64 `json:"maxCRLBytes,omitempty" toml:"maxCRLBytes,omitempty" yaml:"maxCRLBytes,omitempty" export:"true"`
	// CRL Distribution Point whitelist config
	Whitelist *CRLHTTPWhitelist `json:"whitelist,omitempty" toml:"whitelist,omitempty" yaml:"whitelist,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// CRL defines the parameters of the client authentication regarding client certificate expiry based on CRLs.
type CRL struct {
	// Mode denotates how CRL validation will be handled.
	// Available values are: "NOOP", "Lax", "Strict".
	// +kubebuilder:validation:Enum=NOOP;Lax;Strict
	// Default: NOOP.
	Mode string `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty" export:"true"`
	// Interval between files reload, setting it to 0 will turn off reload
	// Default: 1H
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	ReloadInterval *intstr.IntOrString `json:"reloadInterval,omitempty" toml:"reloadInterval,omitempty" yaml:"reloadInterval,omitempty" export:"true"`
	// HTTP config for Load mode HTTP
	HTTP *CRLHTTP `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// Expiry defines the parameters of the client authentication regarding client certificate expiry.
type Expiry struct {
	// CRL based expiry configuration
	CRL *CRL `json:"crl,omitempty" toml:"crl,omitempty" yaml:"crl,omitempty" export:"true"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSOptionList is a collection of TLSOption resources.
type TLSOptionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of TLSOption.
	Items []TLSOption `json:"items"`
}
