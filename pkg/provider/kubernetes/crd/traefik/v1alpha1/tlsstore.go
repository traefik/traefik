package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSStore is the CRD implementation of a Traefik TLS Store.
// For the time being, only the TLSStore named default is supported.
// This means that you cannot have two stores that are named default in different Kubernetes namespaces.
// More info: https://doc.traefik.io/traefik/v2.7/https/tls/#certificates-stores
type TLSStore struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec TLSStoreSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// TLSStoreSpec defines the desired state of a TLSStore.
type TLSStoreSpec struct {
	// DefaultCertificate defines the default certificate configuration.
	DefaultCertificate DefaultCertificate `json:"defaultCertificate"`
}

// +k8s:deepcopy-gen=true

// DefaultCertificate holds the default certificate configuration.
type DefaultCertificate struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretName string `json:"secretName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TLSStoreList is a collection of TLSStore resources.
type TLSStoreList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of TLSStore.
	Items []TLSStore `json:"items"`
}
