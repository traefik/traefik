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

// NamespacedProvider is implemented by providers that support namespace-scoped configurations,
// where each configured namespace results in a dedicated provider instance.
// This enables clear identification of which namespace each provider instance serves during
// startup logging and operational monitoring.
type NamespacedProvider interface {
	Provider

	// Namespace returns the specific namespace this provider instance is configured for.
	Namespace() string
}
