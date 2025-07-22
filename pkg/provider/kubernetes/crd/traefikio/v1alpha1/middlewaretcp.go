package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareTCP is the CRD implementation of a Traefik TCP middleware.
// More info: https://doc.traefik.io/traefik/v3.5/middlewares/overview/
type MiddlewareTCP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareTCPSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareTCPSpec defines the desired state of a MiddlewareTCP.
type MiddlewareTCPSpec struct {
	// InFlightConn defines the InFlightConn middleware configuration.
	InFlightConn *dynamic.TCPInFlightConn `json:"inFlightConn,omitempty"`
	// IPWhiteList defines the IPWhiteList middleware configuration.
	// This middleware accepts/refuses connections based on the client IP.
	// Deprecated: please use IPAllowList instead.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/tcp/ipwhitelist/
	IPWhiteList *dynamic.TCPIPWhiteList `json:"ipWhiteList,omitempty"`
	// IPAllowList defines the IPAllowList middleware configuration.
	// This middleware accepts/refuses connections based on the client IP.
	// More info: https://doc.traefik.io/traefik/v3.5/middlewares/tcp/ipallowlist/
	IPAllowList *dynamic.TCPIPAllowList `json:"ipAllowList,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareTCPList is a collection of MiddlewareTCP resources.
type MiddlewareTCPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of MiddlewareTCP.
	Items []MiddlewareTCP `json:"items"`
}
