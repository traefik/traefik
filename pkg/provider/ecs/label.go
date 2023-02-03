package ecs

import (
	"github.com/traefik/traefik/v3/pkg/config/label"
)

// configuration Contains information from the labels that are globals (not related to the dynamic configuration) or specific to the provider.
type configuration struct {
	Enable bool
}

func (p *Provider) getConfiguration(instance ecsInstance) (configuration, error) {
	conf := configuration{
		Enable: p.ExposedByDefault,
	}

	err := label.Decode(instance.Labels, &conf, "traefik.ecs.", "traefik.enable")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
