package docker

import (
	"fmt"

	"github.com/traefik/traefik/v2/pkg/config/label"
)

const (
	labelDockerComposeProject = "com.docker.compose.project"
	labelDockerComposeService = "com.docker.compose.service"
)

// configuration Contains information from the labels that are globals (not related to the dynamic configuration)
// or specific to the provider.
type configuration struct {
	Enable bool
	Docker specificConfiguration
}

type specificConfiguration struct {
	Network string
	LBSwarm bool
}

func (p *Provider) getConfiguration(container dockerData) (configuration, error) {
	conf := configuration{
		Enable: p.ExposedByDefault,
		Docker: specificConfiguration{
			Network: p.Network,
		},
	}

	err := label.Decode(container.Labels, &conf, "traefik.docker.", "traefik.enable")
	if err != nil {
		return configuration{}, err
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
