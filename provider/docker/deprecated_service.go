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
// Deprecated
func (p Provider) getServiceFrontendRuleV1(container dockerData, serviceName string) string {
	if value, ok := getServiceLabelsV1(container, serviceName)[label.SuffixFrontendRule]; ok {
		return value
	}
	return p.getFrontendRuleV1(container)
}

// Check if for the given container, we find labels that are defining services
// Deprecated
func hasServicesV1(container dockerData) bool {
	return len(label.ExtractServiceProperties(container.Labels)) > 0
}

// Gets array of service names for a given container
// Deprecated
func getServiceNamesV1(container dockerData) []string {
	labelServiceProperties := label.ExtractServiceProperties(container.Labels)
	keys := make([]string, 0, len(labelServiceProperties))
	for k := range labelServiceProperties {
		keys = append(keys, k)
	}
	return keys
}

// checkServiceLabelPort checks if all service names have a port service label
// or if port container label exists for default value
// Deprecated
func checkServiceLabelPortV1(container dockerData) error {
	// If port container label is present, there is a default values for all ports, use it for the check
	_, err := strconv.Atoi(container.Labels[label.TraefikPort])
	if err != nil {
		serviceLabelPorts := make(map[string]struct{})
		serviceLabels := make(map[string]struct{})
		for lbl := range container.Labels {
			// Get all port service labels
			portLabel := extractServicePortV1(lbl)
			if len(portLabel) > 0 {
				serviceLabelPorts[portLabel[0]] = struct{}{}
			}
			// Get only one instance of all service names from service labels
			servicesLabelNames := label.FindSegmentSubmatch(lbl)

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

// Deprecated
func extractServicePortV1(labelName string) []string {
	if strings.HasPrefix(labelName, label.TraefikFrontend+".") ||
		strings.HasPrefix(labelName, label.TraefikBackend+".") {
		return nil
	}

	return label.PortRegexp.FindStringSubmatch(labelName)
}

// Extract backend from labels for a given service and a given docker container
// Deprecated
func getServiceBackendNameV1(container dockerData, serviceName string) string {
	if value, ok := getServiceLabelsV1(container, serviceName)[label.SuffixBackend]; ok {
		return provider.Normalize(container.ServiceName + "-" + value)
	}
	return provider.Normalize(container.ServiceName + "-" + getBackendNameV1(container) + "-" + serviceName)
}

// Extract port from labels for a given service and a given docker container
// Deprecated
func getServicePortV1(container dockerData, serviceName string) string {
	if value, ok := getServiceLabelsV1(container, serviceName)[label.SuffixPort]; ok {
		return value
	}
	return getPortV1(container)
}

// Service label functions

// Deprecated
func getFuncServiceSliceStringLabelV1(labelSuffix string) func(container dockerData, serviceName string) []string {
	return func(container dockerData, serviceName string) []string {
		serviceLabels := getServiceLabelsV1(container, serviceName)
		return getServiceSliceValueV1(container, serviceLabels, labelSuffix)
	}
}

// Deprecated
func getFuncServiceStringLabelV1(labelSuffix string, defaultValue string) func(container dockerData, serviceName string) string {
	return func(container dockerData, serviceName string) string {
		serviceLabels := getServiceLabelsV1(container, serviceName)
		return getServiceStringValueV1(container, serviceLabels, labelSuffix, defaultValue)
	}
}

// Deprecated
func getFuncServiceBoolLabelV1(labelSuffix string, defaultValue bool) func(container dockerData, serviceName string) bool {
	return func(container dockerData, serviceName string) bool {
		serviceLabels := getServiceLabelsV1(container, serviceName)
		return getServiceBoolValueV1(container, serviceLabels, labelSuffix, defaultValue)
	}
}

// Deprecated
func getFuncServiceIntLabelV1(labelSuffix string, defaultValue int) func(container dockerData, serviceName string) int {
	return func(container dockerData, serviceName string) int {
		return getServiceIntLabelV1(container, serviceName, labelSuffix, defaultValue)
	}
}

// Deprecated
func getServiceStringValueV1(container dockerData, serviceLabels map[string]string, labelSuffix string, defaultValue string) string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		return value
	}
	return label.GetStringValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

// Deprecated
func getServiceSliceValueV1(container dockerData, serviceLabels map[string]string, labelSuffix string) []string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		return label.SplitAndTrimString(value, ",")
	}
	return label.GetSliceStringValue(container.Labels, label.Prefix+labelSuffix)
}

// Deprecated
func getServiceBoolValueV1(container dockerData, serviceLabels map[string]string, labelSuffix string, defaultValue bool) bool {
	if rawValue, ok := serviceLabels[labelSuffix]; ok {
		value, err := strconv.ParseBool(rawValue)
		if err == nil {
			return value
		}
	}
	return label.GetBoolValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

// Deprecated
func getServiceIntLabelV1(container dockerData, serviceName string, labelSuffix string, defaultValue int) int {
	if rawValue, ok := getServiceLabelsV1(container, serviceName)[labelSuffix]; ok {
		value, err := strconv.Atoi(rawValue)
		if err == nil {
			return value
		}
	}
	return label.GetIntValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

// Deprecated
func getServiceLabelsV1(container dockerData, serviceName string) label.SegmentPropertyValues {
	return label.ExtractServiceProperties(container.Labels)[serviceName]
}

// Deprecated
func hasServiceRedirectV1(container dockerData, serviceName string) bool {
	serviceLabels, ok := label.ExtractServiceProperties(container.Labels)[serviceName]
	if !ok || len(serviceLabels) == 0 {
		return false
	}

	value, ok := serviceLabels[label.SuffixFrontendRedirectEntryPoint]
	frep := ok && len(value) > 0
	value, ok = serviceLabels[label.SuffixFrontendRedirectRegex]
	frrg := ok && len(value) > 0
	value, ok = serviceLabels[label.SuffixFrontendRedirectReplacement]
	frrp := ok && len(value) > 0

	return frep || frrg && frrp
}
