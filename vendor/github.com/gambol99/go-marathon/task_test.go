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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasHealthCheckResults(t *testing.T) {
	task := Task{}
	assert.False(t, task.HasHealthCheckResults())
	task.HealthCheckResults = append(task.HealthCheckResults, &HealthCheckResult{})
	assert.True(t, task.HasHealthCheckResults())
}

func TestAllTasks(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	tasks, err := endpoint.Client.AllTasks(nil)
	assert.NoError(t, err)
	if assert.NotNil(t, tasks) {
		assert.Equal(t, len(tasks.Tasks), 2)
	}

	tasks, err = endpoint.Client.AllTasks(&AllTasksOpts{Status: "staging"})
	assert.Nil(t, err)
	if assert.NotNil(t, tasks) {
		assert.Equal(t, len(tasks.Tasks), 0)
	}
}

func TestTasks(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	tasks, err := endpoint.Client.Tasks(fakeAppName)
	assert.NoError(t, err)
	if assert.NotNil(t, tasks) {
		assert.Equal(t, len(tasks.Tasks), 2)
	}
}

func TestKillApplicationTasks(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	tasks, err := endpoint.Client.KillApplicationTasks(fakeAppName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
}

func TestKillTask(t *testing.T) {
	cases := map[string]struct {
		TaskID string
		Result string
	}{
		"CommonApp":           {fakeTaskID, fakeTaskID},
		"GroupApp":            {"fake-group_fake-app.fake-task", "fake-group_fake-app.fake-task"},
		"GroupAppWithSlashes": {"fake-group/fake-app.fake-task", "fake-group_fake-app.fake-task"},
	}
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	for k, tc := range cases {
		task, err := endpoint.Client.KillTask(tc.TaskID, nil)
		assert.NoError(t, err, "TestCase: %s", k)
		assert.Equal(t, tc.Result, task.ID, "TestCase: %s", k)
	}
}

func TestKillTasks(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	err := endpoint.Client.KillTasks([]string{fakeTaskID}, nil)
	assert.NoError(t, err)
}

func TestTaskEndpoints(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	endpoints, err := endpoint.Client.TaskEndpoints(fakeAppNameBroken, 8080, true)
	assert.NoError(t, err)
	assert.NotNil(t, endpoints)
	assert.Equal(t, len(endpoints), 1, t)
	assert.Equal(t, endpoints[0], "10.141.141.10:31045", t)

	endpoints, err = endpoint.Client.TaskEndpoints(fakeAppNameBroken, 8080, false)
	assert.NoError(t, err)
	assert.NotNil(t, endpoints)
	assert.Equal(t, len(endpoints), 2, t)
	assert.Equal(t, endpoints[0], "10.141.141.10:31045", t)
	assert.Equal(t, endpoints[1], "10.141.141.10:31234", t)

	_, err = endpoint.Client.TaskEndpoints(fakeAppNameBroken, 80, true)
	assert.Error(t, err)
}
