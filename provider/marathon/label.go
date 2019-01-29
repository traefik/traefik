package marathon

import (
	"math"
	"strings"

	"github.com/containous/traefik/provider/label"
	"github.com/gambol99/go-marathon"
)

type configuration struct {
	Enable   bool
	Tags     []string
	Marathon specificConfiguration
}

type specificConfiguration struct {
	IPAddressIdx int
}

func (p *Provider) getConfiguration(app marathon.Application) (configuration, error) {
	labels := stringValueMap(app.Labels)

	conf := configuration{
		Enable: p.ExposedByDefault,
		Tags:   nil,
		Marathon: specificConfiguration{
			IPAddressIdx: math.MinInt32,
		},
	}

	err := label.Decode(labels, &conf, "traefik.marathon.", "traefik.enable", "traefik.tags")
	if err != nil {
		return configuration{}, err
	}

	if p.FilterMarathonConstraints && app.Constraints != nil {
		for _, constraintParts := range *app.Constraints {
			conf.Tags = append(conf.Tags, strings.Join(constraintParts, ":"))
		}
	}

	return conf, nil
}

func stringValueMap(mp *map[string]string) map[string]string {
	if mp != nil {
		return *mp
	}
	return make(map[string]string)
}
