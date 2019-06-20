package rancher

import (
	"github.com/containous/traefik/pkg/config/label"
)

type configuration struct {
	Enable bool
	Tags   []string
}

func (p *Provider) getConfiguration(service rancherData) (configuration, error) {
	conf := configuration{
		Enable: p.ExposedByDefault,
	}

	err := label.Decode(service.Labels, &conf, "traefik.rancher.", "traefik.enable", "traefik.tags")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
