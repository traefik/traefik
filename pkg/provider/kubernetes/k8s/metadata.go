package k8s

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

// NewMetadata returns the metadata of an object built from the given parts of Kubernetes resources.
func NewMetadata(refs ...dynamic.KubernetesRef) *dynamic.Metadata {
	metadata := &dynamic.Metadata{Kubernetes: &dynamic.KubernetesMetadata{}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}

// NewServiceMetadata returns the metadata of a service built from the given parts of Kubernetes resources.
func NewServiceMetadata(refs ...dynamic.KubernetesRef) *dynamic.ServiceMetadata {
	metadata := &dynamic.ServiceMetadata{Kubernetes: &dynamic.KubernetesServiceMetadata{}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}
