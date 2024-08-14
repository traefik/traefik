package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSOption is the CRD implementation of a Traefik TLS Option, allowing to configure some parameters of the TLS connection.
// More info: https://doc.traefik.io/traefik/v2.11/https/tls/#tls-options
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
	// +kubebuilder:default=VersionTLS10
	// +kubebuilder:example=VersionTLS12
	MinVersion string `json:"minVersion,omitempty"`
	// MaxVersion defines the maximum TLS version that Traefik will accept.
	// Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13.
	// Default: None.
	// +kubebuilder:default=none
	// +kubebuilder:example=VersionTLS13
	MaxVersion string `json:"maxVersion,omitempty"`
	// CipherSuites defines the list of supported cipher suites for TLS versions up to TLS 1.2.
	// More info: https://doc.traefik.io/traefik/v2.11/https/tls/#cipher-suites
	// +kubebuilder:example=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256;TLS_RSA_WITH_AES_256_GCM_SHA384
	CipherSuites []string `json:"cipherSuites,omitempty"`
	// CurvePreferences defines the preferred elliptic curves in a specific order.
	// More info: https://doc.traefik.io/traefik/v2.11/https/tls/#curve-preferences
	// +kubebuilder:example=CurveP521;CurveP384
	CurvePreferences []string `json:"curvePreferences,omitempty"`
	// ClientAuth defines the server's policy for TLS Client Authentication.
	ClientAuth ClientAuth `json:"clientAuth,omitempty"`
	// SniStrict defines whether Traefik allows connections from clients connections that do not specify a server_name extension.
	// +kubebuilder:default=false
	// +kubebuilder:example=true
	SniStrict bool `json:"sniStrict,omitempty"`
	// PreferServerCipherSuites defines whether the server chooses a cipher suite among his own instead of among the client's.
	// It is enabled automatically when minVersion or maxVersion is set.
	// Deprecated: https://github.com/golang/go/issues/45430
	// +kubebuilder:default=false
	// +kubebuilder:example=true
	PreferServerCipherSuites bool `json:"preferServerCipherSuites,omitempty"`
	// ALPNProtocols defines the list of supported application level protocols for the TLS handshake, in order of preference.
	// More info: https://doc.traefik.io/traefik/v2.11/https/tls/#alpn-protocols
	// +kubebuilder:example={"foobar"}
	ALPNProtocols []string `json:"alpnProtocols,omitempty"`
}

// +k8s:deepcopy-gen=true

// ClientAuth holds the TLS client authentication configuration.
type ClientAuth struct {
	// SecretNames defines the names of the referenced Kubernetes Secret storing certificate details.
	// +kubebuilder:example=secret-ca1;secret-ca2
	SecretNames []string `json:"secretNames,omitempty"`
	// ClientAuthType defines the client authentication type to apply.
	// +kubebuilder:validation:Enum=NoClientCert;RequestClientCert;RequireAnyClientCert;VerifyClientCertIfGiven;RequireAndVerifyClientCert
	// +kubebuilder:example=VerifyClientCertIfGiven
	ClientAuthType string `json:"clientAuthType,omitempty"`
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
