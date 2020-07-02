package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSOption is a specification for a TLSOption resource.
type TLSOption struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSOptionSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSOptionSpec configures TLS for an entry point.
type TLSOptionSpec struct {
	MinVersion               string     `json:"minVersion,omitempty"`
	MaxVersion               string     `json:"maxVersion,omitempty"`
	CipherSuites             []string   `json:"cipherSuites,omitempty"`
	CurvePreferences         []string   `json:"curvePreferences,omitempty"`
	ClientAuth               ClientAuth `json:"clientAuth,omitempty"`
	SniStrict                bool       `json:"sniStrict,omitempty"`
	PreferServerCipherSuites bool       `json:"preferServerCipherSuites,omitempty"`
}

// +k8s:deepcopy-gen=true

// ClientAuth defines the parameters of the client authentication part of the TLS connection, if any.
type ClientAuth struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretNames []string `json:"secretNames"`
	// ClientAuthType defines the client authentication type to apply.
	// The available values are: "NoClientCert", "RequestClientCert", "VerifyClientCertIfGiven" and "RequireAndVerifyClientCert".
	ClientAuthType string `json:"clientAuthType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSOptionList is a list of TLSOption resources.
type TLSOptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TLSOption `json:"items"`
}
