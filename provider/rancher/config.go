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
		"getDomain": getFuncString(label.TraefikDomain, p.Domain), // FIXME dead ?

		// Backend functions
		"getPort":                     getFuncString(label.TraefikPort, ""),
		"getProtocol":                 getFuncString(label.TraefikProtocol, label.DefaultProtocol),
		"getWeight":                   getFuncString(label.TraefikWeight, label.DefaultWeight),
		"hasCircuitBreakerLabel":      hasFunc(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncString(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncString(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncInt64(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncString(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getSticky":                   getSticky,
		"hasStickinessLabel":          hasFunc(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncString(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		"hasHealthCheckLabels":        hasFunc(label.TraefikBackendHealthCheckPath),
		"getHealthCheckPath":          getFuncString(label.TraefikBackendHealthCheckPath, ""),
		"getHealthCheckPort":          getFuncInt(label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort),
		"getHealthCheckInterval":      getFuncString(label.TraefikBackendHealthCheckInterval, ""),

		// Frontend functions
		"getBackend":                 getBackend,
		"getPriority":                getFuncString(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":          getFuncString(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":             getFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":             getFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":               getFuncSliceString(label.TraefikFrontendAuthBasic),
		"getWhitelistSourceRange":    getFuncSliceString(label.TraefikFrontendWhitelistSourceRange),
		"getFrontendRule":            p.getFrontendRule,
		"hasRedirect":                hasRedirect,
		"getRedirectEntryPoint":      getFuncString(label.TraefikFrontendRedirectEntryPoint, label.DefaultFrontendRedirectEntryPoint),
		"getRedirectRegex":           getFuncString(label.TraefikFrontendRedirectRegex, ""),
		"getRedirectReplacement":     getFuncString(label.TraefikFrontendRedirectReplacement, ""),
		"hasErrorPages":              hasPrefixFunc(label.Prefix + label.BaseFrontendErrorPage),
		"getErrorPages":              getErrorPages,
		"hasRateLimits":              hasFunc(label.TraefikFrontendRateLimitExtractorFunc),
		"getRateLimitsExtractorFunc": getFuncString(label.TraefikFrontendRateLimitExtractorFunc, ""),
		"getRateLimits":              getRateLimits,
		// Headers
		"hasHeaders":                        hasPrefixFunc(label.TraefikFrontendHeaders),
		"hasRequestHeaders":                 hasFunc(label.TraefikFrontendRequestHeaders),
		"getRequestHeaders":                 getFuncMap(label.TraefikFrontendRequestHeaders),
		"hasResponseHeaders":                hasFunc(label.TraefikFrontendResponseHeaders),
		"getResponseHeaders":                getFuncMap(label.TraefikFrontendResponseHeaders),
		"getAllowedHostsHeaders":            getFuncSliceString(label.TraefikFrontendAllowedHosts),
		"getHostsProxyHeaders":              getFuncSliceString(label.TraefikFrontendHostsProxyHeaders),
		"getSSLRedirectHeaders":             getFuncBool(label.TraefikFrontendSSLRedirect, false),
		"getSSLTemporaryRedirectHeaders":    getFuncBool(label.TraefikFrontendSSLTemporaryRedirect, false),
		"getSSLHostHeaders":                 getFuncString(label.TraefikFrontendSSLHost, ""),
		"hasSSLProxyHeaders":                hasFunc(label.TraefikFrontendSSLProxyHeaders),
		"getSSLProxyHeaders":                getFuncMap(label.TraefikFrontendSSLProxyHeaders),
		"getSTSSecondsHeaders":              getFuncInt64(label.TraefikFrontendSTSSeconds, 0),
		"getSTSIncludeSubdomainsHeaders":    getFuncBool(label.TraefikFrontendSTSIncludeSubdomains, false),
		"getSTSPreloadHeaders":              getFuncBool(label.TraefikFrontendSTSPreload, false),
		"getForceSTSHeaderHeaders":          getFuncBool(label.TraefikFrontendForceSTSHeader, false),
		"getFrameDenyHeaders":               getFuncBool(label.TraefikFrontendFrameDeny, false),
		"getCustomFrameOptionsValueHeaders": getFuncString(label.TraefikFrontendCustomFrameOptionsValue, ""),
		"getContentTypeNosniffHeaders":      getFuncBool(label.TraefikFrontendContentTypeNosniff, false),
		"getBrowserXSSFilterHeaders":        getFuncBool(label.TraefikFrontendBrowserXSSFilter, false),
		"getContentSecurityPolicyHeaders":   getFuncString(label.TraefikFrontendContentSecurityPolicy, ""),
		"getPublicKeyHeaders":               getFuncString(label.TraefikFrontendPublicKey, ""),
		"getReferrerPolicyHeaders":          getFuncString(label.TraefikFrontendReferrerPolicy, ""),
		"getIsDevelopmentHeaders":           getFuncBool(label.TraefikFrontendIsDevelopment, false),
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
		Frontends: frontends,
		Backends:  backends,
		Domain:    p.Domain,
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
	frep := label.Has(service.Labels, label.TraefikFrontendRedirectEntryPoint)
	frrg := label.Has(service.Labels, label.TraefikFrontendRedirectRegex)
	frrp := label.Has(service.Labels, label.TraefikFrontendRedirectReplacement)

	return frep || frrg && frrp
}

func getErrorPages(service rancherData) map[string]*types.ErrorPage {
	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(service.Labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimits(service rancherData) map[string]*types.Rate {
	prefix := label.Prefix + label.BaseFrontendRateLimit
	return label.ParseRateSets(service.Labels, prefix, label.RegexpFrontendRateLimit)
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

func getFuncMap(labelName string) func(service rancherData) map[string]string {
	return func(service rancherData) map[string]string {
		return label.GetMapValue(service.Labels, labelName)
	}
}

func hasFunc(labelName string) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.Has(service.Labels, labelName)
	}
}

func hasPrefixFunc(prefix string) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.HasPrefix(service.Labels, prefix)
	}
}
