package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSStore is the CRD implementation of a Traefik "TLS Store". 
// Traefik currently only uses the TLS Store named "default". This means that if you have two stores that are named default in different kubernetes namespaces, they may be randomly chosen. For the time being, please only configure one TLSSTore named default.
// More info: https://doc.traefik.io/traefik/https/tls/#certificates-stores
type TLSStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSStoreSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSStoreSpec defines the desired state of TLSStore.
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
