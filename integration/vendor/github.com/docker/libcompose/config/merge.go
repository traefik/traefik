package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/docker/docker/pkg/urlutil"
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

// Merge merges a compose file into an existing set of service configs
func Merge(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte, options *ParseOptions) (string, map[string]*ServiceConfig, map[string]*VolumeConfig, map[string]*NetworkConfig, error) {
	if options == nil {
		options = &defaultParseOptions
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return "", nil, nil, nil, err
	}

	var serviceConfigs map[string]*ServiceConfig
	var volumeConfigs map[string]*VolumeConfig
	var networkConfigs map[string]*NetworkConfig
	if config.Version == "2" {
		var err error
		serviceConfigs, err = MergeServicesV2(existingServices, environmentLookup, resourceLookup, file, bytes, options)
		if err != nil {
			return "", nil, nil, nil, err
		}
		volumeConfigs, err = ParseVolumes(bytes)
		if err != nil {
			return "", nil, nil, nil, err
		}
		networkConfigs, err = ParseNetworks(bytes)
		if err != nil {
			return "", nil, nil, nil, err
		}
	} else {
		serviceConfigsV1, err := MergeServicesV1(existingServices, environmentLookup, resourceLookup, file, bytes, options)
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

	return config.Version, serviceConfigs, volumeConfigs, networkConfigs, nil
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
	envFiles := serviceData["env_file"].([]interface{})
	if len(envFiles) == 0 {
		return serviceData, nil
	}

	if resourceLookup == nil {
		return nil, fmt.Errorf("Can not use env_file in file %s no mechanism provided to load files", inFile)
	}

	var vars []interface{}
	if _, ok := serviceData["environment"]; ok {
		vars = serviceData["environment"].([]interface{})
	}

	for i := len(envFiles) - 1; i >= 0; i-- {
		envFile := envFiles[i].(string)
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
			key := strings.SplitAfter(line, "=")[0]

			found := false
			for _, v := range vars {
				if strings.HasPrefix(v.(string), key) {
					found = true
					break
				}
			}

			if !found {
				vars = append(vars, line)
			}
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	serviceData["environment"] = vars

	delete(serviceData, "env_file")

	return serviceData, nil
}

func mergeConfig(baseService, serviceData RawService) RawService {
	for k, v := range serviceData {
		// Image and build are mutually exclusive in merge
		if k == "image" {
			delete(baseService, "build")
		} else if k == "build" {
			delete(baseService, "image")
		}
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
