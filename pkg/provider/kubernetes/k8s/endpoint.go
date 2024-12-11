package k8s

import (
	v1 "k8s.io/api/discovery/v1"
	"k8s.io/utils/ptr"
)

// EndpointServing returns true if the endpoint is still serving the service.
func EndpointServing(endpoint v1.Endpoint) bool {
	return ptr.Deref(endpoint.Conditions.Ready, false) || ptr.Deref(endpoint.Conditions.Serving, false)
}
