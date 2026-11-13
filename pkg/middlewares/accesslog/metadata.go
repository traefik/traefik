package accesslog

import (
	"slices"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// addRouterMetadataFields adds the fields of the Kubernetes Ingress and IngressRoute that the router comes from.
func addRouterMetadataFields(core CoreLogData, metadata *dynamic.Metadata) {
	if metadata == nil || metadata.Kubernetes == nil {
		return
	}

	if ref, ok := firstKubernetesRef(metadata.Kubernetes.Sources, "networking.k8s.io", "Ingress"); ok {
		core[KubernetesIngressNamespace] = ref.Namespace
		core[KubernetesIngressName] = ref.Name
	} else if ref, ok = firstKubernetesRef(metadata.Kubernetes.Sources, "traefik.io", "IngressRoute"); ok {
		core[KubernetesIngressRouteNamespace] = ref.Namespace
		core[KubernetesIngressRouteName] = ref.Name
	}
}

// addServiceMetadataFields adds the fields of the Kubernetes Service that the service comes from.
func addServiceMetadataFields(core CoreLogData, metadata *dynamic.ServiceMetadata) {
	if metadata == nil || metadata.Kubernetes == nil {
		return
	}

	ref, ok := firstKubernetesRef(metadata.Kubernetes.Sources, "", "Service")
	if !ok {
		return
	}

	core[KubernetesServiceNamespace] = ref.Namespace
	core[KubernetesServiceName] = ref.Name
	core[KubernetesServicePort] = metadata.Kubernetes.ServicePort
}

// firstKubernetesRef returns the first source of the given group and kind.
// The sources are sorted, so the same source is returned for each request.
func firstKubernetesRef(sources dynamic.KubernetesSources, group, kind string) (dynamic.KubernetesRef, bool) {
	i := slices.IndexFunc(sources, func(ref dynamic.KubernetesRef) bool {
		return ref.Group == group && ref.Kind == kind
	})
	if i < 0 {
		return dynamic.KubernetesRef{}, false
	}

	return sources[i], true
}
