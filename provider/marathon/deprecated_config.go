package marathon

import (
	"errors"
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
	"github.com/gambol99/go-marathon"
)

func (p *Provider) buildConfigurationV1(applications *marathon.Applications) *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getBackend":   p.getBackendNameV1,
		"getDomain":    getFuncStringServiceV1(label.SuffixDomain, p.Domain), // see https://github.com/containous/traefik/pull/1693
		"getSubDomain": p.getSubDomain,                                       // see https://github.com/containous/traefik/pull/1693

		// Backend functions
		"getBackendServer": p.getBackendServerV1,
		"getPort":          getPortV1,
		"getServers":       p.getServersV1,

		"getWeight":                   getFuncIntServiceV1(label.SuffixWeight, label.DefaultWeight),
		"getProtocol":                 getFuncStringServiceV1(label.SuffixProtocol, label.DefaultProtocol),
		"hasCircuitBreakerLabels":     hasFuncV1(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncStringV1(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabels":       hasLoadBalancerLabelsV1,
		"getLoadBalancerMethod":       getFuncStringV1(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"getSticky":                   getStickyV1,
		"hasStickinessLabel":          hasFuncV1(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncStringV1(label.TraefikBackendLoadBalancerStickinessCookieName, ""),
		"hasMaxConnLabels":            hasMaxConnLabelsV1,
		"getMaxConnExtractorFunc":     getFuncStringV1(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getMaxConnAmount":            getFuncInt64V1(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"hasHealthCheckLabels":        hasFuncV1(label.TraefikBackendHealthCheckPath),
		"getHealthCheckPath":          getFuncStringV1(label.TraefikBackendHealthCheckPath, ""),
		"getHealthCheckInterval":      getFuncStringV1(label.TraefikBackendHealthCheckInterval, ""),

		// Frontend functions
		"getServiceNames":         getServiceNamesV1,
		"getServiceNameSuffix":    getSegmentNameSuffix,
		"getPassHostHeader":       getFuncBoolServiceV1(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":          getFuncBoolServiceV1(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":             getFuncIntServiceV1(label.SuffixFrontendPriority, label.DefaultFrontendPriority),
		"getEntryPoints":          getFuncSliceStringServiceV1(label.SuffixFrontendEntryPoints),
		"getFrontendRule":         p.getFrontendRuleV1,
		"getFrontendName":         p.getFrontendNameV1,
		"getBasicAuth":            getFuncSliceStringServiceV1(label.SuffixFrontendAuthBasic),
		"getWhitelistSourceRange": getFuncSliceStringServiceV1(label.SuffixFrontendWhitelistSourceRange),
		"getWhiteList":            getWhiteListV1,
	}

	filteredApps := fun.Filter(p.applicationFilter, applications.Apps).([]marathon.Application)
	for i, app := range filteredApps {
		filteredApps[i].Tasks = fun.Filter(func(task *marathon.Task) bool {
			filtered := p.taskFilter(*task, app)
			if filtered {
				logIllegalServicesV1(*task, app)
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

	configuration, err := p.GetConfiguration("templates/marathon-v1.tmpl", MarathonFuncMap, templateObjects)
	if err != nil {
		log.Errorf("Failed to render Marathon configuration template: %v", err)
	}
	return configuration
}

// logIllegalServicesV1 logs illegal service configurations.
// While we cannot filter on the service level, they will eventually get
// rejected once the server configuration is rendered.
// Deprecated
func logIllegalServicesV1(task marathon.Task, app marathon.Application) {
	for _, serviceName := range getServiceNamesV1(app) {
		// Check for illegal/missing ports.
		if _, err := processPortsV1(app, task, serviceName); err != nil {
			log.Warnf("%s has an illegal configuration: no proper port available", identifierV1(app, task, serviceName))
			continue
		}

		// Check for illegal port label combinations.
		labels := getLabelsV1(app, serviceName)
		hasPortLabel := label.Has(labels, getLabelNameV1(serviceName, label.SuffixPort))
		hasPortIndexLabel := label.Has(labels, getLabelNameV1(serviceName, label.SuffixPortIndex))
		if hasPortLabel && hasPortIndexLabel {
			log.Warnf("%s has both port and port index specified; port will take precedence", identifierV1(app, task, serviceName))
		}
	}
}

// Deprecated
func (p *Provider) getBackendNameV1(application marathon.Application, serviceName string) string {
	labels := getLabelsV1(application, serviceName)
	lblBackend := getLabelNameV1(serviceName, label.SuffixBackend)
	value := label.GetStringValue(labels, lblBackend, "")
	if len(value) > 0 {
		return provider.Normalize("backend" + value)
	}
	return provider.Normalize("backend" + application.ID + getSegmentNameSuffix(serviceName))
}

// Deprecated
func (p *Provider) getFrontendNameV1(application marathon.Application, serviceName string) string {
	return provider.Normalize("frontend" + application.ID + getSegmentNameSuffix(serviceName))
}

// getFrontendRuleV1 returns the frontend rule for the specified application, using
// its label. If service is provided, it will look for serviceName label before generic one.
// It returns a default one (Host) if the label is not present.
// Deprecated
func (p *Provider) getFrontendRuleV1(application marathon.Application, serviceName string) string {
	labels := getLabelsV1(application, serviceName)
	lblFrontendRule := getLabelNameV1(serviceName, label.SuffixFrontendRule)
	if value := label.GetStringValue(labels, lblFrontendRule, ""); len(value) > 0 {
		return value
	}

	if p.MarathonLBCompatibility {
		if value := label.GetStringValue(stringValueMap(application.Labels), labelLbCompatibility, ""); len(value) > 0 {
			return "Host:" + value
		}
	}

	domain := label.GetStringValue(labels, label.SuffixDomain, p.Domain)
	if len(serviceName) > 0 {
		return "Host:" + strings.ToLower(provider.Normalize(serviceName)) + "." + p.getSubDomain(application.ID) + "." + domain
	}
	return "Host:" + p.getSubDomain(application.ID) + "." + domain
}

// Deprecated
func (p *Provider) getBackendServerV1(task marathon.Task, application marathon.Application) string {
	networks := application.Networks
	var hostFlag bool

	if networks == nil {
		hostFlag = application.IPAddressPerTask == nil
	} else {
		hostFlag = (*networks)[0].Mode != marathon.ContainerNetworkMode
	}

	if hostFlag || p.ForceTaskHostname {
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
		ipAddressIdx := label.GetIntValue(stringValueMap(application.Labels), labelIPAddressIdx, math.MinInt32)

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

// getServiceNamesV1 returns a list of service names for a given application
// An empty name "" will be added if no service specific properties exist,
// as an indication that there are no sub-services, but only main application
// Deprecated
func getServiceNamesV1(application marathon.Application) []string {
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

// Deprecated
func identifierV1(app marathon.Application, task marathon.Task, serviceName string) string {
	id := fmt.Sprintf("Marathon task %s from application %s", task.ID, app.ID)
	if serviceName != "" {
		id += fmt.Sprintf(" (service: %s)", serviceName)
	}
	return id
}

// Deprecated
func hasLoadBalancerLabelsV1(application marathon.Application) bool {
	method := label.Has(stringValueMap(application.Labels), label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(stringValueMap(application.Labels), label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(stringValueMap(application.Labels), label.TraefikBackendLoadBalancerStickiness)
	return method || sticky || stickiness
}

// Deprecated
func hasMaxConnLabelsV1(application marathon.Application) bool {
	mca := label.Has(stringValueMap(application.Labels), label.TraefikBackendMaxConnAmount)
	mcef := label.Has(stringValueMap(application.Labels), label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getStickyV1(application marathon.Application) bool {
	if label.Has(stringValueMap(application.Labels), label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}
	return label.GetBoolValue(stringValueMap(application.Labels), label.TraefikBackendLoadBalancerSticky, false)
}

// Deprecated
func getPortV1(task marathon.Task, application marathon.Application, serviceName string) string {
	port, err := processPortsV1(application, task, serviceName)
	if err != nil {
		log.Errorf("Unable to process ports for %s: %s", identifierV1(application, task, serviceName), err)
		return ""
	}

	return strconv.Itoa(port)
}

// processPortsV1 returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
// Deprecated
func processPortsV1(application marathon.Application, task marathon.Task, serviceName string) (int, error) {
	labels := getLabelsV1(application, serviceName)
	lblPort := getLabelNameV1(serviceName, label.SuffixPort)

	if label.Has(labels, lblPort) {
		port := label.GetIntValue(labels, lblPort, 0)

		if port <= 0 {
			return 0, fmt.Errorf("explicitly specified port %d must be larger than zero", port)
		} else if port > 0 {
			return port, nil
		}
	}

	ports := retrieveAvailablePortsV1(application, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	lblPortIndex := getLabelNameV1(serviceName, label.SuffixPortIndex)
	portIndex := label.GetIntValue(labels, lblPortIndex, 0)
	if portIndex < 0 || portIndex > len(ports)-1 {
		return 0, fmt.Errorf("index %d must be within range (0, %d)", portIndex, len(ports)-1)
	}
	return ports[portIndex], nil
}

// Deprecated
func retrieveAvailablePortsV1(application marathon.Application, task marathon.Task) []int {
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

// Deprecated
func (p *Provider) getServersV1(application marathon.Application, serviceName string) map[string]types.Server {
	var servers map[string]types.Server

	for _, task := range application.Tasks {
		host := p.getBackendServerV1(*task, application)
		if len(host) == 0 {
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		labels := getLabelsV1(application, serviceName)

		port := getPortV1(*task, application, serviceName)
		protocol := label.GetStringValue(labels, getLabelNameV1(serviceName, label.SuffixProtocol), label.DefaultProtocol)

		serverName := provider.Normalize("server-" + task.ID + getSegmentNameSuffix(serviceName))
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%v", protocol, host, port),
			Weight: label.GetIntValue(labels, getLabelNameV1(serviceName, label.SuffixWeight), label.DefaultWeight),
		}
	}

	return servers
}

// Deprecated
func getWhiteListV1(application marathon.Application, serviceName string) *types.WhiteList {
	labels := getLabelsV1(application, serviceName)

	ranges := label.GetSliceStringValue(labels, getLabelNameV1(serviceName, label.SuffixFrontendWhiteListSourceRange))
	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: label.GetBoolValue(labels, getLabelNameV1(serviceName, label.SuffixFrontendWhiteListUseXForwardedFor), false),
		}
	}

	return nil
}

// Label functions

// Deprecated
func getLabelsV1(application marathon.Application, serviceName string) map[string]string {
	if len(serviceName) > 0 {
		return label.ExtractServicePropertiesP(application.Labels)[serviceName]
	}

	if application.Labels != nil {
		return *application.Labels
	}

	return make(map[string]string)
}

// Deprecated
func getLabelNameV1(serviceName string, suffix string) string {
	if len(serviceName) != 0 {
		return suffix
	}
	return label.Prefix + suffix
}

// Deprecated
func hasFuncV1(labelName string) func(application marathon.Application) bool {
	return func(application marathon.Application) bool {
		return label.Has(stringValueMap(application.Labels), labelName)
	}
}

// Deprecated
func getFuncStringServiceV1(labelName string, defaultValue string) func(application marathon.Application, serviceName string) string {
	return func(application marathon.Application, serviceName string) string {
		labels := getLabelsV1(application, serviceName)
		lbName := getLabelNameV1(serviceName, labelName)
		return label.GetStringValue(labels, lbName, defaultValue)
	}
}

// Deprecated
func getFuncBoolServiceV1(labelName string, defaultValue bool) func(application marathon.Application, serviceName string) bool {
	return func(application marathon.Application, serviceName string) bool {
		labels := getLabelsV1(application, serviceName)
		lbName := getLabelNameV1(serviceName, labelName)
		return label.GetBoolValue(labels, lbName, defaultValue)
	}
}

// Deprecated
func getFuncIntServiceV1(labelName string, defaultValue int) func(application marathon.Application, serviceName string) int {
	return func(application marathon.Application, serviceName string) int {
		labels := getLabelsV1(application, serviceName)
		lbName := getLabelNameV1(serviceName, labelName)
		return label.GetIntValue(labels, lbName, defaultValue)
	}
}

// Deprecated
func getFuncSliceStringServiceV1(labelName string) func(application marathon.Application, serviceName string) []string {
	return func(application marathon.Application, serviceName string) []string {
		labels := getLabelsV1(application, serviceName)
		return label.GetSliceStringValue(labels, getLabelNameV1(serviceName, labelName))
	}
}

// Deprecated
func getFuncStringV1(labelName string, defaultValue string) func(application marathon.Application) string {
	return func(application marathon.Application) string {
		return label.GetStringValue(stringValueMap(application.Labels), labelName, defaultValue)
	}
}

// Deprecated
func getFuncInt64V1(labelName string, defaultValue int64) func(application marathon.Application) int64 {
	return func(application marathon.Application) int64 {
		return label.GetInt64Value(stringValueMap(application.Labels), labelName, defaultValue)
	}
}
