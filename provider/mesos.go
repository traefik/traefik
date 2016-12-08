package provider

import (
	"errors"
	"strconv"
	"strings"
	"text/template"

	"fmt"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/mesos/mesos-go/detector"
	_ "github.com/mesos/mesos-go/detector/zoo" // Registers the ZK detector
	"github.com/mesosphere/mesos-dns/detect"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/mesosphere/mesos-dns/util"
)

var _ Provider = (*Mesos)(nil)

//Mesos holds configuration of the mesos provider.
type Mesos struct {
	BaseProvider
	Endpoint           string `description:"Mesos server endpoint. You can also specify multiple endpoint for Mesos"`
	Domain             string `description:"Default domain used"`
	ExposedByDefault   bool   `description:"Expose Mesos apps by default"`
	GroupsAsSubDomains bool   `description:"Convert Mesos groups to subdomains"`
	ZkDetectionTimeout int    `description:"Zookeeper timeout (in seconds)"`
	RefreshSeconds     int    `description:"Polling interval (in seconds)"`
	IPSources          string `description:"IPSources (e.g. host, docker, mesos, rkt)"` // e.g. "host", "docker", "mesos", "rkt"
	StateTimeoutSecond int    `description:"HTTP Timeout (in seconds)"`
	Masters            []string
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Mesos) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	operation := func() error {

		// initialize logging
		logging.SetupLogs()

		log.Debugf("%s", provider.IPSources)

		var zk string
		var masters []string

		if strings.HasPrefix(provider.Endpoint, "zk://") {
			zk = provider.Endpoint
		} else {
			masters = strings.Split(provider.Endpoint, ",")
		}

		errch := make(chan error)

		changed := detectMasters(zk, masters)
		reload := time.NewTicker(time.Second * time.Duration(provider.RefreshSeconds))
		zkTimeout := time.Second * time.Duration(provider.ZkDetectionTimeout)
		timeout := time.AfterFunc(zkTimeout, func() {
			if zkTimeout > 0 {
				errch <- fmt.Errorf("master detection timed out after %s", zkTimeout)
			}
		})

		defer reload.Stop()
		defer util.HandleCrash()

		if !provider.Watch {
			reload.Stop()
			timeout.Stop()
		}

		for {
			select {
			case <-reload.C:
				configuration := provider.loadMesosConfig()
				if configuration != nil {
					configurationChan <- types.ConfigMessage{
						ProviderName:  "mesos",
						Configuration: configuration,
					}
				}
			case masters := <-changed:
				if len(masters) == 0 || masters[0] == "" {
					// no leader
					timeout.Reset(zkTimeout)
				} else {
					timeout.Stop()
				}
				log.Debugf("new masters detected: %v", masters)
				provider.Masters = masters
				configuration := provider.loadMesosConfig()
				if configuration != nil {
					configurationChan <- types.ConfigMessage{
						ProviderName:  "mesos",
						Configuration: configuration,
					}
				}
			case err := <-errch:
				log.Errorf("%s", err)
			}
		}
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("mesos connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to mesos server %+v", err)
	}
	return nil
}

