package reloadable

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

// Reloadable is an interface for providers that support reloading their configuration.
type Reloadable interface {
	ReloadConfig(configurationChan chan<- dynamic.Message) error
}
