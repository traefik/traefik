package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteTCPSpec is a specification for a IngressRouteTCPSpec resource.
type IngressRouteTCPSpec struct {
	Routes      []RouteTCP `json:"routes"`
	EntryPoints []string   `json:"entryPoints"`
	TLS         *TLSTCP    `json:"tls,omitempty"`
}

// RouteTCP contains the set of routes.
type RouteTCP struct {
	Match    string       `json:"match"`
	Services []ServiceTCP `json:"services,omitempty"`
}

// TLSTCP contains the TLS certificates configuration of the routes.
// To enable Let's Encrypt, use an empty TLS struct,
// e.g. in YAML:
//
//	 tls: {} # inline format
//
//	 tls:
//	   secretName: # block format
type TLSTCP struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretName  string `json:"secretName"`
	Passthrough bool   `json:"passthrough"`
	// Options is a reference to a TLSOption, that specifies the parameters of the TLS connection.
	Options *TLSOptionTCPRef `json:"options"`
	// Store is a reference to a TLSStore, that specifies the parameters of the TLS store.
	Store        *TLSStoreTCPRef `json:"store"`
	CertResolver string          `json:"certResolver"`
	Domains      []types.Domain  `json:"domains,omitempty"`
}

// TLSOptionTCPRef is a ref to the TLSOption resources.
type TLSOptionTCPRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// TLSStoreTCPRef is a ref to the TLSStore resources.
type TLSStoreTCPRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ServiceTCP defines an upstream to proxy traffic.
type ServiceTCP struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	Port             int32  `json:"port"`
	Weight           *int   `json:"weight,omitempty"`
	TerminationDelay *int   `json:"terminationDelay,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteTCP is an Ingress CRD specification.
type IngressRouteTCP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteTCPSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteTCPList is a list of IngressRoutes.
type IngressRouteTCPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IngressRouteTCP `json:"items"`
}
