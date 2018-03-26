package rancher

import (
	"fmt"
	"math"
	"strconv"
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
		"getDomain": getFuncString(label.TraefikDomain, p.Domain),

		// Backend functions
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getBuffering":      getBuffering,
		"getServers":        getServers,

		// TODO Deprecated [breaking]
		"getPort": getFuncString(label.TraefikPort, ""),
		// TODO Deprecated [breaking]
		"getProtocol": getFuncString(label.TraefikProtocol, label.DefaultProtocol),
		// TODO Deprecated [breaking]
		"getWeight": getFuncInt(label.TraefikWeight, label.DefaultWeightInt),
		// TODO Deprecated [breaking]
		"hasCircuitBreakerLabel": hasFunc(label.TraefikBackendCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"getCircuitBreakerExpression": getFuncString(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"hasLoadBalancerLabel": hasLoadBalancerLabel,
		// TODO Deprecated [breaking]
		"getLoadBalancerMethod": getFuncString(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		// TODO Deprecated [breaking]
		"hasMaxConnLabels": hasMaxConnLabels,
		// TODO Deprecated [breaking]
		"getMaxConnAmount": getFuncInt64(label.TraefikBackendMaxConnAmount, 0),
		// TODO Deprecated [breaking]
		"getMaxConnExtractorFunc": getFuncString(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		// TODO Deprecated [breaking]
		"getSticky": getSticky,
		// TODO Deprecated [breaking]
		"hasStickinessLabel": hasFunc(label.TraefikBackendLoadBalancerStickiness),
		// TODO Deprecated [breaking]
		"getStickinessCookieName": getFuncString(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),

		// Frontend functions
		"getBackend":        getBackendName, // TODO Deprecated [breaking] replaced by getBackendName
		"getBackendName":    getBackendName,
		"getFrontendRule":   p.getFrontendRule,
		"getPriority":       getFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getPassHostHeader": getFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    getFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":    getFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      getFuncSliceString(label.TraefikFrontendAuthBasic),
		"getErrorPages":     getErrorPages,
		"getRateLimit":      getRateLimit,
		"getRedirect":       getRedirect,
		"getHeaders":        getHeaders,
		"getWhiteList":      getWhiteList,

		// TODO Deprecated [breaking]
		"getWhitelistSourceRange": getFuncSliceString(label.TraefikFrontendWhitelistSourceRange),
	}

	// filter services
	filteredServices := fun.Filter(p.serviceFilter, services).([]rancherData)

	frontends := map[string]rancherData{}
	backends := map[string]rancherData{}

	for _, service := range filteredServices {
		frontendName := p.getFrontendName(service)
		frontends[frontendName] = service
		backendName := getBackendName(service)
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
		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded && service.State != upgrading {
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
func getSticky(service rancherData) bool {
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

func getBackendName(service rancherData) string {
	backend := label.GetStringValue(service.Labels, label.TraefikBackend, service.Name)
	return provider.Normalize(backend)
}

func getCircuitBreaker(service rancherData) *types.CircuitBreaker {
	circuitBreaker := label.GetStringValue(service.Labels, label.TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func getLoadBalancer(service rancherData) *types.LoadBalancer {
	if !label.HasPrefix(service.Labels, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := label.GetStringValue(service.Labels, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getSticky(service),
	}

	if label.GetBoolValue(service.Labels, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := label.GetStringValue(service.Labels, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

func getMaxConn(service rancherData) *types.MaxConn {
	amount := label.GetInt64Value(service.Labels, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := label.GetStringValue(service.Labels, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getHealthCheck(service rancherData) *types.HealthCheck {
	path := label.GetStringValue(service.Labels, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := label.GetIntValue(service.Labels, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort)
	interval := label.GetStringValue(service.Labels, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getBuffering(service rancherData) *types.Buffering {
	if !label.HasPrefix(service.Labels, label.TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  label.GetInt64Value(service.Labels, label.TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: label.GetInt64Value(service.Labels, label.TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  label.GetInt64Value(service.Labels, label.TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: label.GetInt64Value(service.Labels, label.TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      label.GetStringValue(service.Labels, label.TraefikBackendBufferingRetryExpression, ""),
	}
}

func getServers(service rancherData) map[string]types.Server {
	var servers map[string]types.Server

	for index, ip := range service.Containers {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(service.Labels, label.TraefikProtocol, label.DefaultProtocol)
		port := label.GetStringValue(service.Labels, label.TraefikPort, "")
		weight := label.GetIntValue(service.Labels, label.TraefikWeight, label.DefaultWeightInt)

		serverName := "server-" + strconv.Itoa(index)
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%s", protocol, ip, port),
			Weight: weight,
		}
	}

	return servers
}

func getWhiteList(service rancherData) *types.WhiteList {
	ranges := label.GetSliceStringValue(service.Labels, label.TraefikFrontendWhiteListSourceRange)

	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: label.GetBoolValue(service.Labels, label.TraefikFrontendWhiteListUseXForwardedFor, false),
		}
	}

	return nil
}

func getRedirect(service rancherData) *types.Redirect {
	permanent := label.GetBoolValue(service.Labels, label.TraefikFrontendRedirectPermanent, false)

	if label.Has(service.Labels, label.TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: label.GetStringValue(service.Labels, label.TraefikFrontendRedirectEntryPoint, ""),
			Permanent:  permanent,
		}
	}

	if label.Has(service.Labels, label.TraefikFrontendRedirectRegex) &&
		label.Has(service.Labels, label.TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       label.GetStringValue(service.Labels, label.TraefikFrontendRedirectRegex, ""),
			Replacement: label.GetStringValue(service.Labels, label.TraefikFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func getErrorPages(service rancherData) map[string]*types.ErrorPage {
	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(service.Labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimit(service rancherData) *types.RateLimit {
	extractorFunc := label.GetStringValue(service.Labels, label.TraefikFrontendRateLimitExtractorFunc, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	prefix := label.Prefix + label.BaseFrontendRateLimit
	limits := label.ParseRateSets(service.Labels, prefix, label.RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func getHeaders(service rancherData) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    label.GetMapValue(service.Labels, label.TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   label.GetMapValue(service.Labels, label.TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         label.GetMapValue(service.Labels, label.TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            label.GetSliceStringValue(service.Labels, label.TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       label.GetSliceStringValue(service.Labels, label.TraefikFrontendHostsProxyHeaders),
		STSSeconds:              label.GetInt64Value(service.Labels, label.TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             label.GetBoolValue(service.Labels, label.TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    label.GetBoolValue(service.Labels, label.TraefikFrontendSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    label.GetBoolValue(service.Labels, label.TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              label.GetBoolValue(service.Labels, label.TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          label.GetBoolValue(service.Labels, label.TraefikFrontendForceSTSHeader, false),
		FrameDeny:               label.GetBoolValue(service.Labels, label.TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      label.GetBoolValue(service.Labels, label.TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        label.GetBoolValue(service.Labels, label.TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           label.GetBoolValue(service.Labels, label.TraefikFrontendIsDevelopment, false),
		SSLHost:                 label.GetStringValue(service.Labels, label.TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: label.GetStringValue(service.Labels, label.TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   label.GetStringValue(service.Labels, label.TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               label.GetStringValue(service.Labels, label.TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          label.GetStringValue(service.Labels, label.TraefikFrontendReferrerPolicy, ""),
		CustomBrowserXSSValue:   label.GetStringValue(service.Labels, label.TraefikFrontendCustomBrowserXSSValue, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Label functions

func getFuncString(labelName string, defaultValue string) func(service rancherData) string {
	return func(service rancherData) string {
		return label.GetStringValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncBool(labelName string, defaultValue bool) func(service rancherData) bool {
	return func(service rancherData) bool {
		return label.GetBoolValue(service.Labels, labelName, defaultValue)
	}
}

func getFuncInt(labelName string, defaultValue int) func(service rancherData) int {
	return func(service rancherData) int {
		return label.GetIntValue(service.Labels, labelName, defaultValue)
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
