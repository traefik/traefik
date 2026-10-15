package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDP is the CRD implementation of a Traefik UDP middleware.
// More info: https://doc.traefik.io/traefik/v3.7/reference/routing-configuration/udp/middlewares/overview/
type MiddlewareUDP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareUDPSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareUDPSpec defines the desired state of a MiddlewareUDP.
type MiddlewareUDPSpec struct {
	// AccessGovernance defines the access authorization, authentication, audit, and metering configuration.
	AccessGovernance *AccessGovernance `json:"accessGovernance,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDPList is a collection of MiddlewareUDP resources.
type MiddlewareUDPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of MiddlewareUDP.
	Items []MiddlewareUDP `json:"items"`
}
