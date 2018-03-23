package consulcatalog

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

func (p *Provider) buildConfiguration(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,

		// Backend functions
		"getBackend":              getNodeBackendName, // TODO Deprecated [breaking] getBackend -> getNodeBackendName
		"getNodeBackendName":      getNodeBackendName,
		"getServiceBackendName":   getServiceBackendName,
		"getBackendAddress":       getBackendAddress,
		"getBackendName":          getServerName, // TODO Deprecated [breaking] getBackendName -> getServerName
		"getServerName":           getServerName,
		"hasMaxconnAttributes":    p.hasMaxConnAttributes,    // TODO Deprecated [breaking]
		"getSticky":               p.getSticky,               // TODO Deprecated [breaking]
		"hasStickinessLabel":      p.hasStickinessLabel,      // TODO Deprecated [breaking]
		"getStickinessCookieName": p.getStickinessCookieName, // TODO Deprecated [breaking]
		"getWeight":               p.getWeight,               // TODO Deprecated [breaking] Must replaced by a simple: "getWeight": p.getFuncIntAttribute(label.SuffixWeight, 0)
		"getProtocol":             p.getFuncStringAttribute(label.SuffixProtocol, label.DefaultProtocol),
		"getCircuitBreaker":       p.getCircuitBreaker,
		"getLoadBalancer":         p.getLoadBalancer,
		"getMaxConn":              p.getMaxConn,
		"getHealthCheck":          p.getHealthCheck,
		"getBuffering":            p.getBuffering,

		// Frontend functions
		"getFrontendRule":        p.getFrontendRule,
		"getBasicAuth":           p.getFuncSliceAttribute(label.SuffixFrontendAuthBasic),
		"getEntryPoints":         getEntryPoints,                                           // TODO Deprecated [breaking]
		"getFrontEndEntryPoints": p.getFuncSliceAttribute(label.SuffixFrontendEntryPoints), // TODO [breaking] rename to getEntryPoints when getEntryPoints will be removed
		"getPriority":            p.getFuncIntAttribute(label.SuffixFrontendPriority, label.DefaultFrontendPriorityInt),
		"getPassHostHeader":      p.getFuncBoolAttribute(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":         p.getFuncBoolAttribute(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getWhiteList":           p.getWhiteList,
		"getRedirect":            p.getRedirect,
		"hasErrorPages":          p.getFuncHasAttributePrefix(label.BaseFrontendErrorPage),
		"getErrorPages":          p.getErrorPages,
		"hasRateLimit":           p.getFuncHasAttributePrefix(label.BaseFrontendRateLimit),
		"getRateLimit":           p.getRateLimit,
		"getHeaders":             p.getHeaders,
	}

	var allNodes []*api.ServiceEntry
	var services []*serviceUpdate
	for _, info := range catalog {
		if len(info.Nodes) > 0 {
			services = append(services, info.Service)
			allNodes = append(allNodes, info.Nodes...)
		}
	}
	// Ensure a stable ordering of nodes so that identical configurations may be detected
	sort.Sort(nodeSorter(allNodes))

	templateObjects := struct {
		Services []*serviceUpdate
		Nodes    []*api.ServiceEntry
	}{
		Services: services,
		Nodes:    allNodes,
	}

	configuration, err := p.GetConfiguration("templates/consul_catalog.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

func (p *Provider) setupFrontEndRuleTemplate() {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,
	}
	tmpl := template.New("consul catalog frontend rule").Funcs(FuncMap)
	p.frontEndRuleTemplate = tmpl
}

// Specific functions

func (p *Provider) getFrontendRule(service serviceUpdate) string {
	customFrontendRule := p.getAttribute(label.SuffixFrontendRule, service.Attributes, "")
	if customFrontendRule == "" {
		customFrontendRule = p.FrontEndRule
	}

	tmpl := p.frontEndRuleTemplate
	tmpl, err := tmpl.Parse(customFrontendRule)
	if err != nil {
		log.Errorf("Failed to parse Consul Catalog custom frontend rule: %v", err)
		return ""
	}

	templateObjects := struct {
		ServiceName string
		Domain      string
		Attributes  []string
	}{
		ServiceName: service.ServiceName,
		Domain:      p.Domain,
		Attributes:  service.Attributes,
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Errorf("Failed to execute Consul Catalog custom frontend rule template: %v", err)
		return ""
	}

	return buffer.String()
}

// Deprecated
func (p *Provider) hasMaxConnAttributes(attributes []string) bool {
	amount := p.getAttribute(label.SuffixBackendMaxConnAmount, attributes, "")
	extractorFunc := p.getAttribute(label.SuffixBackendMaxConnExtractorFunc, attributes, "")
	return amount != "" && extractorFunc != ""
}

// Deprecated
func getEntryPoints(list string) []string {
	return strings.Split(list, ",")
}

func getNodeBackendName(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func getServiceBackendName(service *serviceUpdate) string {
	return strings.ToLower(service.ServiceName)
}

func getBackendAddress(node *api.ServiceEntry) string {
	if node.Service.Address != "" {
		return node.Service.Address
	}
	return node.Node.Address
}

func getServerName(node *api.ServiceEntry, index int) string {
	serviceName := node.Service.Service + node.Service.Address + strconv.Itoa(node.Service.Port)
	// TODO sort tags ?
	serviceName += strings.Join(node.Service.Tags, "")

	hash := sha1.New()
	_, err := hash.Write([]byte(serviceName))
	if err != nil {
		// Impossible case
		log.Error(err)
	} else {
		serviceName = base64.URLEncoding.EncodeToString(hash.Sum(nil))
	}

	// unique int at the end
	return provider.Normalize(node.Service.Service + "-" + strconv.Itoa(index) + "-" + serviceName)
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func (p *Provider) getSticky(tags []string) string {
	stickyTag := p.getAttribute(label.SuffixBackendLoadBalancerSticky, tags, "")
	if len(stickyTag) > 0 {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	} else {
		stickyTag = "false"
	}
	return stickyTag
}

// Deprecated
func (p *Provider) hasStickinessLabel(tags []string) bool {
	stickinessTag := p.getAttribute(label.SuffixBackendLoadBalancerStickiness, tags, "")
	return len(stickinessTag) > 0 && strings.EqualFold(strings.TrimSpace(stickinessTag), "true")
}

// Deprecated
func (p *Provider) getStickinessCookieName(tags []string) string {
	return p.getAttribute(label.SuffixBackendLoadBalancerStickinessCookieName, tags, "")
}

// Deprecated
func (p *Provider) getWeight(tags []string) int {
	weight := p.getIntAttribute(label.SuffixWeight, tags, label.DefaultWeightInt)

	// Deprecated
	deprecatedWeightTag := "backend." + label.SuffixWeight
	if p.hasAttribute(deprecatedWeightTag, tags) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.",
			p.getPrefixedName(deprecatedWeightTag), p.getPrefixedName(label.SuffixWeight))

		weight = p.getIntAttribute(deprecatedWeightTag, tags, label.DefaultWeightInt)
	}

	return weight
}

func (p *Provider) getCircuitBreaker(tags []string) *types.CircuitBreaker {
	circuitBreaker := p.getAttribute(label.SuffixBackendCircuitBreakerExpression, tags, "")

	if p.hasAttribute(label.SuffixBackendCircuitBreaker, tags) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.",
			p.getPrefixedName(label.SuffixBackendCircuitBreaker), p.getPrefixedName(label.SuffixBackendCircuitBreakerExpression))

		circuitBreaker = p.getAttribute(label.SuffixBackendCircuitBreaker, tags, "")
	}

	if len(circuitBreaker) == 0 {
		return nil
	}

	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func (p *Provider) getLoadBalancer(tags []string) *types.LoadBalancer {
	rawSticky := p.getSticky(tags)
	sticky, err := strconv.ParseBool(rawSticky)
	if err != nil {
		log.Debugf("Invalid sticky value: %s", rawSticky)
		sticky = false
	}

	method := p.getAttribute(label.SuffixBackendLoadBalancerMethod, tags, label.DefaultBackendLoadBalancerMethod)

	// Deprecated
	deprecatedMethodTag := "backend.loadbalancer"
	if p.hasAttribute(deprecatedMethodTag, tags) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.",
			p.getPrefixedName(deprecatedMethodTag), p.getPrefixedName(label.SuffixWeight))

		method = p.getAttribute(deprecatedMethodTag, tags, label.SuffixBackendLoadBalancerMethod)
	}

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: sticky,
	}

	if p.getBoolAttribute(label.SuffixBackendLoadBalancerStickiness, tags, false) {
		lb.Stickiness = &types.Stickiness{
			CookieName: p.getAttribute(label.SuffixBackendLoadBalancerStickinessCookieName, tags, ""),
		}
	}

	return lb
}

