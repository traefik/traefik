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

func (p *Provider) buildConfigurationV1(tasks []state.Task) *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getDomain": getFuncStringValueV1(label.TraefikDomain, p.Domain),
		"getID":     getIDV1,

		// Backend functions
		"getBackendName": getBackendNameV1,
		"getHost":        p.getHostV1,
		"getProtocol":    getFuncApplicationStringValueV1(label.TraefikProtocol, label.DefaultProtocol),
		"getWeight":      getFuncApplicationIntValueV1(label.TraefikWeight, label.DefaultWeight),
		"getBackend":     getBackendV1,
		"getPort":        p.getPort,

		// Frontend functions
		"getFrontendBackend": getBackendNameV1,
		"getFrontEndName":    getFrontendNameV1,
		"getEntryPoints":     getFuncSliceStringValueV1(label.TraefikFrontendEntryPoints),
		"getBasicAuth":       getFuncSliceStringValueV1(label.TraefikFrontendAuthBasic),
		"getPriority":        getFuncIntValueV1(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":  getFuncBoolValueV1(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getFrontendRule":    p.getFrontendRuleV1,
	}

	// filter tasks
	filteredTasks := fun.Filter(func(task state.Task) bool {
		return taskFilterV1(task, p.ExposedByDefault)
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

	configuration, err := p.GetConfiguration("templates/mesos-v1.tmpl", mesosFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

// Deprecated
func taskFilterV1(task state.Task, exposedByDefaultFlag bool) bool {
	if len(task.DiscoveryInfo.Ports.DiscoveryPorts) == 0 {
		log.Debugf("Filtering Mesos task without port %s", task.Name)
		return false
	}

	if !isEnabledV1(task, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled Mesos task %s", task.DiscoveryInfo.Name)
		return false
	}

	// filter indeterminable task port
	portIndexLabel := getStringValueV1(task, label.TraefikPortIndex, "")
	portValueLabel := getStringValueV1(task, label.TraefikPort, "")
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

// Deprecated
func getIDV1(task state.Task) string {
	return provider.Normalize(task.ID)
}

// Deprecated
func getBackendV1(task state.Task, apps []state.Task) string {
	_, err := getApplicationV1(task, apps)
	if err != nil {
		log.Error(err)
		return ""
	}
	return getBackendNameV1(task)
}

// Deprecated
func getBackendNameV1(task state.Task) string {
	if value := getStringValueV1(task, label.TraefikBackend, ""); len(value) > 0 {
		return value
	}
	return provider.Normalize(task.DiscoveryInfo.Name)
}

// Deprecated
func getFrontendNameV1(task state.Task) string {
	// TODO task.ID -> task.Name + task.ID
	return provider.Normalize(task.ID)
}

// Deprecated
func (p *Provider) getPort(task state.Task, applications []state.Task) string {
	_, err := getApplicationV1(task, applications)
	if err != nil {
		log.Error(err)
		return ""
	}

	plv := getIntValueV1(task, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	if pv := getStringValueV1(task, label.TraefikPort, ""); len(pv) > 0 {
		return pv
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

// getFrontendRuleV1 returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
// Deprecated
func (p *Provider) getFrontendRuleV1(task state.Task) string {
	if v := getStringValueV1(task, label.TraefikFrontendRule, ""); len(v) > 0 {
		return v
	}

	domain := getStringValueV1(task, label.TraefikDomain, p.Domain)
	return "Host:" + strings.ToLower(strings.Replace(p.getSubDomain(task.DiscoveryInfo.Name), "_", "-", -1)) + "." + domain
}

// Deprecated
func (p *Provider) getHostV1(task state.Task) string {
	return task.IP(strings.Split(p.IPSources, ",")...)
}

// Deprecated
func isEnabledV1(task state.Task, exposedByDefault bool) bool {
	return getBoolValueV1(task, label.TraefikEnable, exposedByDefault)
}

// Label functions

// Deprecated
func getFuncApplicationStringValueV1(labelName string, defaultValue string) func(task state.Task, applications []state.Task) string {
	return func(task state.Task, applications []state.Task) string {
		_, err := getApplicationV1(task, applications)
		if err != nil {
			log.Error(err)
			return defaultValue
		}

		return getStringValueV1(task, labelName, defaultValue)
	}
}

// Deprecated
func getFuncApplicationIntValueV1(labelName string, defaultValue int) func(task state.Task, applications []state.Task) int {
	return func(task state.Task, applications []state.Task) int {
		_, err := getApplicationV1(task, applications)
		if err != nil {
			log.Error(err)
			return defaultValue
		}

		return getIntValueV1(task, labelName, defaultValue, math.MaxInt32)
	}
}

// Deprecated
func getFuncStringValueV1(labelName string, defaultValue string) func(task state.Task) string {
	return func(task state.Task) string {
		return getStringValueV1(task, labelName, defaultValue)
	}
}

// Deprecated
func getFuncBoolValueV1(labelName string, defaultValue bool) func(task state.Task) bool {
	return func(task state.Task) bool {
		return getBoolValueV1(task, labelName, defaultValue)
	}
}

// Deprecated
func getFuncIntValueV1(labelName string, defaultValue int) func(task state.Task) int {
	return func(task state.Task) int {
		return getIntValueV1(task, labelName, defaultValue, math.MaxInt32)
	}
}

// Deprecated
func getFuncSliceStringValueV1(labelName string) func(task state.Task) []string {
	return func(task state.Task) []string {
		return getSliceStringValueV1(task, labelName)
	}
}

// Deprecated
func getStringValueV1(task state.Task, labelName string, defaultValue string) string {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName && len(lbl.Value) > 0 {
			return lbl.Value
		}
	}
	return defaultValue
}

// Deprecated
func getBoolValueV1(task state.Task, labelName string, defaultValue bool) bool {
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

// Deprecated
func getIntValueV1(task state.Task, labelName string, defaultValue int, maxValue int) int {
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

// Deprecated
func getSliceStringValueV1(task state.Task, labelName string) []string {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
			return label.SplitAndTrimString(lbl.Value, ",")
		}
	}
	return nil
}

// Deprecated
func getApplicationV1(task state.Task, apps []state.Task) (state.Task, error) {
	for _, app := range apps {
		if app.DiscoveryInfo.Name == task.DiscoveryInfo.Name {
			return app, nil
		}
	}
	return state.Task{}, fmt.Errorf("unable to get Mesos application from task %s", task.DiscoveryInfo.Name)
}
