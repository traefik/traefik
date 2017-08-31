package project

import (
	"github.com/docker/libcompose/config"
	"gopkg.in/yaml.v2"
)

// ExportedConfig holds config attribute that will be exported
type ExportedConfig struct {
	Version  string                           `yaml:"version,omitempty"`
	Services map[string]*config.ServiceConfig `yaml:"services"`
	Volumes  map[string]*config.VolumeConfig  `yaml:"volumes"`
	Networks map[string]*config.NetworkConfig `yaml:"networks"`
}

// Config validates and print the compose file.
func (p *Project) Config() (string, error) {
	cfg := ExportedConfig{
		Version:  "2.0",
		Services: p.ServiceConfigs.All(),
		Volumes:  p.VolumeConfigs,
		Networks: p.NetworkConfigs,
	}

	bytes, err := yaml.Marshal(cfg)
	return string(bytes), err
}
