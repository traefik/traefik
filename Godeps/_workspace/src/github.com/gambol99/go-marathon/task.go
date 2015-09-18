/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Tasks struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID                string               `json:"id"`
	AppID             string               `json:"appId"`
	Host              string               `json:"host"`
	HealthCheckResult []*HealthCheckResult `json:"healthCheckResults"`
	Ports             []int                `json:"ports"`
	ServicePorts      []int                `json:"servicePorts"`
	StagedAt          string               `json:"stagedAt"`
	StartedAt         string               `json:"startedAt"`
	Version           string               `json:"version"`
}

func (task Task) String() string {
	return fmt.Sprintf("id: %s, application: %s, host: %s, ports: %s, created: %s",
		task.ID, task.AppID, task.Host, task.Ports, task.StartedAt)
}

// Check if the task has any health checks
func (task *Task) HasHealthCheckResults() bool {
	if task.HealthCheckResult == nil || len(task.HealthCheckResult) <= 0 {
		return false
	}
	return true
}

// Retrieve all the tasks currently running
func (client *Client) AllTasks() (*Tasks, error) {
	tasks := new(Tasks)
	if err := client.apiGet(MARATHON_API_TASKS, nil, tasks); err != nil {
		return nil, err
	} else {
		return tasks, nil
	}
}

// Retrieve a list of tasks for an application
//		application_id:		the id for the application
func (client *Client) Tasks(id string) (*Tasks, error) {
	tasks := new(Tasks)
	if err := client.apiGet(fmt.Sprintf("%s/%s/tasks", MARATHON_API_APPS, trimRootPath(id)), nil, tasks); err != nil {
		return nil, err
	} else {
		return tasks, nil
	}
}

// Retrieve an array of task ids currently running in marathon
func (client *Client) ListTasks() ([]string, error) {
	if tasks, err := client.AllTasks(); err != nil {
		return nil, err
	} else {
		list := make([]string, 0)
		for _, task := range tasks.Tasks {
			list = append(list, task.ID)
		}
		return list, nil
	}
}

// Kill all tasks relating to an application
//		application_id:		the id for the application
//      host:				kill only those tasks on a specific host (optional)
//		scale:              Scale the app down (i.e. decrement its instances setting by the number of tasks killed) after killing the specified tasks
func (client *Client) KillApplicationTasks(id, hostname string, scale bool) (*Tasks, error) {
	var options struct {
		Host  string `json:"host"`
		Scale bool   `json:bool`
	}
	options.Host = hostname
	options.Scale = scale
	tasks := new(Tasks)
	client.log("KillApplicationTasks() Killing application tasks for: %s, hostname: %s, scale: %t", id, hostname, scale)
	if err := client.apiDelete(fmt.Sprintf("%s/%s/tasks", MARATHON_API_APPS, trimRootPath(id)), &options, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Kill the task associated with a given ID
// 	task_id:		the id for the task
// 	scale:		Scale the app down
func (client *Client) KillTask(taskId string, scale bool) (*Task, error) {
	var options struct {
		Scale bool  `json:bool`
	}
	options.Scale = scale
	task := new(Task)
	appName := taskId[0:strings.LastIndex(taskId, ".")]
	client.log("KillTask Killing task `%s` for: %s, scale: %t", taskId, appName, scale)
	if err := client.apiDelete(fmt.Sprintf("%s/%s/tasks/%s", MARATHON_API_APPS, appName, taskId), &options, task); err != nil {
		return nil, err
	}
	return task, nil
}

// Kill tasks associated with given array of ids
// 	tasks: 	the array of task ids
// 	scale: 	Scale the app down
func (client *Client) KillTasks(tasks []string, scale bool) error {
	v := url.Values{}
	v.Add("scale", strconv.FormatBool(scale))
	var post struct {
		TaskIDs []string `json:"ids"`
	}
	post.TaskIDs = tasks
	client.log("KillTasks Killing %d tasks", len(tasks))
	return client.apiPost(fmt.Sprintf("%s/delete?%s", MARATHON_API_TASKS, v.Encode()), &post, nil)
}

// Get the endpoints i.e. HOST_IP:DYNAMIC_PORT for a specific application service
// I.e. a container running apache, might have ports 80/443 (translated to X dynamic ports), but i want
// port 80 only and i only want those whom have passed the health check
//
// Note: I've NO IDEA how to associate the health_check_result to the actual port, I don't think it's
// possible at the moment, however, given marathon will fail and restart an application even if one of x ports of a task is
// down, the per port check is redundant??? .. personally, I like it anyhow, but hey
//

//		name:		the identifier for the application
//		port:		the container port you are interested in
//		health: 	whether to check the health or not
func (client *Client) TaskEndpoints(name string, port int, health_check bool) ([]string, error) {
	/* step: get the application details */
	if application, err := client.Application(name); err != nil {
		return nil, err
	} else {
		/* step: we need to get the port index of the service we are interested in */
		if port_index, err := application.Container.Docker.ServicePortIndex(port); err != nil {
			return nil, err
		} else {
			list := make([]string, 0)
			/* step: do we have any tasks? */
			if application.Tasks == nil || len(application.Tasks) <= 0 {
				return list, nil
			}

			/* step: iterate the tasks and extract the dynamic ports */
			for _, task := range application.Tasks {
				/* step: if we are checking health the 'service' has a health check? */
				if health_check && application.HasHealthChecks() {
					/*
						check: does the task have a health check result, if NOT, it's because the
						health of the task hasn't yet been performed, hence we assume it as DOWN
					*/
					if task.HasHealthCheckResults() == false {
						client.log("TaskEndpoints() The task: %s for application: %s hasn't been checked yet, skipping", task, application)
						continue
					}

					/* step: check the health results then */
					skip_endpoint := false
					for _, health := range task.HealthCheckResult {
						if health.Alive == false {
							client.log("TaskEndpoints() The task: %s for application: %s failed health checks", task, application)
							skip_endpoint = true
						}
					}

					if skip_endpoint == true {
						continue
					}
				}
				/* else we can just add it */
				list = append(list, fmt.Sprintf("%s:%d", task.Host, task.Ports[port_index]))
			}
			return list, nil
		}
	}
}
