package consulcatalog

import (
	"github.com/traefik/traefik/v2/pkg/config/label"
)

// configuration Contains information from the labels that are globals (not related to the dynamic configuration) or specific to the provider.
type configuration struct {
	Enable        bool
	ConsulCatalog specificConfiguration
}

type specificConfiguration struct {
	Connect bool
}

func (p *Provider) getConfiguration(labels map[string]string) (configuration, error) {
	conf := configuration{
		Enable:        p.ExposedByDefault,
		ConsulCatalog: specificConfiguration{Connect: p.ConnectByDefault},
	}

	err := label.Decode(labels, &conf, "traefik.consulcatalog.", "traefik.enable")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
