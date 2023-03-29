package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteSpec defines the desired state of IngressRoute.
type IngressRouteSpec struct {
	// Routes defines the list of routes.
	Routes []Route `json:"routes"`
	// EntryPoints defines the list of entry point names to bind to.
	// Entry points have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/entrypoints/
	// Default: all.
	EntryPoints []string `json:"entryPoints,omitempty"`
	// TLS defines the TLS configuration.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/routers/#tls
	TLS *TLS `json:"tls,omitempty"`
}

// Route holds the HTTP route configuration.
type Route struct {
	// Match defines the router's rule.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/routers/#rule
	Match string `json:"match"`
	// Kind defines the kind of the route.
	// Rule is the only supported kind.
	// +kubebuilder:validation:Enum=Rule
	Kind string `json:"kind"`
	// Priority defines the router's priority.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/routers/#priority
	Priority int `json:"priority,omitempty"`
	// Services defines the list of Service.
	// It can contain any combination of TraefikService and/or reference to a Kubernetes Service.
	Services []Service `json:"services,omitempty"`
	// Middlewares defines the list of references to Middleware resources.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/providers/kubernetes-crd/#kind-middleware
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
}

// TLS holds the TLS configuration.
// More info: https://doc.traefik.io/traefik/v2.10/routing/routers/#tls
type TLS struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretName string `json:"secretName,omitempty"`
	// Options defines the reference to a TLSOption, that specifies the parameters of the TLS connection.
	// If not defined, the `default` TLSOption is used.
	// More info: https://doc.traefik.io/traefik/v2.10/https/tls/#tls-options
	Options *TLSOptionRef `json:"options,omitempty"`
	// Store defines the reference to the TLSStore, that will be used to store certificates.
	// Please note that only `default` TLSStore can be used.
	Store *TLSStoreRef `json:"store,omitempty"`
	// CertResolver defines the name of the certificate resolver to use.
	// Cert resolvers have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v2.10/https/acme/#certificate-resolvers
	CertResolver string `json:"certResolver,omitempty"`
	// Domains defines the list of domains that will be used to issue certificates.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/routers/#domains
	Domains []types.Domain `json:"domains,omitempty"`
}

// TLSOptionRef is a reference to a TLSOption resource.
type TLSOptionRef struct {
	// Name defines the name of the referenced TLSOption.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/providers/kubernetes-crd/#kind-tlsoption
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced TLSOption.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/providers/kubernetes-crd/#kind-tlsoption
	Namespace string `json:"namespace,omitempty"`
}

// TLSStoreRef is a reference to a TLSStore resource.
type TLSStoreRef struct {
	// Name defines the name of the referenced TLSStore.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/providers/kubernetes-crd/#kind-tlsstore
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced TLSStore.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/providers/kubernetes-crd/#kind-tlsstore
	Namespace string `json:"namespace,omitempty"`
}

// LoadBalancerSpec defines the desired state of LoadBalancer.
// It can reference either a Kubernetes Service object (a load-balancer of servers),
// or a TraefikService object (a load-balancer of Traefik services).
type LoadBalancerSpec struct {
	// Name defines the name of the referenced Kubernetes Service or TraefikService.
	// The differentiation between the two is specified in the Kind field.
	Name string `json:"name"`
	// Kind defines the kind of the Service.
	// +kubebuilder:validation:Enum=Service;TraefikService
	Kind string `json:"kind,omitempty"`
	// Namespace defines the namespace of the referenced Kubernetes Service or TraefikService.
	Namespace string `json:"namespace,omitempty"`
	// Sticky defines the sticky sessions configuration.
	// More info: https://doc.traefik.io/traefik/v2.10/routing/services/#sticky-sessions
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
	// Port defines the port of a Kubernetes Service.
	// This can be a reference to a named port.
	Port intstr.IntOrString `json:"port,omitempty"`
	// Scheme defines the scheme to use for the request to the upstream Kubernetes Service.
	// It defaults to https when Kubernetes Service port is 443, http otherwise.
	Scheme string `json:"scheme,omitempty"`
	// Strategy defines the load balancing strategy between the servers.
	// RoundRobin is the only supported value at the moment.
	Strategy string `json:"strategy,omitempty"`
	// PassHostHeader defines whether the client Host header is forwarded to the upstream Kubernetes Service.
	// By default, passHostHeader is true.
	PassHostHeader *bool `json:"passHostHeader,omitempty"`
	// ResponseForwarding defines how Traefik forwards the response from the upstream Kubernetes Service to the client.
	ResponseForwarding *dynamic.ResponseForwarding `json:"responseForwarding,omitempty"`
	// ServersTransport defines the name of ServersTransport resource to use.
	// It allows to configure the transport between Traefik and your servers.
	// Can only be used on a Kubernetes Service.
	ServersTransport string `json:"serversTransport,omitempty"`
	// Weight defines the weight and should only be specified when Name references a TraefikService object
	// (and to be precise, one that embeds a Weighted Round Robin).
	Weight *int `json:"weight,omitempty"`
	// NativeLB controls, when creating the load-balancer,
	// whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.
	// The Kubernetes Service itself does load-balance to the pods.
	// By default, NativeLB is false.
	NativeLB bool `json:"nativeLB,omitempty"`
}

// Service defines an upstream HTTP service to proxy traffic to.
type Service struct {
	LoadBalancerSpec `json:",inline"`
}

// MiddlewareRef is a reference to a Middleware resource.
type MiddlewareRef struct {
	// Name defines the name of the referenced Middleware resource.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Middleware resource.
	Namespace string `json:"namespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRoute is the CRD implementation of a Traefik HTTP Router.
type IngressRoute struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteList is a collection of IngressRoute.
type IngressRouteList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of IngressRoute.
	Items []IngressRoute `json:"items"`
}
