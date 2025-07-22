package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// TraefikService is the CRD implementation of a Traefik Service.
// TraefikService object allows to:
// - Apply weight to Services on load-balancing
// - Mirror traffic on services
// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-traefikservice
type TraefikService struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec TraefikServiceSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TraefikServiceList is a collection of TraefikService resources.
type TraefikServiceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of TraefikService.
	Items []TraefikService `json:"items"`
}

// +k8s:deepcopy-gen=true

// TraefikServiceSpec defines the desired state of a TraefikService.
type TraefikServiceSpec struct {
	// Weighted defines the Weighted Round Robin configuration.
	Weighted *WeightedRoundRobin `json:"weighted,omitempty"`
	// Mirroring defines the Mirroring service configuration.
	Mirroring *Mirroring `json:"mirroring,omitempty"`
}

// +k8s:deepcopy-gen=true

// Mirroring holds the mirroring service configuration.
// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#mirroring-service
type Mirroring struct {
	LoadBalancerSpec `json:",inline"`

	// MirrorBody defines whether the body of the request should be mirrored.
	// Default value is true.
	MirrorBody *bool `json:"mirrorBody,omitempty"`
	// MaxBodySize defines the maximum size allowed for the body of the request.
	// If the body is larger, the request is not mirrored.
	// Default value is -1, which means unlimited size.
	MaxBodySize *int64 `json:"maxBodySize,omitempty"`
	// Mirrors defines the list of mirrors where Traefik will duplicate the traffic.
	Mirrors []MirrorService `json:"mirrors,omitempty"`
}

// +k8s:deepcopy-gen=true

// MirrorService holds the mirror configuration.
type MirrorService struct {
	LoadBalancerSpec `json:",inline"`

	// Percent defines the part of the traffic to mirror.
	// Supported values: 0 to 100.
	Percent int `json:"percent,omitempty"`
}

// +k8s:deepcopy-gen=true

// WeightedRoundRobin holds the weighted round-robin configuration.
// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#weighted-round-robin-service
type WeightedRoundRobin struct {
	// Services defines the list of Kubernetes Service and/or TraefikService to load-balance, with weight.
	Services []Service `json:"services,omitempty"`
	// Sticky defines whether sticky sessions are enabled.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#stickiness-and-load-balancing
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
}
