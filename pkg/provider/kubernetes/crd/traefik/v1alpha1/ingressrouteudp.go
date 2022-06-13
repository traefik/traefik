package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteUDPSpec defines the desired state of a IngressRouteUDP.
type IngressRouteUDPSpec struct {
	// List of routes
	Routes []RouteUDP `json:"routes"`
	// List of entry points to use on this IngressRouteTCP
	// They are defined in your static configuration. (Default: use all entrypoints)
	EntryPoints []string `json:"entryPoints,omitempty"`
}

// RouteUDP contains the set of routes.
type RouteUDP struct {
	// List of Kubernetes Services
	Services []ServiceUDP `json:"services,omitempty"`
}

// TLSOptionUDPRef is a ref to the TLSOption resources.
type TLSOptionUDPRef struct {
	// Name of the TLSOption. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsoption
	Name string `json:"name"`
	// Namespace of the TLSOption. More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-tlsoption
	Namespace string `json:"namespace,omitempty"`
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
