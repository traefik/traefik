package docker

import (
	"fmt"
	"strings"

	"github.com/containous/traefik/log"
)

// Gets the entry for a service label searching in all labels of the given container
func getContainerServiceLabel(container dockerData, serviceName string, entry string) (string, bool) {
	value, ok := extractServicesLabels(container.Labels)[serviceName][entry]
	return value, ok
}

// Extract the service labels from container labels of dockerData struct
func extractServicesLabels(labels map[string]string) labelServiceProperties {
	v := make(labelServiceProperties)

	for index, serviceProperty := range labels {
		matches := servicesPropertiesRegexp.FindStringSubmatch(index)
		if matches != nil {
			result := make(map[string]string)
			for i, name := range servicesPropertiesRegexp.SubexpNames() {
				if i != 0 {
					result[name] = matches[i]
				}
			}
			serviceName := result["service_name"]
			if _, ok := v[serviceName]; !ok {
				v[serviceName] = make(map[string]string)
			}
			v[serviceName][result["property_name"]] = serviceProperty
		}
	}

	return v
}

func hasLabel(label string) func(container dockerData) bool {
	return func(container dockerData) bool {
		lbl, err := getLabel(container, label)
		return err == nil && len(lbl) > 0
	}
}

func getLabel(container dockerData, label string) (string, error) {
	for key, value := range container.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", fmt.Errorf("label not found: %s", label)
}

func getLabels(container dockerData, labels []string) (map[string]string, error) {
	var globalErr error
	foundLabels := map[string]string{}
	for _, label := range labels {
		foundLabel, err := getLabel(container, label)
		// Error out only if one of them is defined.
		if err != nil {
			globalErr = fmt.Errorf("label not found: %s", label)
			continue
		}
		foundLabels[label] = foundLabel

	}
	return foundLabels, globalErr
}

func getBoolHeader(container dockerData, label string) bool {
	lbl, err := getLabel(container, label)
	return err == nil && len(lbl) > 0 && strings.EqualFold(strings.TrimSpace(lbl), "true")
}

func getSliceStringHeaders(container dockerData, containerType string) []string {
	value := []string{}
	if label, err := getLabel(container, containerType); err == nil {
		for _, sublabels := range strings.Split(label, ",") {
			if len(sublabels) == 0 {
				log.Warnf("Could not load header %v, skipping", sublabels)
			} else {
				value = append(value, sublabels)
			}
		}
	}
	if len(value) == 0 {
		log.Errorf("Could not load %v headers", containerType)
	}
	return value
}

func parseCustomHeaders(container dockerData, containerType string) map[string]string {
	customHeaders := make(map[string]string)
	if label, err := getLabel(container, containerType); err == nil {
		for _, headers := range strings.Split(label, ",") {
			pair := strings.Split(headers, ":")
			if len(pair) != 2 {
				log.Warnf("Could not load header %v, skipping...", pair)
			} else {
				customHeaders[pair[0]] = pair[1]
			}
		}
	}
	if len(customHeaders) == 0 {
		log.Errorf("Could not load any custom headers")
	}
	return customHeaders
}
