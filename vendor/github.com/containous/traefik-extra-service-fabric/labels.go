package servicefabric

import (
	"strconv"
	"strings"
)

func getFuncBoolLabel(labelName string, defaultValue bool) func(service ServiceItemExtended) bool {
	return func(service ServiceItemExtended) bool {
		return getBoolValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncServiceStringLabel(service ServiceItemExtended, labelName string, defaultValue string) string {
	return getStringValue(service.Labels, labelName, defaultValue)
}

func hasFuncService(service ServiceItemExtended, labelName string) bool {
	return hasLabel(service.Labels, labelName)
}

func getServiceLabelsWithPrefix(service ServiceItemExtended, prefix string) map[string]string {
	results := make(map[string]string)
	for k, v := range service.Labels {
		if strings.HasPrefix(k, prefix) {
			results[k] = v
		}
	}
	return results
}

// must be replace by label.Has()
// Deprecated
func hasLabel(labels map[string]string, labelName string) bool {
	value, ok := labels[labelName]
	return ok && len(value) > 0
}

// must be replace by label.GetStringValue()
// Deprecated
func getStringValue(labels map[string]string, labelName string, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}

// must be replace by label.GetBoolValue()
// Deprecated
func getBoolValue(labels map[string]string, labelName string, defaultValue bool) bool {
	rawValue, ok := labels[labelName]
	if ok {
		v, err := strconv.ParseBool(rawValue)
		if err == nil {
			return v
		}
	}
	return defaultValue
}
