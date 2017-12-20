package ecs

import (
	"math"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
func (p *Provider) buildConfiguration(services map[string][]ecsInstance) (*types.Configuration, error) {
	var ecsFuncMap = template.FuncMap{
		// Backend functions
		"getProtocol":                 getFuncStringValue(label.TraefikProtocol, label.DefaultProtocol),
		"getHost":                     getHost,
		"getPort":                     getPort,
		"getWeight":                   getFuncStringValue(label.TraefikWeight, label.DefaultWeight),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncFirstStringValue(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"getSticky":                   getSticky,
		"hasStickinessLabel":          getFuncFirstBoolValue(label.TraefikBackendLoadBalancerStickiness, false),
		"getStickinessCookieName":     getFuncFirstStringValue(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		"hasHealthCheckLabels":        hasFuncFirst(label.TraefikBackendHealthCheckPath),
		"getHealthCheckPath":          getFuncFirstStringValue(label.TraefikBackendHealthCheckPath, ""),
		"getHealthCheckPort":          getFuncFirstIntValue(label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort),
		"getHealthCheckInterval":      getFuncFirstStringValue(label.TraefikBackendHealthCheckInterval, ""),
		"hasCircuitBreakerLabel":      hasFuncFirst(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncFirstStringValue(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncFirstInt64Value(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncFirstStringValue(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),

		// Frontend functions
		"filterFrontends":            filterFrontends,
		"getFrontendRule":            p.getFrontendRule,
		"getPassHostHeader":          getFuncStringValue(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":             getFuncBoolValue(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":                getFuncStringValue(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getBasicAuth":               getFuncSliceString(label.TraefikFrontendAuthBasic),
		"getEntryPoints":             getFuncSliceString(label.TraefikFrontendEntryPoints),
		"getWhitelistSourceRange":    getFuncSliceString(label.TraefikFrontendWhitelistSourceRange),
		"hasRedirect":                hasRedirect,
		"getRedirectEntryPoint":      getFuncStringValue(label.TraefikFrontendRedirectEntryPoint, label.DefaultFrontendRedirectEntryPoint),
		"getRedirectRegex":           getFuncStringValue(label.TraefikFrontendRedirectRegex, ""),
		"getRedirectReplacement":     getFuncStringValue(label.TraefikFrontendRedirectReplacement, ""),
		"hasErrorPages":              hasPrefixFuncLabel(label.Prefix + label.BaseFrontendErrorPage),
		"getErrorPages":              getErrorPages,
		"hasRateLimits":              hasFuncLabel(label.TraefikFrontendRateLimitExtractorFunc),
		"getRateLimitsExtractorFunc": getFuncStringValue(label.TraefikFrontendRateLimitExtractorFunc, ""),
		"getRateLimits":              getRateLimits,
		// Headers
		"hasRequestHeaders":                 hasFuncLabel(label.TraefikFrontendRequestHeaders),
		"getRequestHeaders":                 getFuncMapValue(label.TraefikFrontendRequestHeaders),
		"hasResponseHeaders":                hasFuncLabel(label.TraefikFrontendResponseHeaders),
		"getResponseHeaders":                getFuncMapValue(label.TraefikFrontendResponseHeaders),
		"hasAllowedHostsHeaders":            hasFuncLabel(label.TraefikFrontendAllowedHosts),
		"getAllowedHostsHeaders":            getFuncSliceString(label.TraefikFrontendAllowedHosts),
		"hasHostsProxyHeaders":              hasFuncLabel(label.TraefikFrontendHostsProxyHeaders),
		"getHostsProxyHeaders":              getFuncSliceString(label.TraefikFrontendHostsProxyHeaders),
		"hasSSLRedirectHeaders":             hasFuncLabel(label.TraefikFrontendSSLRedirect),
		"getSSLRedirectHeaders":             getFuncBoolValue(label.TraefikFrontendSSLRedirect, false),
		"hasSSLTemporaryRedirectHeaders":    hasFuncLabel(label.TraefikFrontendSSLTemporaryRedirect),
		"getSSLTemporaryRedirectHeaders":    getFuncBoolValue(label.TraefikFrontendSSLTemporaryRedirect, false),
		"hasSSLHostHeaders":                 hasFuncLabel(label.TraefikFrontendSSLHost),
		"getSSLHostHeaders":                 getFuncStringValue(label.TraefikFrontendSSLHost, ""),
		"hasSSLProxyHeaders":                hasFuncLabel(label.TraefikFrontendSSLProxyHeaders),
		"getSSLProxyHeaders":                getFuncMapValue(label.TraefikFrontendSSLProxyHeaders),
		"hasSTSSecondsHeaders":              hasFuncLabel(label.TraefikFrontendSTSSeconds),
		"getSTSSecondsHeaders":              getFuncInt64Value(label.TraefikFrontendSTSSeconds, 0),
		"hasSTSIncludeSubdomainsHeaders":    hasFuncLabel(label.TraefikFrontendSTSIncludeSubdomains),
		"getSTSIncludeSubdomainsHeaders":    getFuncBoolValue(label.TraefikFrontendSTSIncludeSubdomains, false),
		"hasSTSPreloadHeaders":              hasFuncLabel(label.TraefikFrontendSTSPreload),
		"getSTSPreloadHeaders":              getFuncBoolValue(label.TraefikFrontendSTSPreload, false),
		"hasForceSTSHeaderHeaders":          hasFuncLabel(label.TraefikFrontendForceSTSHeader),
		"getForceSTSHeaderHeaders":          getFuncBoolValue(label.TraefikFrontendForceSTSHeader, false),
		"hasFrameDenyHeaders":               hasFuncLabel(label.TraefikFrontendFrameDeny),
		"getFrameDenyHeaders":               getFuncBoolValue(label.TraefikFrontendFrameDeny, false),
		"hasCustomFrameOptionsValueHeaders": hasFuncLabel(label.TraefikFrontendCustomFrameOptionsValue),
		"getCustomFrameOptionsValueHeaders": getFuncStringValue(label.TraefikFrontendCustomFrameOptionsValue, ""),
		"hasContentTypeNosniffHeaders":      hasFuncLabel(label.TraefikFrontendContentTypeNosniff),
		"getContentTypeNosniffHeaders":      getFuncBoolValue(label.TraefikFrontendContentTypeNosniff, false),
		"hasBrowserXSSFilterHeaders":        hasFuncLabel(label.TraefikFrontendBrowserXSSFilter),
		"getBrowserXSSFilterHeaders":        getFuncBoolValue(label.TraefikFrontendBrowserXSSFilter, false),
		"hasContentSecurityPolicyHeaders":   hasFuncLabel(label.TraefikFrontendContentSecurityPolicy),
		"getContentSecurityPolicyHeaders":   getFuncStringValue(label.TraefikFrontendContentSecurityPolicy, ""),
		"hasPublicKeyHeaders":               hasFuncLabel(label.TraefikFrontendPublicKey),
		"getPublicKeyHeaders":               getFuncStringValue(label.TraefikFrontendPublicKey, ""),
		"hasReferrerPolicyHeaders":          hasFuncLabel(label.TraefikFrontendReferrerPolicy),
		"getReferrerPolicyHeaders":          getFuncStringValue(label.TraefikFrontendReferrerPolicy, ""),
		"hasIsDevelopmentHeaders":           hasFuncLabel(label.TraefikFrontendIsDevelopment),
		"getIsDevelopmentHeaders":           getFuncBoolValue(label.TraefikFrontendIsDevelopment, false),
	}
	return p.GetConfiguration("templates/ecs.tmpl", ecsFuncMap, struct {
		Services map[string][]ecsInstance
	}{
		services,
	})
}

func (p *Provider) getFrontendRule(i ecsInstance) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + p.Domain
	return getStringValue(i, label.TraefikFrontendRule, defaultRule)
}

// TODO: Deprecated
// Deprecated replaced by Stickiness
func getSticky(instances []ecsInstance) string {
	if hasFirst(instances, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return getFirstStringValue(instances, label.TraefikBackendLoadBalancerSticky, "false")
}

func getHost(i ecsInstance) string {
	return *i.machine.PrivateIpAddress
}

func getPort(i ecsInstance) string {
	if value := getStringValue(i, label.TraefikPort, ""); len(value) > 0 {
		return value
	}
	return strconv.FormatInt(*i.container.NetworkBindings[0].HostPort, 10)
}

func filterFrontends(instances []ecsInstance) []ecsInstance {
	byName := make(map[string]struct{})

	return fun.Filter(func(i ecsInstance) bool {
		_, found := byName[i.Name]
		if !found {
			byName[i.Name] = struct{}{}
		}
		return !found
	}, instances).([]ecsInstance)
}

func hasLoadBalancerLabel(instances []ecsInstance) bool {
	method := hasFirst(instances, label.TraefikBackendLoadBalancerMethod)
	sticky := hasFirst(instances, label.TraefikBackendLoadBalancerSticky)
	stickiness := hasFirst(instances, label.TraefikBackendLoadBalancerStickiness)
	cookieName := hasFirst(instances, label.TraefikBackendLoadBalancerStickinessCookieName)

	return method || sticky || stickiness || cookieName
}

func hasMaxConnLabels(instances []ecsInstance) bool {
	mca := hasFirst(instances, label.TraefikBackendMaxConnAmount)
	mcef := hasFirst(instances, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

func hasRedirect(instance ecsInstance) bool {
	return hasLabel(instance, label.TraefikFrontendRedirectEntryPoint) ||
		hasLabel(instance, label.TraefikFrontendRedirectRegex) && hasLabel(instance, label.TraefikFrontendRedirectReplacement)
}

func getErrorPages(instance ecsInstance) map[string]*types.ErrorPage {
	labels := mapPToMap(instance.containerDefinition.DockerLabels)

	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimits(instance ecsInstance) map[string]*types.Rate {
	labels := mapPToMap(instance.containerDefinition.DockerLabels)

	prefix := label.Prefix + label.BaseFrontendRateLimit
	return label.ParseRateSets(labels, prefix, label.RegexpFrontendRateLimit)
}

// Label functions

func hasFuncLabel(labelName string) func(i ecsInstance) bool {
	return func(i ecsInstance) bool {
		return hasLabel(i, labelName)
	}
}

func hasPrefixFuncLabel(prefix string) func(i ecsInstance) bool {
	return func(i ecsInstance) bool {
		return hasPrefix(i, prefix)
	}
}

func getFuncStringValue(labelName string, defaultValue string) func(i ecsInstance) string {
	return func(i ecsInstance) string {
		return getStringValue(i, labelName, defaultValue)
	}
}

func getFuncBoolValue(labelName string, defaultValue bool) func(i ecsInstance) bool {
	return func(i ecsInstance) bool {
		return getBoolValue(i, labelName, defaultValue)
	}
}

func getFuncInt64Value(labelName string, defaultValue int64) func(i ecsInstance) int64 {
	return func(i ecsInstance) int64 {
		return getInt64Value(i, labelName, defaultValue)
	}
}

func getFuncSliceString(labelName string) func(i ecsInstance) []string {
	return func(i ecsInstance) []string {
		return getSliceString(i, labelName)
	}
}

func getFuncMapValue(labelName string) func(i ecsInstance) map[string]string {
	return func(i ecsInstance) map[string]string {
		return getMapString(i, labelName)
	}
}

func hasFuncFirst(labelName string) func(instances []ecsInstance) bool {
	return func(instances []ecsInstance) bool {
		return hasFirst(instances, labelName)
	}
}

func getFuncFirstStringValue(labelName string, defaultValue string) func(instances []ecsInstance) string {
	return func(instances []ecsInstance) string {
		return getFirstStringValue(instances, labelName, defaultValue)
	}
}

func getFuncFirstIntValue(labelName string, defaultValue int) func(instances []ecsInstance) int {
	return func(instances []ecsInstance) int {
		if len(instances) < 0 {
			return defaultValue
		}
		return getIntValue(instances[0], labelName, defaultValue)
	}
}

func getFuncFirstInt64Value(labelName string, defaultValue int64) func(instances []ecsInstance) int64 {
	return func(instances []ecsInstance) int64 {
		if len(instances) < 0 {
			return defaultValue
		}
		return getInt64Value(instances[0], labelName, defaultValue)
	}
}

func getFuncFirstBoolValue(labelName string, defaultValue bool) func(instances []ecsInstance) bool {
	return func(instances []ecsInstance) bool {
		if len(instances) < 0 {
			return defaultValue
		}
		return getBoolValue(instances[0], labelName, defaultValue)
	}
}

func hasLabel(i ecsInstance, labelName string) bool {
	value, ok := i.containerDefinition.DockerLabels[labelName]
	return ok && value != nil && len(*value) > 0
}

func hasPrefix(i ecsInstance, prefix string) bool {
	for name, value := range i.containerDefinition.DockerLabels {
		if strings.HasPrefix(name, prefix) && value != nil && len(*value) > 0 {
			return true
		}
	}
	return false
}

func getStringValue(i ecsInstance, labelName string, defaultValue string) string {
	if v, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if v == nil {
			return defaultValue
		}
		if len(*v) == 0 {
			return defaultValue
		}
		return *v
	}
	return defaultValue
}

func getBoolValue(i ecsInstance, labelName string, defaultValue bool) bool {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.ParseBool(*rawValue)
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

func getIntValue(i ecsInstance, labelName string, defaultValue int) int {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.Atoi(*rawValue)
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

func getInt64Value(i ecsInstance, labelName string, defaultValue int64) int64 {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.ParseInt(*rawValue, 10, 64)
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

func getSliceString(i ecsInstance, labelName string) []string {
	if value, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if value == nil {
			return nil
		}
		if len(*value) == 0 {
			return nil
		}
		return label.SplitAndTrimString(*value, ",")
	}
	return nil
}

func getMapString(i ecsInstance, labelName string) map[string]string {
	if value, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if value == nil {
			return nil
		}
		if len(*value) == 0 {
			return nil
		}
		return label.ParseMapValue(labelName, *value)
	}
	return nil
}

func hasFirst(instances []ecsInstance, labelName string) bool {
	if len(instances) == 0 {
		return false
	}
	return hasLabel(instances[0], labelName)
}

func getFirstStringValue(instances []ecsInstance, labelName string, defaultValue string) string {
	if len(instances) == 0 {
		return defaultValue
	}
	return getStringValue(instances[0], labelName, defaultValue)
}

func mapPToMap(src map[string]*string) map[string]string {
	result := make(map[string]string)
	for key, value := range src {
		if value != nil && len(*value) > 0 {
			result[key] = *value
		}
	}
	return result
}

func isEnabled(i ecsInstance, exposedByDefault bool) bool {
	return getBoolValue(i, label.TraefikEnable, exposedByDefault)
}
