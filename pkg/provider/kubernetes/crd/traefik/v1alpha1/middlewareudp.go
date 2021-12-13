package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDP is a specification for a MiddlewareUDP resource.
type MiddlewareUDP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareUDPSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareUDPSpec holds the MiddlewareUDPSpec configuration.
type MiddlewareUDPSpec struct {
	IPWhiteList *dynamic.UDPIPWhiteList `json:"ipWhiteList,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDPList is a list of MiddlewareUDPList resources.
type MiddlewareUDPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MiddlewareUDP `json:"items"`
}
