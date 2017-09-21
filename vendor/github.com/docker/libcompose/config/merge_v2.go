package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/utils"
)

// MergeServicesV2 merges a v2 compose file into an existing set of service configs
func MergeServicesV2(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, datas RawServiceMap, options *ParseOptions) (map[string]*ServiceConfig, error) {
	if options.Validate {
		if err := validateV2(datas); err != nil {
			return nil, err
		}
	}

	for name, data := range datas {
		data, err := parseV2(resourceLookup, environmentLookup, file, data, datas, options)
		if err != nil {
			logrus.Errorf("Failed to parse service %s: %v", name, err)
			return nil, err
		}

		if serviceConfig, ok := existingServices.Get(name); ok {
			var rawExistingService RawService
			if err := utils.Convert(serviceConfig, &rawExistingService); err != nil {
				return nil, err
			}

			data = mergeConfig(rawExistingService, data)
		}

		datas[name] = data
	}

	if options.Validate {
		var errs []string
		for name, data := range datas {
			err := validateServiceConstraintsv2(data, name)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) != 0 {
			return nil, fmt.Errorf(strings.Join(errs, "\n"))
		}
	}

	serviceConfigs := make(map[string]*ServiceConfig)
	if err := utils.Convert(datas, &serviceConfigs); err != nil {
		return nil, err
	}

	return serviceConfigs, nil
}

func parseV2(resourceLookup ResourceLookup, environmentLookup EnvironmentLookup, inFile string, serviceData RawService, datas RawServiceMap, options *ParseOptions) (RawService, error) {
	serviceData, err := readEnvFile(resourceLookup, inFile, serviceData)
	if err != nil {
		return nil, err
	}

	serviceData = resolveContextV2(inFile, serviceData)

	value, ok := serviceData["extends"]
	if !ok {
		return serviceData, nil
	}

	mapValue, ok := value.(map[interface{}]interface{})
	if !ok {
		return serviceData, nil
	}

	if resourceLookup == nil {
		return nil, fmt.Errorf("Can not use extends in file %s no mechanism provided to files", inFile)
	}

	file := asString(mapValue["file"])
	service := asString(mapValue["service"])

	if service == "" {
		return serviceData, nil
	}

	var baseService RawService

	if file == "" {
		if serviceData, ok := datas[service]; ok {
			baseService, err = parseV2(resourceLookup, environmentLookup, inFile, serviceData, datas, options)
		} else {
			return nil, fmt.Errorf("Failed to find service %s to extend", service)
		}
	} else {
		bytes, resolved, err := resourceLookup.Lookup(file, inFile)
		if err != nil {
			logrus.Errorf("Failed to lookup file %s: %v", file, err)
			return nil, err
		}

		config, err := CreateConfig(bytes)
		if err != nil {
			return nil, err
		}
		baseRawServices := config.Services

		if options.Interpolate {
			if err = InterpolateRawServiceMap(&baseRawServices, environmentLookup); err != nil {
				return nil, err
			}
		}

		if options.Validate {
			if err := validateV2(baseRawServices); err != nil {
				return nil, err
			}
		}

		baseService, ok = baseRawServices[service]
		if !ok {
			return nil, fmt.Errorf("Failed to find service %s in file %s", service, file)
		}

		baseService, err = parseV2(resourceLookup, environmentLookup, resolved, baseService, baseRawServices, options)
	}

	if err != nil {
		return nil, err
	}

	baseService = clone(baseService)

	logrus.Debugf("Merging %#v, %#v", baseService, serviceData)

	for _, k := range noMerge {
		if _, ok := baseService[k]; ok {
			source := file
			if source == "" {
				source = inFile
			}
			return nil, fmt.Errorf("Cannot extend service '%s' in %s: services with '%s' cannot be extended", service, source, k)
		}
	}

	baseService = mergeConfig(baseService, serviceData)

	logrus.Debugf("Merged result %#v", baseService)

	return baseService, nil
}

func resolveContextV2(inFile string, serviceData RawService) RawService {
	if _, ok := serviceData["build"]; !ok {
		return serviceData
	}
	var build map[interface{}]interface{}
	if buildAsString, ok := serviceData["build"].(string); ok {
		build = map[interface{}]interface{}{
			"context": buildAsString,
		}
	} else {
		build = serviceData["build"].(map[interface{}]interface{})
	}
	context := asString(build["context"])
	if context == "" {
		return serviceData
	}

	if IsValidRemote(context) {
		return serviceData
	}

	current := path.Dir(inFile)

	if context == "." {
		context = current
	} else {
		current = path.Join(current, context)
	}
	if _, ok := serviceData["build"].(string); ok {
		//build is specified as a string containing a path to the build context
		serviceData["build"] = current
	} else {
		//build is specified as an object with the path specified under context
		build["context"] = current
	}

	return serviceData
}
