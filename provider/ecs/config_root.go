package ecs

import (
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration(instances []ecsInstance) (*types.Configuration, error) {
	if p.TemplateVersion == 1 {
		return p.buildConfigurationV1(instances)
	}
	return p.buildConfigurationV2(instances)
}
