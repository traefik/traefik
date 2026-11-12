package dynamic

import (
	"cmp"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +k8s:deepcopy-gen=true

// Metadata stores provider-specific runtime object metadata.
type Metadata struct {
	Kubernetes *KubernetesMetadata `json:"kubernetes,omitempty"`
}

// Merge adds the Kubernetes sources of other.
func (m *Metadata) Merge(other *Metadata) {
	if m.Kubernetes == nil || other.Kubernetes == nil {
		return
	}

	m.Kubernetes.Sources.Add(other.Kubernetes.Sources...)
}

// +k8s:deepcopy-gen=true

type ServiceMetadata struct {
	Kubernetes *KubernetesServiceMetadata `json:"kubernetes,omitempty"`
}

// Merge adds the Kubernetes sources of other.
func (m *ServiceMetadata) Merge(other *ServiceMetadata) {
	if m.Kubernetes == nil || other.Kubernetes == nil {
		return
	}

	m.Kubernetes.Sources.Add(other.Kubernetes.Sources...)
}

// +k8s:deepcopy-gen=true

type KubernetesMetadata struct {
	// Sources are the parts of the Kubernetes resources that the provider reads
	// to build the object.
	Sources KubernetesSources `json:"sources"`
}

// +k8s:deepcopy-gen=true

type KubernetesServiceMetadata struct {
	KubernetesMetadata

	ServicePort string `json:"servicePort,omitempty"`
}

// KubernetesSources are the parts of the Kubernetes resources that a provider reads to build an object.
type KubernetesSources []KubernetesRef

// +k8s:deepcopy-gen=true

// KubernetesRef is a part of a Kubernetes resource.
type KubernetesRef struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version"`
	Kind    string `json:"kind"`

	ResourceVersion string `json:"resourceVersion"`

	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`

	// JSON Pointer, empty for the whole resource.
	Path string `json:"path,omitempty"`
}

// MakeKubernetesRef returns the reference to the part of v at path.
func MakeKubernetesRef(v client.Object, path string) KubernetesRef {
	gvk := v.GetObjectKind().GroupVersionKind()

	return KubernetesRef{
		Group:           gvk.Group,
		Version:         gvk.Version,
		Kind:            gvk.Kind,
		ResourceVersion: v.GetResourceVersion(),
		Namespace:       v.GetNamespace(),
		Name:            v.GetName(),
		Path:            path,
	}
}

// Add records the given parts. The sources stay sorted, with no duplicates,
// whatever their current order, as configurations are compared with reflect.DeepEqual.
// We expect a small number of items in KubernetesSources.
func (s *KubernetesSources) Add(refs ...KubernetesRef) {
	*s = slices.Concat(*s, refs)
	slices.SortFunc(*s, compareKubernetesRefs)
	*s = slices.CompactFunc(*s, func(a, b KubernetesRef) bool {
		return compareKubernetesRefs(a, b) == 0
	})
}

func compareKubernetesRefs(a, b KubernetesRef) int {
	return cmp.Or(
		cmp.Compare(a.Group, b.Group),
		cmp.Compare(a.Version, b.Version),
		cmp.Compare(a.Kind, b.Kind),
		cmp.Compare(a.Namespace, b.Namespace),
		cmp.Compare(a.Name, b.Name),
		cmp.Compare(a.Path, b.Path),
	)
}
