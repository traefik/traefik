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

// TLSOptionSpec configures TLS for an entry point
type TLSOptionSpec struct {
	MinVersion   string   `json:"minversion"`
	CipherSuites []string `json:"ciphersuites"`
	ClientCA     ClientCA `json:"clientca"`
	SniStrict    bool     `json:"snistrict"`
}

// +k8s:deepcopy-gen=true

// ClientCA defines traefik CA files for an entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided
type ClientCA struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretNames []string `json:"secretnames"`
	// Optional indicates if ClientCA are mandatory or have just to be analyzed if provided
	Optional bool `json:"optional"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSOptionList is a list of TLSOption resources.
type TLSOptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TLSOption `json:"items"`
}
