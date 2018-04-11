package servicefabric

import (
	"strings"

	"github.com/containous/traefik/provider/label"
)

// SF Specific Traefik Labels
const (
	traefikSFGroupName                   = "traefik.servicefabric.groupname"
	traefikSFGroupWeight                 = "traefik.servicefabric.groupweight"
	traefikSFEnableLabelOverrides        = "traefik.servicefabric.enablelabeloverrides"
	traefikSFEnableLabelOverridesDefault = true
)

func getFuncBoolLabel(labelName string, defaultValue bool) func(service ServiceItemExtended) bool {
	return func(service ServiceItemExtended) bool {
		return label.GetBoolValue(service.Labels, labelName, defaultValue)
	}
}

func getServiceStringLabel(service ServiceItemExtended, labelName string, defaultValue string) string {
	return label.GetStringValue(service.Labels, labelName, defaultValue)
}

func getFuncServiceStringLabel(labelName string, defaultValue string) func(service ServiceItemExtended) string {
	return func(service ServiceItemExtended) string {
		return label.GetStringValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncServiceIntLabel(labelName string, defaultValue int) func(service ServiceItemExtended) int {
	return func(service ServiceItemExtended) int {
		return label.GetIntValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncServiceBoolLabel(labelName string, defaultValue bool) func(service ServiceItemExtended) bool {
	return func(service ServiceItemExtended) bool {
		return label.GetBoolValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncServiceSliceStringLabel(labelName string) func(service ServiceItemExtended) []string {
	return func(service ServiceItemExtended) []string {
		return label.GetSliceStringValue(service.Labels, labelName)
	}
}

func hasService(service ServiceItemExtended, labelName string) bool {
	return label.Has(service.Labels, labelName)
}

func getFuncServiceLabelWithPrefix(labelName string) func(service ServiceItemExtended) map[string]string {
	return func(service ServiceItemExtended) map[string]string {
		return getServiceLabelsWithPrefix(service, labelName)
	}
}

func getFuncServicesGroupedByLabel(labelName string) func(services []ServiceItemExtended) map[string][]ServiceItemExtended {
	return func(services []ServiceItemExtended) map[string][]ServiceItemExtended {
		return getServices(services, labelName)
	}
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
