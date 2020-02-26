package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteUDPSpec is a specification for a IngressRouteUDPSpec resource.
type IngressRouteUDPSpec struct {
	Routes      []RouteUDP `json:"routes"`
	EntryPoints []string   `json:"entryPoints"`
}

// RouteUDP contains the set of routes.
type RouteUDP struct {
	Services []ServiceUDP `json:"services,omitempty"`
}

// TLSOptionUDPRef is a ref to the TLSOption resources.
type TLSOptionUDPRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ServiceUDP defines an upstream to proxy traffic.
type ServiceUDP struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Port      int32  `json:"port"`
	Weight    *int   `json:"weight,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteUDP is an Ingress CRD specification.
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
