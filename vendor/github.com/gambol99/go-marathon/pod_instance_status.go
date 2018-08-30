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

// PodInstanceState is the state of a specific pod instance
type PodInstanceState string

const (
	// PodInstanceStatePending is when an instance is pending scheduling
	PodInstanceStatePending PodInstanceState = "PENDING"

	// PodInstanceStateStaging is when an instance is staged to be scheduled
	PodInstanceStateStaging PodInstanceState = "STAGING"

	// PodInstanceStateStable is when an instance is stably running
	PodInstanceStateStable PodInstanceState = "STABLE"

	// PodInstanceStateDegraded is when an instance is degraded status
	PodInstanceStateDegraded PodInstanceState = "DEGRADED"

	// PodInstanceStateTerminal is when an instance is terminal
	PodInstanceStateTerminal PodInstanceState = "TERMINAL"
)

// PodInstanceStatus is the status of a pod instance
type PodInstanceStatus struct {
	AgentHostname string              `json:"agentHostname,omitempty"`
	Conditions    []*StatusCondition  `json:"conditions,omitempty"`
	Containers    []*ContainerStatus  `json:"containers,omitempty"`
	ID            string              `json:"id,omitempty"`
	LastChanged   string              `json:"lastChanged,omitempty"`
	LastUpdated   string              `json:"lastUpdated,omitempty"`
	Message       string              `json:"message,omitempty"`
	Networks      []*PodNetworkStatus `json:"networks,omitempty"`
	Resources     *Resources          `json:"resources,omitempty"`
	SpecReference string              `json:"specReference,omitempty"`
	Status        PodInstanceState    `json:"status,omitempty"`
	StatusSince   string              `json:"statusSince,omitempty"`
}

// PodNetworkStatus is the networks attached to a pod instance
type PodNetworkStatus struct {
	Addresses []string `json:"addresses,omitempty"`
	Name      string   `json:"name,omitempty"`
}

// StatusCondition describes info about a status change
type StatusCondition struct {
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Reason      string `json:"reason,omitempty"`
	LastChanged string `json:"lastChanged,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

// ContainerStatus contains all status information for a container instance
type ContainerStatus struct {
	Conditions  []*StatusCondition         `json:"conditions,omitempty"`
	ContainerID string                     `json:"containerId,omitempty"`
	Endpoints   []*PodEndpoint             `json:"endpoints,omitempty"`
	LastChanged string                     `json:"lastChanged,omitempty"`
	LastUpdated string                     `json:"lastUpdated,omitempty"`
	Message     string                     `json:"message,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Resources   *Resources                 `json:"resources,omitempty"`
	Status      string                     `json:"status,omitempty"`
	StatusSince string                     `json:"statusSince,omitempty"`
	Termination *ContainerTerminationState `json:"termination,omitempty"`
}

// ContainerTerminationState describes why a container terminated
type ContainerTerminationState struct {
	ExitCode int    `json:"exitCode,omitempty"`
	Message  string `json:"message,omitempty"`
}
