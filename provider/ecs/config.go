package ecs

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
func (p *Provider) buildConfiguration(services map[string][]ecsInstance) (*types.Configuration, error) {
	var ecsFuncMap = template.FuncMap{
		// Backend functions
		"getHost":           getHost,
		"getPort":           getPort,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getBuffering":      getBuffering,
		"getServers":        getServers,

		// TODO Deprecated [breaking]
		"getProtocol": getFuncStringValue(label.TraefikProtocol, label.DefaultProtocol),
		// TODO Deprecated [breaking]
		"getWeight": getFuncIntValue(label.TraefikWeight, label.DefaultWeightInt),
		// TODO Deprecated [breaking]
		"getLoadBalancerMethod": getFuncFirstStringValue(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		// TODO Deprecated [breaking]
		"getSticky": getSticky,
		// TODO Deprecated [breaking]
		"hasStickinessLabel": getFuncFirstBoolValue(label.TraefikBackendLoadBalancerStickiness, false),
		// TODO Deprecated [breaking]
		"getStickinessCookieName": getFuncFirstStringValue(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		// TODO Deprecated [breaking]
		"hasHealthCheckLabels": hasFuncFirst(label.TraefikBackendHealthCheckPath),
		// TODO Deprecated [breaking]
		"getHealthCheckPath": getFuncFirstStringValue(label.TraefikBackendHealthCheckPath, ""),
		// TODO Deprecated [breaking]
		"getHealthCheckInterval": getFuncFirstStringValue(label.TraefikBackendHealthCheckInterval, ""),

		// Frontend functions
		"filterFrontends":   filterFrontends,
		"getFrontendRule":   p.getFrontendRule,
		"getPassHostHeader": getFuncBoolValue(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    getFuncBoolValue(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":       getFuncIntValue(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getBasicAuth":      getFuncSliceString(label.TraefikFrontendAuthBasic),
		"getEntryPoints":    getFuncSliceString(label.TraefikFrontendEntryPoints),
		"getRedirect":       getRedirect,
		"getErrorPages":     getErrorPages,
		"getRateLimit":      getRateLimit,
		"getHeaders":        getHeaders,
		"getWhiteList":      getWhiteList,
	}
	return p.GetConfiguration("templates/ecs.tmpl", ecsFuncMap, struct {
		Services map[string][]ecsInstance
	}{
		Services: services,
	})
}

func (p *Provider) getFrontendRule(i ecsInstance) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + p.Domain
	return getStringValue(i, label.TraefikFrontendRule, defaultRule)
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(instances []ecsInstance) bool {
	if hasFirst(instances, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return getFirstBoolValue(instances, label.TraefikBackendLoadBalancerSticky, false)
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getStickyOne(instance ecsInstance) bool {
	if hasLabel(instance, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return getBoolValue(instance, label.TraefikBackendLoadBalancerSticky, false)
}

func getHost(i ecsInstance) string {
	return aws.StringValue(i.machine.PrivateIpAddress)
}

func getPort(i ecsInstance) string {
	if value := getStringValue(i, label.TraefikPort, ""); len(value) > 0 {
		return value
	}
	return strconv.FormatInt(aws.Int64Value(i.container.NetworkBindings[0].HostPort), 10)
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

func getCircuitBreaker(instance ecsInstance) *types.CircuitBreaker {
	expression := getStringValue(instance, label.TraefikBackendCircuitBreakerExpression, "")
	if len(expression) == 0 {
		return nil
	}

	return &types.CircuitBreaker{Expression: expression}
}

func getLoadBalancer(instance ecsInstance) *types.LoadBalancer {
	if !hasPrefix(instance, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := getStringValue(instance, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getStickyOne(instance),
	}

	if getBoolValue(instance, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := getStringValue(instance, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

func getMaxConn(instance ecsInstance) *types.MaxConn {
	amount := getInt64Value(instance, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := getStringValue(instance, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getHealthCheck(instance ecsInstance) *types.HealthCheck {
	path := getStringValue(instance, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := getIntValue(instance, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort)
	interval := getStringValue(instance, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getBuffering(instance ecsInstance) *types.Buffering {
	if !hasPrefix(instance, label.TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  getInt64Value(instance, label.TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: getInt64Value(instance, label.TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  getInt64Value(instance, label.TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: getInt64Value(instance, label.TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      getStringValue(instance, label.TraefikBackendBufferingRetryExpression, ""),
	}
}

func getServers(instances []ecsInstance) map[string]types.Server {
	var servers map[string]types.Server

	for _, instance := range instances {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := getStringValue(instance, label.TraefikProtocol, label.DefaultProtocol)
		host := getHost(instance)
		port := getPort(instance)

		serverName := provider.Normalize(fmt.Sprintf("server-%s-%s", instance.Name, instance.ID))
		servers[serverName] = types.Server{
			URL:      fmt.Sprintf("%s://%s:%s", protocol, host, port),
			Priority: getIntValue(instance, label.TraefikPriority, 0),
			Weight:   getIntValue(instance, label.TraefikWeight, 0),
		}
	}

	return servers
}

func getWhiteList(instance ecsInstance) *types.WhiteList {
	ranges := getSliceString(instance, label.TraefikFrontendWhiteListSourceRange)
	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: getBoolValue(instance, label.TraefikFrontendWhiteListUseXForwardedFor, false),
		}
	}

	return nil
}

func getRedirect(instance ecsInstance) *types.Redirect {
	permanent := getBoolValue(instance, label.TraefikFrontendRedirectPermanent, false)

	if hasLabel(instance, label.TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: getStringValue(instance, label.TraefikFrontendRedirectEntryPoint, ""),
			Permanent:  permanent,
		}
	}

	if hasLabel(instance, label.TraefikFrontendRedirectRegex) &&
		hasLabel(instance, label.TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       getStringValue(instance, label.TraefikFrontendRedirectRegex, ""),
			Replacement: getStringValue(instance, label.TraefikFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func getErrorPages(instance ecsInstance) map[string]*types.ErrorPage {
	labels := mapPToMap(instance.containerDefinition.DockerLabels)
	if len(labels) == 0 {
		return nil
	}

	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimit(instance ecsInstance) *types.RateLimit {
	extractorFunc := getStringValue(instance, label.TraefikFrontendRateLimitExtractorFunc, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	labels := mapPToMap(instance.containerDefinition.DockerLabels)
	prefix := label.Prefix + label.BaseFrontendRateLimit
	limits := label.ParseRateSets(labels, prefix, label.RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func getHeaders(instance ecsInstance) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    getMapString(instance, label.TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   getMapString(instance, label.TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         getMapString(instance, label.TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            getSliceString(instance, label.TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       getSliceString(instance, label.TraefikFrontendHostsProxyHeaders),
		STSSeconds:              getInt64Value(instance, label.TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             getBoolValue(instance, label.TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    getBoolValue(instance, label.TraefikFrontendSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    getBoolValue(instance, label.TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              getBoolValue(instance, label.TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          getBoolValue(instance, label.TraefikFrontendForceSTSHeader, false),
		FrameDeny:               getBoolValue(instance, label.TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      getBoolValue(instance, label.TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        getBoolValue(instance, label.TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           getBoolValue(instance, label.TraefikFrontendIsDevelopment, false),
		SSLHost:                 getStringValue(instance, label.TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: getStringValue(instance, label.TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   getStringValue(instance, label.TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               getStringValue(instance, label.TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          getStringValue(instance, label.TraefikFrontendReferrerPolicy, ""),
		CustomBrowserXSSValue:   getStringValue(instance, label.TraefikFrontendCustomBrowserXSSValue, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Label functions

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

func getFuncIntValue(labelName string, defaultValue int) func(i ecsInstance) int {
	return func(i ecsInstance) int {
		return getIntValue(i, labelName, defaultValue)
	}
}

func getFuncSliceString(labelName string) func(i ecsInstance) []string {
	return func(i ecsInstance) []string {
		return getSliceString(i, labelName)
	}
}

// Deprecated
func hasFuncFirst(labelName string) func(instances []ecsInstance) bool {
	return func(instances []ecsInstance) bool {
		return hasFirst(instances, labelName)
	}
}

// Deprecated
func getFuncFirstStringValue(labelName string, defaultValue string) func(instances []ecsInstance) string {
	return func(instances []ecsInstance) string {
		return getFirstStringValue(instances, labelName, defaultValue)
	}
}

// Deprecated
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
	return ok && value != nil && len(aws.StringValue(value)) > 0
}

func hasPrefix(i ecsInstance, prefix string) bool {
	for name, value := range i.containerDefinition.DockerLabels {
		if strings.HasPrefix(name, prefix) && value != nil && len(aws.StringValue(value)) > 0 {
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
		if len(aws.StringValue(v)) == 0 {
			return defaultValue
		}
		return aws.StringValue(v)
	}
	return defaultValue
}

func getBoolValue(i ecsInstance, labelName string, defaultValue bool) bool {
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

func getIntValue(i ecsInstance, labelName string, defaultValue int) int {
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

func getInt64Value(i ecsInstance, labelName string, defaultValue int64) int64 {
	rawValue, ok := i.containerDefinition.DockerLabels[labelName]
	if ok {
		if rawValue != nil {
			v, err := strconv.ParseInt(aws.StringValue(rawValue), 10, 64)
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
		if len(aws.StringValue(value)) == 0 {
			return nil
		}
		return label.SplitAndTrimString(aws.StringValue(value), ",")
	}
	return nil
}

func getMapString(i ecsInstance, labelName string) map[string]string {
	if value, ok := i.containerDefinition.DockerLabels[labelName]; ok {
		if value == nil {
			return nil
		}
		if len(aws.StringValue(value)) == 0 {
			return nil
		}
		return label.ParseMapValue(labelName, aws.StringValue(value))
	}
	return nil
}

// Deprecated
func hasFirst(instances []ecsInstance, labelName string) bool {
	if len(instances) == 0 {
		return false
	}
	return hasLabel(instances[0], labelName)
}

// Deprecated
func getFirstStringValue(instances []ecsInstance, labelName string, defaultValue string) string {
	if len(instances) == 0 {
		return defaultValue
	}
	return getStringValue(instances[0], labelName, defaultValue)
}

// Deprecated
func getFirstBoolValue(instances []ecsInstance, labelName string, defaultValue bool) bool {
	if len(instances) == 0 {
		return defaultValue
	}
	return getBoolValue(instances[0], labelName, defaultValue)
}

func mapPToMap(src map[string]*string) map[string]string {
	result := make(map[string]string)
	for key, value := range src {
		if value != nil && len(aws.StringValue(value)) > 0 {
			result[key] = aws.StringValue(value)
		}
	}
	return result
}

func isEnabled(i ecsInstance, exposedByDefault bool) bool {
	return getBoolValue(i, label.TraefikEnable, exposedByDefault)
}
