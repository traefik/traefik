package mesos

import (
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records/state"
)

func (p *Provider) buildConfiguration(tasks []state.Task) *types.Configuration {
	if p.TemplateVersion == 1 {
		return p.buildConfigurationV1(tasks)
	}
	return p.buildConfigurationV2(tasks)
}
