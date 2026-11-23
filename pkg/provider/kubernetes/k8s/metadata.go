package k8s

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

// NewMetadata returns the metadata of an object built for the resource of primary,
// from primary and the other given parts of Kubernetes resources.
func NewMetadata(primary dynamic.KubernetesRef, refs ...dynamic.KubernetesRef) *dynamic.Metadata {
	metadata := &dynamic.Metadata{Kubernetes: &dynamic.KubernetesMetadata{}}
	setSources(metadata.Kubernetes, primary, refs)

	return metadata
}

// NewServiceMetadata returns the metadata of a service built for the resource of primary,
// from primary and the other given parts of Kubernetes resources.
func NewServiceMetadata(primary dynamic.KubernetesRef, refs ...dynamic.KubernetesRef) *dynamic.ServiceMetadata {
	metadata := &dynamic.ServiceMetadata{Kubernetes: &dynamic.KubernetesServiceMetadata{}}
	setSources(&metadata.Kubernetes.KubernetesMetadata, primary, refs)

	return metadata
}

func setSources(metadata *dynamic.KubernetesMetadata, primary dynamic.KubernetesRef, refs []dynamic.KubernetesRef) {
	resource := primary
	resource.Path = ""

	metadata.Primary = &resource
	metadata.Sources.Add(primary)
	metadata.Sources.Add(refs...)
}
