/*
Copyright 2017 The go-marathon Authors All rights reserved.

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
	"time"
)

// PodInstance is the representation of an instance as returned by deleting an instance
type PodInstance struct {
	InstanceID          PodInstanceID              `json:"instanceId"`
	AgentInfo           PodAgentInfo               `json:"agentInfo"`
	TasksMap            map[string]PodTask         `json:"tasksMap"`
	RunSpecVersion      time.Time                  `json:"runSpecVersion"`
	State               PodInstanceStateHistory    `json:"state"`
	UnreachableStrategy EnabledUnreachableStrategy `json:"unreachableStrategy"`
}

// PodInstanceStateHistory is the pod instance's state
type PodInstanceStateHistory struct {
	Condition   PodTaskCondition `json:"condition"`
	Since       time.Time        `json:"since"`
	ActiveSince time.Time        `json:"activeSince"`
}

// PodInstanceID contains the instance ID
type PodInstanceID struct {
	ID string `json:"idString"`
}

// PodAgentInfo contains info about the agent the instance is running on
type PodAgentInfo struct {
	Host       string   `json:"host"`
	AgentID    string   `json:"agentId"`
	Attributes []string `json:"attributes"`
}

// PodTask contains the info about the specific task within the instance
type PodTask struct {
	TaskID         string        `json:"taskId"`
	RunSpecVersion time.Time     `json:"runSpecVersion"`
	Status         PodTaskStatus `json:"status"`
}

// PodTaskStatus is the current status of the task
type PodTaskStatus struct {
	StagedAt    time.Time        `json:"stagedAt"`
	StartedAt   time.Time        `json:"startedAt"`
	MesosStatus string           `json:"mesosStatus"`
	Condition   PodTaskCondition `json:"condition"`
	NetworkInfo PodNetworkInfo   `json:"networkInfo"`
}

// PodTaskCondition contains a string representation of the condition
type PodTaskCondition struct {
	Str string `json:"str"`
}

// PodNetworkInfo contains the network info for a task
type PodNetworkInfo struct {
	HostName    string      `json:"hostName"`
	HostPorts   []int       `json:"hostPorts"`
	IPAddresses []IPAddress `json:"ipAddresses"`
}

// DeletePodInstances deletes all instances of the named pod
func (r *marathonClient) DeletePodInstances(name string, instances []string) ([]*PodInstance, error) {
	uri := buildPodInstancesURI(name)
	var result []*PodInstance
	if err := r.apiDelete(uri, instances, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeletePodInstance deletes a specific instance of a pod
func (r *marathonClient) DeletePodInstance(name, instance string) (*PodInstance, error) {
	uri := fmt.Sprintf("%s/%s", buildPodInstancesURI(name), instance)
	result := new(PodInstance)
	if err := r.apiDelete(uri, nil, result); err != nil {
		return nil, err
	}

	return result, nil
}

func buildPodInstancesURI(path string) string {
	return fmt.Sprintf("%s/%s::instances", marathonAPIPods, trimRootPath(path))
}
