// Package thrift provides a Thrift client for the add service.
package thrift

import (
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/addsvc"
	"github.com/go-kit/kit/ratelimit"
)

// New returns an AddService backed by a Thrift server described by the provided
// client. The caller is responsible for constructing the client, and eventually
// closing the underlying transport.
func New(client *thriftadd.AddServiceClient) addsvc.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.

	limiter := ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(100, 100))

	// Thrift does not currently have tracer bindings, so we skip tracing.

	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = addsvc.MakeThriftSumEndpoint(client)
		sumEndpoint = limiter(sumEndpoint)
		sumEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Sum",
			Timeout: 30 * time.Second,
		}))(sumEndpoint)
	}

	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = addsvc.MakeThriftConcatEndpoint(client)
		concatEndpoint = limiter(concatEndpoint)
		concatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Concat",
			Timeout: 30 * time.Second,
		}))(concatEndpoint)
	}

	return addsvc.Endpoints{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}
