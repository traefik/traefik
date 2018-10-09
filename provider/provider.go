package provider

import (
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- config.Message, pool *safe.Pool) error
	Init(constraints types.Constraints) error
}
