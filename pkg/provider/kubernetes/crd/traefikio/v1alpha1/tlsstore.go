package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TLSStore is the CRD implementation of a Traefik TLS Store.
// For the time being, only the TLSStore named default is supported.
// This means that you cannot have two stores that are named default in different Kubernetes namespaces.
// More info: https://doc.traefik.io/traefik/v3.5/https/tls/#certificates-stores
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
	DefaultCertificate *Certificate `json:"defaultCertificate,omitempty"`

	// DefaultGeneratedCert defines the default generated certificate configuration.
	DefaultGeneratedCert *tls.GeneratedCert `json:"defaultGeneratedCert,omitempty"`

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

// TLSStoreList is a collection of TLSStore resources.
type TLSStoreList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of TLSStore.
	Items []TLSStore `json:"items"`
}
