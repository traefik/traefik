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

import "time"

// ReadinessCheck represents a readiness check.
type ReadinessCheck struct {
	Name                    *string `json:"name,omitempty"`
	Protocol                string  `json:"protocol,omitempty"`
	Path                    string  `json:"path,omitempty"`
	PortName                string  `json:"portName,omitempty"`
	IntervalSeconds         int     `json:"intervalSeconds,omitempty"`
	TimeoutSeconds          int     `json:"timeoutSeconds,omitempty"`
	HTTPStatusCodesForReady *[]int  `json:"httpStatusCodesForReady,omitempty"`
	PreserveLastResponse    *bool   `json:"preserveLastResponse,omitempty"`
}

// SetName sets the name on the readiness check.
func (rc *ReadinessCheck) SetName(name string) *ReadinessCheck {
	rc.Name = &name
	return rc
}

// SetProtocol sets the protocol on the readiness check.
func (rc *ReadinessCheck) SetProtocol(proto string) *ReadinessCheck {
	rc.Protocol = proto
	return rc
}

// SetPath sets the path on the readiness check.
func (rc *ReadinessCheck) SetPath(p string) *ReadinessCheck {
	rc.Path = p
	return rc
}

// SetPortName sets the port name on the readiness check.
func (rc *ReadinessCheck) SetPortName(name string) *ReadinessCheck {
	rc.PortName = name
	return rc
}

// SetInterval sets the interval on the readiness check.
func (rc *ReadinessCheck) SetInterval(interval time.Duration) *ReadinessCheck {
	secs := int(interval.Seconds())
	rc.IntervalSeconds = secs
	return rc
}

// SetTimeout sets the timeout on the readiness check.
func (rc *ReadinessCheck) SetTimeout(timeout time.Duration) *ReadinessCheck {
	secs := int(timeout.Seconds())
	rc.TimeoutSeconds = secs
	return rc
}

// SetHTTPStatusCodesForReady sets the HTTP status codes for ready on the
// readiness check.
func (rc *ReadinessCheck) SetHTTPStatusCodesForReady(codes []int) *ReadinessCheck {
	rc.HTTPStatusCodesForReady = &codes
	return rc
}

// SetPreserveLastResponse sets the preserve last response flag on the
// readiness check.
func (rc *ReadinessCheck) SetPreserveLastResponse(preserve bool) *ReadinessCheck {
	rc.PreserveLastResponse = &preserve
	return rc
}

// ReadinessLastResponse holds the result of the last response embedded in a
// readiness check result.
type ReadinessLastResponse struct {
	Body        string `json:"body"`
	ContentType string `json:"contentType"`
	Status      int    `json:"status"`
}

// ReadinessCheckResult is the result of a readiness check.
type ReadinessCheckResult struct {
	Name         string                `json:"name"`
	TaskID       string                `json:"taskId"`
	Ready        bool                  `json:"ready"`
	LastResponse ReadinessLastResponse `json:"lastResponse,omitempty"`
}
