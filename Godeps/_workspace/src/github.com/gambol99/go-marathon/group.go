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

type Group struct {
	ID           string         `json:"id"`
	Apps         []*Application `json:"apps"`
	Dependencies []string       `json:"dependencies"`
	Groups       []*Group       `json:"groups"`
}

type Groups struct {
	ID           string         `json:"id"`
	Apps         []*Application `json:"apps"`
	Dependencies []string       `json:"dependencies"`
	Groups       []*Group       `json:"groups"`
}

// Create a new Application Group
//		name:	the name of the group
func NewApplicationGroup(name string) *Group {
	return &Group{
		ID:           name,
		Apps:         make([]*Application, 0),
		Dependencies: make([]string, 0),
		Groups:       make([]*Group, 0),
	}
}

// Specify the name of the group
// 		name:	the name of the group
func (group *Group) Name(name string) *Group {
	group.ID = validateID(name)
	return group
}

// Add a application to the group in question
// 		application:	a pointer to the Application
func (group *Group) App(application *Application) *Group {
	if group.Apps == nil {
		group.Apps = make([]*Application, 0)
	}
	group.Apps = append(group.Apps, application)
	return group
}

// Retrieve a list of all the groups from marathon
func (client *Client) Groups() (*Groups, error) {
	groups := new(Groups)
	if err := client.apiGet(MARATHON_API_GROUPS, "", groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// Retrieve the configuration of a specific group from marathon
//		name:	the identifier for the group
func (client *Client) Group(name string) (*Group, error) {
	group := new(Group)
	if err := client.apiGet(fmt.Sprintf("%s/%s", MARATHON_API_GROUPS, trimRootPath(name)), nil, group); err != nil {
		return nil, err
	}
	return group, nil
}

// Check if the group exists in marathon
// 		name:	the identifier for the group
func (client *Client) HasGroup(name string) (bool, error) {
	uri := fmt.Sprintf("%s/%s", MARATHON_API_GROUPS, trimRootPath(name))
	status, _, err := client.apiCall(HTTP_GET, uri, "", nil)
	if err == nil {
		return true, nil
	} else if status == 404 {
		return false, nil
	} else {
		return false, err
	}
}

// Create a new group in marathon
//		group:	a pointer the Group structure defining the group
func (client *Client) CreateGroup(group *Group, wait_on_running bool) error {
	if err := client.apiPost(MARATHON_API_GROUPS, group, nil); err != nil {
		return err
	}
	if wait_on_running {
		return client.WaitOnGroup(group.ID, 0)
	}
	return nil
}

// Waits for all the applications in a group to be deployed
// 		group:		the identifier for the group
//		timeout: 	a duration of time to wait before considering it failed (all tasks in all apps running defined as deployed)
func (client *Client) WaitOnGroup(name string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = time.Duration(500) * time.Second
	}
	err := deadline(timeout, func(stop_channel chan bool) error {
		var flick AtomicSwitch
		go func() {
			<-stop_channel
			close(stop_channel)
			flick.SwitchOn()
		}()
		for !flick.IsSwitched() {
			if group, err := client.Group(name); err != nil {
				continue
			} else {
				all_running := true
				// for each of the application, check if the tasks and running
				for _, appID := range group.Apps {
					// Arrrgghhh!! .. so we can't use application instances from the Application struct like with app wait on as it
					// appears the instance count is not set straight away!! .. it defaults to zero and changes probably at the
					// dependencies gets deployed. Which is probably how it internally handles dependencies ..
					// step: grab the application
					application, err := client.Application(appID.ID)
					if err != nil {
						all_running = false
						break
					}

					if application.Tasks == nil {
						all_running = false
					} else if len(application.Tasks) != appID.Instances {
						all_running = false
					} else if application.TasksRunning != appID.Instances {
						all_running = false
					} else if len(application.Deployments()) > 0 {
						all_running = false
					}
				}
				// has anyone toggle the flag?
				if all_running {
					return nil
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
		return nil
	})
	return err
}

// Delete a group from marathon
// 		name:	the identifier for the group
func (client *Client) DeleteGroup(name string) (*DeploymentID, error) {
	version := new(DeploymentID)
	uri := fmt.Sprintf("%s/%s", MARATHON_API_GROUPS, trimRootPath(name))
	if err := client.apiDelete(uri, nil, version); err != nil {
		return nil, err
	}
	return version, nil
}

// Update the parameters of a groups
// 		name:	the identifier for the group
//      group:  the group structure with the new params
func (client *Client) UpdateGroup(name string, group *Group) (*DeploymentID, error) {
	deploymentId := new(DeploymentID)
	uri := fmt.Sprintf("%s/%s", MARATHON_API_GROUPS, trimRootPath(name))
	if err := client.apiPut(uri, group, deploymentId); err != nil {
		return nil, err
	}
	return deploymentId, nil
}
