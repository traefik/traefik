package rancher

import (
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfigurationV1(services []rancherData) *types.Configuration {
	var RancherFuncMap = template.FuncMap{
		"getDomain": getFuncStringV1(label.TraefikDomain, p.Domain),

		// Backend functions
		"getPort":                     getFuncStringV1(label.TraefikPort, ""),
		"getProtocol":                 getFuncStringV1(label.TraefikProtocol, label.DefaultProtocol),
		"getWeight":                   getFuncIntV1(label.TraefikWeight, label.DefaultWeight),
		"hasCircuitBreakerLabel":      hasFuncV1(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncStringV1(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncStringV1(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncInt64V1(label.TraefikBackendMaxConnAmount, 0),
		"getMaxConnExtractorFunc":     getFuncStringV1(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getSticky":                   getStickyV1,
		"hasStickinessLabel":          hasFuncV1(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncStringV1(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),

		// Frontend functions
		"getBackend":             getBackendNameV1,
		"getFrontendRule":        p.getFrontendRuleV1,
		"getPriority":            getFuncIntV1(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":      getFuncBoolV1(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getEntryPoints":         getFuncSliceStringV1(label.TraefikFrontendEntryPoints),
		"getBasicAuth":           getFuncSliceStringV1(label.TraefikFrontendAuthBasic),
		"hasRedirect":            hasRedirect,
		"getRedirectEntryPoint":  getRedirectEntryPoint,
		"getRedirectRegex":       getRedirectRegex,
		"getRedirectReplacement": getRedirectReplacement,
	}

	// filter services
	filteredServices := fun.Filter(p.serviceFilterV1, services).([]rancherData)

	frontends := map[string]rancherData{}
	backends := map[string]rancherData{}

	for _, service := range filteredServices {
		frontendName := p.getFrontendNameV1(service)
		frontends[frontendName] = service
		backendName := getBackendNameV1(service)
		backends[backendName] = service
	}

	templateObjects := struct {
		Frontends map[string]rancherData
		Backends  map[string]rancherData
		Domain    string
	}{
		Frontends: frontends,
		Backends:  backends,
		Domain:    p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/rancher-v1.tmpl", RancherFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

// Deprecated
func (p *Provider) serviceFilterV1(service rancherData) bool {
	if service.Labels[label.TraefikPort] == "" {
		log.Debugf("Filtering service %s without traefik.port label", service.Name)
		return false
	}

	if !label.IsEnabled(service.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled service %s", service.Name)
		return false
	}

	constraintTags := label.GetSliceStringValue(service.Labels, label.TraefikTags)
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering service %s with constraint %s", service.Name, failingConstraint.String())
		}
		return false
	}

	// Only filter services by Health (HealthState) and State if EnableServiceHealthFilter is true
	if p.EnableServiceHealthFilter {

		if service.Health != "" && service.Health != healthy && service.Health != updatingHealthy {
			log.Debugf("Filtering service %s with healthState of %s", service.Name, service.Health)
			return false
		}
		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded && service.State != upgrading {
			log.Debugf("Filtering service %s with state of %s", service.Name, service.State)
			return false
		}
	}

	return true
}

// Deprecated
func (p *Provider) getFrontendRuleV1(service rancherData) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(service.Name, "/", ".", -1)) + "." + p.Domain
	return label.GetStringValue(service.Labels, label.TraefikFrontendRule, defaultRule)
}

// Deprecated
func (p *Provider) getFrontendNameV1(service rancherData) string {
	return provider.Normalize(p.getFrontendRuleV1(service))
}

// Deprecated
func getBackendNameV1(service rancherData) string {
	backend := label.GetStringValue(service.Labels, label.TraefikBackend, service.Name)
	return provider.Normalize(backend)
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getStickyV1(service rancherData) bool {
	if label.Has(service.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return label.GetBoolValue(service.Labels, label.TraefikBackendLoadBalancerSticky, false)
}

// Deprecated
func hasLoadBalancerLabel(service rancherData) bool {
	method := label.Has(service.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(service.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(service.Labels, label.TraefikBackendLoadBalancerStickiness)
	cookieName := label.Has(service.Labels, label.TraefikBackendLoadBalancerStickinessCookieName)
	return method || sticky || stickiness || cookieName
}

// Deprecated
func hasMaxConnLabels(service rancherData) bool {
	mca := label.Has(service.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.Has(service.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

func hasRedirect(service rancherData) bool {
	value := label.GetStringValue(service.Labels, label.TraefikFrontendRedirectEntryPoint, "")
	frep := len(value) > 0
	value = label.GetStringValue(service.Labels, label.TraefikFrontendRedirectRegex, "")
	frrg := len(value) > 0
	value = label.GetStringValue(service.Labels, label.TraefikFrontendRedirectReplacement, "")
	frrp := len(value) > 0

	return frep || frrg && frrp
}

func getRedirectEntryPoint(service rancherData) string {
	value := label.GetStringValue(service.Labels, label.TraefikFrontendRedirectEntryPoint, "")
	if len(value) == 0 {
		return ""
	}
	return value
}

func getRedirectRegex(service rancherData) string {
	value := label.GetStringValue(service.Labels, label.TraefikFrontendRedirectRegex, "")
	if len(value) == 0 {
		return ""
	}
	return value
}

func getRedirectReplacement(service rancherData) string {
	value := label.GetStringValue(service.Labels, label.TraefikFrontendRedirectReplacement, "")
	if len(value) == 0 {
		return ""
	}
	return value
}

// Label functions

// Deprecated
func getFuncStringV1(labelName string, defaultValue string) func(service rancherData) string {
	return func(service rancherData) string {
		return label.GetStringValue(service.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncBoolV1(labelName string, defaultValue bool) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.GetBoolValue(service.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncIntV1(labelName string, defaultValue int) func(service rancherData) int {
	return func(service rancherData) int {
		return label.GetIntValue(service.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncInt64V1(labelName string, defaultValue int64) func(service rancherData) int64 {
	return func(service rancherData) int64 {
		return label.GetInt64Value(service.Labels, labelName, defaultValue)
	}
}

// Deprecated
func getFuncSliceStringV1(labelName string) func(service rancherData) []string {
	return func(service rancherData) []string {
		return label.GetSliceStringValue(service.Labels, labelName)
	}
}

// Deprecated
func hasFuncV1(labelName string) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.Has(service.Labels, labelName)
	}
}
