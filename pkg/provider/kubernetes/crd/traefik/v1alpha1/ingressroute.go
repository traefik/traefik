package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteSpec defines the desired state of IngressRoute.
type IngressRouteSpec struct {
	// List of routes
	Routes      []Route  `json:"routes"`
  // List of entry points to use on this IngressRoute
  // They are defined in your static configuration. (Default: use all entrypoints)
	EntryPoints []string `json:"entryPoints,omitempty"`
	// Define TLS certificate configuration
	// To enable Let's Encrypt, use an empty TLS struct,
	// e.g. in YAML:
	//
	//	 tls: {} # inline format
	//
	//	 tls:
	//	   secretName: # block format
	TLS         *TLS     `json:"tls,omitempty"`
}

// Route contains the set of routes.
type Route struct {
  // Defines the rule of the underlying HTTP Router. More info: https://doc.traefik.io/traefik/routing/routers/#rule
	Match string `json:"match"`
	// Kind of the route. Rule is the only supported kind.
	// +kubebuilder:validation:Enum=Rule
	Kind        string          `json:"kind"`
	// Priority is used to disambiguate rules of the same length, for route matching
	Priority    int             `json:"priority,omitempty"`
	// List of any combination of TraefikService and/or reference to a kubernetes Service
	Services    []Service       `json:"services,omitempty"`
	// List of references to Middleware. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-middleware
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// TLS contains the TLS certificates configuration of the routes.
// To enable Let's Encrypt, use an empty TLS struct,
// e.g. in YAML:
//
//	 tls: {} # inline format
//
//	 tls:
//	   secretName: # block format
type TLS struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretName string `json:"secretName,omitempty"`
	// Options is a reference to a TLSOption, that specifies the parameters of the TLS connection.
  // Default: use TLSOption named 'default'. More info: https://doc.traefik.io/traefik/https/tls/#tls-options
	Options *TLSOptionRef `json:"options,omitempty"`
	// Store is a reference to a TLSStore, that specifies the parameters of the TLS store.
	Store        *TLSStoreRef   `json:"store,omitempty"`
	// Name of certificate resolver to use. They are defined in static configuration. More info: https://doc.traefik.io/traefik/https/acme/#certificate-resolvers
	CertResolver string         `json:"certResolver,omitempty"`
	// List of Domains. Uses Host in in rule by default. Useful to get a wildcard certificate. More info: https://doc.traefik.io/traefik/routing/routers/#domains
	Domains      []types.Domain `json:"domains,omitempty"`
}

// TLSOptionRef is a ref to the TLSOption resources.
type TLSOptionRef struct {
	// Name of the TLSOption. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsoption
	Name      string `json:"name"`
	// Namespace of the TLSOption. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsoption
	Namespace string `json:"namespace,omitempty"`
}

// TLSStoreRef is a ref to the TLSStore resource.
type TLSStoreRef struct {
	// Name of the TLSStore. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsstore
	Name      string `json:"name"`
	// Namespace of the TLSSTore. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsstore
	Namespace string `json:"namespace,omitempty"`
}

// LoadBalancerSpec can reference either a Kubernetes Service object (a load-balancer of servers),
// or a TraefikService object (a traefik load-balancer of services).
type LoadBalancerSpec struct {
	// Name is a reference to a Kubernetes Service object (for a load-balancer of servers),
	// or to a TraefikService object (service load-balancer, mirroring, etc).
	// The differentiation between the two is specified in the Kind field.
	Name      string `json:"name"`
	// Kind of the Service. Supported values: Service or TraefikService
	// +kubebuilder:validation:Enum=Service;TraefikService
	Kind      string          `json:"kind,omitempty"`
	// Namespace of this reference
	Namespace string          `json:"namespace,omitempty"`
	// Enable and configure sticky session on this service. More info: https://doc.traefik.io/traefik/routing/services/#sticky-sessions
	Sticky    *dynamic.Sticky `json:"sticky,omitempty"`

	// Defines the port of a Kubernetes service. This can be a reference to a named port.
	// Can only be used on a kubernetes Services
	Port               intstr.IntOrString          `json:"port,omitempty"`
	// Define explicitly the scheme. Supported values: http,https,h2c.
	// Default to https when kubernetes service port is 443, http otherwise.
	// Can only be used on a kubernetes Services.
	Scheme             string                      `json:"scheme,omitempty"`
	// Strategy defines the load balancing strategy between the servers. It defaults
	// to RoundRobin. It's the only supported value at the moment.
	// Can only be used on a kubernetes Services.
	Strategy           string                      `json:"strategy,omitempty"`
	// The passHostHeader allows to forward client Host header to server.
	// By default, passHostHeader is true.
	// Can only be used on a kubernetes Services.
	PassHostHeader     *bool                       `json:"passHostHeader,omitempty"`
	// Defines how Traefik forwards the response from the backend server to the client.
	// Can only be used on a kubernetes Services.
	ResponseForwarding *dynamic.ResponseForwarding `json:"responseForwarding,omitempty"`
	// ServersTransport allows to configure the transport between Traefik and your servers.
	// Can only be used on a kubernetes Services.
	ServersTransport   string                      `json:"serversTransport,omitempty"`

	// Weight should only be specified when Name references a TraefikService object
	// (and to be precise, one that embeds a Weighted Round Robin).
	Weight *int `json:"weight,omitempty"`
}

// Service defines an upstream to proxy traffic.
type Service struct {
	LoadBalancerSpec `json:",inline"`
}

// MiddlewareRef is a ref to the Middleware resources.
type MiddlewareRef struct {
	// Name of the Middleware
	Name      string `json:"name"`
	// Namespace of the Middleware
	Namespace string `json:"namespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRoute is the CRD implementation of a Traefik HTTP Router
type IngressRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteList is a list of IngressRoutes.
type IngressRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IngressRoute `json:"items"`
}
