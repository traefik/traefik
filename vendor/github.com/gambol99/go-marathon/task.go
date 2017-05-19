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
	"strings"
)

// Tasks is a collection of marathon tasks
type Tasks struct {
	Tasks []Task `json:"tasks"`
}

// Task is the definition for a marathon task
type Task struct {
	ID                 string               `json:"id"`
	AppID              string               `json:"appId"`
	Host               string               `json:"host"`
	HealthCheckResults []*HealthCheckResult `json:"healthCheckResults"`
	Ports              []int                `json:"ports"`
	ServicePorts       []int                `json:"servicePorts"`
	SlaveID            string               `json:"slaveId"`
	StagedAt           string               `json:"stagedAt"`
	StartedAt          string               `json:"startedAt"`
	State              string               `json:"state"`
	IPAddresses        []*IPAddress         `json:"ipAddresses"`
	Version            string               `json:"version"`
}

// IPAddress represents a task's IP address and protocol.
type IPAddress struct {
	IPAddress string `json:"ipAddress"`
	Protocol  string `json:"protocol"`
}

// AllTasksOpts contains a payload for AllTasks method
//		status:		Return only those tasks whose status matches this parameter.
//				If not specified, all tasks are returned. Possible values: running, staging. Default: none.
type AllTasksOpts struct {
	Status string `url:"status,omitempty"`
}

// KillApplicationTasksOpts contains a payload for KillApplicationTasks method
//		host:		kill only those tasks on a specific host (optional)
//		scale:		Scale the app down (i.e. decrement its instances setting by the number of tasks killed) after killing the specified tasks
type KillApplicationTasksOpts struct {
	Host  string `url:"host,omitempty"`
	Scale bool   `url:"scale,omitempty"`
	Force bool   `url:"force,omitempty"`
}

// KillTaskOpts contains a payload for task killing methods
//		scale:		Scale the app down
type KillTaskOpts struct {
	Scale bool `url:"scale,omitempty"`
	Force bool `url:"force,omitempty"`
}

// HasHealthCheckResults checks if the task has any health checks
func (r *Task) HasHealthCheckResults() bool {
	return r.HealthCheckResults != nil && len(r.HealthCheckResults) > 0
}

// AllTasks lists tasks of all applications.
//		opts: 		AllTasksOpts request payload
func (r *marathonClient) AllTasks(opts *AllTasksOpts) (*Tasks, error) {
	path, err := addOptions(marathonAPITasks, opts)
	if err != nil {
		return nil, err
	}

	tasks := new(Tasks)
	if err := r.apiGet(path, nil, tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Tasks retrieves a list of tasks for an application
//		id:		the id of the application
func (r *marathonClient) Tasks(id string) (*Tasks, error) {
	tasks := new(Tasks)
	if err := r.apiGet(fmt.Sprintf("%s/%s/tasks", marathonAPIApps, trimRootPath(id)), nil, tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// KillApplicationTasks kills all tasks relating to an application
//		id:		the id of the application
//		opts: 		KillApplicationTasksOpts request payload
func (r *marathonClient) KillApplicationTasks(id string, opts *KillApplicationTasksOpts) (*Tasks, error) {
	path := fmt.Sprintf("%s/%s/tasks", marathonAPIApps, trimRootPath(id))
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, err
	}

	tasks := new(Tasks)
	if err := r.apiDelete(path, nil, tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// KillTask kills the task associated with a given ID
// 	taskID:		the id for the task
//	opts:		KillTaskOpts request payload
func (r *marathonClient) KillTask(taskID string, opts *KillTaskOpts) (*Task, error) {
	appName := taskID[0:strings.LastIndex(taskID, ".")]
	appName = strings.Replace(appName, "_", "/", -1)
	taskID = strings.Replace(taskID, "/", "_", -1)

	path := fmt.Sprintf("%s/%s/tasks/%s", marathonAPIApps, appName, taskID)
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, err
	}

	wrappedTask := new(struct {
		Task Task `json:"task"`
	})

	if err := r.apiDelete(path, nil, wrappedTask); err != nil {
		return nil, err
	}

	return &wrappedTask.Task, nil
}

// KillTasks kills tasks associated with given array of ids
//	tasks:		the array of task ids
//	opts:		KillTaskOpts request payload
func (r *marathonClient) KillTasks(tasks []string, opts *KillTaskOpts) error {
	path := fmt.Sprintf("%s/delete", marathonAPITasks)
	path, err := addOptions(path, opts)
	if err != nil {
		return nil
	}

	var post struct {
		IDs []string `json:"ids"`
	}
	post.IDs = tasks

	return r.apiPost(path, &post, nil)
}

// TaskEndpoints gets the endpoints i.e. HOST_IP:DYNAMIC_PORT for a specific application service
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
func (r *marathonClient) TaskEndpoints(name string, port int, healthCheck bool) ([]string, error) {
	// step: get the application details
	application, err := r.Application(name)
	if err != nil {
		return nil, err
	}

	// step: we need to get the port index of the service we are interested in
	portIndex, err := application.Container.Docker.ServicePortIndex(port)
	if err != nil {
		return nil, err
	}

	// step: do we have any tasks?
	if application.Tasks == nil || len(application.Tasks) == 0 {
		return nil, nil
	}

	// step: if we are checking health the 'service' has a health check?
	healthCheck = healthCheck && application.HasHealthChecks()

	// step: iterate the tasks and extract the dynamic ports
	var list []string
	for _, task := range application.Tasks {
		if !healthCheck || task.allHealthChecksAlive() {
			endpoint := fmt.Sprintf("%s:%d", task.Host, task.Ports[portIndex])
			list = append(list, endpoint)
		}
	}

	return list, nil
}

func (r *Task) allHealthChecksAlive() bool {
	// check: does the task have a health check result, if NOT, it's because the
	// health of the task hasn't yet been performed, hence we assume it as DOWN
	if !r.HasHealthCheckResults() {
		return false
	}
	// step: check the health results then
	for _, check := range r.HealthCheckResults {
		if check.Alive == false {
			return false
		}
	}

	return true
}
