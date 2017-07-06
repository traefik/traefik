package config

import (
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
)

// ConvertServices converts a set of v1 service configs to v2 service configs
func ConvertServices(v1Services map[string]*ServiceConfigV1) (map[string]*ServiceConfig, error) {
	v2Services := make(map[string]*ServiceConfig)
	replacementFields := make(map[string]*ServiceConfig)

	for name, service := range v1Services {
		replacementFields[name] = &ServiceConfig{
			Build: yaml.Build{
				Context:    service.Build,
				Dockerfile: service.Dockerfile,
			},
			Logging: Log{
				Driver:  service.LogDriver,
				Options: service.LogOpt,
			},
			NetworkMode: service.Net,
		}

		v1Services[name].Build = ""
		v1Services[name].Dockerfile = ""
		v1Services[name].LogDriver = ""
		v1Services[name].LogOpt = nil
		v1Services[name].Net = ""
	}

	if err := utils.Convert(v1Services, &v2Services); err != nil {
		return nil, err
	}

	for name := range v2Services {
		v2Services[name].Build = replacementFields[name].Build
		v2Services[name].Logging = replacementFields[name].Logging
		v2Services[name].NetworkMode = replacementFields[name].NetworkMode
	}

	return v2Services, nil
}
