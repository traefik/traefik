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

type HealthCheck struct {
	Command                *Command `json:"command,omitempty"`
	Protocol               string   `json:"protocol,omitempty"`
	Path                   string   `json:"path,omitempty"`
	GracePeriodSeconds     int      `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds        int      `json:"intervalSeconds,omitempty"`
	PortIndex              int      `json:"portIndex,omitempty"`
	MaxConsecutiveFailures int      `json:"maxConsecutiveFailures,omitempty"`
	TimeoutSeconds         int      `json:"timeoutSeconds,omitempty"`
}

func NewDefaultHealthCheck() *HealthCheck {
	return &HealthCheck{
		Protocol:               "HTTP",
		Path:                   "",
		GracePeriodSeconds:     30,
		IntervalSeconds:        10,
		PortIndex:              0,
		MaxConsecutiveFailures: 3,
		TimeoutSeconds:         5,
	}
}

type HealthCheckResult struct {
	Alive               bool   `json:"alive"`
	ConsecutiveFailures int    `json:"consecutiveFailures"`
	FirstSuccess        string `json:"firstSuccess"`
	LastFailure         string `json:"lastFailure"`
	LastSuccess         string `json:"lastSuccess"`
	TaskID              string `json:"taskId"`
}

type Command struct {
	Value string `json:"value"`
}
