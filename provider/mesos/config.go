package mesos

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records/state"
)

type taskData struct {
	state.Task
	TraefikLabels map[string]string
}

func (p *Provider) buildConfigurationV2(tasks []state.Task) *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getDomain": label.GetFuncString(label.TraefikDomain, p.Domain),
		"getID":     getID,

		// Backend functions
		"getBackendName":    getBackendName,
		"getCircuitBreaker": label.GetCircuitBreaker,
		"getLoadBalancer":   label.GetLoadBalancer,
		"getMaxConn":        label.GetMaxConn,
		"getHealthCheck":    label.GetHealthCheck,
		"getBuffering":      label.GetBuffering,
		"getServers":        p.getServers,
		"getHost":           p.getHost,
		"getServerPort":     p.getServerPort,

		// Frontend functions
		"getFrontEndName":   getFrontendName,
		"getEntryPoints":    label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      label.GetFuncSliceString(label.TraefikFrontendAuthBasic),
		"getPriority":       label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader": label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":    label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getFrontendRule":   p.getFrontendRule,
		"getRedirect":       label.GetRedirect,
		"getErrorPages":     label.GetErrorPages,
		"getRateLimit":      label.GetRateLimit,
		"getHeaders":        label.GetHeaders,
		"getWhiteList":      label.GetWhiteList,
	}

	// filter tasks
	appsTasks := make(map[string][]taskData)
	for _, task := range tasks {
		data := taskData{
			Task:          task,
			TraefikLabels: extractLabels(task),
		}
		if taskFilter(data, p.ExposedByDefault) {
			if _, ok := appsTasks[task.DiscoveryInfo.Name]; !ok {
				appsTasks[task.DiscoveryInfo.Name] = []taskData{data}
			} else {
				appsTasks[task.DiscoveryInfo.Name] = append(appsTasks[task.DiscoveryInfo.Name], data)
			}
		}
	}

	templateObjects := struct {
		ApplicationsTasks map[string][]taskData
		Domain            string
	}{
		ApplicationsTasks: appsTasks,
		Domain:            p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/mesos.tmpl", mesosFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func taskFilter(task taskData, exposedByDefaultFlag bool) bool {
	if len(task.DiscoveryInfo.Ports.DiscoveryPorts) == 0 {
		log.Debugf("Filtering Mesos task without port %s", task.Name)
		return false
	}

	if !isEnabled(task, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled Mesos task %s", task.DiscoveryInfo.Name)
		return false
	}

	// filter indeterminable task port
	portIndexLabel := label.GetStringValue(task.TraefikLabels, label.TraefikPortIndex, "")
	portValueLabel := label.GetStringValue(task.TraefikLabels, label.TraefikPort, "")
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

func getID(task taskData) string {
	return provider.Normalize(task.ID)
}

func getBackendName(task taskData) string {
	return label.GetStringValue(task.TraefikLabels, label.TraefikBackend, provider.Normalize(task.DiscoveryInfo.Name))
}

func getFrontendName(task taskData) string {
	// TODO task.ID -> task.Name + task.ID
	return provider.Normalize(task.ID)
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

// getFrontendRule returns the frontend rule for the specified application, using it's label.
// It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(task taskData) string {
	if v := label.GetStringValue(task.TraefikLabels, label.TraefikFrontendRule, ""); len(v) > 0 {
		return v
	}

	domain := label.GetStringValue(task.TraefikLabels, label.TraefikDomain, p.Domain)
	return "Host:" + strings.ToLower(strings.Replace(p.getSubDomain(task.DiscoveryInfo.Name), "_", "-", -1)) + "." + domain
}

func (p *Provider) getServers(tasks []taskData) map[string]types.Server {
	var servers map[string]types.Server

	for _, task := range tasks {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(task.TraefikLabels, label.TraefikProtocol, label.DefaultProtocol)
		host := p.getHost(task)
		port := p.getServerPort(task)

		serverName := "server-" + getID(task)
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%s", protocol, host, port),
			Weight: getIntValue(task.TraefikLabels, label.TraefikWeight, label.DefaultWeight, math.MaxInt32),
		}
	}

	return servers
}

func (p *Provider) getHost(task taskData) string {
	return task.IP(strings.Split(p.IPSources, ",")...)
}

func (p *Provider) getServerPort(task taskData) string {
	plv := getIntValue(task.TraefikLabels, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	if pv := label.GetStringValue(task.TraefikLabels, label.TraefikPort, ""); len(pv) > 0 {
		return pv
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

func isEnabled(task taskData, exposedByDefault bool) bool {
	return label.GetBoolValue(task.TraefikLabels, label.TraefikEnable, exposedByDefault)
}

// Label functions

func getIntValue(labels map[string]string, labelName string, defaultValue int, maxValue int) int {
	value := label.GetIntValue(labels, labelName, defaultValue)
	if value <= maxValue {
		return value
	}
	log.Warnf("The value %q for %q exceed the max authorized value %q, falling back to %v.", value, labelName, maxValue, defaultValue)
	return defaultValue
}

func extractLabels(task state.Task) map[string]string {
	labels := make(map[string]string)
	for _, lbl := range task.Labels {
		labels[lbl.Key] = lbl.Value
	}
	return labels
}
