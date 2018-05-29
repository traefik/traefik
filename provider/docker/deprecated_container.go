package docker

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/docker/go-connections/nat"
)

// Specific functions

// Deprecated
func (p Provider) getFrontendNameV1(container dockerData, idx int) string {
	return provider.Normalize(p.getFrontendRuleV1(container) + "-" + strconv.Itoa(idx))
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
// Deprecated
func (p Provider) getFrontendRuleV1(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikFrontendRule, ""); len(value) != 0 {
		return value
	}

	domain := label.GetStringValue(container.Labels, label.TraefikDomain, p.Domain)

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return "Host:" + getSubDomain(values[labelDockerComposeService]+"."+values[labelDockerComposeProject]) + "." + domain
	}

	if len(domain) > 0 {
		return "Host:" + getSubDomain(container.ServiceName) + "." + domain
	}

	return ""
}

// Deprecated
func getBackendNameV1(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikBackend, ""); len(value) != 0 {
		return provider.Normalize(value)
	}

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return provider.Normalize(values[labelDockerComposeService] + "_" + values[labelDockerComposeProject])
	}

	return provider.Normalize(container.ServiceName)
}

// Deprecated
func getPortV1(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikPort, ""); len(value) != 0 {
		return value
	}

	// See iteration order in https://blog.golang.org/go-maps-in-action
	var ports []nat.Port
	for port := range container.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	less := func(i, j nat.Port) bool {
		return i.Int() < j.Int()
	}
	nat.Sort(ports, less)

	if len(ports) > 0 {
		min := ports[0]
		return min.Port()
	}

	return ""
}

// replaced by Stickiness
// Deprecated
func getStickyV1(container dockerData) bool {
	if label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}

	return label.GetBoolValue(container.Labels, label.TraefikBackendLoadBalancerSticky, false)
}

// Deprecated
func hasLoadBalancerLabelV1(container dockerData) bool {
	method := label.Has(container.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickiness)
	cookieName := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickinessCookieName)
	return method || sticky || stickiness || cookieName
}

// Deprecated
func hasMaxConnLabelsV1(container dockerData) bool {
	mca := label.Has(container.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.Has(container.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

// Deprecated
func hasRedirectV1(container dockerData) bool {
	return hasLabelV1(label.TraefikFrontendRedirectEntryPoint)(container) ||
		hasLabelV1(label.TraefikFrontendRedirectReplacement)(container) && hasLabelV1(label.TraefikFrontendRedirectRegex)(container)
}

// Deprecated
func hasHeadersV1(container dockerData) bool {
	for key := range container.Labels {
		if strings.HasPrefix(key, label.Prefix+"frontend.headers.") {
			return true
		}
	}
	return false
}

// Label functions

// Deprecated
func getFuncStringLabelV1(labelName string, defaultValue string) func(container dockerData) string {
	return func(container dockerData) string {
		return label.GetStringValue(container.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncBoolLabelV1(labelName string, defaultValue bool) func(container dockerData) bool {
	return func(container dockerData) bool {
		return label.GetBoolValue(container.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncSliceStringLabelV1(labelName string) func(container dockerData) []string {
	return func(container dockerData) []string {
		return label.GetSliceStringValue(container.Labels, labelName)
	}
}

// Deprecated
func getFuncIntLabelV1(labelName string, defaultValue int) func(container dockerData) int {
	return func(container dockerData) int {
		return label.GetIntValue(container.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncInt64LabelV1(labelName string, defaultValue int64) func(container dockerData) int64 {
	return func(container dockerData) int64 {
		return label.GetInt64Value(container.Labels, labelName, defaultValue)
	}
}

// Deprecated
func hasFuncV1(labelName string) func(container dockerData) bool {
	return func(container dockerData) bool {
		return label.Has(container.Labels, labelName)
	}
}

// Deprecated
func hasLabelV1(label string) func(container dockerData) bool {
	return func(container dockerData) bool {
		lbl, err := getLabelV1(container, label)
		return err == nil && len(lbl) > 0
	}
}

// Deprecated
func getLabelV1(container dockerData, label string) (string, error) {
	if value, ok := container.Labels[label]; ok {
		return value, nil
	}
	return "", fmt.Errorf("label not found: %s", label)
}

// Deprecated
func getFuncMapLabelV1(labelName string) func(container dockerData) map[string]string {
	return func(container dockerData) map[string]string {
		return parseMapLabelV1(container, labelName)
	}
}

// Deprecated
func parseMapLabelV1(container dockerData, labelName string) map[string]string {
	if parts, err := getLabelV1(container, labelName); err == nil {
		if len(parts) == 0 {
			log.Errorf("Could not load %q", labelName)
			return nil
		}

		values := make(map[string]string)
		for _, headers := range strings.Split(parts, "||") {
			pair := strings.SplitN(headers, ":", 2)
			if len(pair) != 2 {
				log.Warnf("Could not load %q: %v, skipping...", labelName, pair)
			} else {
				values[http.CanonicalHeaderKey(strings.TrimSpace(pair[0]))] = strings.TrimSpace(pair[1])
			}
		}

		if len(values) == 0 {
			log.Errorf("Could not load %q", labelName)
			return nil
		}
		return values
	}

	return nil
}
