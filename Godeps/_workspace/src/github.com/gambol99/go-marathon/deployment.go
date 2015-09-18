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
	"time"
)

type Deployment struct {
	ID             string              `json:"id"`
	Version        string              `json:"version"`
	CurrentStep    int                 `json:"currentStep"`
	TotalSteps     int                 `json:"totalSteps"`
	AffectedApps   []string            `json:"affectedApps"`
	Steps          [][]*DeploymentStep `json:"steps"`
	CurrentActions []*DeploymentStep   `json:"currentActions"`
}

type DeploymentID struct {
	DeploymentID string `json:"deploymentId"`
	Version      string `json:"version"`
}

type DeploymentStep struct {
	Action string `json:"action"`
	App    string `json:"app"`
}

type DeploymentPlan struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Original struct {
		Apps         []*Application `json:"apps"`
		Dependencies []string       `json:"dependencies"`
		Groups       []*Group       `json:"groups"`
		ID           string         `json:"id"`
		Version      string         `json:"version"`
	} `json:"original"`
	Steps  []*DeploymentStep `json:"steps"`
	Target struct {
		Apps         []*Application `json:"apps"`
		Dependencies []string       `json:"dependencies"`
		Groups       []*Group       `json:"groups"`
		ID           string         `json:"id"`
		Version      string         `json:"version"`
	} `json:"target"`
}

// Retrieve a list of current deployments
func (client *Client) Deployments() ([]*Deployment, error) {
	var deployments []*Deployment
	if err := client.apiGet(MARATHON_API_DEPLOYMENTS, nil, &deployments); err != nil {
		return nil, err
	} else {
		return deployments, nil
	}
}

// Delete a current deployment from marathon
// 	id:		the deployment id you wish to delete
// 	force:	whether or not to force the deletion
func (client *Client) DeleteDeployment(id string, force bool) (*DeploymentID, error) {
	deployment := new(DeploymentID)
	if err := client.apiDelete(fmt.Sprintf("%s/%s", MARATHON_API_DEPLOYMENTS, id), nil, deployment); err != nil {
		return nil, err
	} else {
		return deployment, nil
	}
}

// Check to see if a deployment exists
// 	id:		the deployment id you are looking for
func (client *Client) HasDeployment(id string) (bool, error) {
	deployments, err := client.Deployments()
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

// Wait of a deployment to finish
//  version:    the version of the application
// 	timeout:	the timeout to wait for the deployment to take, otherwise return an error
func (client *Client) WaitOnDeployment(id string, timeout time.Duration) error {
	if found, err := client.HasDeployment(id); err != nil {
		return err
	} else if !found {
		return nil
	}

	client.log("WaitOnDeployment() Waiting for deployment: %s to finish", id)
	now_time := time.Now()
	stop_time := now_time.Add(timeout)
	if timeout <= 0 {
		stop_time = now_time.Add(time.Duration(900) * time.Second)
	}

	// step: a somewhat naive implementation, but it will work
	for {
		if time.Now().After(stop_time) {
			return ErrTimeoutError
		}
		found, err := client.HasDeployment(id)
		if err != nil {
			client.log("WaitOnDeployment() Failed to get the deployments list, error: %s", err)
		}
		if !found {
			return nil
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
}
