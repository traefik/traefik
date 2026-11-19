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
			desc: "no primary",
			metadata: newMetadata(nil,
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{},
		},
		{
			desc: "Ingress primary",
			metadata: newMetadata(
				&dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{
				KubernetesIngressNamespace: "default",
				KubernetesIngressName:      "ingress",
			},
		},
		{
			desc: "Ingress primary among several Ingresses",
			metadata: newMetadata(
				&dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "main"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "main", Path: "/spec/rules/0/http/paths/0"},
				dynamic.KubernetesRef{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "canary", Path: "/spec/rules/0/http/paths/0"},
			),
			expected: CoreLogData{
				KubernetesIngressNamespace: "default",
				KubernetesIngressName:      "main",
			},
		},
		{
			desc: "IngressRoute primary with a parent IngressRoute",
			metadata: newMetadata(
				&dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child"},
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child", Path: "/spec/routes/0"},
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "alpha", Path: "/spec/routes"},
			),
			expected: CoreLogData{
				KubernetesIngressRouteNamespace: "default",
				KubernetesIngressRouteName:      "child",
			},
		},
		{
			desc: "Ingress kind of another group",
			metadata: newMetadata(
				&dynamic.KubernetesRef{Group: "example.com", Version: "v1", Kind: "Ingress", Namespace: "default", Name: "ingress"},
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
			desc: "no primary",
			metadata: newServiceMetadata("80", nil,
				dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
			),
			expected: CoreLogData{},
		},
		{
			desc: "TraefikService primary",
			metadata: newServiceMetadata("",
				&dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr"},
				dynamic.KubernetesRef{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr", Path: "/spec"},
			),
			expected: CoreLogData{},
		},
		{
			desc: "Service primary",
			metadata: newServiceMetadata("80",
				&dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami"},
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
			desc: "Service primary with a fallback Service",
			metadata: newServiceMetadata("80",
				&dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami"},
				dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
				dynamic.KubernetesRef{Version: "v1", Kind: "Service", Namespace: "default", Name: "fallback", Path: "/spec/ports/0"},
			),
			expected: CoreLogData{
				KubernetesServiceNamespace: "default",
				KubernetesServiceName:      "whoami",
				KubernetesServicePort:      "80",
			},
		},
		{
			desc: "Service kind of another group",
			metadata: newServiceMetadata("80",
				&dynamic.KubernetesRef{Group: "serving.knative.dev", Version: "v1", Kind: "Service", Namespace: "default", Name: "knative"},
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

func newMetadata(primary *dynamic.KubernetesRef, refs ...dynamic.KubernetesRef) *dynamic.Metadata {
	metadata := &dynamic.Metadata{Kubernetes: &dynamic.KubernetesMetadata{Primary: primary}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}

func newServiceMetadata(port string, primary *dynamic.KubernetesRef, refs ...dynamic.KubernetesRef) *dynamic.ServiceMetadata {
	metadata := &dynamic.ServiceMetadata{Kubernetes: &dynamic.KubernetesServiceMetadata{Primary: primary, ServicePort: port}}
	metadata.Kubernetes.Sources.Add(refs...)

	return metadata
}
