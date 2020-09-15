package rancher

import (
	"github.com/traefik/traefik/v2/pkg/config/label"
)

type configuration struct {
	Enable bool
}

func (p *Provider) getConfiguration(service rancherData) (configuration, error) {
	conf := configuration{
		Enable: p.ExposedByDefault,
	}

	err := label.Decode(service.Labels, &conf, "traefik.rancher.", "traefik.enable")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
