package docker

import (
	"errors"
	"strconv"
	"strings"

	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
)

// Specific functions

// Extract rule from labels for a given service and a given docker container
func (p Provider) getServiceFrontendRule(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixFrontendRule]; ok {
		return value
	}
	return p.getFrontendRule(container)
}

// Check if for the given container, we find labels that are defining services
func hasServices(container dockerData) bool {
	return len(label.ExtractServiceProperties(container.Labels)) > 0
}

// Gets array of service names for a given container
func getServiceNames(container dockerData) []string {
	labelServiceProperties := label.ExtractServiceProperties(container.Labels)
	keys := make([]string, 0, len(labelServiceProperties))
	for k := range labelServiceProperties {
		keys = append(keys, k)
	}
	return keys
}

// checkServiceLabelPort checks if all service names have a port service label
// or if port container label exists for default value
func checkServiceLabelPort(container dockerData) error {
	// If port container label is present, there is a default values for all ports, use it for the check
	_, err := strconv.Atoi(container.Labels[label.TraefikPort])
	if err != nil {
		serviceLabelPorts := make(map[string]struct{})
		serviceLabels := make(map[string]struct{})
		for lbl := range container.Labels {
			// Get all port service labels
			portLabel := label.PortRegexp.FindStringSubmatch(lbl)
			if len(portLabel) > 0 {
				serviceLabelPorts[portLabel[0]] = struct{}{}
			}
			// Get only one instance of all service names from service labels
			servicesLabelNames := label.ServicesPropertiesRegexp.FindStringSubmatch(lbl)
			if len(servicesLabelNames) > 0 {
				serviceLabels[strings.Split(servicesLabelNames[0], ".")[1]] = struct{}{}
			}
		}
		// If the number of service labels is different than the number of port services label
		// there is an error
		if len(serviceLabels) == len(serviceLabelPorts) {
			for labelPort := range serviceLabelPorts {
				_, err = strconv.Atoi(container.Labels[labelPort])
				if err != nil {
					break
				}
			}
		} else {
			err = errors.New("port service labels missing, please use traefik.port as default value or define all port service labels")
		}
	}
	return err
}

// Extract backend from labels for a given service and a given docker container
func getServiceBackend(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixFrontendBackend]; ok {
		return container.ServiceName + "-" + value
	}
	return strings.TrimPrefix(container.ServiceName, "/") + "-" + getBackend(container) + "-" + provider.Normalize(serviceName)
}

// Extract port from labels for a given service and a given docker container
func getServicePort(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixPort]; ok {
		return value
	}
	return getPort(container)
}

// Service label functions

func getFuncServiceMapLabel(labelSuffix string) func(container dockerData, serviceName string) map[string]string {
	return func(container dockerData, serviceName string) map[string]string {
		return getServiceMapLabel(container, serviceName, labelSuffix)
	}
}

func getFuncServiceSliceStringLabel(labelSuffix string) func(container dockerData, serviceName string) []string {
	return func(container dockerData, serviceName string) []string {
		return getServiceSliceStringLabel(container, serviceName, labelSuffix)
	}
}

func getFuncServiceStringLabel(labelSuffix string, defaultValue string) func(container dockerData, serviceName string) string {
	return func(container dockerData, serviceName string) string {
		return getServiceStringLabel(container, serviceName, labelSuffix, defaultValue)
	}
}

func hasFuncServiceLabel(labelSuffix string) func(container dockerData, serviceName string) bool {
	return func(container dockerData, serviceName string) bool {
		return hasServiceLabel(container, serviceName, labelSuffix)
	}
}

func hasServiceLabel(container dockerData, serviceName string, labelSuffix string) bool {
	value, ok := getServiceLabels(container, serviceName)[labelSuffix]
	if ok && len(value) > 0 {
		return true
	}
	return label.Has(container.Labels, label.Prefix+labelSuffix)
}

func getServiceMapLabel(container dockerData, serviceName string, labelSuffix string) map[string]string {
	if value, ok := getServiceLabels(container, serviceName)[labelSuffix]; ok {
		lblName := label.GetServiceLabel(labelSuffix, serviceName)
		return label.ParseMapValue(lblName, value)
	}
	return label.GetMapValue(container.Labels, label.Prefix+labelSuffix)
}

func getServiceSliceStringLabel(container dockerData, serviceName string, labelSuffix string) []string {
	if value, ok := getServiceLabels(container, serviceName)[labelSuffix]; ok {
		return label.SplitAndTrimString(value, ",")
	}
	return label.GetSliceStringValue(container.Labels, label.Prefix+labelSuffix)
}

func getServiceStringLabel(container dockerData, serviceName string, labelSuffix string, defaultValue string) string {
	if value, ok := getServiceLabels(container, serviceName)[labelSuffix]; ok {
		return value
	}
	return label.GetStringValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

func getServiceLabels(container dockerData, serviceName string) label.ServicePropertyValues {
	return label.ExtractServiceProperties(container.Labels)[serviceName]
}
