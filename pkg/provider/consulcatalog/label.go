package consulcatalog

import (
	"github.com/traefik/traefik/v2/pkg/config/label"
)

// configuration contains information from the labels that are globals (not related to the dynamic configuration) or specific to the provider.
type configuration struct {
	Enable        bool
	ConsulCatalog specificConfiguration
}

type specificConfiguration struct {
	Connect bool // <prefix>.consulcatalog.connect is the corresponding label.
	Canary  bool // <prefix>.consulcatalog.canary is the corresponding label.
}

// getExtraConf returns a configuration with settings which are not part of the dynamic configuration (e.g. "<prefix>.enable").
func (p *Provider) getExtraConf(labels map[string]string) (configuration, error) {
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
