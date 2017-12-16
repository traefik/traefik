package rancher

import (
	"math"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration(services []rancherData) *types.Configuration {

	var RancherFuncMap = template.FuncMap{
		"getPort":                     getFuncString(label.TraefikPort, ""),
		"getBackend":                  getBackend,
		"getWeight":                   getFuncString(label.TraefikWeight, label.DefaultWeight),
		"getDomain":                   getFuncString(label.TraefikDomain, p.Domain),
		"getProtocol":                 getFuncString(label.TraefikProtocol, label.DefaultProtocol),
		"getPassHostHeader":           getFuncString(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":              getFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":                 getFuncString(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getEntryPoints":              getFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":                getFuncSliceString(label.TraefikFrontendAuthBasic),
		"getFrontendRule":             p.getFrontendRule,
		"hasCircuitBreakerLabel":      hasFunc(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncString(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncString(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncInt64(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncString(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getSticky":                   getSticky, // deprecated
		"hasStickinessLabel":          hasFunc(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncString(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		"hasRedirect":                 hasRedirect,
		"getRedirectEntryPoint":       getFuncString(label.TraefikFrontendRedirectEntryPoint, label.DefaultFrontendRedirectEntryPoint),
		"getRedirectRegex":            getFuncString(label.TraefikFrontendRedirectRegex, ""),
		"getRedirectReplacement":      getFuncString(label.TraefikFrontendRedirectReplacement, ""),
		"getWhitelistSourceRange":     getFuncSliceString(label.TraefikFrontendWhitelistSourceRange),
		"hasHealthCheckLabels":        hasFunc(label.TraefikBackendHealthCheckPath),
		"getHealthCheckPath":          getFuncString(label.TraefikBackendHealthCheckPath, ""),
		"getHealthCheckPort":          getFuncInt(label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort),
		"getHealthCheckInterval":      getFuncString(label.TraefikBackendHealthCheckInterval, ""),
	}

	// filter services
	filteredServices := fun.Filter(p.serviceFilter, services).([]rancherData)

	frontends := map[string]rancherData{}
	backends := map[string]rancherData{}

	for _, service := range filteredServices {
		frontendName := p.getFrontendName(service)
		frontends[frontendName] = service
		backendName := getBackend(service)
		backends[backendName] = service
	}

	templateObjects := struct {
		Frontends map[string]rancherData
		Backends  map[string]rancherData
		Domain    string
	}{
		frontends,
		backends,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/rancher.tmpl", RancherFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p *Provider) serviceFilter(service rancherData) bool {

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

		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded {
			log.Debugf("Filtering service %s with state of %s", service.Name, service.State)
			return false
		}
	}

	return true
}

func (p *Provider) getFrontendRule(service rancherData) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(service.Name, "/", ".", -1)) + "." + p.Domain
	return label.GetStringValue(service.Labels, label.TraefikFrontendRule, defaultRule)
}

func (p *Provider) getFrontendName(service rancherData) string {
	return provider.Normalize(p.getFrontendRule(service))
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(service rancherData) string {
	if label.Has(service.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return label.GetStringValue(service.Labels, label.TraefikBackendLoadBalancerSticky, "false")
}

func hasLoadBalancerLabel(service rancherData) bool {
	method := label.Has(service.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(service.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(service.Labels, label.TraefikBackendLoadBalancerStickiness)
	cookieName := label.Has(service.Labels, label.TraefikBackendLoadBalancerStickinessCookieName)
	return method || sticky || stickiness || cookieName
}

func hasMaxConnLabels(service rancherData) bool {
	mca := label.Has(service.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.Has(service.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

func getBackend(service rancherData) string {
	backend := label.GetStringValue(service.Labels, label.TraefikBackend, service.Name)
	return provider.Normalize(backend)
}

func hasRedirect(service rancherData) bool {
	return label.Has(service.Labels, label.TraefikFrontendRedirectEntryPoint) ||
		label.Has(service.Labels, label.TraefikFrontendRedirectRegex) && label.Has(service.Labels, label.TraefikFrontendRedirectReplacement)
}

// Label functions

func getFuncString(labelName string, defaultValue string) func(service rancherData) string {
	return func(service rancherData) string {
		return label.GetStringValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncInt(labelName string, defaultValue int) func(service rancherData) int {
	return func(service rancherData) int {
		return label.GetIntValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncBool(labelName string, defaultValue bool) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.GetBoolValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncInt64(labelName string, defaultValue int64) func(service rancherData) int64 {
	return func(service rancherData) int64 {
		return label.GetInt64Value(service.Labels, labelName, defaultValue)
	}
}

func getFuncSliceString(labelName string) func(service rancherData) []string {
	return func(service rancherData) []string {
		return label.GetSliceStringValue(service.Labels, labelName)
	}
}

func hasFunc(labelName string) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.Has(service.Labels, labelName)
	}
}
