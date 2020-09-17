package marathon

import (
	"math"

	"github.com/gambol99/go-marathon"
	"github.com/traefik/traefik/v2/pkg/config/label"
)

type configuration struct {
	Enable   bool
	Marathon specificConfiguration
}

type specificConfiguration struct {
	IPAddressIdx int
}

func (p *Provider) getConfiguration(app marathon.Application) (configuration, error) {
	labels := stringValueMap(app.Labels)

	conf := configuration{
		Enable: p.ExposedByDefault,
		Marathon: specificConfiguration{
			IPAddressIdx: math.MinInt32,
		},
	}

	err := label.Decode(labels, &conf, "traefik.marathon.", "traefik.enable")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}

func stringValueMap(mp *map[string]string) map[string]string {
	if mp != nil {
		return *mp
	}
	return make(map[string]string)
}
