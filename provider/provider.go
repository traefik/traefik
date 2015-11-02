package provider

import "github.com/emilevauge/traefik/types"

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to traefik
	// using the given configuration channel.
	Provide(configurationChan chan<- types.ConfigMessage) error
}
