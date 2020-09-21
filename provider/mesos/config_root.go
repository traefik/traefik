package mesos

import (
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/traefik/traefik/types"
)

func (p *Provider) buildConfiguration(tasks []state.Task) *types.Configuration {
	if p.TemplateVersion == 1 {
		return p.buildConfigurationV1(tasks)
	}
	return p.buildConfigurationV2(tasks)
}
