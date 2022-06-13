package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSOption is the CRD implementation of a Traefik "TLS Option", allowing to configure some parameters of the TLS connection. More info: https://doc.traefik.io/traefik/https/tls/#tls-options
type TLSOption struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSOptionSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSOptionSpec defines the desired state of TLSOption.
type TLSOptionSpec struct {
	// Defines the minimum TLS version that Traefik will accept. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13. Default: VersionTLS10
	MinVersion string `json:"minVersion,omitempty"`
	// Defines the maximum TLS version that Traefik will accept. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13. Default: None.
	MaxVersion string `json:"maxVersion,omitempty"`
	// List of supported cipher suites for TLS versions up to TLS 1.2. More info: https://doc.traefik.io/traefik/https/tls/#cipher-suites
	CipherSuites []string `json:"cipherSuites,omitempty"`
	// This option allows to set the preferred elliptic curves in a specific order. More info: https://doc.traefik.io/traefik/https/tls/#curve-preferences
	CurvePreferences []string   `json:"curvePreferences,omitempty"`
	ClientAuth       ClientAuth `json:"clientAuth,omitempty"`
	// If true, Traefik won't allow connections from clients connections that do not specify a server_name extension. Default: false.
	SniStrict bool `json:"sniStrict,omitempty"`
	// This option allows the server to choose its most preferred cipher suite instead of the client's.
	// It is enabled automatically when minVersion or maxVersion are set.
	PreferServerCipherSuites bool `json:"preferServerCipherSuites,omitempty"`
	// List of supported application level protocols for the TLS handshake, in order of preference. More info: https://doc.traefik.io/traefik/https/tls/#alpn-protocols
	ALPNProtocols []string `json:"alpnProtocols,omitempty"`
}

// +k8s:deepcopy-gen=true

// ClientAuth defines the parameters of the client authentication part of the TLS connection, if any.
type ClientAuth struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretNames []string `json:"secretNames,omitempty"`
	// +kubebuilder:validation:Enum=NoClientCert;RequestClientCert;RequireAnyClientCert;VerifyClientCertIfGiven;RequireAndVerifyClientCert
	// ClientAuthType defines the client authentication type to apply.
	ClientAuthType string `json:"clientAuthType,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSOptionList is a list of TLSOption resources.
type TLSOptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TLSOption `json:"items"`
}
