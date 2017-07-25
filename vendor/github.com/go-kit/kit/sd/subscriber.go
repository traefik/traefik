package sd

import "github.com/go-kit/kit/endpoint"

// Subscriber listens to a service discovery system and yields a set of
// identical endpoints on demand. An error indicates a problem with connectivity
// to the service discovery system, or within the system itself; a subscriber
// may yield no endpoints without error.
type Subscriber interface {
	Endpoints() ([]endpoint.Endpoint, error)
}
