package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TraefikService is the specification for a service (that an IngressRoute refers
// to) that is usually not a terminal service (i.e. not a pod of servers), as
// opposed to a Kubernetes Service. That is to say, it usually refers to other
// (children) services, which themselves can be TraefikServices or Services.
type TraefikService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ServiceSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TraefikServiceList is a list of TraefikService resources.
type TraefikServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TraefikService `json:"items"`
}

// +k8s:deepcopy-gen=true

// ServiceSpec defines whether a TraefikService is a load-balancer of services or a
// mirroring service.
type ServiceSpec struct {
	Weighted  *WeightedRoundRobin `json:"weighted,omitempty"`
	Mirroring *Mirroring          `json:"mirroring,omitempty"`
}

// +k8s:deepcopy-gen=true

// Mirroring defines a mirroring service, which is composed of a main
// load-balancer, and a list of mirrors.
type Mirroring struct {
	LoadBalancerSpec
	MaxBodySize *int64
	Mirrors     []MirrorService `json:"mirrors,omitempty"`
}

// +k8s:deepcopy-gen=true

// MirrorService defines one of the mirrors of a Mirroring service.
type MirrorService struct {
	LoadBalancerSpec
	Percent int `json:"percent,omitempty"`
}

// +k8s:deepcopy-gen=true

// WeightedRoundRobin defines a load-balancer of services.
type WeightedRoundRobin struct {
	Services []Service       `json:"services,omitempty"`
	Sticky   *dynamic.Sticky `json:"sticky,omitempty"`
}
