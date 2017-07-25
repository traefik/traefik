package sd

import "github.com/go-kit/kit/endpoint"

// FixedSubscriber yields a fixed set of services.
type FixedSubscriber []endpoint.Endpoint

// Endpoints implements Subscriber.
func (s FixedSubscriber) Endpoints() ([]endpoint.Endpoint, error) { return s, nil }