func (p *Provider) getMaxConn(tags []string) *types.MaxConn {
	amount := p.getInt64Attribute(label.SuffixBackendMaxConnAmount, tags, math.MinInt64)
	extractorFunc := p.getAttribute(label.SuffixBackendMaxConnExtractorFunc, tags, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func (p *Provider) getHealthCheck(tags []string) *types.HealthCheck {
	path := p.getAttribute(label.SuffixBackendHealthCheckPath, tags, "")

	if len(path) == 0 {
		return nil
	}

	port := p.getIntAttribute(label.SuffixBackendHealthCheckPort, tags, label.DefaultBackendHealthCheckPort)
	interval := p.getAttribute(label.SuffixBackendHealthCheckInterval, tags, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func (p *Provider) getBuffering(tags []string) *types.Buffering {
	if !p.hasAttributePrefix(label.SuffixBackendBuffering, tags) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  p.getInt64Attribute(label.SuffixBackendBufferingMaxRequestBodyBytes, tags, 0),
		MaxResponseBodyBytes: p.getInt64Attribute(label.SuffixBackendBufferingMaxResponseBodyBytes, tags, 0),
		MemRequestBodyBytes:  p.getInt64Attribute(label.SuffixBackendBufferingMemRequestBodyBytes, tags, 0),
		MemResponseBodyBytes: p.getInt64Attribute(label.SuffixBackendBufferingMemResponseBodyBytes, tags, 0),
		RetryExpression:      p.getAttribute(label.SuffixBackendBufferingRetryExpression, tags, ""),
	}
}

func (p *Provider) getWhiteList(tags []string) *types.WhiteList {
	ranges := p.getSliceAttribute(label.SuffixFrontendWhiteListSourceRange, tags)

	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: p.getBoolAttribute(label.SuffixFrontendWhiteListUseXForwardedFor, tags, false),
		}
	}

	return nil
}

func (p *Provider) getRedirect(tags []string) *types.Redirect {
	permanent := p.getBoolAttribute(label.SuffixFrontendRedirectPermanent, tags, false)

	if p.hasAttribute(label.SuffixFrontendRedirectEntryPoint, tags) {
		return &types.Redirect{
			EntryPoint: p.getAttribute(label.SuffixFrontendRedirectEntryPoint, tags, ""),
			Permanent:  permanent,
		}
	}

	if p.hasAttribute(label.SuffixFrontendRedirectRegex, tags) && p.hasAttribute(label.SuffixFrontendRedirectReplacement, tags) {
		return &types.Redirect{
			Regex:       p.getAttribute(label.SuffixFrontendRedirectRegex, tags, ""),
			Replacement: p.getAttribute(label.SuffixFrontendRedirectReplacement, tags, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func (p *Provider) getErrorPages(tags []string) map[string]*types.ErrorPage {
	labels := p.parseTagsToNeutralLabels(tags)

	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(labels, prefix, label.RegexpFrontendErrorPage)
}

func (p *Provider) getRateLimit(tags []string) *types.RateLimit {
	extractorFunc := p.getAttribute(label.SuffixFrontendRateLimitExtractorFunc, tags, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	labels := p.parseTagsToNeutralLabels(tags)

	prefix := label.Prefix + label.BaseFrontendRateLimit
	limits := label.ParseRateSets(labels, prefix, label.RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func (p *Provider) getHeaders(tags []string) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    p.getMapAttribute(label.SuffixFrontendRequestHeaders, tags),
		CustomResponseHeaders:   p.getMapAttribute(label.SuffixFrontendResponseHeaders, tags),
		SSLProxyHeaders:         p.getMapAttribute(label.SuffixFrontendHeadersSSLProxyHeaders, tags),
		AllowedHosts:            p.getSliceAttribute(label.SuffixFrontendHeadersAllowedHosts, tags),
		HostsProxyHeaders:       p.getSliceAttribute(label.SuffixFrontendHeadersHostsProxyHeaders, tags),
		SSLHost:                 p.getAttribute(label.SuffixFrontendHeadersSSLHost, tags, ""),
		CustomFrameOptionsValue: p.getAttribute(label.SuffixFrontendHeadersCustomFrameOptionsValue, tags, ""),
		ContentSecurityPolicy:   p.getAttribute(label.SuffixFrontendHeadersContentSecurityPolicy, tags, ""),
		PublicKey:               p.getAttribute(label.SuffixFrontendHeadersPublicKey, tags, ""),
		ReferrerPolicy:          p.getAttribute(label.SuffixFrontendHeadersReferrerPolicy, tags, ""),
		CustomBrowserXSSValue:   p.getAttribute(label.SuffixFrontendHeadersCustomBrowserXSSValue, tags, ""),
		STSSeconds:              p.getInt64Attribute(label.SuffixFrontendHeadersSTSSeconds, tags, 0),
		SSLRedirect:             p.getBoolAttribute(label.SuffixFrontendHeadersSSLRedirect, tags, false),
		SSLTemporaryRedirect:    p.getBoolAttribute(label.SuffixFrontendHeadersSSLTemporaryRedirect, tags, false),
		STSIncludeSubdomains:    p.getBoolAttribute(label.SuffixFrontendHeadersSTSIncludeSubdomains, tags, false),
		STSPreload:              p.getBoolAttribute(label.SuffixFrontendHeadersSTSPreload, tags, false),
		ForceSTSHeader:          p.getBoolAttribute(label.SuffixFrontendHeadersForceSTSHeader, tags, false),
		FrameDeny:               p.getBoolAttribute(label.SuffixFrontendHeadersFrameDeny, tags, false),
		ContentTypeNosniff:      p.getBoolAttribute(label.SuffixFrontendHeadersContentTypeNosniff, tags, false),
		BrowserXSSFilter:        p.getBoolAttribute(label.SuffixFrontendHeadersBrowserXSSFilter, tags, false),
		IsDevelopment:           p.getBoolAttribute(label.SuffixFrontendHeadersIsDevelopment, tags, false),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Base functions

func (p *Provider) parseTagsToNeutralLabels(tags []string) map[string]string {
	var labels map[string]string

	for _, tag := range tags {
		if strings.HasPrefix(tag, p.Prefix) {

			parts := strings.SplitN(tag, "=", 2)
			if len(parts) == 2 {
				if labels == nil {
					labels = make(map[string]string)
				}

				// replace custom prefix by the generic prefix
				key := label.Prefix + strings.TrimPrefix(parts[0], p.Prefix+".")
				labels[key] = parts[1]
			}
		}
	}

	return labels
}

func (p *Provider) getFuncStringAttribute(name string, defaultValue string) func(tags []string) string {
	return func(tags []string) string {
		return p.getAttribute(name, tags, defaultValue)
	}
}

func (p *Provider) getFuncSliceAttribute(name string) func(tags []string) []string {
	return func(tags []string) []string {
		return p.getSliceAttribute(name, tags)
	}
}

func (p *Provider) getMapAttribute(name string, tags []string) map[string]string {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return nil
	}

	return label.ParseMapValue(p.getPrefixedName(name), rawValue)
}

func (p *Provider) getFuncIntAttribute(name string, defaultValue int) func(tags []string) int {
	return func(tags []string) int {
		return p.getIntAttribute(name, tags, defaultValue)
	}
}

func (p *Provider) getFuncBoolAttribute(name string, defaultValue bool) func(tags []string) bool {
	return func(tags []string) bool {
		return p.getBoolAttribute(name, tags, defaultValue)
	}
}

func (p *Provider) getFuncHasAttributePrefix(name string) func(tags []string) bool {
	return func(tags []string) bool {
		return p.hasAttributePrefix(name, tags)
	}
}

func (p *Provider) getInt64Attribute(name string, tags []string, defaultValue int64) int64 {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) getIntAttribute(name string, tags []string, defaultValue int) int {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) getSliceAttribute(name string, tags []string) []string {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return nil
	}
	return label.SplitAndTrimString(rawValue, ",")
}

func (p *Provider) getBoolAttribute(name string, tags []string, defaultValue bool) bool {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) hasAttribute(name string, tags []string) bool {
	return hasTag(p.getPrefixedName(name), tags)
}

func (p *Provider) hasAttributePrefix(name string, tags []string) bool {
	return hasTagPrefix(p.getPrefixedName(name), tags)
}

func (p *Provider) getAttribute(name string, tags []string, defaultValue string) string {
	return getTag(p.getPrefixedName(name), tags, defaultValue)
}

func (p *Provider) getPrefixedName(name string) string {
	if len(p.Prefix) > 0 && len(name) > 0 {
		return p.Prefix + "." + name
	}
	return name
}

func hasTag(name string, tags []string) bool {
	lowerName := strings.ToLower(name)

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)

		// Given the nature of Consul tags, which could be either singular markers, or key=value pairs
		if strings.HasPrefix(lowerTag, lowerName+"=") || lowerTag == lowerName {
			return true
		}
	}
	return false
}

func hasTagPrefix(name string, tags []string) bool {
	lowerName := strings.ToLower(name)

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)

		if strings.HasPrefix(lowerTag, lowerName) {
			return true
		}
	}
	return false
}

func getTag(name string, tags []string, defaultValue string) string {
	lowerName := strings.ToLower(name)

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)

		// Given the nature of Consul tags, which could be either singular markers, or key=value pairs
		if strings.HasPrefix(lowerTag, lowerName+"=") || lowerTag == lowerName {
			// In case, where a tag might be a key=value, try to split it by the first '='
			kv := strings.SplitN(tag, "=", 2)

			// If the returned result is a key=value pair, return the 'value' component
			if len(kv) == 2 {
				return kv[1]
			}
			// If the returned result is a singular marker, return the 'key' component
			return kv[0]
		}

	}
	return defaultValue
}
