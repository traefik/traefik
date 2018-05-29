package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"reflect"

	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/libcompose/utils"
	composeYaml "github.com/docker/libcompose/yaml"
	"gopkg.in/yaml.v2"
)

var (
	noMerge = []string{
		"links",
		"volumes_from",
	}
	defaultParseOptions = ParseOptions{
		Interpolate: true,
		Validate:    true,
	}
)

// CreateConfig unmarshals bytes to config and creates config based on version
func CreateConfig(bytes []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	if config.Version != "2" {
		var baseRawServices RawServiceMap
		if err := yaml.Unmarshal(bytes, &baseRawServices); err != nil {
			return nil, err
		}
		config.Services = baseRawServices
	}

	if config.Volumes == nil {
		config.Volumes = make(map[string]interface{})
	}
	if config.Networks == nil {
		config.Networks = make(map[string]interface{})
	}

	return &config, nil
}

// Merge merges a compose file into an existing set of service configs
func Merge(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte, options *ParseOptions) (string, map[string]*ServiceConfig, map[string]*VolumeConfig, map[string]*NetworkConfig, error) {
	if options == nil {
		options = &defaultParseOptions
	}

	config, err := CreateConfig(bytes)
	if err != nil {
		return "", nil, nil, nil, err
	}
	baseRawServices := config.Services

	for service, data := range baseRawServices {
		for key, value := range data {
			//check for "extends" key and check whether it is string or not
			if key == "extends" && reflect.TypeOf(value).Kind() == reflect.String {
				//converting string to map
				extendMap := make(map[interface{}]interface{})
				extendMap["service"] = value
				baseRawServices[service][key] = extendMap
			}
		}
	}

	if options.Interpolate {
		if err := InterpolateRawServiceMap(&baseRawServices, environmentLookup); err != nil {
			return "", nil, nil, nil, err
		}

		for k, v := range config.Volumes {
			if err := Interpolate(k, &v, environmentLookup); err != nil {
				return "", nil, nil, nil, err
			}
			config.Volumes[k] = v
		}

		for k, v := range config.Networks {
			if err := Interpolate(k, &v, environmentLookup); err != nil {
				return "", nil, nil, nil, err
			}
			config.Networks[k] = v
		}
	}

	if options.Preprocess != nil {
		var err error
		baseRawServices, err = options.Preprocess(baseRawServices)
		if err != nil {
			return "", nil, nil, nil, err
		}
	}

	var serviceConfigs map[string]*ServiceConfig
	if config.Version == "2" {
		var err error
		serviceConfigs, err = MergeServicesV2(existingServices, environmentLookup, resourceLookup, file, baseRawServices, options)
		if err != nil {
			return "", nil, nil, nil, err
		}
	} else {
		serviceConfigsV1, err := MergeServicesV1(existingServices, environmentLookup, resourceLookup, file, baseRawServices, options)
		if err != nil {
			return "", nil, nil, nil, err
		}
		serviceConfigs, err = ConvertServices(serviceConfigsV1)
		if err != nil {
			return "", nil, nil, nil, err
		}
	}

	adjustValues(serviceConfigs)

	if options.Postprocess != nil {
		var err error
		serviceConfigs, err = options.Postprocess(serviceConfigs)
		if err != nil {
			return "", nil, nil, nil, err
		}
	}

	var volumes map[string]*VolumeConfig
	var networks map[string]*NetworkConfig
	if err := utils.Convert(config.Volumes, &volumes); err != nil {
		return "", nil, nil, nil, err
	}
	if err := utils.Convert(config.Networks, &networks); err != nil {
		return "", nil, nil, nil, err
	}

	return config.Version, serviceConfigs, volumes, networks, nil
}

// InterpolateRawServiceMap replaces varialbse in raw service map struct based on environment lookup
func InterpolateRawServiceMap(baseRawServices *RawServiceMap, environmentLookup EnvironmentLookup) error {
	for k, v := range *baseRawServices {
		for k2, v2 := range v {
			if err := Interpolate(k2, &v2, environmentLookup); err != nil {
				return err
			}
			(*baseRawServices)[k][k2] = v2
		}
	}
	return nil
}

func adjustValues(configs map[string]*ServiceConfig) {
	// yaml parser turns "no" into "false" but that is not valid for a restart policy
	for _, v := range configs {
		if v.Restart == "false" {
			v.Restart = "no"
		}
	}
}

func readEnvFile(resourceLookup ResourceLookup, inFile string, serviceData RawService) (RawService, error) {
	if _, ok := serviceData["env_file"]; !ok {
		return serviceData, nil
	}

	var envFiles composeYaml.Stringorslice

	if err := utils.Convert(serviceData["env_file"], &envFiles); err != nil {
		return nil, err
	}

	if len(envFiles) == 0 {
		return serviceData, nil
	}

	if resourceLookup == nil {
		return nil, fmt.Errorf("Can not use env_file in file %s no mechanism provided to load files", inFile)
	}

	var vars composeYaml.MaporEqualSlice

	if _, ok := serviceData["environment"]; ok {
		if err := utils.Convert(serviceData["environment"], &vars); err != nil {
			return nil, err
		}
	}

	for i := len(envFiles) - 1; i >= 0; i-- {
		envFile := envFiles[i]
		content, _, err := resourceLookup.Lookup(envFile, inFile)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(content))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if len(line) > 0 && !strings.HasPrefix(line, "#") {
				key := strings.SplitAfter(line, "=")[0]

				found := false
				for _, v := range vars {
					if strings.HasPrefix(v, key) {
						found = true
						break
					}
				}

				if !found {
					vars = append(vars, line)
				}
			}
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	serviceData["environment"] = vars

	return serviceData, nil
}

func mergeConfig(baseService, serviceData RawService) RawService {
	for k, v := range serviceData {
		existing, ok := baseService[k]
		if ok {
			baseService[k] = merge(existing, v)
		} else {
			baseService[k] = v
		}
	}

	return baseService
}

// IsValidRemote checks if the specified string is a valid remote (for builds)
func IsValidRemote(remote string) bool {
	return urlutil.IsGitURL(remote) || urlutil.IsURL(remote)
}
