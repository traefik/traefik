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

// Group is a marathon application group
type Group struct {
	ID           string         `json:"id"`
	Apps         []*Application `json:"apps"`
	Dependencies []string       `json:"dependencies"`
	Groups       []*Group       `json:"groups"`
}

// Groups is a collection of marathon application groups
type Groups struct {
	ID           string         `json:"id"`
	Apps         []*Application `json:"apps"`
	Dependencies []string       `json:"dependencies"`
	Groups       []*Group       `json:"groups"`
}

// GetGroupOpts contains a payload for Group and Groups method
//		embed:		Embeds nested resources that match the supplied path.
// 					You can specify this parameter multiple times with different values
type GetGroupOpts struct {
	Embed []string `url:"embed,omitempty"`
}

// DeleteGroupOpts contains a payload for DeleteGroup method
//		force:		overrides a currently running deployment.
type DeleteGroupOpts struct {
	Force bool `url:"force,omitempty"`
}

// UpdateGroupOpts contains a payload for UpdateGroup method
//		force:		overrides a currently running deployment.
type UpdateGroupOpts struct {
	Force bool `url:"force,omitempty"`
}

// NewApplicationGroup create a new application group
//		name:			the name of the group
func NewApplicationGroup(name string) *Group {
	return &Group{
		ID:           name,
		Apps:         make([]*Application, 0),
		Dependencies: make([]string, 0),
		Groups:       make([]*Group, 0),
	}
}

// Name sets the name of the group
// 		name:	the name of the group
func (r *Group) Name(name string) *Group {
	r.ID = validateID(name)
	return r
}

// App add a application to the group in question
// 		application:	a pointer to the Application
func (r *Group) App(application *Application) *Group {
	if r.Apps == nil {
		r.Apps = make([]*Application, 0)
	}
	r.Apps = append(r.Apps, application)
	return r
}

// Groups retrieves a list of all the groups from marathon
func (r *marathonClient) Groups() (*Groups, error) {
	groups := new(Groups)
	if err := r.apiGet(marathonAPIGroups, "", groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// Group retrieves the configuration of a specific group from marathon
//		name:			the identifier for the group
func (r *marathonClient) Group(name string) (*Group, error) {
	group := new(Group)
	if err := r.apiGet(fmt.Sprintf("%s/%s", marathonAPIGroups, trimRootPath(name)), nil, group); err != nil {
		return nil, err
	}
	return group, nil
}

// GroupsBy retrieves a list of all the groups from marathon by embed options
//		opts:		GetGroupOpts request payload
func (r *marathonClient) GroupsBy(opts *GetGroupOpts) (*Groups, error) {
	u, err := addOptions(marathonAPIGroups, opts)
	if err != nil {
		return nil, err
	}
	groups := new(Groups)
	if err := r.apiGet(u, "", groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GroupBy retrieves the configuration of a specific group from marathon
//		name:			the identifier for the group
//		opts:			GetGroupOpts request payload
func (r *marathonClient) GroupBy(name string, opts *GetGroupOpts) (*Group, error) {
	u, err := addOptions(fmt.Sprintf("%s/%s", marathonAPIGroups, trimRootPath(name)), opts)
	if err != nil {
		return nil, err
	}
	group := new(Group)
	if err := r.apiGet(u, nil, group); err != nil {
		return nil, err
	}
	return group, nil
}

// HasGroup checks if the group exists in marathon
// 		name:			the identifier for the group
func (r *marathonClient) HasGroup(name string) (bool, error) {
	uri := fmt.Sprintf("%s/%s", marathonAPIGroups, trimRootPath(name))
	err := r.apiCall("GET", uri, "", nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.ErrCode == ErrCodeNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateGroup creates a new group in marathon
//		group:			a pointer the Group structure defining the group
func (r *marathonClient) CreateGroup(group *Group) error {
	return r.apiPost(marathonAPIGroups, group, nil)
}

// WaitOnGroup waits for all the applications in a group to be deployed
// 		group:			the identifier for the group
//		timeout: 		a duration of time to wait before considering it failed (all tasks in all apps running defined as deployed)
func (r *marathonClient) WaitOnGroup(name string, timeout time.Duration) error {
	err := deadline(timeout, func(stop_channel chan bool) error {
		var flick atomicSwitch
		go func() {
			<-stop_channel
			close(stop_channel)
			flick.SwitchOn()
		}()
		for !flick.IsSwitched() {
			if group, err := r.Group(name); err != nil {
				continue
			} else {
				allRunning := true
				// for each of the application, check if the tasks and running
				for _, appID := range group.Apps {
					// Arrrgghhh!! .. so we can't use application instances from the Application struct like with app wait on as it
					// appears the instance count is not set straight away!! .. it defaults to zero and changes probably at the
					// dependencies gets deployed. Which is probably how it internally handles dependencies ..
					// step: grab the application
					application, err := r.Application(appID.ID)
					if err != nil {
						allRunning = false
						break
					}

					if application.Tasks == nil {
						allRunning = false
					} else if len(application.Tasks) != *appID.Instances {
						allRunning = false
					} else if application.TasksRunning != *appID.Instances {
						allRunning = false
					} else if len(application.DeploymentIDs()) > 0 {
						allRunning = false
					}
				}
				// has anyone toggle the flag?
				if allRunning {
					return nil
				}
			}
			time.Sleep(r.config.PollingWaitTime)
		}
		return nil
	})

	return err
}

// DeleteGroup deletes a group from marathon
//		name:			the identifier for the group
//		force:			used to force the delete operation in case of blocked deployment
func (r *marathonClient) DeleteGroup(name string, force bool) (*DeploymentID, error) {
	version := new(DeploymentID)
	uri := fmt.Sprintf("%s/%s", marathonAPIGroups, trimRootPath(name))
	if force {
		uri = uri + "?force=true"
	}
	if err := r.apiDelete(uri, nil, version); err != nil {
		return nil, err
	}

	return version, nil
}

// UpdateGroup updates the parameters of a groups
//		name:			the identifier for the group
//		group:  		the group structure with the new params
//		force:			used to force the update operation in case of blocked deployment
func (r *marathonClient) UpdateGroup(name string, group *Group, force bool) (*DeploymentID, error) {
	deploymentID := new(DeploymentID)
	uri := fmt.Sprintf("%s/%s", marathonAPIGroups, trimRootPath(name))
	if force {
		uri = uri + "?force=true"
	}
	if err := r.apiPut(uri, group, deploymentID); err != nil {
		return nil, err
	}

	return deploymentID, nil
}
