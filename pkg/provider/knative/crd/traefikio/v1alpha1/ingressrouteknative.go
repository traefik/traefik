package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteKnativeSpec IngressRouteSpec defines the desired state of IngressRoute.
type IngressRouteKnativeSpec struct {
	// Routes defines the list of routes.
	Routes []KnativeRoute `json:"routes"`
	// EntryPoints defines the list of entry point names to bind to.
	// Entry points have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/entrypoints/
	// Default: all.
	EntryPoints []string `json:"entryPoints,omitempty"`
	// TLS defines the TLS configuration.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/routers/#tls
	TLS *traefikv1alpha1.TLS `json:"tls,omitempty"`
}

// KnativeRoute Route holds the HTTP route configuration.
type KnativeRoute struct {
	// Match defines the router's rule.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/routers/#rule
	Match string `json:"match"`
	// Kind defines the kind of the route.
	// Rule is the only supported kind.
	// +kubebuilder:validation:Enum=Rule
	Kind string `json:"kind,omitempty"`
	// Priority defines the router's priority.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/routers/#priority
	Priority int `json:"priority,omitempty"`
	// Syntax defines the router's rule syntax.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/routers/#rulesyntax
	Syntax string `json:"syntax,omitempty"`
	// Services defines the list of Service.
	// It can contain any combination of TraefikService and/or reference to a Kubernetes Service.
	Services []ServiceKnativeSpec `json:"services,omitempty"`
	// Middlewares defines the list of references to Middleware resources.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/providers/kubernetes-crd/#kind-middleware
	Middlewares []traefikv1alpha1.MiddlewareRef `json:"middlewares,omitempty"`
}

// ServiceKnativeSpec LoadBalancerSpec defines the desired state of LoadBalancer.
// It can reference either a Kubernetes Service object (a load-balancer of servers),
// or a TraefikService object (a load-balancer of Traefik services).
type ServiceKnativeSpec struct {
	// Name defines the name of the referenced Kubernetes Service or TraefikService.
	// The differentiation between the two is specified in the Kind field.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Kubernetes Service or TraefikService.
	Namespace string `json:"namespace,omitempty"`
	// Sticky defines the sticky sessions configuration.
	// More info: https://doc.traefik.io/traefik/v3.2/routing/services/#sticky-sessions
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
	// Port defines the port of a Kubernetes Service.
	// This can be a reference to a named port.
	Port intstr.IntOrString `json:"port,omitempty"`
	// Strategy defines the load balancing strategy between the servers.
	// RoundRobin is the only supported value at the moment.
	Strategy string `json:"strategy,omitempty"`
	// PassHostHeader defines whether the client Host header is forwarded to the upstream Kubernetes Service.
	// By default, passHostHeader is true.
	PassHostHeader *bool `json:"passHostHeader,omitempty"`
	// ResponseForwarding defines how Traefik forwards the response from the upstream Kubernetes Service to the client.
	ResponseForwarding *ResponseForwarding `json:"responseForwarding,omitempty"`
	// ServersTransport defines the name of ServersTransport resource to use.
	// It allows to configure the transport between Traefik and your servers.
	// Can only be used on a Kubernetes Service.
	ServersTransport string `json:"serversTransport,omitempty"`
}

type ResponseForwarding struct {
	// FlushInterval defines the interval, in milliseconds, in between flushes to the client while copying the response body.
	// A negative value means to flush immediately after each write to the client.
	// This configuration is ignored when ReverseProxy recognizes a response as a streaming response;
	// for such responses, writes are flushed to the client immediately.
	// Default: 100ms
	FlushInterval string `json:"flushInterval,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRouteKnative IngressRoute is the CRD implementation of a Traefik HTTP Router.
type IngressRouteKnative struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteKnativeSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteKnativeList IngressRouteList is a collection of IngressRoute.
type IngressRouteKnativeList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of IngressRoute.
	Items []IngressRouteKnative `json:"items"`
}
