package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteTCPSpec defines the desired state of IngressRouteTCP.
type IngressRouteTCPSpec struct {
	// Routes defines the list of routes.
	Routes []RouteTCP `json:"routes"`
	// EntryPoints defines the list of entry point names to bind to.
	// Entry points have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/entrypoints/
	// Default: all.
	EntryPoints []string `json:"entryPoints,omitempty"`
	// TLS defines the TLS configuration on a layer 4 / TCP Route.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#tls_1
	TLS *TLSTCP `json:"tls,omitempty"`
}

// RouteTCP holds the TCP route configuration.
type RouteTCP struct {
	// Match defines the router's rule.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#rule_1
	Match string `json:"match"`
	// Priority defines the router's priority.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#priority_1
	// +kubebuilder:validation:Maximum=9223372036854774807
	Priority int `json:"priority,omitempty"`
	// Syntax defines the router's rule syntax.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#rulesyntax_1
	// +kubebuilder:validation:Enum=v3;v2
	// Deprecated: Please do not use this field and rewrite the router rules to use the v3 syntax.
	Syntax string `json:"syntax,omitempty"`
	// Services defines the list of TCP services.
	Services []ServiceTCP `json:"services,omitempty"`
	// Middlewares defines the list of references to MiddlewareTCP resources.
	Middlewares []ObjectReference `json:"middlewares,omitempty"`
}

// TLSTCP holds the TLS configuration for an IngressRouteTCP.
// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#tls_1
type TLSTCP struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretName string `json:"secretName,omitempty"`
	// Passthrough defines whether a TLS router will terminate the TLS connection.
	Passthrough bool `json:"passthrough,omitempty"`
	// Options defines the reference to a TLSOption, that specifies the parameters of the TLS connection.
	// If not defined, the `default` TLSOption is used.
	// More info: https://doc.traefik.io/traefik/v3.5/https/tls/#tls-options
	Options *ObjectReference `json:"options,omitempty"`
	// Store defines the reference to the TLSStore, that will be used to store certificates.
	// Please note that only `default` TLSStore can be used.
	Store *ObjectReference `json:"store,omitempty"`
	// CertResolver defines the name of the certificate resolver to use.
	// Cert resolvers have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/https/acme/#certificate-resolvers
	CertResolver string `json:"certResolver,omitempty"`
	// Domains defines the list of domains that will be used to issue certificates.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#domains
	Domains []types.Domain `json:"domains,omitempty"`
}

// ServiceTCP defines an upstream TCP service to proxy traffic to.
type ServiceTCP struct {
	// Name defines the name of the referenced Kubernetes Service.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Kubernetes Service.
	Namespace string `json:"namespace,omitempty"`
	// Port defines the port of a Kubernetes Service.
	// This can be a reference to a named port.
	// +kubebuilder:validation:XIntOrString
	Port intstr.IntOrString `json:"port"`
	// Weight defines the weight used when balancing requests between multiple Kubernetes Service.
	// +kubebuilder:validation:Minimum=0
	Weight *int `json:"weight,omitempty"`
	// TerminationDelay defines the deadline that the proxy sets, after one of its connected peers indicates
	// it has closed the writing capability of its connection, to close the reading capability as well,
	// hence fully terminating the connection.
	// It is a duration in milliseconds, defaulting to 100.
	// A negative value means an infinite deadline (i.e. the reading capability is never closed).
	// Deprecated: TerminationDelay will not be supported in future APIVersions, please use ServersTransport to configure the TerminationDelay instead.
	TerminationDelay *int `json:"terminationDelay,omitempty"`
	// ProxyProtocol defines the PROXY protocol configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#proxy-protocol
	ProxyProtocol *dynamic.ProxyProtocol `json:"proxyProtocol,omitempty"`
	// ServersTransport defines the name of ServersTransportTCP resource to use.
	// It allows to configure the transport between Traefik and your servers.
	// Can only be used on a Kubernetes Service.
	ServersTransport string `json:"serversTransport,omitempty"`
	// TLS determines whether to use TLS when dialing with the backend.
	TLS bool `json:"tls,omitempty"`
	// NativeLB controls, when creating the load-balancer,
	// whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.
	// The Kubernetes Service itself does load-balance to the pods.
	// By default, NativeLB is false.
	NativeLB *bool `json:"nativeLB,omitempty"`
	// NodePortLB controls, when creating the load-balancer,
	// whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is NodePort.
	// It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.
	// By default, NodePortLB is false.
	NodePortLB bool `json:"nodePortLB,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRouteTCP is the CRD implementation of a Traefik TCP Router.
type IngressRouteTCP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteTCPSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteTCPList is a collection of IngressRouteTCP.
type IngressRouteTCPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of IngressRouteTCP.
	Items []IngressRouteTCP `json:"items"`
}
