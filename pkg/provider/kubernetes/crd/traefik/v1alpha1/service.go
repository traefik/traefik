package v1alpha1

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TraefikService is the CRD implementation of a "Traefik Service".
// TraefikService object allows to:
//  (a) Apply weight to Services on load-balancing
//  (b) Mirror traffic on services
// More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#kind-traefikservice
type TraefikService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec TraefikServiceSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TraefikServiceList is a list of TraefikService resources.
type TraefikServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TraefikService `json:"items"`
}

// +k8s:deepcopy-gen=true

// TraefikServiceSpec defines the desired state of a TraefikService.
type TraefikServiceSpec struct {
	Weighted  *WeightedRoundRobin `json:"weighted,omitempty"`
	Mirroring *Mirroring          `json:"mirroring,omitempty"`
}

// +k8s:deepcopy-gen=true

// Mirroring defines a mirroring service, which is composed of a main load-balancer, and a list of mirrors.
// More info: https://doc.traefik.io/traefik/routing/services/#mirroring-service
type Mirroring struct {
	LoadBalancerSpec `json:",inline"`

	// MaxBodySize defines the maximum size allowed for the body of the request.
	// If the body is larger, the request is not mirrored.
	// Default value is -1, which means unlimited size.
	MaxBodySize *int64 `json:"maxBodySize,omitempty"`
	// Mirrors defines the list of mirrors where Traefik will duplicate the traffic.
	Mirrors []MirrorService `json:"mirrors,omitempty"`
}

// +k8s:deepcopy-gen=true

// MirrorService defines one of the mirrors of a Mirroring service.
type MirrorService struct {
	LoadBalancerSpec `json:",inline"`

	// Percent defines the part of the traffic to mirror.
	// Supported values: 0 to 100.
	Percent int `json:"percent,omitempty"`
}

// +k8s:deepcopy-gen=true

// WeightedRoundRobin allows to apply weight to services on load-balancing.
type WeightedRoundRobin struct {
	// Services defines the list of Kubernetes Service and/or TraefikService to load-balance, with weight.
	Services []Service `json:"services,omitempty"`
	// Sticky defines whether sticky sessions are enabled.
	// More info: https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/#stickiness-and-load-balancing
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
}
