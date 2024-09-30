package k8s

import (
	"k8s.io/api/discovery/v1"
)

// EndpointServing returns true if the endpoint is still serving the service, regardless of its ready status.
func EndpointServing(endpoint v1.Endpoint) bool {
	return endpoint.Conditions.Ready != nil && (*endpoint.Conditions.Ready || (endpoint.Conditions.Serving != nil && *endpoint.Conditions.Serving))
}
