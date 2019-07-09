package provider

import (
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/safe"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error
	Init() error
}
