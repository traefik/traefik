package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteTCPSpec defines the desired state of IngressRouteTCP.
type IngressRouteTCPSpec struct {
	// List of routes
	Routes      []RouteTCP `json:"routes"`
  // List of entry points to use on this IngressRouteTCP
  // They are defined in your static configuration. (Default: use all entrypoints)
	EntryPoints []string   `json:"entryPoints,omitempty"`
	// Define TLS certificate configuration on a layer 4 / TCP Route
	// To enable Let's Encrypt, use an empty TLS struct,
	// e.g. in YAML:
	//
	//   tls: {} # inline format
	//
	//   tls:
	//     secretName: # block format
	TLS         *TLSTCP    `json:"tls,omitempty"`
}

// RouteTCP contains the set of routes.
type RouteTCP struct {
	// Defines the rule of the underlying TCP Router. More info: https://doc.traefik.io/traefik/routing/routers/#rule_1
	Match    string       `json:"match"`
	// Priority is used to disambiguate rules of the same length, for route matching
	Priority int          `json:"priority,omitempty"`
	// List of Kubernetes Service
	Services []ServiceTCP `json:"services,omitempty"`
	// Middlewares contains references to MiddlewareTCP resources.
	Middlewares []ObjectReference `json:"middlewares,omitempty"`
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
	SecretName  string `json:"secretName,omitempty"`
	// A TLS router will terminate the TLS connection by default.
	// However, the passthrough option can be specified to set whether the requests
	// should be forwarded "as is", keeping all data encrypted.
	// Default: false
	Passthrough bool   `json:"passthrough,omitempty"`
	// Options is a reference to a TLSOption, that specifies the parameters of the TLS connection.
  // Default: use TLSOption named 'default'. More info: https://doc.traefik.io/traefik/https/tls/#tls-options
	Options *ObjectReference `json:"options,omitempty"`
	// Store is a reference to a TLSStore, that specifies the parameters of the TLS store.
	Store        *ObjectReference `json:"store,omitempty"`
	// Name of certificate resolver to use. They are defined in static configuration. More info: https://doc.traefik.io/traefik/https/acme/#certificate-resolvers
	CertResolver string           `json:"certResolver,omitempty"`
	// List of Domains. Uses Host in in rule by default. Useful to get a wildcard certificate. More info: https://doc.traefik.io/traefik/routing/routers/#domains
	Domains      []types.Domain   `json:"domains,omitempty"`
}

// ServiceTCP defines an upstream to proxy traffic.
type ServiceTCP struct {
	// Name is a reference to a Kubernetes Service object
	Name             string                 `json:"name"`
	// Namespace of the Kubernetes Service object
	Namespace        string                 `json:"namespace,omitempty"`
	// Defines the port of a Kubernetes service. This can be a reference to a named port.
	Port             intstr.IntOrString     `json:"port"`
	// Used when balancing requests between multiple services
	Weight           *int                   `json:"weight,omitempty"`
	// corresponds to the deadline that the proxy sets, after one of its connected peers indicates
	// it has closed the writing capability of its connection, to close the reading capability as well,
	// hence fully terminating the connection.
	// It is a duration in milliseconds, defaulting to 100.
	// A negative value means an infinite deadline (i.e. the reading capability is never closed).
	TerminationDelay *int                   `json:"terminationDelay,omitempty"`
	// Defines the PROXY protocol configuration. More info: https://doc.traefik.io/traefik/routing/services/#proxy-protocol
	ProxyProtocol    *dynamic.ProxyProtocol `json:"proxyProtocol,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRouteTCP is the CRD implementation of a Traefik TCP Router
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
