package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSStore is a specification for a TLSStore resource.
type TLSStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSStoreSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSStoreSpec configures a TLSStore resource.
type TLSStoreSpec struct {
	// DefaultCertificate is the name of the secret holding the default key/certificate pair for the store.
	DefaultCertificate *Certificate `json:"defaultCertificate,omitempty"`
	// Certificates is a list of secret names, each secret holding a key/certificate pair to add to the store.
	Certificates []Certificate `json:"certificates,omitempty"`
}

// +k8s:deepcopy-gen=true

// Certificate holds a secret name for the TLSStore resource.
type Certificate struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretName string `json:"secretName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSStoreList is a list of TLSStore resources.
type TLSStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TLSStore `json:"items"`
}
