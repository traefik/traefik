package accesslog

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

// addRouterMetadataFields adds the fields of the Kubernetes Ingress and IngressRoute that the router comes from.
func addRouterMetadataFields(core CoreLogData, metadata *dynamic.Metadata) {
	if metadata == nil || metadata.Kubernetes == nil || metadata.Kubernetes.Primary == nil {
		return
	}

	primary := metadata.Kubernetes.Primary
	switch {
	case primary.Group == "networking.k8s.io" && primary.Kind == "Ingress":
		core[KubernetesIngressNamespace] = primary.Namespace
		core[KubernetesIngressName] = primary.Name
	case primary.Group == "traefik.io" && primary.Kind == "IngressRoute":
		core[KubernetesIngressRouteNamespace] = primary.Namespace
		core[KubernetesIngressRouteName] = primary.Name
	}
}

// addServiceMetadataFields adds the fields of the Kubernetes Service that the service comes from.
func addServiceMetadataFields(core CoreLogData, metadata *dynamic.ServiceMetadata) {
	if metadata == nil || metadata.Kubernetes == nil || metadata.Kubernetes.Primary == nil {
		return
	}

	primary := metadata.Kubernetes.Primary
	if primary.Group != "" || primary.Kind != "Service" {
		return
	}

	core[KubernetesServiceNamespace] = primary.Namespace
	core[KubernetesServiceName] = primary.Name
	core[KubernetesServicePort] = metadata.Kubernetes.ServicePort
}
