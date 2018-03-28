package ecs

import (
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration(services map[string][]ecsInstance) (*types.Configuration, error) {
	if p.TemplateVersion == 1 {
		return p.buildConfigurationV1(services)
	}
	return p.buildConfigurationV2(services)
}