func (provider *Mesos) loadMesosConfig() *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getBackend":         provider.getBackend,
		"getPort":            provider.getPort,
		"getHost":            provider.getHost,
		"getWeight":          provider.getWeight,
		"getDomain":          provider.getDomain,
		"getProtocol":        provider.getProtocol,
		"getPassHostHeader":  provider.getPassHostHeader,
		"getPriority":        provider.getPriority,
		"getEntryPoints":     provider.getEntryPoints,
		"getFrontendRule":    provider.getFrontendRule,
		"getFrontendBackend": provider.getFrontendBackend,
		"getID":              provider.getID,
		"getFrontEndName":    provider.getFrontEndName,
		"replace":            replace,
	}

	t := records.NewRecordGenerator(time.Duration(provider.StateTimeoutSecond) * time.Second)
	sj, err := t.FindMaster(provider.Masters...)
	if err != nil {
		log.Errorf("Failed to create a client for mesos, error: %s", err)
		return nil
	}
	tasks := provider.taskRecords(sj)

	//filter tasks
	filteredTasks := fun.Filter(func(task state.Task) bool {
		return mesosTaskFilter(task, provider.ExposedByDefault)
	}, tasks).([]state.Task)

	filteredApps := []state.Task{}
	for _, value := range filteredTasks {
		if !taskInSlice(value, filteredApps) {
			filteredApps = append(filteredApps, value)
		}
	}

	templateObjects := struct {
		Applications []state.Task
		Tasks        []state.Task
		Domain       string
	}{
		filteredApps,
		filteredTasks,
		provider.Domain,
	}

	configuration, err := provider.getConfiguration("templates/mesos.tmpl", mesosFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func taskInSlice(a state.Task, list []state.Task) bool {
	for _, b := range list {
		if b.DiscoveryInfo.Name == a.DiscoveryInfo.Name {
			return true
		}
	}
	return false
}

// labels returns all given Status.[]Labels' values whose keys are equal
// to the given key
func labels(task state.Task, key string) string {
	for _, l := range task.Labels {
		if l.Key == key {
			return l.Value
		}
	}
	return ""
}

func mesosTaskFilter(task state.Task, exposedByDefaultFlag bool) bool {
	if len(task.DiscoveryInfo.Ports.DiscoveryPorts) == 0 {
		log.Debugf("Filtering mesos task without port %s", task.Name)
		return false
	}
	if !isMesosApplicationEnabled(task, exposedByDefaultFlag) {
		log.Debugf("Filtering disabled mesos task %s", task.DiscoveryInfo.Name)
		return false
	}

	//filter indeterminable task port
	portIndexLabel := labels(task, "traefik.portIndex")
	portValueLabel := labels(task, "traefik.port")
	if portIndexLabel != "" && portValueLabel != "" {
		log.Debugf("Filtering mesos task %s specifying both traefik.portIndex and traefik.port labels", task.Name)
		return false
	}
	if portIndexLabel != "" {
		index, err := strconv.Atoi(labels(task, "traefik.portIndex"))
		if err != nil || index < 0 || index > len(task.DiscoveryInfo.Ports.DiscoveryPorts)-1 {
			log.Debugf("Filtering mesos task %s with unexpected value for traefik.portIndex label", task.Name)
			return false
		}
	}
	if portValueLabel != "" {
		port, err := strconv.Atoi(labels(task, "traefik.port"))
		if err != nil {
			log.Debugf("Filtering mesos task %s with unexpected value for traefik.port label", task.Name)
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
			log.Debugf("Filtering mesos task %s without a matching port for traefik.port label", task.Name)
			return false
		}
	}

	//filter healthchecks
	if task.Statuses != nil && len(task.Statuses) > 0 && task.Statuses[0].Healthy != nil && !*task.Statuses[0].Healthy {
		log.Debugf("Filtering mesos task %s with bad healthcheck", task.DiscoveryInfo.Name)
		return false

	}
	return true
}

func getMesos(task state.Task, apps []state.Task) (state.Task, error) {
	for _, application := range apps {
		if application.DiscoveryInfo.Name == task.DiscoveryInfo.Name {
			return application, nil
		}
	}
	return state.Task{}, errors.New("Application not found: " + task.DiscoveryInfo.Name)
}

func isMesosApplicationEnabled(task state.Task, exposedByDefault bool) bool {
	return exposedByDefault && labels(task, "traefik.enable") != "false" || labels(task, "traefik.enable") == "true"
}

func (provider *Mesos) getLabel(task state.Task, label string) (string, error) {
	for _, tmpLabel := range task.Labels {
		if tmpLabel.Key == label {
			return tmpLabel.Value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func (provider *Mesos) getPort(task state.Task, applications []state.Task) string {
	application, err := getMesos(task, applications)
	if err != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return ""
	}

	if portIndexLabel, err := provider.getLabel(application, "traefik.portIndex"); err == nil {
		if index, err := strconv.Atoi(portIndexLabel); err == nil {
			return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[index].Number)
		}
	}
	if portValueLabel, err := provider.getLabel(application, "traefik.port"); err == nil {
		return portValueLabel
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

func (provider *Mesos) getWeight(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return "0"
	}

	if label, err := provider.getLabel(application, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (provider *Mesos) getDomain(task state.Task) string {
	if label, err := provider.getLabel(task, "traefik.domain"); err == nil {
		return label
	}
	return provider.Domain
}

func (provider *Mesos) getProtocol(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return "http"
	}
	if label, err := provider.getLabel(application, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (provider *Mesos) getPassHostHeader(task state.Task) string {
	if passHostHeader, err := provider.getLabel(task, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "false"
}

func (provider *Mesos) getPriority(task state.Task) string {
	if priority, err := provider.getLabel(task, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (provider *Mesos) getEntryPoints(task state.Task) []string {
	if entryPoints, err := provider.getLabel(task, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (provider *Mesos) getFrontendRule(task state.Task) string {
	if label, err := provider.getLabel(task, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host:" + strings.ToLower(strings.Replace(provider.getSubDomain(task.DiscoveryInfo.Name), "_", "-", -1)) + "." + provider.Domain
}

func (provider *Mesos) getBackend(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return ""
	}
	return provider.getFrontendBackend(application)
}

func (provider *Mesos) getFrontendBackend(task state.Task) string {
	if label, err := provider.getLabel(task, "traefik.backend"); err == nil {
		return label
	}
	return "-" + cleanupSpecialChars(task.DiscoveryInfo.Name)
}

func (provider *Mesos) getHost(task state.Task) string {
	return task.IP(strings.Split(provider.IPSources, ",")...)
}

func (provider *Mesos) getID(task state.Task) string {
	return cleanupSpecialChars(task.ID)
}

func (provider *Mesos) getFrontEndName(task state.Task) string {
	return strings.Replace(cleanupSpecialChars(task.ID), "/", "-", -1)
}

func cleanupSpecialChars(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(s, ".", "-", -1), ":", "-", -1), "_", "-", -1)
}

func detectMasters(zk string, masters []string) <-chan []string {
	changed := make(chan []string, 1)
	if zk != "" {
		log.Debugf("Starting master detector for ZK ", zk)
		if md, err := detector.New(zk); err != nil {
			log.Errorf("failed to create master detector: %v", err)
		} else if err := md.Detect(detect.NewMasters(masters, changed)); err != nil {
			log.Errorf("failed to initialize master detector: %v", err)
		}
	} else {
		changed <- masters
	}
	return changed
}

func (provider *Mesos) taskRecords(sj state.State) []state.Task {
	var p []state.Task // == nil
	for _, f := range sj.Frameworks {
		for _, task := range f.Tasks {
			for _, slave := range sj.Slaves {
				if task.SlaveID == slave.ID {
					task.SlaveIP = slave.Hostname
				}
			}

			// only do running and discoverable tasks
			if task.State == "TASK_RUNNING" {
				p = append(p, task)
			}
		}
	}

	return p
}

// ErrorFunction A function definition that returns an error
// to be passed to the Ignore or Panic error handler
type ErrorFunction func() error

// Ignore Calls an ErrorFunction, and ignores the result.
// This allows us to be more explicit when there is no error
// handling to be done, for example in defers
func Ignore(f ErrorFunction) {
	_ = f()
}
func (provider *Mesos) getSubDomain(name string) string {
	if provider.GroupsAsSubDomains {
		splitedName := strings.Split(strings.TrimPrefix(name, "/"), "/")
		reverseStringSlice(&splitedName)
		reverseName := strings.Join(splitedName, ".")
		return reverseName
	}
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}
