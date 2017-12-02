package mesos

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/records/state"
)

func (p *Provider) buildConfiguration() *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getBackend":         getBackend,
		"getPort":            p.getPort,
		"getHost":            p.getHost,
		"getWeight":          getFuncApplicationStringValue(label.TraefikWeight, label.DefaultWeight),
		"getDomain":          getFuncStringValue(label.TraefikDomain, p.Domain),
		"getProtocol":        getFuncApplicationStringValue(label.TraefikProtocol, label.DefaultProtocol),
		"getPassHostHeader":  getFuncStringValue(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPriority":        getFuncStringValue(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getEntryPoints":     getFuncSliceStringValue(label.TraefikFrontendEntryPoints),
		"getFrontendRule":    p.getFrontendRule,
		"getFrontendBackend": getFrontendBackend,
		"getID":              getID,
		"getFrontEndName":    getFrontEndName,
	}

	rg := records.NewRecordGenerator(time.Duration(p.StateTimeoutSecond) * time.Second)
	st, err := rg.FindMaster(p.Masters...)
	if err != nil {
		log.Errorf("Failed to create a client for Mesos, error: %v", err)
		return nil
	}
	tasks := taskRecords(st)

	// filter tasks
	filteredTasks := fun.Filter(func(task state.Task) bool {
		return taskFilter(task, p.ExposedByDefault)
	}, tasks).([]state.Task)

	uniqueApps := make(map[string]state.Task)
	for _, value := range filteredTasks {
		if _, ok := uniqueApps[value.DiscoveryInfo.Name]; !ok {
			uniqueApps[value.DiscoveryInfo.Name] = value
		}
	}
	var filteredApps []state.Task
	for _, value := range uniqueApps {
		filteredApps = append(filteredApps, value)
	}

	templateObjects := struct {
		Applications []state.Task
		Tasks        []state.Task
		Domain       string
	}{
		Applications: filteredApps,
		Tasks:        filteredTasks,
		Domain:       p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/mesos.tmpl", mesosFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func taskRecords(st state.State) []state.Task {
	var tasks []state.Task
	for _, f := range st.Frameworks {
		for _, task := range f.Tasks {
			for _, slave := range st.Slaves {
				if task.SlaveID == slave.ID {
					task.SlaveIP = slave.Hostname
				}
			}

			// only do running and discoverable tasks
			if task.State == "TASK_RUNNING" {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks
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

	//filter healthChecks
	if task.Statuses != nil && len(task.Statuses) > 0 && task.Statuses[0].Healthy != nil && !*task.Statuses[0].Healthy {
		log.Debugf("Filtering Mesos task %s with bad healthCheck", task.DiscoveryInfo.Name)
		return false

	}
	return true
}

func getID(task state.Task) string {
	return provider.Normalize(task.ID)
}

func getBackend(task state.Task, apps []state.Task) string {
	application, err := getApplication(task, apps)
	if err != nil {
		log.Error(err)
		return ""
	}
	return getFrontendBackend(application)
}

func getFrontendBackend(task state.Task) string {
	if value := getStringValue(task, label.TraefikBackend, ""); len(value) > 0 {
		return value
	}
	return "-" + provider.Normalize(task.DiscoveryInfo.Name)
}

func getFrontEndName(task state.Task) string {
	return provider.Normalize(task.ID)
}

func (p *Provider) getPort(task state.Task, applications []state.Task) string {
	application, err := getApplication(task, applications)
	if err != nil {
		log.Error(err)
		return ""
	}

	plv := getIntValue(application, label.TraefikPortIndex, math.MinInt32, len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1)
	if plv >= 0 {
		return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[plv].Number)
	}

	if pv := getStringValue(application, label.TraefikPort, ""); len(pv) > 0 {
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

// Label functions

func getFuncApplicationStringValue(labelName string, defaultValue string) func(task state.Task, applications []state.Task) string {
	return func(task state.Task, applications []state.Task) string {
		app, err := getApplication(task, applications)
		if err == nil {
			return getStringValue(app, labelName, defaultValue)
		}
		log.Error(err)
		return defaultValue
	}
}

func getFuncStringValue(labelName string, defaultValue string) func(task state.Task) string {
	return func(task state.Task) string {
		return getStringValue(task, labelName, defaultValue)
	}
}

func getFuncSliceStringValue(labelName string) func(task state.Task) []string {
	return func(task state.Task) []string {
		return getSliceStringValue(task, labelName)
	}
}

func getStringValue(task state.Task, labelName string, defaultValue string) string {
	for _, lbl := range task.Labels {
		if lbl.Key == labelName {
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

func getApplication(task state.Task, apps []state.Task) (state.Task, error) {
	for _, app := range apps {
		if app.DiscoveryInfo.Name == task.DiscoveryInfo.Name {
			return app, nil
		}
	}
	return state.Task{}, fmt.Errorf("unable to get Mesos application from task %s", task.DiscoveryInfo.Name)
}

func isEnabled(task state.Task, exposedByDefault bool) bool {
	return getBoolValue(task, label.TraefikEnable, exposedByDefault)
}
