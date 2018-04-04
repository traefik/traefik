package consulcatalog

import "github.com/containous/traefik/types"

func (p *Provider) buildConfiguration(catalog []catalogUpdate) *types.Configuration {
	if p.TemplateVersion == 1 {
		return p.buildConfigurationV1(catalog)
	}
	return p.buildConfigurationV2(catalog)
}
