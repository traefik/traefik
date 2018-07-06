/*
Copyright 2014 The go-marathon Authors All rights reserved.

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
	"encoding/json"
	"fmt"
	"time"
)

// Deployment is a marathon deployment definition
type Deployment struct {
	ID             string              `json:"id"`
	Version        string              `json:"version"`
	CurrentStep    int                 `json:"currentStep"`
	TotalSteps     int                 `json:"totalSteps"`
	AffectedApps   []string            `json:"affectedApps"`
	AffectedPods   []string            `json:"affectedPods"`
	Steps          [][]*DeploymentStep `json:"-"`
	XXStepsRaw     json.RawMessage     `json:"steps"` // Holds raw steps JSON to unmarshal later
	CurrentActions []*DeploymentStep   `json:"currentActions"`
}

// DeploymentID is the identifier for a application deployment
type DeploymentID struct {
	DeploymentID string `json:"deploymentId"`
	Version      string `json:"version"`
}

// DeploymentStep is a step in the application deployment plan
type DeploymentStep struct {
	Action                string                  `json:"action"`
	App                   string                  `json:"app"`
	ReadinessCheckResults *[]ReadinessCheckResult `json:"readinessCheckResults,omitempty"`
}

// StepActions is a series of deployment steps
type StepActions struct {
	Actions []struct {
		Action string `json:"action"` // 1.1.2 and after
		Type   string `json:"type"`   // 1.1.1 and before
		App    string `json:"app"`
	}
}

// DeploymentPlan is a collection of steps for application deployment
type DeploymentPlan struct {
	ID       string         `json:"id"`
	Version  string         `json:"version"`
	Original *Group         `json:"original"`
	Target   *Group         `json:"target"`
	Steps    []*StepActions `json:"steps"`
}

// Deployments retrieves a list of current deployments
func (r *marathonClient) Deployments() ([]*Deployment, error) {
	var deployments []*Deployment
	err := r.apiGet(marathonAPIDeployments, nil, &deployments)
	if err != nil {
		return nil, err
	}
	// Allows loading of deployment steps from the Marathon v1.X API
	// Implements a fix for issue https://github.com/gambol99/go-marathon/issues/153
	for _, deployment := range deployments {
		// Unmarshal pre-v1.X step
		if err := json.Unmarshal(deployment.XXStepsRaw, &deployment.Steps); err != nil {
			deployment.Steps = make([][]*DeploymentStep, 0)
			var steps []*StepActions
			// Unmarshal v1.X Marathon step
			if err := json.Unmarshal(deployment.XXStepsRaw, &steps); err != nil {
				return nil, err
			}
			for stepIndex, step := range steps {
				deployment.Steps = append(deployment.Steps, make([]*DeploymentStep, len(step.Actions)))
				for actionIndex, action := range step.Actions {
					var stepAction string
					if action.Type != "" {
						stepAction = action.Type
					} else {
						stepAction = action.Action
					}
					deployment.Steps[stepIndex][actionIndex] = &DeploymentStep{
						Action: stepAction,
						App:    action.App,
					}
				}
			}
		}
	}
	return deployments, nil
}

// DeleteDeployment delete a current deployment from marathon
// 	id:		the deployment id you wish to delete
// 	force:	whether or not to force the deletion
func (r *marathonClient) DeleteDeployment(id string, force bool) (*DeploymentID, error) {
	path := fmt.Sprintf("%s/%s", marathonAPIDeployments, id)

	// if force=true, no body is returned
	if force {
		path += "?force=true"
		return nil, r.apiDelete(path, nil, nil)
	}

	deployment := new(DeploymentID)
	err := r.apiDelete(path, nil, deployment)

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// HasDeployment checks to see if a deployment exists
// 	id:		the deployment id you are looking for
func (r *marathonClient) HasDeployment(id string) (bool, error) {
	deployments, err := r.Deployments()
	if err != nil {
		return false, err
	}
	for _, deployment := range deployments {
		if deployment.ID == id {
			return true, nil
		}
	}
	return false, nil
}

// WaitOnDeployment waits on a deployment to finish
//  version:		the version of the application
// 	timeout:		the timeout to wait for the deployment to take, otherwise return an error
func (r *marathonClient) WaitOnDeployment(id string, timeout time.Duration) error {
	if found, err := r.HasDeployment(id); err != nil {
		return err
	} else if !found {
		return nil
	}

	nowTime := time.Now()
	stopTime := nowTime.Add(timeout)
	if timeout <= 0 {
		stopTime = nowTime.Add(time.Duration(900) * time.Second)
	}

	// step: a somewhat naive implementation, but it will work
	for {
		if time.Now().After(stopTime) {
			return ErrTimeoutError
		}
		found, err := r.HasDeployment(id)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		time.Sleep(r.config.PollingWaitTime)
	}
}
