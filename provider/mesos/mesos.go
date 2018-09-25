package mesos

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/mesos/mesos-go/detector"
	"github.com/mesosphere/mesos-dns/records"
	"github.com/mesosphere/mesos-dns/records/state"

	// Register mesos zoo the detector
	_ "github.com/mesos/mesos-go/detector/zoo"
	"github.com/mesosphere/mesos-dns/detect"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/util"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint           string `description:"Mesos server endpoint. You can also specify multiple endpoint for Mesos"`
	Domain             string `description:"Default domain used"`
	ExposedByDefault   bool   `description:"Expose Mesos apps by default" export:"true"`
	GroupsAsSubDomains bool   `description:"Convert Mesos groups to subdomains" export:"true"`
	ZkDetectionTimeout int    `description:"Zookeeper timeout (in seconds)" export:"true"`
	RefreshSeconds     int    `description:"Polling interval (in seconds)" export:"true"`
	IPSources          string `description:"IPSources (e.g. host, docker, mesos, netinfo)" export:"true"`
	StateTimeoutSecond int    `description:"HTTP Timeout (in seconds)" export:"true"`
	Masters            []string
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// Provide allows the mesos provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
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
				tasks := p.getTasks()
				configuration := p.buildConfiguration(tasks)
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
				tasks := p.getTasks()
				configuration := p.buildConfiguration(tasks)
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
		log.Errorf("Mesos connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Mesos server %+v", err)
	}
	return nil
}

func detectMasters(zk string, masters []string) <-chan []string {
	changed := make(chan []string, 1)
	if zk != "" {
		log.Debugf("Starting master detector for ZK %s", zk)
		if md, err := detector.New(zk); err != nil {
			log.Errorf("Failed to create master detector: %v", err)
		} else if err := md.Detect(detect.NewMasters(masters, changed)); err != nil {
			log.Errorf("Failed to initialize master detector: %v", err)
		}
	} else {
		changed <- masters
	}
	return changed
}

func (p *Provider) getTasks() []state.Task {
	rg := records.NewRecordGenerator(time.Duration(p.StateTimeoutSecond) * time.Second)

	st, err := rg.FindMaster(p.Masters...)
	if err != nil {
		log.Errorf("Failed to create a client for Mesos, error: %v", err)
		return nil
	}

	return taskRecords(st)
}

func taskRecords(st state.State) []state.Task {
	var tasks []state.Task
	for _, f := range st.Frameworks {
		for _, task := range f.Tasks {
			for _, slave := range st.Slaves {
				if task.SlaveID == slave.ID {
					task.SlaveIP = slave.PID.Host
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
