package servicefabric

import "strings"

func hasServiceLabel(service ServiceItemExtended, key string) bool {
	_, exists := service.Labels[key]
	return exists
}

func getFuncBoolLabel(labelName string) func(service ServiceItemExtended) bool {
	return func(service ServiceItemExtended) bool {
		return getBoolLabel(service, labelName)
	}
}

func getBoolLabel(service ServiceItemExtended, labelName string) bool {
	value, exists := service.Labels[labelName]
	return exists && strings.EqualFold(strings.TrimSpace(value), "true")
}

func getServiceLabelValue(service ServiceItemExtended, key string) string {
	return service.Labels[key]
}
