package docker

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/label"
)

const (
	labelDockerComposeProject = "com.docker.compose.project"
	labelDockerComposeService = "com.docker.compose.service"
)

// configuration contains information from the labels that are globals (not related to the dynamic configuration)
// or specific to the provider.
type configuration struct {
	Enable  bool
	Network string
	LBSwarm bool
}

type labelConfiguration struct {
	Enable bool
	Docker *specificConfiguration
	Swarm  *specificConfiguration
}

type specificConfiguration struct {
	Network *string
	LBSwarm bool
}

func (p *Shared) extractDockerLabels(container dockerData) (configuration, error) {
	conf := labelConfiguration{Enable: p.ExposedByDefault}
	if err := label.Decode(container.Labels, &conf, "traefik.docker.", "traefik.enable"); err != nil {
		return configuration{}, fmt.Errorf("decoding Docker labels: %w", err)
	}

	network := p.Network
	if conf.Docker != nil && conf.Docker.Network != nil {
		network = *conf.Docker.Network
	}

	return configuration{
		Enable:  conf.Enable,
		Network: network,
	}, nil
}

func (p *Shared) extractSwarmLabels(container dockerData) (configuration, error) {
	labelConf := labelConfiguration{Enable: p.ExposedByDefault}
	if err := label.Decode(container.Labels, &labelConf, "traefik.enable", "traefik.docker.", "traefik.swarm."); err != nil {
		return configuration{}, fmt.Errorf("decoding Swarm labels: %w", err)
	}

	if labelConf.Docker != nil && labelConf.Swarm != nil {
		return configuration{}, errors.New("both Docker and Swarm labels are defined")
	}

	conf := configuration{
		Enable:  labelConf.Enable,
		Network: p.Network,
	}

	if labelConf.Docker != nil {
		log.Warn().Msg("Labels traefik.docker.* for Swarm provider are deprecated. Please use traefik.swarm.* labels instead")

		conf.LBSwarm = labelConf.Docker.LBSwarm

		if labelConf.Docker.Network != nil {
			conf.Network = *labelConf.Docker.Network
		}
	}

	if labelConf.Swarm != nil {
		conf.LBSwarm = labelConf.Swarm.LBSwarm

		if labelConf.Swarm.Network != nil {
			conf.Network = *labelConf.Swarm.Network
		}
	}

	return conf, nil
}

// getStringMultipleStrict get multiple string values associated to several labels.
// Fail if one label is missing.
func getStringMultipleStrict(labels map[string]string, labelNames ...string) (map[string]string, error) {
	foundLabels := map[string]string{}
	for _, name := range labelNames {
		value := getStringValue(labels, name, "")
		// Error out only if one of them is not defined.
		if len(value) == 0 {
			return nil, fmt.Errorf("label not found: %s", name)
		}
		foundLabels[name] = value
	}
	return foundLabels, nil
}

// getStringValue get string value associated to a label.
func getStringValue(labels map[string]string, labelName, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}
