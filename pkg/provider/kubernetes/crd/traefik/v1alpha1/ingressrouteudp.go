package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteUDPSpec defines the desired state of a IngressRouteUDP.
type IngressRouteUDPSpec struct {
	// Routes defines the list of routes.
	Routes []RouteUDP `json:"routes"`
	// EntryPoints defines the list of entry points to bind to.
	// They are set in static configuration. More info: https://doc.traefik.io/traefik/routing/entrypoints/
	// Default: all entrypoints.
	EntryPoints []string `json:"entryPoints,omitempty"`
}

// RouteUDP contains the set of routes.
type RouteUDP struct {
	// Services defines the list of Kubernetes Services.
	Services []ServiceUDP `json:"services,omitempty"`
}

// ServiceUDP defines an upstream to proxy traffic.
type ServiceUDP struct {
	// Name is a reference to a Kubernetes Service object
	Name string `json:"name"`
	// Namespace of the Kubernetes Service object
	Namespace string `json:"namespace,omitempty"`
	// Defines the port of a Kubernetes service. This can be a reference to a named port.
	Port intstr.IntOrString `json:"port"`
	// Used when balancing requests between multiple services
	Weight *int `json:"weight,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRouteUDP is a CRD implementation of a Traefik UDP Router.
type IngressRouteUDP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteUDPSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteUDPList is a list of IngressRoutes.
type IngressRouteUDPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IngressRouteUDP `json:"items"`
}
