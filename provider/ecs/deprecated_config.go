package ecs

import (
	"strconv"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
// Deprecated
func (p *Provider) buildConfigurationV1(instances []ecsInstance) (*types.Configuration, error) {
	instances = fun.Filter(p.filterInstance, instances).([]ecsInstance)

	services := make(map[string][]ecsInstance)
	for _, instance := range instances {
		if serviceInstances, ok := services[instance.Name]; ok {
			services[instance.Name] = append(serviceInstances, instance)
		} else {
			services[instance.Name] = []ecsInstance{instance}
		}
	}

	var ecsFuncMap = template.FuncMap{
		// Backend functions
		"getHost": getHost,
		"getPort": getPort,

		"getProtocol":             getFuncStringValueV1(label.TraefikProtocol, label.DefaultProtocol),
		"getWeight":               getFuncIntValueV1(label.TraefikWeight, label.DefaultWeightInt),
		"getLoadBalancerMethod":   getFuncFirstStringValueV1(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"getLoadBalancerSticky":   getStickyV1,
		"hasStickinessLabel":      getFuncFirstBoolValueV1(label.TraefikBackendLoadBalancerStickiness, false),
		"getStickinessCookieName": getFuncFirstStringValueV1(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		"hasHealthCheckLabels":    hasFuncFirstV1(label.TraefikBackendHealthCheckPath),
		"getHealthCheckPath":      getFuncFirstStringValueV1(label.TraefikBackendHealthCheckPath, ""),
		"getHealthCheckInterval":  getFuncFirstStringValueV1(label.TraefikBackendHealthCheckInterval, ""),

		// Frontend functions
		"filterFrontends":   filterFrontends,
		"getFrontendRule":   p.getFrontendRule,
		"getPassHostHeader": getFuncBoolValueV1(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    getFuncBoolValueV1(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":       getFuncIntValueV1(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getBasicAuth":      getFuncSliceStringV1(label.TraefikFrontendAuthBasic),
		"getEntryPoints":    getFuncSliceStringV1(label.TraefikFrontendEntryPoints),
	}

	return p.GetConfiguration("templates/ecs-v1.tmpl", ecsFuncMap, struct {
		Services map[string][]ecsInstance
	}{
		Services: services,
	})
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getStickyV1(instances []ecsInstance) bool {
	if hasFirstV1(instances, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return getFirstBoolValueV1(instances, label.TraefikBackendLoadBalancerSticky, false)
}

// Label functions

// Deprecated
func getFuncStringValueV1(labelName string, defaultValue string) func(i ecsInstance) string {
	return func(i ecsInstance) string {
		return getStringValueV1(i, labelName, defaultValue)
	}
}

// Deprecated
func getFuncBoolValueV1(labelName string, defaultValue bool) func(i ecsInstance) bool {
	return func(i ecsInstance) bool {
		return getBoolValueV1(i, labelName, defaultValue)
	}
}

// Deprecated
func getFuncIntValueV1(labelName string, defaultValue int) func(i ecsInstance) int {
	return func(i ecsInstance) int {
		return getIntValueV1(i, labelName, defaultValue)
	}
}

// Deprecated
func getFuncSliceStringV1(labelName string) func(i ecsInstance) []string {
	return func(i ecsInstance) []string {
		return getSliceStringV1(i, labelName)
	}
}

// Deprecated
func hasLabelV1(i ecsInstance, labelName string) bool {
	value, ok := i.containerDefinition.DockerLabels[labelName]
	return ok && value != nil && len(aws.StringValue(value)) > 0
}

// Deprecated
func getStringValueV1(i ecsInstance, labelName string, defaultValue string) string {
	if v, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if v == nil {
			return defaultValue
		}
		if len(aws.StringValue(v)) == 0 {
			return defaultValue
		}
		return aws.StringValue(v)
	}
	return defaultValue
}

// Deprecated
func getBoolValueV1(i ecsInstance, labelName string, defaultValue bool) bool {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.ParseBool(aws.StringValue(rawValue))
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

// Deprecated
func getIntValueV1(i ecsInstance, labelName string, defaultValue int) int {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.Atoi(aws.StringValue(rawValue))
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

// Deprecated
func getSliceStringV1(i ecsInstance, labelName string) []string {
	if value, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if value == nil {
			return nil
		}
		if len(aws.StringValue(value)) == 0 {
			return nil
		}
		return label.SplitAndTrimString(aws.StringValue(value), ",")
	}
	return nil
}

// Deprecated
func hasFuncFirstV1(labelName string) func(instances []ecsInstance) bool {
	return func(instances []ecsInstance) bool {
		return hasFirstV1(instances, labelName)
	}
}

// Deprecated
func getFuncFirstStringValueV1(labelName string, defaultValue string) func(instances []ecsInstance) string {
	return func(instances []ecsInstance) string {
		return getFirstStringValueV1(instances, labelName, defaultValue)
	}
}

// Deprecated
func getFuncFirstBoolValueV1(labelName string, defaultValue bool) func(instances []ecsInstance) bool {
	return func(instances []ecsInstance) bool {
		if len(instances) < 0 {
			return defaultValue
		}
		return getBoolValueV1(instances[0], labelName, defaultValue)
	}
}

// Deprecated
func hasFirstV1(instances []ecsInstance, labelName string) bool {
	if len(instances) == 0 {
		return false
	}
	return hasLabelV1(instances[0], labelName)
}

// Deprecated
func getFirstStringValueV1(instances []ecsInstance, labelName string, defaultValue string) string {
	if len(instances) == 0 {
		return defaultValue
	}
	return getStringValueV1(instances[0], labelName, defaultValue)
}

// Deprecated
func getFirstBoolValueV1(instances []ecsInstance, labelName string, defaultValue bool) bool {
	if len(instances) == 0 {
		return defaultValue
	}
	return getBoolValueV1(instances[0], labelName, defaultValue)
}
