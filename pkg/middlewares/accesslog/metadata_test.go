package accesslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestAddRouterMetadataFields(t *testing.T) {
	testCases := []struct {
		desc     string
		metadata *dynamic.Metadata
		expected CoreLogData
	}{
		{
			desc:     "nil metadata",
			expected: CoreLogData{},
		},
		{
			desc:     "no Kubernetes metadata",
			metadata: &dynamic.Metadata{},
			expected: CoreLogData{},
		},
		{
			desc: "no matching source",
			metadata: newMetadata(
				dynamic.KubernetesRef{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute", Namespace: "default", Name: "route", Path: "/spec/rules/0"},
				dynamic.KubernetesRef{Version: "v1", Kind: "ConfigMap", Namespace: "default", Name: "config", Path: "/data"},
			),
			expected: CoreLogData{},
		},
		{
			desc: "Ingress",
			metadata: newMetadata(
				dynamic.KubernetesRef{Version: "v1", Kind: "ConfigMap", Namespace: "default", Name: "config", Path: "/data"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{
				KubernetesIngressNamespace: "default",
				KubernetesIngressName:      "ingress",
			},
		},
		{
			desc: "first of several Ingresses",
			metadata: newMetadata(
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "main", Path: "/spec/rules/0/http/paths/0"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "canary", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{
				KubernetesIngressNamespace: "default",
				KubernetesIngressName:      "canary",
			},
		},
		{
			desc: "IngressRoute",
			metadata: newMetadata(
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "route", Path: "/spec/entryPoints"},
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "route", Path: "/spec/routes/0"},
			),
			expected: CoreLogData{
				KubernetesIngressRouteNamespace: "default",
				KubernetesIngressRouteName:      "route",
			},
		},
		{
			desc: "Ingress before IngressRoute",
			metadata: newMetadata(
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "route", Path: "/spec/routes/0"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{
				KubernetesIngressNamespace: "default",
				KubernetesIngressName:      "ingress",
			},
		},
		{
			desc: "Ingress kind of another group",
			metadata: newMetadata(
				dynamic.KubernetesRef{Group: "example.com", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress"},
			),
			expected: CoreLogData{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			core := CoreLogData{}
			addRouterMetadataFields(core, test.metadata)

			assert.Equal(t, test.expected, core)
		})
	}
}

func TestAddServiceMetadataFields(t *testing.T) {
	testCases := []struct {
		desc     string
		metadata *dynamic.ServiceMetadata
		expected CoreLogData
	}{
		{
			desc:     "nil metadata",
			expected: CoreLogData{},
		},
		{
			desc:     "no Kubernetes metadata",
			metadata: &dynamic.ServiceMetadata{},
			expected: CoreLogData{},
		},
		{
			desc: "no Service source",
			metadata: newServiceMetadata("",
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr", Path: "/spec"},
			),
			expected: CoreLogData{},
		},
		{
			desc: "Service",
			metadata: newServiceMetadata("80",
				dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress", Path: "/spec/rules/0/http/paths/0/backend"},
			),
			expected: CoreLogData{
				KubernetesServiceNamespace: "default",
				KubernetesServiceName:      "whoami",
				KubernetesServicePort:      "80",
			},
		},
		{
			desc: "ExternalName Service",
			metadata: newServiceMetadata("http",
				dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "external", Path: "/spec/externalName"},
			),
			expected: CoreLogData{
				KubernetesServiceNamespace: "default",
				KubernetesServiceName:      "external",
				KubernetesServicePort:      "http",
			},
		},
		{
			desc: "Service kind of another group",
			metadata: newServiceMetadata("80",
				dynamic.KubernetesRef{Group: "serving.knative.dev", Version: "v1", Kind: "Service", Namespace: "default", Name: "knative"},
			),
			expected: CoreLogData{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			core := CoreLogData{}
			addServiceMetadataFields(core, test.metadata)

			assert.Equal(t, test.expected, core)
		})
	}
}

func newMetadata(refs ...dynamic.KubernetesRef) *dynamic.Metadata {
	metadata := &dynamic.Metadata{Kubernetes: &dynamic.KubernetesMetadata{}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}

func newServiceMetadata(port string, refs ...dynamic.KubernetesRef) *dynamic.ServiceMetadata {
	metadata := &dynamic.ServiceMetadata{Kubernetes: &dynamic.KubernetesServiceMetadata{ServicePort: port}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}
