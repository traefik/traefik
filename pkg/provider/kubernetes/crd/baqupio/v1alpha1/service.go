package v1alpha1

import (
	"github.com/baqupio/baqup/v3/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// BaqupService is the CRD implementation of a Baqup Service.
// BaqupService object allows to:
// - Apply weight to Services on load-balancing
// - Mirror traffic on services
// More info: https://doc.baqup.io/baqup/v3.6/reference/routing-configuration/kubernetes/crd/http/baqupservice/
type BaqupService struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec BaqupServiceSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BaqupServiceList is a collection of BaqupService resources.
type BaqupServiceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of BaqupService.
	Items []BaqupService `json:"items"`
}

// +k8s:deepcopy-gen=true

// BaqupServiceSpec defines the desired state of a BaqupService.
type BaqupServiceSpec struct {
	// Weighted defines the Weighted Round Robin configuration.
	Weighted *WeightedRoundRobin `json:"weighted,omitempty"`
	// Mirroring defines the Mirroring service configuration.
	Mirroring *Mirroring `json:"mirroring,omitempty"`
	// HighestRandomWeight defines the highest random weight service configuration.
	HighestRandomWeight *HighestRandomWeight `json:"highestRandomWeight,omitempty"`
}

// +k8s:deepcopy-gen=true

// Mirroring holds the mirroring service configuration.
// More info: https://doc.baqup.io/baqup/v3.6/reference/routing-configuration/http/load-balancing/service/#mirroring
type Mirroring struct {
	LoadBalancerSpec `json:",inline"`

	// MirrorBody defines whether the body of the request should be mirrored.
	// Default value is true.
	MirrorBody *bool `json:"mirrorBody,omitempty"`
	// MaxBodySize defines the maximum size allowed for the body of the request.
	// If the body is larger, the request is not mirrored.
	// Default value is -1, which means unlimited size.
	MaxBodySize *int64 `json:"maxBodySize,omitempty"`
	// Mirrors defines the list of mirrors where Baqup will duplicate the traffic.
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
// More info: https://doc.baqup.io/baqup/v3.6/reference/routing-configuration/http/load-balancing/service/#weighted-round-robin-wrr
type WeightedRoundRobin struct {
	// Services defines the list of Kubernetes Service and/or BaqupService to load-balance, with weight.
	Services []Service `json:"services,omitempty"`
	// Sticky defines whether sticky sessions are enabled.
	// More info: https://doc.baqup.io/baqup/v3.6/reference/routing-configuration/kubernetes/crd/http/baqupservice/#stickiness-and-load-balancing
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
}

// +k8s:deepcopy-gen=true

// HighestRandomWeight holds the highest random weight configuration.
// More info: https://doc.baqup.io/baqup/v3.6/routing/services/#highest-random-configuration
type HighestRandomWeight struct {
	// Services defines the list of Kubernetes Service and/or BaqupService to load-balance, with weight.
	Services []Service `json:"services,omitempty"`
}
