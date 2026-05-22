package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

func TestBackendEndpointsCacheKey(t *testing.T) {
	backend := netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: "whoami",
			Port: netv1.ServiceBackendPort{Name: "http"},
		},
	}

	base := backendEndpointsCacheKey("testing", backend, IngressConfig{})

	// The same service and config must reuse the same cache entry.
	assert.Equal(t, base, backendEndpointsCacheKey("testing", backend, IngressConfig{}))

	// getBackendEndpoints branches on ServiceUpstream and DefaultBackend, so those must not
	// collide with the base entry for the same service.
	assert.NotEqual(t, base, backendEndpointsCacheKey("testing", backend, IngressConfig{ServiceUpstream: ptr.To(true)}))
	assert.NotEqual(t, base, backendEndpointsCacheKey("testing", backend, IngressConfig{DefaultBackend: ptr.To("fallback")}))

	// A different port of the same service is a distinct backend.
	otherPort := netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: "whoami",
			Port: netv1.ServiceBackendPort{Name: "https"},
		},
	}
	assert.NotEqual(t, base, backendEndpointsCacheKey("testing", otherPort, IngressConfig{}))

	// A backend without a Service cannot be keyed and must bypass the cache.
	assert.Empty(t, backendEndpointsCacheKey("testing", netv1.IngressBackend{}, IngressConfig{}))
}
