package marathon

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
)

const defaultService = ""

type appData struct {
	marathon.Application
	SegmentLabels map[string]string
	SegmentName   string
	LinkedApps    []*appData
}

func (p *Provider) buildConfigurationV2(applications *marathon.Applications) *types.Configuration {
	var MarathonFuncMap = template.FuncMap{
		"getDomain":      label.GetFuncString(label.TraefikDomain, p.Domain), // see https://github.com/containous/traefik/pull/1693
		"getSubDomain":   p.getSubDomain,                                     // see https://github.com/containous/traefik/pull/1693
		"getBackendName": p.getBackendName,

		// Backend functions
		"getPort":               getPort,
		"getCircuitBreaker":     label.GetCircuitBreaker,
		"getLoadBalancer":       label.GetLoadBalancer,
		"getMaxConn":            label.GetMaxConn,
		"getHealthCheck":        label.GetHealthCheck,
		"getBuffering":          label.GetBuffering,
		"getResponseForwarding": label.GetResponseForwarding,
		"getServers":            p.getServers,

		// Frontend functions
		"getSegmentNameSuffix": getSegmentNameSuffix,
		"getFrontendRule":      p.getFrontendRule,
		"getFrontendName":      p.getFrontendName,
		"getPassHostHeader":    label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":       label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPassTLSClientCert": label.GetTLSClientCert,
		"getPriority":          label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getEntryPoints":       label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":         label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":              label.GetAuth,
		"getRedirect":          label.GetRedirect,
		"getErrorPages":        label.GetErrorPages,
		"getRateLimit":         label.GetRateLimit,
		"getHeaders":           label.GetHeaders,
		"getWhiteList":         label.GetWhiteList,
	}

	apps := make(map[string]*appData)
	for _, app := range applications.Apps {
		if p.applicationFilter(app) {
			// Tasks
			var filteredTasks []*marathon.Task
			for _, task := range app.Tasks {
				if p.taskFilter(*task, app) {
					filteredTasks = append(filteredTasks, task)
					logIllegalServices(*task, app)
				}
			}

			app.Tasks = filteredTasks

			// segments
			segmentProperties := label.ExtractTraefikLabels(stringValueMap(app.Labels))
			for segmentName, labels := range segmentProperties {
				data := &appData{
					Application:   app,
					SegmentLabels: labels,
					SegmentName:   segmentName,
				}

				backendName := p.getBackendName(*data)
				if baseApp, ok := apps[backendName]; ok {
					baseApp.LinkedApps = append(baseApp.LinkedApps, data)
				} else {
					apps[backendName] = data
				}
			}
		}
	}

	templateObjects := struct {
		Applications map[string]*appData
		Domain       string
	}{
		Applications: apps,
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
	if !label.IsEnabled(stringValueMap(app.Labels), p.ExposedByDefault) {
		log.Debugf("Filtering disabled Marathon application %s", app.ID)
		return false
	}

	// Filter by constraints.
	constraintTags := label.GetSliceStringValue(stringValueMap(app.Labels), label.TraefikTags)
	if p.MarathonLBCompatibility {
		if haGroup := label.GetStringValue(stringValueMap(app.Labels), labelLbCompatibilityGroup, ""); len(haGroup) > 0 {
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

// logIllegalServices logs illegal service configurations.
// While we cannot filter on the service level, they will eventually get
// rejected once the server configuration is rendered.
func logIllegalServices(task marathon.Task, app marathon.Application) {
	segmentProperties := label.ExtractTraefikLabels(stringValueMap(app.Labels))
	for segmentName, labels := range segmentProperties {
		// Check for illegal/missing ports.
		if _, err := processPorts(app, task, labels); err != nil {
			log.Warnf("%s has an illegal configuration: no proper port available", identifier(app, task, segmentName))
			continue
		}

		// Check for illegal port label combinations.
		hasPortLabel := label.Has(labels, label.TraefikPort)
		hasPortIndexLabel := label.Has(labels, label.TraefikPortIndex)
		if hasPortLabel && hasPortIndexLabel {
			log.Warnf("%s has both port and port index specified; port will take precedence", identifier(app, task, segmentName))
		}
	}
}

func getSegmentNameSuffix(serviceName string) string {
	if len(serviceName) > 0 {
		return "-service-" + provider.Normalize(serviceName)
	}
	return ""
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

func (p *Provider) getBackendName(app appData) string {
	value := label.GetStringValue(app.SegmentLabels, label.TraefikBackend, "")
	if len(value) > 0 {
		return provider.Normalize("backend" + value)
	}

	return provider.Normalize("backend" + app.ID + getSegmentNameSuffix(app.SegmentName))
}

func (p *Provider) getFrontendName(app appData) string {
	return provider.Normalize("frontend" + app.ID + getSegmentNameSuffix(app.SegmentName))
}

// getFrontendRule returns the frontend rule for the specified application, using
// its label. If service is provided, it will look for serviceName label before generic one.
// It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(app appData) string {
	if value := label.GetStringValue(app.SegmentLabels, label.TraefikFrontendRule, ""); len(value) > 0 {
		return value
	}

	if p.MarathonLBCompatibility {
		if value := label.GetStringValue(stringValueMap(app.Labels), labelLbCompatibility, ""); len(value) > 0 {
			return "Host:" + value
		}
	}

	domain := label.GetStringValue(app.SegmentLabels, label.TraefikDomain, p.Domain)
	if len(domain) > 0 {
		domain = "." + domain
	}

	if len(app.SegmentName) > 0 {
		return "Host:" + strings.ToLower(provider.Normalize(app.SegmentName)) + "." + p.getSubDomain(app.ID) + domain
	}
	return "Host:" + p.getSubDomain(app.ID) + domain
}

func getPort(task marathon.Task, app appData) string {
	port, err := processPorts(app.Application, task, app.SegmentLabels)
	if err != nil {
		log.Errorf("Unable to process ports for %s: %s", identifier(app.Application, task, app.SegmentName), err)
		return ""
	}

	return strconv.Itoa(port)
}

// processPorts returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
func processPorts(app marathon.Application, task marathon.Task, labels map[string]string) (int, error) {
	if label.Has(labels, label.TraefikPort) {
		port := label.GetIntValue(labels, label.TraefikPort, 0)

		if port <= 0 {
			return 0, fmt.Errorf("explicitly specified port %d must be larger than zero", port)
		} else if port > 0 {
			return port, nil
		}
	}

	ports := retrieveAvailablePorts(app, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	portIndex := label.GetIntValue(labels, label.TraefikPortIndex, 0)
	if portIndex < 0 || portIndex > len(ports)-1 {
		return 0, fmt.Errorf("index %d must be within range (0, %d)", portIndex, len(ports)-1)
	}
	return ports[portIndex], nil
}

func retrieveAvailablePorts(app marathon.Application, task marathon.Task) []int {
	// Using default port configuration
	if len(task.Ports) > 0 {
		return task.Ports
	}

	// Using port definition if available
	if app.PortDefinitions != nil && len(*app.PortDefinitions) > 0 {
		var ports []int
		for _, def := range *app.PortDefinitions {
			if def.Port != nil {
				ports = append(ports, *def.Port)
			}
		}
		return ports
	}

	// If using IP-per-task using this port definition
	if app.IPAddressPerTask != nil && app.IPAddressPerTask.Discovery != nil && len(*(app.IPAddressPerTask.Discovery.Ports)) > 0 {
		var ports []int
		for _, def := range *(app.IPAddressPerTask.Discovery.Ports) {
			ports = append(ports, def.Number)
		}
		return ports
	}

	return []int{}
}

func identifier(app marathon.Application, task marathon.Task, segmentName string) string {
	id := fmt.Sprintf("Marathon task %s from application %s", task.ID, app.ID)
	if segmentName != "" {
		id += fmt.Sprintf(" (segment: %s)", segmentName)
	}
	return id
}

func (p *Provider) getServers(app appData) map[string]types.Server {
	var servers map[string]types.Server

	for _, task := range app.Tasks {
		name, server, err := p.getServer(app, *task)
		if err != nil {
			log.Error(err)
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		servers[name] = *server
	}

	for _, linkedApp := range app.LinkedApps {
		for _, task := range linkedApp.Tasks {
			name, server, err := p.getServer(*linkedApp, *task)
			if err != nil {
				log.Error(err)
				continue
			}

			if servers == nil {
				servers = make(map[string]types.Server)
			}

			servers[name] = *server
		}
	}

	return servers
}

func (p *Provider) getServer(app appData, task marathon.Task) (string, *types.Server, error) {
	host, err := p.getServerHost(task, app)
	if len(host) == 0 {
		return "", nil, err
	}

	port := getPort(task, app)
	protocol := label.GetStringValue(app.SegmentLabels, label.TraefikProtocol, label.DefaultProtocol)

	serverName := provider.Normalize("server-" + app.ID + "-" + task.ID + getSegmentNameSuffix(app.SegmentName))

	return serverName, &types.Server{
		URL:    fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(host, port)),
		Weight: label.GetIntValue(app.SegmentLabels, label.TraefikWeight, label.DefaultWeight),
	}, nil
}

func (p *Provider) getServerHost(task marathon.Task, app appData) (string, error) {
	networks := app.Networks
	var hostFlag bool

	if networks == nil {
		hostFlag = app.IPAddressPerTask == nil
	} else {
		hostFlag = (*networks)[0].Mode != marathon.ContainerNetworkMode
	}

	if hostFlag || p.ForceTaskHostname {
		if len(task.Host) == 0 {
			return "", fmt.Errorf("host is undefined for task %q app %q", task.ID, app.ID)
		}
		return task.Host, nil
	}

	numTaskIPAddresses := len(task.IPAddresses)
	switch numTaskIPAddresses {
	case 0:
		return "", fmt.Errorf("missing IP address for Marathon application %s on task %s", app.ID, task.ID)
	case 1:
		return task.IPAddresses[0].IPAddress, nil
	default:
		ipAddressIdx := label.GetIntValue(stringValueMap(app.Labels), labelIPAddressIdx, math.MinInt32)

		if ipAddressIdx == math.MinInt32 {
			return "", fmt.Errorf("found %d task IP addresses but missing IP address index for Marathon application %s on task %s",
				numTaskIPAddresses, app.ID, task.ID)
		}
		if ipAddressIdx < 0 || ipAddressIdx > numTaskIPAddresses {
			return "", fmt.Errorf("cannot use IP address index to select from %d task IP addresses for Marathon application %s on task %s",
				numTaskIPAddresses, app.ID, task.ID)
		}

		return task.IPAddresses[ipAddressIdx].IPAddress, nil
	}
}
