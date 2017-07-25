// Package client provides a profilesvc client based on a predefined Consul
// service name and relevant tags. Users must only provide the address of a
// Consul server.
package client

import (
	"io"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/profilesvc"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

// New returns a service that's load-balanced over instances of profilesvc found
// in the provided Consul server. The mechanism of looking up profilesvc
// instances in Consul is hard-coded into the client.
func New(consulAddr string, logger log.Logger) (profilesvc.Service, error) {
	apiclient, err := consulapi.NewClient(&consulapi.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}

	// As the implementer of profilesvc, we declare and enforce these
	// parameters for all of the profilesvc consumers.
	var (
		consulService = "profilesvc"
		consulTags    = []string{"prod"}
		passingOnly   = true
		retryMax      = 3
		retryTimeout  = 500 * time.Millisecond
	)

	var (
		sdclient  = consul.NewClient(apiclient)
		endpoints profilesvc.Endpoints
	)
	{
		factory := factoryFor(profilesvc.MakePostProfileEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostProfileEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakeGetProfileEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetProfileEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakePutProfileEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PutProfileEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakePatchProfileEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PatchProfileEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakeDeleteProfileEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteProfileEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakeGetAddressesEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressesEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakeGetAddressEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakePostAddressEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostAddressEndpoint = retry
	}
	{
		factory := factoryFor(profilesvc.MakeDeleteAddressEndpoint)
		subscriber := consul.NewSubscriber(sdclient, factory, logger, consulService, consulTags, passingOnly)
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteAddressEndpoint = retry
	}

	return endpoints, nil
}

func factoryFor(makeEndpoint func(profilesvc.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		service, err := profilesvc.MakeClientEndpoints(instance)
		if err != nil {
			return nil, nil, err
		}
		return makeEndpoint(service), nil, nil
	}
}
