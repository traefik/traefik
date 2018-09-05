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

// TaskLostBehaviorType sets action taken when the resident task is lost
type TaskLostBehaviorType string

const (
	// TaskLostBehaviorTypeWaitForever indicates to not take any action when the resident task is lost
	TaskLostBehaviorTypeWaitForever TaskLostBehaviorType = "WAIT_FOREVER"
	// TaskLostBehaviorTypeRelaunchAfterTimeout indicates to try relaunching the lost resident task on
	// another node after the relaunch escalation timeout has elapsed
	TaskLostBehaviorTypeRelaunchAfterTimeout TaskLostBehaviorType = "RELAUNCH_AFTER_TIMEOUT"
)

// Residency defines how terminal states of tasks with local persistent volumes are handled
type Residency struct {
	TaskLostBehavior                 TaskLostBehaviorType `json:"taskLostBehavior,omitempty"`
	RelaunchEscalationTimeoutSeconds int                  `json:"relaunchEscalationTimeoutSeconds,omitempty"`
}

// SetTaskLostBehavior sets the residency behavior
func (r *Residency) SetTaskLostBehavior(behavior TaskLostBehaviorType) *Residency {
	r.TaskLostBehavior = behavior
	return r
}

// SetRelaunchEscalationTimeout sets the residency relaunch escalation timeout with seconds precision
func (r *Residency) SetRelaunchEscalationTimeout(timeout time.Duration) *Residency {
	r.RelaunchEscalationTimeoutSeconds = int(timeout.Seconds())
	return r
}
