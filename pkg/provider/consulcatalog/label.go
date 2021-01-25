package consulcatalog

import (
	"github.com/traefik/traefik/v2/pkg/config/label"
)

// configuration Contains information from the labels that are globals (not related to the dynamic configuration) or specific to the provider.
type configuration struct {
	Enable  bool
	Connect bool
}

func (p *Provider) getConfiguration(item itemData) (configuration, error) {
	conf := configuration{
		Enable:  p.ExposedByDefault,
		Connect: p.ConnectByDefault,
	}

	err := label.Decode(item.Labels, &conf, "traefik.consulcatalog.", "traefik.enable")
	if err != nil {
		return configuration{}, err
	}
	err = label.Decode(item.Labels, &conf, "traefik.consulcatalog.", "traefik.connect")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
