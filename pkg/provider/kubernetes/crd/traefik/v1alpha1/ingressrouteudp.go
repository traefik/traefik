package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteUDPSpec defines the desired state of a IngressRouteUDP.
type IngressRouteUDPSpec struct {
	// Routes defines the list of routes.
	Routes []RouteUDP `json:"routes"`
	// EntryPoints defines the list of entry point names to bind to.
	// Entry points have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.0/routing/entrypoints/
	// Default: all.
	EntryPoints []string `json:"entryPoints,omitempty"`
}

// RouteUDP holds the UDP route configuration.
type RouteUDP struct {
	// Services defines the list of UDP services.
	Services []ServiceUDP `json:"services,omitempty"`
}

// ServiceUDP defines an upstream UDP service to proxy traffic to.
type ServiceUDP struct {
	// Name defines the name of the referenced Kubernetes Service.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Kubernetes Service.
	Namespace string `json:"namespace,omitempty"`
	// Port defines the port of a Kubernetes Service.
	// This can be a reference to a named port.
	Port intstr.IntOrString `json:"port"`
	// Weight defines the weight used when balancing requests between multiple Kubernetes Service.
	Weight *int `json:"weight,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRouteUDP is a CRD implementation of a Traefik UDP Router.
type IngressRouteUDP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteUDPSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteUDPList is a collection of IngressRouteUDP.
type IngressRouteUDPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of IngressRouteUDP.
	Items []IngressRouteUDP `json:"items"`
}
