package mesos

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
	"github.com/mesosphere/mesos-dns/records/state"
)

func (p *Provider) buildConfiguration(tasks []state.Task) *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getDomain": getFuncStringValue(label.TraefikDomain, p.Domain),
		"getID":     getID,

		// Backend functions
		"getBackendName":    getBackendName,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getBuffering":      getBuffering,
		"getServers":        p.getServers,
		"getHost":           p.getHost,
		"getServerPort":     p.getServerPort,

		// TODO Deprecated [breaking]
		"getProtocol": getFuncApplicationStringValue(label.TraefikProtocol, label.DefaultProtocol),
		// TODO Deprecated [breaking]
		"getWeight": getFuncApplicationStringValue(label.TraefikWeight, label.DefaultWeight),
		// TODO Deprecated [breaking] replaced by getBackendName
		"getBackend": getBackend,
		// TODO Deprecated [breaking]
		"getPort": p.getPort,

		// Frontend functions
		"getFrontEndName":         getFrontendName,
		"getEntryPoints":          getFuncSliceStringValue(label.TraefikFrontendEntryPoints),
		"getBasicAuth":            getFuncSliceStringValue(label.TraefikFrontendAuthBasic),
		"getWhitelistSourceRange": getFuncSliceStringValue(label.TraefikFrontendWhitelistSourceRange),
		"getPriority":             getFuncStringValue(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":       getFuncBoolValue(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":          getFuncBoolValue(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getFrontendRule":         p.getFrontendRule,
		"getRedirect":             getRedirect,
		"getErrorPages":           getErrorPages,
		"getRateLimit":            getRateLimit,
		"getHeaders":              getHeaders,

		// TODO Deprecated [breaking]
		"getFrontendBackend": getBackendName,
	}

	// filter tasks
	filteredTasks := fun.Filter(func(task state.Task) bool {
		return taskFilter(task, p.ExposedByDefault)
	}, tasks).([]state.Task)

	// Deprecated
	var filteredApps []state.Task
	uniqueApps := make(map[string]struct{})
	for _, task := range filteredTasks {
		if _, ok := uniqueApps[task.DiscoveryInfo.Name]; !ok {
			uniqueApps[task.DiscoveryInfo.Name] = struct{}{}
			filteredApps = append(filteredApps, task)
		}
	}

	appsTasks := make(map[string][]state.Task)
	for _, task := range filteredTasks {
		if _, ok := appsTasks[task.DiscoveryInfo.Name]; !ok {
			appsTasks[task.DiscoveryInfo.Name] = []state.Task{task}
		} else {
			appsTasks[task.DiscoveryInfo.Name] = append(appsTasks[task.DiscoveryInfo.Name], task)
		}
	}

	templateObjects := struct {
		ApplicationsTasks map[string][]state.Task
		Applications      []state.Task // Deprecated
		Tasks             []state.Task // Deprecated
		Domain            string
	}{
		ApplicationsTasks: appsTasks,
		Applications:      filteredApps,  // Deprecated
		Tasks:             filteredTasks, // Deprecated
		Domain:            p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/mesos.tmpl", mesosFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func taskFilter(task state.Task, exposedByDefaultFlag bool) bool {
	if len(task.DiscoveryInfo.Ports.DiscoveryPorts) == 0 {
		log.Debugf("Filtering Mesos task without port %s", task.Name)
		return false
	}

	if !isEnabled(task, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled Mesos task %s", task.DiscoveryInfo.Name)
		return false
	}

	// filter indeterminable task port
	portIndexLabel := getStringValue(task, label.TraefikPortIndex, "")
	portValueLabel := getStringValue(task, label.TraefikPort, "")
	if portIndexLabel != "" && portValueLabel != "" {
		log.Debugf("Filtering Mesos task %s specifying both %q' and %q labels", task.Name, label.TraefikPortIndex, label.TraefikPort)
		return false
	}
	if portIndexLabel != "" {
		index, err := strconv.Atoi(portIndexLabel)
		if err != nil || index < 0 || index > len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1 {
			log.Debugf("Filtering Mesos task %s with unexpected value for %q label", task.Name, label.TraefikPortIndex)
			return false
		}
	}
	if portValueLabel != "" {
		port, err := strconv.Atoi(portValueLabel)
		if err != nil {
			log.Debugf("Filtering Mesos task %s with unexpected value for %q label", task.Name, label.TraefikPort)
			return false
		}

		var foundPort bool
		for _, exposedPort := range task.DiscoveryInfo.Ports.DiscoveryPorts {
			if port == exposedPort.Number {
				foundPort = true
				break
			}
		}

		if !foundPort {
			log.Debugf("Filtering Mesos task %s without a matching port for %q label", task.Name, label.TraefikPort)
			return false
		}
	}

	// filter healthChecks
	if task.Statuses != nil && len(task.Statuses) > 0 && task.Statuses[0].Healthy != nil && !*task.Statuses[0].Healthy {
		log.Debugf("Filtering Mesos task %s with bad healthCheck", task.DiscoveryInfo.Name)
		return false

	}
	return true
}

func getID(task state.Task) string {
	return provider.Normalize(task.ID)
}

// Deprecated
func getBackend(task state.Task, apps []state.Task) string {
	_, err := getApplication(task, apps)
	if err != nil {
		log.Error(err)
		return ""
	}
	return getBackendName(task)
}

func getBackendName(task state.Task) string {
	if value := getStringValue(task, label.TraefikBackend, ""); len(value) > 0 {
		return value
	}
	return provider.Normalize(task.DiscoveryInfo.Name)
}

func getFrontendName(task state.Task) string {
	// TODO task.ID -> task.Name + task.ID
	return provider.Normalize(task.ID)
}

func (p *Provider) getServerPort(task state.Task) string {
	plv := getIntValue(task, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	if pv := getStringValue(task, label.TraefikPort, ""); len(pv) > 0 {
		return pv
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

// Deprecated
func (p *Provider) getPort(task state.Task, applications []state.Task) string {
	_, err := getApplication(task, applications)
	if err != nil {
		log.Error(err)
		return ""
	}

	plv := getIntValue(task, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	if pv := getStringValue(task, label.TraefikPort, ""); len(pv) > 0 {
		return pv
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(task state.Task) string {
	if v := getStringValue(task, label.TraefikFrontendRule, ""); len(v) > 0 {
		return v
	}
	return "Host:" + strings.ToLower(strings.Replace(p.getSubDomain(task.DiscoveryInfo.Name), "_", "-", -1)) + "." + p.Domain
}

func (p *Provider) getHost(task state.Task) string {
	return task.IP(strings.Split(p.IPSources, ",")...)
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

func getCircuitBreaker(task state.Task) *types.CircuitBreaker {
	circuitBreaker := getStringValue(task, label.TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func getLoadBalancer(task state.Task) *types.LoadBalancer {
	if !hasPrefix(task, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := getStringValue(task, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
	}

	if getBoolValue(task, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := getStringValue(task, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

func getMaxConn(task state.Task) *types.MaxConn {
	amount := getInt64Value(task, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := getStringValue(task, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getHealthCheck(task state.Task) *types.HealthCheck {
	path := getStringValue(task, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := getIntValue(task, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort, math.MaxInt32)
	interval := getStringValue(task, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getBuffering(task state.Task) *types.Buffering {
	if !hasPrefix(task, label.TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  getInt64Value(task, label.TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: getInt64Value(task, label.TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  getInt64Value(task, label.TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: getInt64Value(task, label.TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      getStringValue(task, label.TraefikBackendBufferingRetryExpression, ""),
	}
}

func (p *Provider) getServers(tasks []state.Task) map[string]types.Server {
	var servers map[string]types.Server

	for _, task := range tasks {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := getStringValue(task, label.TraefikProtocol, label.DefaultProtocol)
		host := p.getHost(task)
		port := p.getServerPort(task)

		serverName := "server-" + getID(task)
		servers[serverName] = types.Server{
			URL:      fmt.Sprintf("%s://%s:%s", protocol, host, port),
			Priority: getIntValue(task, label.TraefikPriority, label.DefaultPriority, math.MaxInt64),
			Weight:   getIntValue(task, label.TraefikWeight, label.DefaultWeightInt, math.MaxInt32),
		}
	}

	return servers
}

func getRedirect(task state.Task) *types.Redirect {
	permanent := getBoolValue(task, label.TraefikFrontendRedirectPermanent, false)

	if hasLabel(task, label.TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: getStringValue(task, label.TraefikFrontendRedirectEntryPoint, ""),
			Permanent:  permanent,
		}
	}

	if hasLabel(task, label.TraefikFrontendRedirectRegex) &&
		hasLabel(task, label.TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       getStringValue(task, label.TraefikFrontendRedirectRegex, ""),
			Replacement: getStringValue(task, label.TraefikFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func getErrorPages(task state.Task) map[string]*types.ErrorPage {
	prefix := label.Prefix + label.BaseFrontendErrorPage
	labels := taskLabelsToMap(task)
	return label.ParseErrorPages(labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimit(task state.Task) *types.RateLimit {
	extractorFunc := getStringValue(task, label.TraefikFrontendRateLimitExtractorFunc, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	labels := taskLabelsToMap(task)
	prefix := label.Prefix + label.BaseFrontendRateLimit
	limits := label.ParseRateSets(labels, prefix, label.RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func getHeaders(task state.Task) *types.Headers {
	labels := taskLabelsToMap(task)

	headers := &types.Headers{
		CustomRequestHeaders:    label.GetMapValue(labels, label.TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   label.GetMapValue(labels, label.TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         label.GetMapValue(labels, label.TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            label.GetSliceStringValue(labels, label.TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       label.GetSliceStringValue(labels, label.TraefikFrontendHostsProxyHeaders),
		STSSeconds:              label.GetInt64Value(labels, label.TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             label.GetBoolValue(labels, label.TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    label.GetBoolValue(labels, label.TraefikFrontendSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    label.GetBoolValue(labels, label.TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              label.GetBoolValue(labels, label.TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          label.GetBoolValue(labels, label.TraefikFrontendForceSTSHeader, false),
		FrameDeny:               label.GetBoolValue(labels, label.TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      label.GetBoolValue(labels, label.TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        label.GetBoolValue(labels, label.TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           label.GetBoolValue(labels, label.TraefikFrontendIsDevelopment, false),
		SSLHost:                 label.GetStringValue(labels, label.TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: label.GetStringValue(labels, label.TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   label.GetStringValue(labels, label.TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               label.GetStringValue(labels, label.TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          label.GetStringValue(labels, label.TraefikFrontendReferrerPolicy, ""),
		CustomBrowserXSSValue:   label.GetStringValue(labels, label.TraefikFrontendCustomBrowserXSSValue, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

func isEnabled(task state.Task, exposedByDefault bool) bool {
	return getBoolValue(task, label.TraefikEnable, exposedByDefault)
}

// Label functions

// Deprecated
func getFuncApplicationStringValue(labelName string, defaultValue string) func(task state.Task, applications []state.Task) string {
	return func(task state.Task, applications []state.Task) string {
		_, err := getApplication(task, applications)
		if err != nil {
			log.Error(err)
			return defaultValue
		}

		return getStringValue(task, labelName, defaultValue)
	}
}

func getFuncStringValue(labelName string, defaultValue string) func(task state.Task) string {
	return func(task state.Task) string {
		return getStringValue(task, labelName, defaultValue)
	}
}

func getFuncBoolValue(labelName string, defaultValue bool) func(task state.Task) bool {
	return func(task state.Task) bool {
		return getBoolValue(task, labelName, defaultValue)
	}
}

func getFuncSliceStringValue(labelName string) func(task state.Task) []string {
	return func(task state.Task) []string {
		return getSliceStringValue(task, labelName)
	}
}

func getStringValue(task state.Task, labelName string, defaultValue string) string {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName && len(lbl.Value) > 0 {
			return lbl.Value
		}
	}
	return defaultValue
}

func getBoolValue(task state.Task, labelName string, defaultValue bool) bool {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
			v, err := strconv.ParseBool(lbl.Value)
			if err == nil {
				return v
			}
		}
	}
	return defaultValue
}

func getIntValue(task state.Task, labelName string, defaultValue int, maxValue int) int {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
			value, err := strconv.Atoi(lbl.Value)
			if err == nil {
				if value <= maxValue {
					return value
				}
				log.Warnf("The value %q for %q exceed the max authorized value %q, falling back to %v.", lbl.Value, labelName, maxValue, defaultValue)
			} else {
				log.Warnf("Unable to parse %q: %q, falling back to %v. %v", labelName, lbl.Value, defaultValue, err)
			}
		}
	}
	return defaultValue
}

func getSliceStringValue(task state.Task, labelName string) []string {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
			return label.SplitAndTrimString(lbl.Value, ",")
		}
	}
	return nil
}

// Deprecated
func getApplication(task state.Task, apps []state.Task) (state.Task, error) {
	for _, app := range apps {
		if app.DiscoveryInfo.Name == task.DiscoveryInfo.Name {
			return app, nil
		}
	}
	return state.Task{}, fmt.Errorf("unable to get Mesos application from task %s", task.DiscoveryInfo.Name)
}

func hasPrefix(task state.Task, prefix string) bool {
	for _, lbl := range task.Labels {
		if strings.HasPrefix(lbl.Key, prefix) {
			return true
		}
	}
	return false
}

func getInt64Value(task state.Task, labelName string, defaultValue int64) int64 {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
			value, err := strconv.ParseInt(lbl.Value, 10, 64)
			if err != nil {
				log.Warnf("Unable to parse %q: %q, falling back to %v. %v", labelName, lbl.Value, defaultValue, err)
			}
			return value
		}
	}
	return defaultValue
}

func hasLabel(task state.Task, label string) bool {
	for _, lbl := range task.Labels {
		if lbl.Key == label {
			return true
		}
	}
	return false
}

func taskLabelsToMap(task state.Task) map[string]string {
	labels := make(map[string]string)
	for _, lbl := range task.Labels {
		labels[lbl.Key] = lbl.Value
	}
	return labels
}
