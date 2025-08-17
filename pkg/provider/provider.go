package provider

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error
	Init() error
}

// NamespaceProvider defines methods for providers that have namespace information.
type NamespaceProvider interface {
	Provider
	// GetNamespace returns the namespace of the provider, if any.
	GetNamespace() string
}
