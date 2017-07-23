package client

import (
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/project"
)

// Factory is a factory to create docker clients.
type Factory interface {
	// Create constructs a Docker client for the given service. The passed in
	// config may be nil in which case a generic client for the project should
	// be returned.
	Create(service project.Service) client.APIClient
}

type defaultFactory struct {
	client client.APIClient
}

// NewDefaultFactory creates and returns the default client factory that uses
// github.com/docker/docker client.
func NewDefaultFactory(opts Options) (Factory, error) {
	client, err := Create(opts)
	if err != nil {
		return nil, err
	}

	return &defaultFactory{
		client: client,
	}, nil
}

func (s *defaultFactory) Create(service project.Service) client.APIClient {
	return s.client
}
