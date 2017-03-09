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

// HealthCheck is the definition for an application health check
type HealthCheck struct {
	Command                *Command `json:"command,omitempty"`
	PortIndex              *int     `json:"portIndex,omitempty"`
	Port                   *int     `json:"port,omitempty"`
	Path                   *string  `json:"path,omitempty"`
	MaxConsecutiveFailures *int     `json:"maxConsecutiveFailures,omitempty"`
	Protocol               string   `json:"protocol,omitempty"`
	GracePeriodSeconds     int      `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds        int      `json:"intervalSeconds,omitempty"`
	TimeoutSeconds         int      `json:"timeoutSeconds,omitempty"`
}

// SetCommand sets the given command on the health check.
func (h HealthCheck) SetCommand(c Command) HealthCheck {
	h.Command = &c
	return h
}

// SetPortIndex sets the given port index on the health check.
func (h HealthCheck) SetPortIndex(i int) HealthCheck {
	h.PortIndex = &i
	return h
}

// SetPort sets the given port on the health check.
func (h HealthCheck) SetPort(i int) HealthCheck {
	h.Port = &i
	return h
}

// SetPath sets the given path on the health check.
func (h HealthCheck) SetPath(p string) HealthCheck {
	h.Path = &p
	return h
}

// SetMaxConsecutiveFailures sets the maximum consecutive failures on the health check.
func (h HealthCheck) SetMaxConsecutiveFailures(i int) HealthCheck {
	h.MaxConsecutiveFailures = &i
	return h
}

// NewDefaultHealthCheck creates a default application health check
func NewDefaultHealthCheck() *HealthCheck {
	portIndex := 0
	path := ""
	maxConsecutiveFailures := 3

	return &HealthCheck{
		Protocol:               "HTTP",
		Path:                   &path,
		PortIndex:              &portIndex,
		MaxConsecutiveFailures: &maxConsecutiveFailures,
		GracePeriodSeconds:     30,
		IntervalSeconds:        10,
		TimeoutSeconds:         5,
	}
}

// HealthCheckResult is the health check result
type HealthCheckResult struct {
	Alive               bool   `json:"alive"`
	ConsecutiveFailures int    `json:"consecutiveFailures"`
	FirstSuccess        string `json:"firstSuccess"`
	LastFailure         string `json:"lastFailure"`
	LastFailureCause    string `json:"lastFailureCause"`
	LastSuccess         string `json:"lastSuccess"`
	TaskID              string `json:"taskId"`
}

// Command is the command health check type
type Command struct {
	Value string `json:"value"`
}
