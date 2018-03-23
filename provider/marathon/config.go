package marathon

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
)

const defaultService = ""

func (p *Provider) buildConfiguration() *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getBackend":   p.getBackend,
		"getDomain":    getFuncStringService(label.SuffixDomain, p.Domain), // see https://github.com/containous/traefik/pull/1693
		"getSubDomain": p.getSubDomain,                                     // see https://github.com/containous/traefik/pull/1693

		// Backend functions
		"getBackendServer":  p.getBackendServer,
		"getPort":           getPort,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getBuffering":      getBuffering,
		"getServers":        p.getServers,

		// TODO Deprecated [breaking]
		"getWeight": getFuncIntService(label.SuffixWeight, label.DefaultWeightInt),
		// TODO Deprecated [breaking]
		"getProtocol": getFuncStringService(label.SuffixProtocol, label.DefaultProtocol),
		// TODO Deprecated [breaking]
		"hasCircuitBreakerLabels": hasFunc(label.TraefikBackendCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"getCircuitBreakerExpression": getFuncString(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"hasLoadBalancerLabels": hasLoadBalancerLabels,
		// TODO Deprecated [breaking]
		"getLoadBalancerMethod": getFuncString(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		// TODO Deprecated [breaking]
		"getSticky": getSticky,
		// TODO Deprecated [breaking]
		"hasStickinessLabel": hasFunc(label.TraefikBackendLoadBalancerStickiness),
		// TODO Deprecated [breaking]
		"getStickinessCookieName": getFuncString(label.TraefikBackendLoadBalancerStickinessCookieName, ""),
		// TODO Deprecated [breaking]
		"hasMaxConnLabels": hasMaxConnLabels,
		// TODO Deprecated [breaking]
		"getMaxConnExtractorFunc": getFuncString(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		// TODO Deprecated [breaking]
		"getMaxConnAmount": getFuncInt64(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		// TODO Deprecated [breaking]
		"hasHealthCheckLabels": hasFunc(label.TraefikBackendHealthCheckPath),
		// TODO Deprecated [breaking]
		"getHealthCheckPath": getFuncString(label.TraefikBackendHealthCheckPath, ""),
		// TODO Deprecated [breaking]
		"getHealthCheckInterval": getFuncString(label.TraefikBackendHealthCheckInterval, ""),

		// Frontend functions
		"getServiceNames":      getServiceNames,
		"getServiceNameSuffix": getServiceNameSuffix,
		"getPassHostHeader":    getFuncBoolService(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":       getFuncBoolService(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":          getFuncIntService(label.SuffixFrontendPriority, label.DefaultFrontendPriorityInt),
		"getEntryPoints":       getFuncSliceStringService(label.SuffixFrontendEntryPoints),
		"getFrontendRule":      p.getFrontendRule,
		"getFrontendName":      p.getFrontendName,
		"getBasicAuth":         getFuncSliceStringService(label.SuffixFrontendAuthBasic),
		"getRedirect":          getRedirect,
		"getErrorPages":        getErrorPages,
		"getRateLimit":         getRateLimit,
		"getHeaders":           getHeaders,
		"getWhiteList":         getWhiteList,

		// TODO Deprecated [breaking]
		"getWhitelistSourceRange": getFuncSliceStringService(label.SuffixFrontendWhitelistSourceRange),
	}

	v := url.Values{}
	v.Add("embed", "apps.tasks")
	v.Add("embed", "apps.deployments")
	v.Add("embed", "apps.readiness")
	applications, err := p.marathonClient.Applications(v)
	if err != nil {
		log.Errorf("Failed to retrieve Marathon applications: %v", err)
		return nil
	}

	filteredApps := fun.Filter(p.applicationFilter, applications.Apps).([]marathon.Application)
	for i, app := range filteredApps {
		filteredApps[i].Tasks = fun.Filter(func(task *marathon.Task) bool {
			filtered := p.taskFilter(*task, app)
			if filtered {
				logIllegalServices(*task, app)
			}
			return filtered
		}, app.Tasks).([]*marathon.Task)
	}

	templateObjects := struct {
		Applications []marathon.Application
		Domain       string
	}{
		Applications: filteredApps,
		Domain:       p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/marathon.tmpl", MarathonFuncMap, templateObjects)
	if err != nil {
		log.Errorf("Failed to render Marathon configuration template: %v", err)
	}
	return configuration
}

func (p *Provider) applicationFilter(app marathon.Application) bool {
	// Filter disabled application.
	if !label.IsEnabledP(app.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled Marathon application %s", app.ID)
		return false
	}

	// Filter by constraints.
	constraintTags := label.GetSliceStringValueP(app.Labels, label.TraefikTags)
	if p.MarathonLBCompatibility {
		if haGroup := label.GetStringValueP(app.Labels, labelLbCompatibilityGroup, ""); len(haGroup) > 0 {
			constraintTags = append(constraintTags, haGroup)
		}
	}
	if p.FilterMarathonConstraints && app.Constraints != nil {
		for _, constraintParts := range *app.Constraints {
			constraintTags = append(constraintTags, strings.Join(constraintParts, ":"))
		}
	}
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering Marathon application %s pruned by %q constraint", app.ID, failingConstraint.String())
		}
		return false
	}

	return true
}

func (p *Provider) taskFilter(task marathon.Task, application marathon.Application) bool {
	if task.State != string(taskStateRunning) {
		return false
	}

	if ready := p.readyChecker.Do(task, application); !ready {
		log.Infof("Filtering unready task %s from application %s", task.ID, application.ID)
		return false
	}

	return true
}

// getFrontendRule returns the frontend rule for the specified application, using
// its label. If service is provided, it will look for serviceName label before generic one.
// It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(application marathon.Application, serviceName string) string {
	labels := getLabels(application, serviceName)
	lblFrontendRule := getLabelName(serviceName, label.SuffixFrontendRule)
	if value := label.GetStringValue(labels, lblFrontendRule, ""); len(value) > 0 {
		return value
	}

	if p.MarathonLBCompatibility {
		if value := label.GetStringValueP(application.Labels, labelLbCompatibility, ""); len(value) > 0 {
			return "Host:" + value
		}
	}

	if len(serviceName) > 0 {
		return "Host:" + strings.ToLower(provider.Normalize(serviceName)) + "." + p.getSubDomain(application.ID) + "." + p.Domain
	}
	return "Host:" + p.getSubDomain(application.ID) + "." + p.Domain
}

func (p *Provider) getBackend(application marathon.Application, serviceName string) string {
	labels := getLabels(application, serviceName)
	lblBackend := getLabelName(serviceName, label.SuffixBackend)
	value := label.GetStringValue(labels, lblBackend, "")
	if len(value) > 0 {
		return provider.Normalize("backend" + value)
	}
	return provider.Normalize("backend" + application.ID + getServiceNameSuffix(serviceName))
}

func (p *Provider) getFrontendName(application marathon.Application, serviceName string) string {
	return provider.Normalize("frontend" + application.ID + getServiceNameSuffix(serviceName))
}

func (p *Provider) getSubDomain(name string) string {
	if p.GroupsAsSubDomains {
		splitedName := strings.Split(strings.TrimPrefix(name, "/"), "/")
		provider.ReverseStringSlice(&splitedName)
		reverseName := strings.Join(splitedName, ".")
		return reverseName
	}
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}

func (p *Provider) getBackendServer(task marathon.Task, application marathon.Application) string {
	if application.IPAddressPerTask == nil || p.ForceTaskHostname {
		return task.Host
	}

	numTaskIPAddresses := len(task.IPAddresses)
	switch numTaskIPAddresses {
	case 0:
		log.Errorf("Missing IP address for Marathon application %s on task %s", application.ID, task.ID)
		return ""
	case 1:
		return task.IPAddresses[0].IPAddress
	default:
		ipAddressIdx := label.GetIntValueP(application.Labels, labelIPAddressIdx, math.MinInt32)

		if ipAddressIdx == math.MinInt32 {
			log.Errorf("Found %d task IP addresses but missing IP address index for Marathon application %s on task %s",
				numTaskIPAddresses, application.ID, task.ID)
			return ""
		}
		if ipAddressIdx < 0 || ipAddressIdx > numTaskIPAddresses {
			log.Errorf("Cannot use IP address index to select from %d task IP addresses for Marathon application %s on task %s",
				numTaskIPAddresses, application.ID, task.ID)
			return ""
		}

		return task.IPAddresses[ipAddressIdx].IPAddress
	}
}

func identifier(app marathon.Application, task marathon.Task, serviceName string) string {
	id := fmt.Sprintf("Marathon task %s from application %s", task.ID, app.ID)
	if serviceName != "" {
		id += fmt.Sprintf(" (service: %s)", serviceName)
	}
	return id
}

// getServiceNames returns a list of service names for a given application
// An empty name "" will be added if no service specific properties exist,
// as an indication that there are no sub-services, but only main application
func getServiceNames(application marathon.Application) []string {
	labelServiceProperties := label.ExtractServicePropertiesP(application.Labels)

	var names []string
	for k := range labelServiceProperties {
		names = append(names, k)
	}

	// An empty name "" will be added if no service specific properties exist,
	// as an indication that there are no sub-services, but only main application
	if len(names) == 0 {
		names = append(names, defaultService)
	}
	return names
}

func getServiceNameSuffix(serviceName string) string {
	if len(serviceName) > 0 {
		return "-service-" + provider.Normalize(serviceName)
	}
	return ""
}

// logIllegalServices logs illegal service configurations.
// While we cannot filter on the service level, they will eventually get
// rejected once the server configuration is rendered.
func logIllegalServices(task marathon.Task, application marathon.Application) {
	for _, serviceName := range getServiceNames(application) {
		// Check for illegal/missing ports.
		if _, err := processPorts(application, task, serviceName); err != nil {
			log.Warnf("%s has an illegal configuration: no proper port available", identifier(application, task, serviceName))
			continue
		}

		// Check for illegal port label combinations.
		labels := getLabels(application, serviceName)
		hasPortLabel := label.Has(labels, getLabelName(serviceName, label.SuffixPort))
		hasPortIndexLabel := label.Has(labels, getLabelName(serviceName, label.SuffixPortIndex))
		if hasPortLabel && hasPortIndexLabel {
			log.Warnf("%s has both port and port index specified; port will take precedence", identifier(application, task, serviceName))
		}
	}
}

// Deprecated
func hasLoadBalancerLabels(application marathon.Application) bool {
	method := label.HasP(application.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.HasP(application.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.HasP(application.Labels, label.TraefikBackendLoadBalancerStickiness)
	return method || sticky || stickiness
}

// Deprecated
func hasMaxConnLabels(application marathon.Application) bool {
	mca := label.HasP(application.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.HasP(application.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(application marathon.Application) bool {
	if label.HasP(application.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return label.GetBoolValueP(application.Labels, label.TraefikBackendLoadBalancerSticky, false)
}

func getPort(task marathon.Task, application marathon.Application, serviceName string) string {
	port, err := processPorts(application, task, serviceName)
	if err != nil {
		log.Errorf("Unable to process ports for %s: %s", identifier(application, task, serviceName), err)
		return ""
	}

	return strconv.Itoa(port)
}

// processPorts returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
func processPorts(application marathon.Application, task marathon.Task, serviceName string) (int, error) {
	labels := getLabels(application, serviceName)
	lblPort := getLabelName(serviceName, label.SuffixPort)

	if label.Has(labels, lblPort) {
		port := label.GetIntValue(labels, lblPort, 0)

		if port <= 0 {
			return 0, fmt.Errorf("explicitly specified port %d must be larger than zero", port)
		} else if port > 0 {
			return port, nil
		}
	}

	ports := retrieveAvailablePorts(application, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	lblPortIndex := getLabelName(serviceName, label.SuffixPortIndex)
	portIndex := label.GetIntValue(labels, lblPortIndex, 0)
	if portIndex < 0 || portIndex > len(ports)-1 {
		return 0, fmt.Errorf("index %d must be within range (0, %d)", portIndex, len(ports)-1)
	}
	return ports[portIndex], nil
}

func retrieveAvailablePorts(application marathon.Application, task marathon.Task) []int {
	// Using default port configuration
	if len(task.Ports) > 0 {
		return task.Ports
	}

	// Using port definition if available
	if application.PortDefinitions != nil && len(*application.PortDefinitions) > 0 {
		var ports []int
		for _, def := range *application.PortDefinitions {
			if def.Port != nil {
				ports = append(ports, *def.Port)
			}
		}
		return ports
	}
	// If using IP-per-task using this port definition
	if application.IPAddressPerTask != nil && len(*(application.IPAddressPerTask.Discovery).Ports) > 0 {
		var ports []int
		for _, def := range *(application.IPAddressPerTask.Discovery).Ports {
			ports = append(ports, def.Number)
		}
		return ports
	}

	return []int{}
}

func getCircuitBreaker(application marathon.Application) *types.CircuitBreaker {
	circuitBreaker := label.GetStringValueP(application.Labels, label.TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func getLoadBalancer(application marathon.Application) *types.LoadBalancer {
	if !label.HasPrefixP(application.Labels, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := label.GetStringValueP(application.Labels, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getSticky(application),
	}

	if label.GetBoolValueP(application.Labels, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := label.GetStringValueP(application.Labels, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

func getMaxConn(application marathon.Application) *types.MaxConn {
	amount := label.GetInt64ValueP(application.Labels, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := label.GetStringValueP(application.Labels, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getHealthCheck(application marathon.Application) *types.HealthCheck {
	path := label.GetStringValueP(application.Labels, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := label.GetIntValueP(application.Labels, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort)
	interval := label.GetStringValueP(application.Labels, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getBuffering(application marathon.Application) *types.Buffering {
	if !label.HasPrefixP(application.Labels, label.TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  label.GetInt64ValueP(application.Labels, label.TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: label.GetInt64ValueP(application.Labels, label.TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  label.GetInt64ValueP(application.Labels, label.TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: label.GetInt64ValueP(application.Labels, label.TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      label.GetStringValueP(application.Labels, label.TraefikBackendBufferingRetryExpression, ""),
	}
}

func (p *Provider) getServers(application marathon.Application, serviceName string) map[string]types.Server {
	var servers map[string]types.Server

	for _, task := range application.Tasks {
		host := p.getBackendServer(*task, application)
		if len(host) == 0 {
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		labels := getLabels(application, serviceName)

		port := getPort(*task, application, serviceName)
		protocol := label.GetStringValue(labels, getLabelName(serviceName, label.SuffixProtocol), label.DefaultProtocol)

		serverName := provider.Normalize("server-" + task.ID + getServiceNameSuffix(serviceName))
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%v", protocol, host, port),
			Weight: label.GetIntValue(labels, getLabelName(serviceName, label.SuffixWeight), label.DefaultWeightInt),
		}
	}

	return servers
}

func getWhiteList(application marathon.Application, serviceName string) *types.WhiteList {
	labels := getLabels(application, serviceName)

	ranges := label.GetSliceStringValue(labels, getLabelName(serviceName, label.SuffixFrontendWhiteListSourceRange))
	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendWhiteListUseXForwardedFor), false),
		}
	}

	return nil
}

func getRedirect(application marathon.Application, serviceName string) *types.Redirect {
	labels := getLabels(application, serviceName)

	permanent := label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendRedirectPermanent), false)

	if label.Has(labels, getLabelName(serviceName, label.SuffixFrontendRedirectEntryPoint)) {
		return &types.Redirect{
			EntryPoint: label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendRedirectEntryPoint), ""),
			Permanent:  permanent,
		}
	}

	if label.Has(labels, getLabelName(serviceName, label.SuffixFrontendRedirectRegex)) &&
		label.Has(labels, getLabelName(serviceName, label.SuffixFrontendRedirectReplacement)) {
		return &types.Redirect{
			Regex:       label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendRedirectRegex), ""),
			Replacement: label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendRedirectReplacement), ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func getErrorPages(application marathon.Application, serviceName string) map[string]*types.ErrorPage {
	labels := getLabels(application, serviceName)
	prefix := getLabelName(serviceName, label.BaseFrontendErrorPage)

	if len(serviceName) > 0 {
		return label.ParseErrorPages(labels, prefix, label.RegexpBaseFrontendErrorPage)
	}
	return label.ParseErrorPages(labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimit(application marathon.Application, serviceName string) *types.RateLimit {
	labels := getLabels(application, serviceName)

	extractorFunc := label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendRateLimitExtractorFunc), "")
	if len(extractorFunc) == 0 {
		return nil
	}

	limits := getRateSet(labels, serviceName)
	if len(limits) == 0 {
		return nil
	}

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func getRateSet(labels map[string]string, serviceName string) map[string]*types.Rate {
	rateSetPrefix := getLabelName(serviceName, label.BaseFrontendRateLimit)

	if len(serviceName) > 0 {
		return label.ParseRateSets(labels, rateSetPrefix, label.RegexpBaseFrontendRateLimit)
	}
	return label.ParseRateSets(labels, rateSetPrefix, label.RegexpFrontendRateLimit)
}

func getHeaders(application marathon.Application, serviceName string) *types.Headers {
	labels := getLabels(application, serviceName)

	headers := &types.Headers{
		CustomRequestHeaders:    label.GetMapValue(labels, getLabelName(serviceName, label.SuffixFrontendRequestHeaders)),
		CustomResponseHeaders:   label.GetMapValue(labels, getLabelName(serviceName, label.SuffixFrontendResponseHeaders)),
		SSLProxyHeaders:         label.GetMapValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSSLProxyHeaders)),
		AllowedHosts:            label.GetSliceStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersAllowedHosts)),
		HostsProxyHeaders:       label.GetSliceStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersHostsProxyHeaders)),
		STSSeconds:              label.GetInt64Value(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSTSSeconds), 0),
		SSLRedirect:             label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSSLRedirect), false),
		SSLTemporaryRedirect:    label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSSLTemporaryRedirect), false),
		STSIncludeSubdomains:    label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSTSIncludeSubdomains), false),
		STSPreload:              label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSTSPreload), false),
		ForceSTSHeader:          label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersForceSTSHeader), false),
		FrameDeny:               label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersFrameDeny), false),
		ContentTypeNosniff:      label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersContentTypeNosniff), false),
		BrowserXSSFilter:        label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersBrowserXSSFilter), false),
		IsDevelopment:           label.GetBoolValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersIsDevelopment), false),
		SSLHost:                 label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersSSLHost), ""),
		CustomFrameOptionsValue: label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersCustomFrameOptionsValue), ""),
		ContentSecurityPolicy:   label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersContentSecurityPolicy), ""),
		PublicKey:               label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersPublicKey), ""),
		ReferrerPolicy:          label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersReferrerPolicy), ""),
		CustomBrowserXSSValue:   label.GetStringValue(labels, getLabelName(serviceName, label.SuffixFrontendHeadersCustomBrowserXSSValue), ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Label functions

func getLabels(application marathon.Application, serviceName string) map[string]string {
	if len(serviceName) > 0 {
		return label.ExtractServicePropertiesP(application.Labels)[serviceName]
	}

	if application.Labels != nil {
		return *application.Labels
	}

	return make(map[string]string)
}

func getLabelName(serviceName string, suffix string) string {
	if len(serviceName) != 0 {
		return suffix
	}
	return label.Prefix + suffix
}

func hasFunc(labelName string) func(application marathon.Application) bool {
	return func(application marathon.Application) bool {
		return label.HasP(application.Labels, labelName)
	}
}

func getFuncStringService(labelName string, defaultValue string) func(application marathon.Application, serviceName string) string {
	return func(application marathon.Application, serviceName string) string {
		labels := getLabels(application, serviceName)
		lbName := getLabelName(serviceName, labelName)
		return label.GetStringValue(labels, lbName, defaultValue)
	}
}

func getFuncBoolService(labelName string, defaultValue bool) func(application marathon.Application, serviceName string) bool {
	return func(application marathon.Application, serviceName string) bool {
		labels := getLabels(application, serviceName)
		lbName := getLabelName(serviceName, labelName)
		return label.GetBoolValue(labels, lbName, defaultValue)
	}
}

func getFuncIntService(labelName string, defaultValue int) func(application marathon.Application, serviceName string) int {
	return func(application marathon.Application, serviceName string) int {
		labels := getLabels(application, serviceName)
		lbName := getLabelName(serviceName, labelName)
		return label.GetIntValue(labels, lbName, defaultValue)
	}
}

func getFuncSliceStringService(labelName string) func(application marathon.Application, serviceName string) []string {
	return func(application marathon.Application, serviceName string) []string {
		labels := getLabels(application, serviceName)
		return label.GetSliceStringValue(labels, getLabelName(serviceName, labelName))
	}
}

func getFuncString(labelName string, defaultValue string) func(application marathon.Application) string {
	return func(application marathon.Application) string {
		return label.GetStringValueP(application.Labels, labelName, defaultValue)
	}
}

func getFuncInt64(labelName string, defaultValue int64) func(application marathon.Application) int64 {
	return func(application marathon.Application) int64 {
		return label.GetInt64ValueP(application.Labels, labelName, defaultValue)
	}
}
