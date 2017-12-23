package mesos

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/mesosphere/mesos-dns/util"
)

var _ provider.Provider = (*Provider)(nil)

//Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint              string `description:"Mesos server endpoint. You can also specify multiple endpoint for Mesos"`
	Domain                string `description:"Default domain used"`
	ExposedByDefault      bool   `description:"Expose Mesos apps by default" export:"true"`
	GroupsAsSubDomains    bool   `description:"Convert Mesos groups to subdomains" export:"true"`
	ZkDetectionTimeout    int    `description:"Zookeeper timeout (in seconds)" export:"true"`
	RefreshSeconds        int    `description:"Polling interval (in seconds)" export:"true"`
	IPSources             string `description:"IPSources (e.g. host, docker, mesos, rkt)" export:"true"`
	StateTimeoutSecond    int    `description:"HTTP Timeout (in seconds)" export:"true"`
	Masters               []string
	Subscribe             bool   `description:"Subscribe to Mesos task events (Mesos 1.1 or later only)" export:"true"`
	SubscribeLabels       string `description:"Filter subscriptions to only tasks with these labels" export:"true"`
	SubscribeFilterLabels []string
}

type mesosEvent struct {
	Type      string `json:"type"`
	TaskAdded struct {
		Task struct {
			TaskID struct {
				Value string `json:"value"`
			} `json:"task_id"`
			Labels struct {
				Labels []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"labels"`
			} `json:"labels"`
		} `json:"task"`
	} `json:"task_added"`
	TaskUpdated struct {
		FrameworkID struct {
			Value string `json:"value"`
		} `json:"framework_id"`
		State  string `json:"state"`
		Status struct {
			TaskID struct {
				Value string `json:"value"`
			} `json:"task_id"`
			AgentID struct {
				Value string `json:"value"`
			} `json:"agent_id"`
		} `json:"status"`
	} `json:"task_updated"`
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
		var reload *time.Ticker
		var taskAddedChan <-chan state.Task
		var taskUpdatedChan <-chan state.Task

		if strings.HasPrefix(p.Endpoint, "zk://") {
			zk = p.Endpoint
		} else {
			masters = strings.Split(p.Endpoint, ",")
		}

		errch := make(chan error)
		subscribed := false

		changed := detectMasters(zk, masters)
		if p.RefreshSeconds == 0 {
			reload = time.NewTicker(time.Second * 100)
		} else {
			reload = time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
		}
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
		if p.RefreshSeconds == 0 {
			reload.Stop()
		}

		if len(p.SubscribeLabels) > 0 {
			p.SubscribeFilterLabels = strings.Split(p.SubscribeLabels, ",")
		}

		runningTasks := make(map[string]string)
		updateTasks := func(tasks []state.Task) {
			for k, v := range runningTasks {
				if v != "ADDED" {
					delete(runningTasks, k)
				}
			}
			for _, t := range tasks {
				runningTasks[t.ID] = "RUNNING"
			}
		}

		for {
			select {
			case <-reload.C:
				configuration, tasks := p.buildConfiguration()
				if configuration != nil {
					updateTasks(tasks)
					configurationChan <- types.ConfigMessage{
						ProviderName:  "mesos",
						Configuration: configuration,
					}
				}
			case task, ok := <-taskAddedChan:
				if ok {
					// Keep track that the ask was added, but wait until
					// it is in the running state to rebuild the configuration
					if _, ok := runningTasks[task.ID]; !ok {
						runningTasks[task.ID] = "ADDED"
					}
				} else {
					// the subscribe listener closed, restart it (maybe
					// there was a leader re-election).
					if len(p.Masters) == 0 {
						subscribed = false
						taskAddedChan = nil
					} else {
						taskAddedChan, taskUpdatedChan = p.subscribeMesos(p.Masters)
						if taskAddedChan == nil {
							subscribed = false
						}
					}
				}
			case task, ok := <-taskUpdatedChan:
				if ok {
					var configuration *types.Configuration
					var tasks []state.Task
					if taskStatus, ok := runningTasks[task.ID]; ok {
						if taskStatus == "ADDED" {
							if task.State == "RUNNING" {
								configuration, tasks = p.buildConfiguration()
							} else {
								delete(runningTasks, task.ID)
							}
						} else {
							if task.State == "EXITED" {
								delete(runningTasks, task.ID)
								configuration, tasks = p.buildConfiguration()
							}
						}
					} else {
						if task.State == "RUNNING" && len(p.SubscribeLabels) == 0 {
							configuration, tasks = p.buildConfiguration()
						}
					}
					if configuration != nil {
						updateTasks(tasks)
						configurationChan <- types.ConfigMessage{
							ProviderName:  "mesos",
							Configuration: configuration,
						}
					}
				} else {
					taskUpdatedChan = nil
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
				if p.Subscribe && len(masters) > 0 && !subscribed {
					taskAddedChan, taskUpdatedChan = p.subscribeMesos(p.Masters)
					subscribed = taskAddedChan != nil
				}
				configuration, tasks := p.buildConfiguration()
				if configuration != nil {
					updateTasks(tasks)
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
		log.Debugf("Starting master detector for ZK ", zk)
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

func (p *Provider) subscribeMesos(masters []string) (<-chan state.Task, <-chan state.Task) {
	var subscribeResp *http.Response
	var masterURI url.URL

	for _, master := range masters {
		log.Debugf("Attempting to connect to master %v", master)
		masterURI = url.URL{
			Scheme: "http",
			Host:   master,
			Path:   "api/v1",
		}
		r, err := http.NewRequest("POST", masterURI.String(), strings.NewReader(`{"type":"SUBSCRIBE"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "application/json")
		if err != nil {
			log.Errorf("Failed to initialize mesos task subscription: %v", err)
			continue
		}
		resp, err := (&http.Client{}).Do(r)
		if err != nil {
			log.Errorf("Failed to start mesos task subscription: %v", err)
			continue
		}
		subscribeResp = resp
		break
	}
	if subscribeResp == nil {
		return nil, nil
	}

	taskAdded := make(chan state.Task)
	taskUpdated := make(chan state.Task)
	go func(resp *http.Response) {
		defer func() {
			close(taskAdded)
			close(taskUpdated)
		}()
		log.Debugf("Starting mesos task subscription on %v", masterURI.String())
		reader := bufio.NewReader(resp.Body)
		readBuff := []byte{}
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Errorf("Error reading length response from mesos subscription: %v", err)
				return
			}
			line = line[:len(line)-1]
			if sz, err := strconv.Atoi(line); err == nil {
				if cap(readBuff) < sz {
					readBuff = make([]byte, sz)
				} else {
					readBuff = readBuff[:sz]
				}
				if n, err := io.ReadFull(reader, readBuff); err != nil {
					log.Errorf("Error reading data response from mesos subscription: %v", err)
					return
				} else if n != sz {
					log.Errorf("Error reading data response from mesos subscription: expected data of length %d but got %d", sz, n)
					return
				}
				var event mesosEvent
				if err := json.Unmarshal(readBuff, &event); err == nil {
					log.Debugf("Received event of type %v from mesos.", event.Type)
					if event.Type == "TASK_ADDED" {
						if len(p.SubscribeFilterLabels) > 0 {
						matchLabel:
							for _, label := range p.SubscribeFilterLabels {
								for _, taskLabel := range event.TaskAdded.Task.Labels.Labels {
									if taskLabel.Key == label {
										taskAdded <- state.Task{
											ID: event.TaskAdded.Task.TaskID.Value,
										}
										break matchLabel
									}
								}
							}
						}
					} else if event.Type == "TASK_UPDATED" {
						t := state.Task{
							ID: event.TaskUpdated.Status.TaskID.Value,
						}
						switch event.TaskUpdated.State {
						case "TASK_RUNNING":
							t.State = "RUNNING"
						case "TASK_FINISHED", "TASK_FAILED", "TASK_KILLED", "TASK_ERROR",
							"TASK_DROPPED", "TASK_GONE":
							t.State = "EXITED"
						}

						if t.State != "" {
							taskUpdated <- t
						}
					}
				} else {
					log.Errorf("Error parsing data response from mesos subscription: %v", err)
					return
				}
			} else {
				log.Errorf("Error parsing length response from mesos subscription: %v", err)
				return
			}
		}
	}(subscribeResp)
	return taskAdded, taskUpdated
}
