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

// PodState defines the state of a pod
type PodState string

const (
	// PodStateDegraded is a degraded pod
	PodStateDegraded PodState = "DEGRADED"

	// PodStateStable is a stable pod
	PodStateStable PodState = "STABLE"

	// PodStateTerminal is a terminal pod
	PodStateTerminal PodState = "TERMINAL"
)

// PodStatus describes the pod status
type PodStatus struct {
	ID                 string                   `json:"id,omitempty"`
	Spec               *Pod                     `json:"spec,omitempty"`
	Status             PodState                 `json:"status,omitempty"`
	StatusSince        string                   `json:"statusSince,omitempty"`
	Message            string                   `json:"message,omitempty"`
	Instances          []*PodInstanceStatus     `json:"instances,omitempty"`
	TerminationHistory []*PodTerminationHistory `json:"terminationHistory,omitempty"`
	LastUpdated        string                   `json:"lastUpdated,omitempty"`
	LastChanged        string                   `json:"lastChanged,omitempty"`
}

// PodTerminationHistory is the termination history of the pod
type PodTerminationHistory struct {
	InstanceID   string                         `json:"instanceId,omitempty"`
	StartedAt    string                         `json:"startedAt,omitempty"`
	TerminatedAt string                         `json:"terminatedAt,omitempty"`
	Message      string                         `json:"message,omitempty"`
	Containers   []*ContainerTerminationHistory `json:"containers,omitempty"`
}

// ContainerTerminationHistory is the termination history of a container in a pod
type ContainerTerminationHistory struct {
	ContainerID    string                     `json:"containerId,omitempty"`
	LastKnownState string                     `json:"lastKnownState,omitempty"`
	Termination    *ContainerTerminationState `json:"termination,omitempty"`
}

// PodStatus retrieves the pod configuration from marathon
func (r *marathonClient) PodStatus(name string) (*PodStatus, error) {
	var podStatus PodStatus

	if err := r.apiGet(buildPodStatusURI(name), nil, &podStatus); err != nil {
		return nil, err
	}

	return &podStatus, nil
}

// PodStatuses retrieves all pod configuration from marathon
func (r *marathonClient) PodStatuses() ([]*PodStatus, error) {
	var podStatuses []*PodStatus

	if err := r.apiGet(buildPodStatusURI(""), nil, &podStatuses); err != nil {
		return nil, err
	}

	return podStatuses, nil
}

// WaitOnPod blocks until a pod to be deployed
func (r *marathonClient) WaitOnPod(name string, timeout time.Duration) error {
	return r.wait(name, timeout, r.PodIsRunning)
}

// PodIsRunning returns whether the pod is stably running
func (r *marathonClient) PodIsRunning(name string) bool {
	podStatus, err := r.PodStatus(name)
	if apiErr, ok := err.(*APIError); ok && apiErr.ErrCode == ErrCodeNotFound {
		return false
	}
	if err == nil && podStatus.Status == PodStateStable {
		return true
	}
	return false
}

func buildPodStatusURI(path string) string {
	return fmt.Sprintf("%s/%s::status", marathonAPIPods, trimRootPath(path))
}
