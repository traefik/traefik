package mesos

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/mesos/mesos-go/detector"
	// Register mesos zoo the detector
	_ "github.com/mesos/mesos-go/detector/zoo"
	"github.com/mesosphere/mesos-dns/detect"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/mesosphere/mesos-dns/util"
)

var _ provider.Provider = (*Provider)(nil)

//Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
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

// Provide allows the mesos provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	operation := func() error {

		// initialize logging
		logging.SetupLogs()

		log.Debugf("%s", p.IPSources)

		var zk string
		var masters []string

		if strings.HasPrefix(p.Endpoint, "zk://") {
			zk = p.Endpoint
		} else {
			masters = strings.Split(p.Endpoint, ",")
		}

		errch := make(chan error)

		changed := detectMasters(zk, masters)
		reload := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
		zkTimeout := time.Second * time.Duration(p.ZkDetectionTimeout)
		timeout := time.AfterFunc(zkTimeout, func() {
			if zkTimeout > 0 {
				errch <- fmt.Errorf("master detection timed out after %s", zkTimeout)
			}
		})

		defer reload.Stop()
		defer util.HandleCrash()

		if !p.Watch {
			reload.Stop()
			timeout.Stop()
		}

		for {
			select {
			case <-reload.C:
				configuration := p.loadMesosConfig()
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
				p.Masters = masters
				configuration := p.loadMesosConfig()
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

func (p *Provider) loadMesosConfig() *types.Configuration {
	var mesosFuncMap = template.FuncMap{
		"getBackend":         p.getBackend,
		"getPort":            p.getPort,
		"getHost":            p.getHost,
		"getWeight":          p.getWeight,
		"getDomain":          p.getDomain,
		"getProtocol":        p.getProtocol,
		"getPassHostHeader":  p.getPassHostHeader,
		"getPriority":        p.getPriority,
		"getEntryPoints":     p.getEntryPoints,
		"getFrontendRule":    p.getFrontendRule,
		"getFrontendBackend": p.getFrontendBackend,
		"getID":              p.getID,
		"getFrontEndName":    p.getFrontEndName,
	}

	t := records.NewRecordGenerator(time.Duration(p.StateTimeoutSecond) * time.Second)
	sj, err := t.FindMaster(p.Masters...)
	if err != nil {
		log.Errorf("Failed to create a client for mesos, error: %s", err)
		return nil
	}
	tasks := p.taskRecords(sj)

	//filter tasks
	filteredTasks := fun.Filter(func(task state.Task) bool {
		return mesosTaskFilter(task, p.ExposedByDefault)
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
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/mesos.tmpl", mesosFuncMap, templateObjects)
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

func (p *Provider) getLabel(task state.Task, label string) (string, error) {
	for _, tmpLabel := range task.Labels {
		if tmpLabel.Key == label {
			return tmpLabel.Value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func (p *Provider) getPort(task state.Task, applications []state.Task) string {
	application, err := getMesos(task, applications)
	if err != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return ""
	}

	if portIndexLabel, err := p.getLabel(application, "traefik.portIndex"); err == nil {
		if index, err := strconv.Atoi(portIndexLabel); err == nil {
			return strconv.Itoa(task.DiscoveryInfo.Ports.DiscoveryPorts[index].Number)
		}
	}
	if portValueLabel, err := p.getLabel(application, "traefik.port"); err == nil {
		return portValueLabel
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		return strconv.Itoa(port.Number)
	}
	return ""
}

func (p *Provider) getWeight(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return "0"
	}

	if label, err := p.getLabel(application, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (p *Provider) getDomain(task state.Task) string {
	if label, err := p.getLabel(task, "traefik.domain"); err == nil {
		return label
	}
	return p.Domain
}

func (p *Provider) getProtocol(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return "http"
	}
	if label, err := p.getLabel(application, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (p *Provider) getPassHostHeader(task state.Task) string {
	if passHostHeader, err := p.getLabel(task, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "false"
}

func (p *Provider) getPriority(task state.Task) string {
	if priority, err := p.getLabel(task, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (p *Provider) getEntryPoints(task state.Task) []string {
	if entryPoints, err := p.getLabel(task, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

// getFrontendRule returns the frontend rule for the specified application, using
// it's label. It returns a default one (Host) if the label is not present.
func (p *Provider) getFrontendRule(task state.Task) string {
	if label, err := p.getLabel(task, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host:" + strings.ToLower(strings.Replace(p.getSubDomain(task.DiscoveryInfo.Name), "_", "-", -1)) + "." + p.Domain
}

func (p *Provider) getBackend(task state.Task, applications []state.Task) string {
	application, errApp := getMesos(task, applications)
	if errApp != nil {
		log.Errorf("Unable to get mesos application from task %s", task.DiscoveryInfo.Name)
		return ""
	}
	return p.getFrontendBackend(application)
}

func (p *Provider) getFrontendBackend(task state.Task) string {
	if label, err := p.getLabel(task, "traefik.backend"); err == nil {
		return label
	}
	return "-" + cleanupSpecialChars(task.DiscoveryInfo.Name)
}

func (p *Provider) getHost(task state.Task) string {
	return task.IP(strings.Split(p.IPSources, ",")...)
}

func (p *Provider) getID(task state.Task) string {
	return cleanupSpecialChars(task.ID)
}

func (p *Provider) getFrontEndName(task state.Task) string {
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

func (p *Provider) taskRecords(sj state.State) []state.Task {
	var tasks []state.Task // == nil
	for _, f := range sj.Frameworks {
		for _, task := range f.Tasks {
			for _, slave := range sj.Slaves {
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

// ErrorFunction A function definition that returns an error
// to be passed to the Ignore or Panic error handler
type ErrorFunction func() error

// Ignore Calls an ErrorFunction, and ignores the result.
// This allows us to be more explicit when there is no error
// handling to be done, for example in defers
func Ignore(f ErrorFunction) {
	_ = f()
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
