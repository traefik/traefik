package mesos

import (
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
	"github.com/mesosphere/mesos-dns/records/state"
)

type taskData struct {
	state.Task
	TraefikLabels map[string]string
	SegmentName   string
}

func (p *Provider) buildConfigurationV2(tasks []state.Task) *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getDomain":           label.GetFuncString(label.TraefikDomain, p.Domain),
		"getSubDomain":        p.getSubDomain,
		"getSegmentSubDomain": p.getSegmentSubDomain,
		"getID":               getID,

		// Backend functions
		"getBackendName":        getBackendName,
		"getCircuitBreaker":     label.GetCircuitBreaker,
		"getLoadBalancer":       label.GetLoadBalancer,
		"getMaxConn":            label.GetMaxConn,
		"getHealthCheck":        label.GetHealthCheck,
		"getBuffering":          label.GetBuffering,
		"getResponseForwarding": label.GetResponseForwarding,
		"getServers":            p.getServers,
		"getHost":               p.getHost,
		"getServerPort":         p.getServerPort,

		// Frontend functions
		"getSegmentNameSuffix": getSegmentNameSuffix,
		"getFrontEndName":      getFrontendName,
		"getEntryPoints":       label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":         label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":              label.GetAuth,
		"getPriority":          label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":    label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":       label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPassTLSClientCert": label.GetTLSClientCert,
		"getFrontendRule":      p.getFrontendRule,
		"getRedirect":          label.GetRedirect,
		"getErrorPages":        label.GetErrorPages,
		"getRateLimit":         label.GetRateLimit,
		"getHeaders":           label.GetHeaders,
		"getWhiteList":         label.GetWhiteList,
	}

	appsTasks := p.filterTasks(tasks)

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

func (p *Provider) filterTasks(tasks []state.Task) map[string][]taskData {
	appsTasks := make(map[string][]taskData)

	for _, task := range tasks {
		taskLabels := label.ExtractTraefikLabels(extractLabels(task))
		for segmentName, traefikLabels := range taskLabels {
			data := taskData{
				Task:          task,
				TraefikLabels: traefikLabels,
				SegmentName:   segmentName,
			}

			if taskFilter(data, p.ExposedByDefault) {
				name := getName(data)
				if _, ok := appsTasks[name]; !ok {
					appsTasks[name] = []taskData{data}
				} else {
					appsTasks[name] = append(appsTasks[name], data)
				}
			}
		}
	}

	return appsTasks
}

func taskFilter(task taskData, exposedByDefaultFlag bool) bool {
	name := getName(task)

	if len(task.DiscoveryInfo.Ports.DiscoveryPorts) == 0 {
		log.Debugf("Filtering Mesos task without port %s", name)
		return false
	}
	if !isEnabled(task, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled Mesos task %s", name)
		return false
	}

	// filter indeterminable task port
	portIndexLabel := label.GetStringValue(task.TraefikLabels, label.TraefikPortIndex, "")
	portNameLabel := label.GetStringValue(task.TraefikLabels, label.TraefikPortName, "")
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
	if portNameLabel != "" {
		var foundPort bool
		for _, exposedPort := range task.DiscoveryInfo.Ports.DiscoveryPorts {
			if portNameLabel == exposedPort.Name {
				foundPort = true
				break
			}
		}

		if !foundPort {
			log.Debugf("Filtering Mesos task %s without a matching port for %q label", task.Name, label.TraefikPortName)
			return false
		}
	}

	// filter healthChecks
	if task.Statuses != nil && len(task.Statuses) > 0 && task.Statuses[0].Healthy != nil && !*task.Statuses[0].Healthy {
		log.Debugf("Filtering Mesos task %s with bad healthCheck", name)
		return false

	}
	return true
}

func getID(task taskData) string {
	return provider.Normalize(task.ID + getSegmentNameSuffix(task.SegmentName))
}

func getName(task taskData) string {
	return provider.Normalize(task.DiscoveryInfo.Name + getSegmentNameSuffix(task.SegmentName))
}

func getBackendName(task taskData) string {
	return label.GetStringValue(task.TraefikLabels, label.TraefikBackend, getName(task))
}

func getFrontendName(task taskData) string {
	// TODO task.ID -> task.Name + task.ID
	return provider.Normalize(task.ID + getSegmentNameSuffix(task.SegmentName))
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
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func (p *Provider) getSegmentSubDomain(task taskData) string {
	subDomain := strings.ToLower(p.getSubDomain(task.DiscoveryInfo.Name))
	if len(task.SegmentName) > 0 {
		subDomain = strings.ToLower(provider.Normalize(task.SegmentName)) + "." + subDomain
	}
	return subDomain
}

// getFrontendRule returns the frontend rule for the specified application, using it's label.
// It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(task taskData) string {
	if v := label.GetStringValue(task.TraefikLabels, label.TraefikFrontendRule, ""); len(v) > 0 {
		return v
	}

	domain := label.GetStringValue(task.TraefikLabels, label.TraefikDomain, p.Domain)
	if len(domain) > 0 {
		domain = "." + domain
	}

	return "Host:" + p.getSegmentSubDomain(task) + domain
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
			URL:    fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(host, port)),
			Weight: getIntValue(task.TraefikLabels, label.TraefikWeight, label.DefaultWeight, math.MaxInt32),
		}
	}

	return servers
}

func (p *Provider) getHost(task taskData) string {
	return task.IP(strings.Split(p.IPSources, ",")...)
}

func (p *Provider) getServerPort(task taskData) string {
	if label.Has(task.TraefikLabels, label.TraefikPort) {
		pv := label.GetIntValue(task.TraefikLabels, label.TraefikPort, 0)
		if pv <= 0 {
			log.Errorf("explicitly specified port %d must be larger than zero", pv)
			return ""
		}
		return strconv.Itoa(pv)
	}

	plv := getIntValue(task.TraefikLabels, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	// Find named port using traefik.portName or the segment name
	if pn := label.GetStringValue(task.TraefikLabels, label.TraefikPortName, task.SegmentName); len(pn) > 0 {
		for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
			if pn == port.Name {
				return strconv.Itoa(port.Number)
			}
		}
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
	log.Warnf("The value %d for %s exceed the max authorized value %d, falling back to %d.", value, labelName, maxValue, defaultValue)
	return defaultValue
}

func extractLabels(task state.Task) map[string]string {
	labels := make(map[string]string)
	for _, lbl := range task.Labels {
		labels[lbl.Key] = lbl.Value
	}
	return labels
}
