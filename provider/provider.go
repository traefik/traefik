package provider

import "github.com/emilevauge/traefik/types"

type Provider interface {
	Provide(configurationChan chan<- types.ConfigMessage) error
}
