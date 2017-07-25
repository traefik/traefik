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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationDependsOn(t *testing.T) {
	app := NewDockerApplication()
	app.DependsOn("fake-app")
	app.DependsOn("fake-app1", "fake-app2")
	assert.Equal(t, 3, len(app.Dependencies))
}

func TestApplicationMemory(t *testing.T) {
	app := NewDockerApplication()
	app.Memory(50.0)
	assert.Equal(t, 50.0, *app.Mem)
}

func TestApplicationString(t *testing.T) {
	app := NewDockerApplication().
		Name("my-app").
		CPU(0.1).
		Memory(64).
		Storage(0.0).
		Count(2).
		AddArgs("/usr/sbin/apache2ctl", "-D", "FOREGROUND").
		AddEnv("NAME", "frontend_http").
		AddEnv("SERVICE_80_NAME", "test_http")
	app.
		Container.Docker.Container("quay.io/gambol99/apache-php:latest").
		Bridged().
		Expose(80).
		Expose(443)
	app, err := app.CheckHTTP("/health", 80, 5)
	assert.Nil(t, err)

	expectedAppJSONBytes, err := ioutil.ReadFile("tests/app-definitions/TestApplicationString-output.json")
	if err != nil {
		panic(err)
	}
	expectedAppJSON := strings.TrimSpace(string(expectedAppJSONBytes))
	assert.Equal(t, expectedAppJSON, app.String())
}

func TestApplicationCount(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Instances)
	app.Count(1)
	assert.Equal(t, 1, *app.Instances)
}

func TestApplicationStorage(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Disk)
	app.Storage(0.10)
	assert.Equal(t, 0.10, *app.Disk)
}

func TestApplicationAllTaskRunning(t *testing.T) {
	app := NewDockerApplication()

	app.Instances = nil
	app.Tasks = nil

	assert.True(t, app.AllTaskRunning())

	var cnt int
	app.Instances = &cnt

	cnt = 0
	assert.True(t, app.AllTaskRunning())

	cnt = 1
	assert.False(t, app.AllTaskRunning())

	app.Tasks = []*Task{}
	app.TasksRunning = 1
	assert.True(t, app.AllTaskRunning())

	cnt = 2
	app.TasksRunning = 1
	assert.False(t, app.AllTaskRunning())
}

func TestApplicationName(t *testing.T) {
	app := NewDockerApplication()
	assert.Equal(t, "", app.ID)
	app.Name(fakeAppName)
	assert.Equal(t, fakeAppName, app.ID)
}

func TestApplicationCommand(t *testing.T) {
	app := NewDockerApplication()
	assert.Equal(t, "", app.ID)
	app.Command("format C:")
	assert.Equal(t, "format C:", *app.Cmd)
}

func TestApplicationCPU(t *testing.T) {
	app := NewDockerApplication()
	assert.Equal(t, 0.0, app.CPUs)
	app.CPU(0.1)
	assert.Equal(t, 0.1, app.CPUs)
}

func TestApplicationSetGPUs(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.GPUs)
	app.SetGPUs(0.1)
	assert.Equal(t, 0.1, *app.GPUs)
}

func TestApplicationEmptyGPUs(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.GPUs)
	app.EmptyGPUs()
	assert.Equal(t, 0.0, *app.GPUs)
}

func TestApplicationArgs(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Args)
	app.AddArgs("-p").AddArgs("option", "-v")
	assert.Equal(t, 3, len(*app.Args))
	assert.Equal(t, "-p", (*app.Args)[0])
	assert.Equal(t, "option", (*app.Args)[1])
	assert.Equal(t, "-v", (*app.Args)[2])

	app.EmptyArgs()
	assert.NotNil(t, app.Args)
	assert.Equal(t, 0, len(*app.Args))
}

func ExampleApplication_AddConstraint() {
	app := NewDockerApplication()

	// add two constraints
	app.AddConstraint("hostname", "UNIQUE").
		AddConstraint("rack_id", "CLUSTER", "rack-1")
}

func TestApplicationConstraints(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Constraints)
	app.AddConstraint("hostname", "UNIQUE").
		AddConstraint("rack_id", "CLUSTER", "rack-1")

	assert.Equal(t, 2, len(*app.Constraints))
	assert.Equal(t, []string{"hostname", "UNIQUE"}, (*app.Constraints)[0])
	assert.Equal(t, []string{"rack_id", "CLUSTER", "rack-1"}, (*app.Constraints)[1])

	app.EmptyConstraints()
	assert.NotNil(t, app.Constraints)
	assert.Equal(t, 0, len(*app.Constraints))
}

func TestApplicationLabels(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Labels)

	app.AddLabel("hello", "world").AddLabel("foo", "bar")
	assert.Equal(t, 2, len(*app.Labels))
	assert.Equal(t, "world", (*app.Labels)["hello"])
	assert.Equal(t, "bar", (*app.Labels)["foo"])

	app.EmptyLabels()
	assert.NotNil(t, app.Labels)
	assert.Equal(t, 0, len(*app.Labels))
}

func TestApplicationEnvs(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Env)

	app.AddEnv("hello", "world").AddEnv("foo", "bar")
	assert.Equal(t, 2, len(*app.Env))
	assert.Equal(t, "world", (*app.Env)["hello"])
	assert.Equal(t, "bar", (*app.Env)["foo"])

	app.EmptyEnvs()
	assert.NotNil(t, app.Env)
	assert.Equal(t, 0, len(*app.Env))
}

func TestApplicationSetExecutor(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Executor)

	app.SetExecutor("executor")
	assert.Equal(t, "executor", *app.Executor)

	app.SetExecutor("")
	assert.Equal(t, "", *app.Executor)
}

func TestApplicationHealthChecks(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.HealthChecks)
	app.AddHealthCheck(HealthCheck{}.SetPath("/check1")).
		AddHealthCheck(HealthCheck{}.SetPath("/check2"))

	assert.Equal(t, 2, len(*app.HealthChecks))
	assert.Equal(t, HealthCheck{}.SetPath("/check1"), (*app.HealthChecks)[0])
	assert.Equal(t, HealthCheck{}.SetPath("/check2"), (*app.HealthChecks)[1])

	app.EmptyHealthChecks()
	assert.NotNil(t, app.HealthChecks)
	assert.Equal(t, 0, len(*app.HealthChecks))
}

func TestApplicationReadinessChecks(t *testing.T) {
	app := NewDockerApplication()
	require.Nil(t, app.HealthChecks)
	rc := ReadinessCheck{}
	rc.SetName("/readiness")
	app.AddReadinessCheck(rc)

	require.Equal(t, 1, len(*app.ReadinessChecks))
	assert.Equal(t, "/readiness", *((*app.ReadinessChecks)[0].Name))

	app.EmptyReadinessChecks()
	require.NotNil(t, app.ReadinessChecks)
	assert.Equal(t, 0, len(*app.ReadinessChecks))
}

func TestApplicationPortDefinitions(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.PortDefinitions)
	app.AddPortDefinition(PortDefinition{Protocol: "tcp", Name: "es"}.SetPort(9201).AddLabel("foo", "bar")).
		AddPortDefinition(PortDefinition{Protocol: "udp,tcp", Name: "syslog"}.SetPort(514))

	assert.Equal(t, 2, len(*app.PortDefinitions))
	assert.Equal(t, PortDefinition{Protocol: "tcp", Name: "es"}.SetPort(9201).AddLabel("foo", "bar"), (*app.PortDefinitions)[0])
	assert.Equal(t, 1, len(*(*app.PortDefinitions)[0].Labels))
	assert.Equal(t, PortDefinition{Protocol: "udp,tcp", Name: "syslog"}.SetPort(514), (*app.PortDefinitions)[1])
	assert.Nil(t, (*app.PortDefinitions)[1].Labels)

	(*app.PortDefinitions)[0].EmptyLabels()
	assert.NotNil(t, (*app.PortDefinitions)[0].Labels)
	assert.Equal(t, 0, len(*(*app.PortDefinitions)[0].Labels))

	app.EmptyPortDefinitions()
	assert.NotNil(t, app.PortDefinitions)
	assert.Equal(t, 0, len(*app.PortDefinitions))
}

func TestHasHealthChecks(t *testing.T) {
	app := NewDockerApplication()
	assert.False(t, app.HasHealthChecks())
	app.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80)
	_, err := app.CheckTCP(80, 10)
	assert.NoError(t, err)
	assert.True(t, app.HasHealthChecks())
}

func TestApplicationCheckTCP(t *testing.T) {
	app := NewDockerApplication()
	assert.False(t, app.HasHealthChecks())
	_, err := app.CheckTCP(80, 10)
	assert.Error(t, err)
	assert.False(t, app.HasHealthChecks())
	app.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80)
	_, err = app.CheckTCP(80, 10)
	assert.NoError(t, err)
	assert.True(t, app.HasHealthChecks())
	check := (*app.HealthChecks)[0]
	assert.Equal(t, "TCP", check.Protocol)
	assert.Equal(t, 10, check.IntervalSeconds)
	assert.Equal(t, 0, *check.PortIndex)
}

func TestApplicationCheckHTTP(t *testing.T) {
	app := NewDockerApplication()
	assert.False(t, app.HasHealthChecks())
	_, err := app.CheckHTTP("/", 80, 10)
	assert.Error(t, err)
	assert.False(t, app.HasHealthChecks())
	app.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80)
	_, err = app.CheckHTTP("/health", 80, 10)
	assert.NoError(t, err)
	assert.True(t, app.HasHealthChecks())
	check := (*app.HealthChecks)[0]
	assert.Equal(t, "HTTP", check.Protocol)
	assert.Equal(t, 10, check.IntervalSeconds)
	assert.Equal(t, "/health", *check.Path)
	assert.Equal(t, 0, *check.PortIndex)
}

func TestCreateApplication(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	application := NewDockerApplication()
	application.Name(fakeAppName)
	app, err := endpoint.Client.CreateApplication(application)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, application.ID, fakeAppName)
	assert.Equal(t, app.Deployments[0]["id"], "f44fd4fc-4330-4600-a68b-99c7bd33014a")
}

func TestUpdateApplication(t *testing.T) {
	for _, force := range []bool{false, true} {
		endpoint := newFakeMarathonEndpoint(t, nil)
		defer endpoint.Close()

		application := NewDockerApplication()
		application.Name(fakeAppName)
		id, err := endpoint.Client.UpdateApplication(application, force)
		assert.NoError(t, err)
		assert.Equal(t, id.DeploymentID, "83b215a6-4e26-4e44-9333-5c385eda6438")
		assert.Equal(t, id.Version, "2014-08-26T07:37:50.462Z")
	}
}

func TestApplications(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	applications, err := endpoint.Client.Applications(nil)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, len(applications.Apps), 2)

	v := url.Values{}
	v.Set("cmd", "nginx")
	applications, err = endpoint.Client.Applications(v)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, len(applications.Apps), 1)
}

func TestApplicationsEmbedTaskStats(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	v := url.Values{}
	v.Set("embed", "apps.taskStats")
	applications, err := endpoint.Client.Applications(v)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, len(applications.Apps), 1)
	assert.NotNil(t, applications.Apps[0].TaskStats)
	assert.Equal(t, applications.Apps[0].TaskStats["startedAfterLastScaling"].Stats.Counts["healthy"], 1)
	assert.Equal(t, applications.Apps[0].TaskStats["startedAfterLastScaling"].Stats.LifeTime["averageSeconds"], 17024.575)
}

func TestApplicationsEmbedReadiness(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	v := url.Values{}
	v.Set("embed", "apps.readiness")
	applications, err := endpoint.Client.Applications(v)
	require.NoError(t, err)
	require.NotNil(t, applications)
	require.Equal(t, len(applications.Apps), 1)
	require.NotNil(t, applications.Apps[0].ReadinessCheckResults)
	require.True(t, len(*applications.Apps[0].ReadinessCheckResults) > 0)
	actualRes := (*applications.Apps[0].ReadinessCheckResults)[0]
	expectedRes := ReadinessCheckResult{
		Name:   "myReadyCheck",
		TaskID: "test_frontend_app1.c9de6033",
		Ready:  false,
		LastResponse: ReadinessLastResponse{
			Body:        "{}",
			ContentType: "application/json",
			Status:      500,
		},
	}
	assert.Equal(t, expectedRes, actualRes)
}

func TestListApplications(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	applications, err := endpoint.Client.ListApplications(nil)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, len(applications), 2)
	assert.Equal(t, applications[0], fakeAppName)
	assert.Equal(t, applications[1], fakeAppNameBroken)
}

func TestApplicationVersions(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	versions, err := endpoint.Client.ApplicationVersions(fakeAppName)
	assert.NoError(t, err)
	assert.NotNil(t, versions)
	assert.NotNil(t, versions.Versions)
	assert.Equal(t, len(versions.Versions), 1)
	assert.Equal(t, versions.Versions[0], "2014-04-04T06:25:31.399Z")
	/* check we get an error on app not there */
	versions, err = endpoint.Client.ApplicationVersions("/not/there")
	assert.Error(t, err)
}

func TestRestartApplication(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	id, err := endpoint.Client.RestartApplication(fakeAppName, false)
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, "83b215a6-4e26-4e44-9333-5c385eda6438", id.DeploymentID)
	assert.Equal(t, "2014-08-26T07:37:50.462Z", id.Version)
	id, err = endpoint.Client.RestartApplication("/not/there", false)
	assert.Error(t, err)
	assert.Nil(t, id)
}

func TestApplicationUris(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Uris)
	app.AddUris("file://uri1.tar.gz").AddUris("file://uri2.tar.gz", "file://uri3.tar.gz")
	assert.Equal(t, 3, len(*app.Uris))
	assert.Equal(t, "file://uri1.tar.gz", (*app.Uris)[0])
	assert.Equal(t, "file://uri2.tar.gz", (*app.Uris)[1])
	assert.Equal(t, "file://uri3.tar.gz", (*app.Uris)[2])

	app.EmptyUris()
	assert.NotNil(t, app.Uris)
	assert.Equal(t, 0, len(*app.Uris))
}

func TestApplicationFetchURIs(t *testing.T) {
	app := NewDockerApplication()
	assert.Nil(t, app.Fetch)
	app.AddFetchURIs(Fetch{URI: "file://uri1.tar.gz"}).
		AddFetchURIs(Fetch{URI: "file://uri2.tar.gz"}, Fetch{URI: "file://uri3.tar.gz"})
	assert.Equal(t, 3, len(*app.Fetch))
	assert.Equal(t, Fetch{URI: "file://uri1.tar.gz"}, (*app.Fetch)[0])
	assert.Equal(t, Fetch{URI: "file://uri2.tar.gz"}, (*app.Fetch)[1])
	assert.Equal(t, Fetch{URI: "file://uri3.tar.gz"}, (*app.Fetch)[2])

	app.EmptyUris()
	assert.NotNil(t, app.Uris)
	assert.Equal(t, 0, len(*app.Uris))
}

func TestSetApplicationVersion(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	deployment, err := endpoint.Client.SetApplicationVersion(fakeAppName, &ApplicationVersion{Version: "2014-08-26T07:37:50.462Z"})
	assert.NoError(t, err)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Version)
	assert.NotNil(t, deployment.DeploymentID)
	assert.Equal(t, deployment.Version, "2014-08-26T07:37:50.462Z")
	assert.Equal(t, deployment.DeploymentID, "83b215a6-4e26-4e44-9333-5c385eda6438")
	_, err = endpoint.Client.SetApplicationVersion("/not/there", &ApplicationVersion{Version: "2014-04-04T06:25:31.399Z"})
	assert.Error(t, err)
}

func TestHasApplicationVersion(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	found, err := endpoint.Client.HasApplicationVersion(fakeAppName, "2014-04-04T06:25:31.399Z")
	assert.NoError(t, err)
	assert.True(t, found)
	found, err = endpoint.Client.HasApplicationVersion(fakeAppName, "###2015-04-04T06:25:31.399Z")
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestDeleteApplication(t *testing.T) {
	for _, force := range []bool{false, true} {
		endpoint := newFakeMarathonEndpoint(t, nil)
		defer endpoint.Close()
		id, err := endpoint.Client.DeleteApplication(fakeAppName, force)
		assert.NoError(t, err)
		assert.NotNil(t, id)
		assert.Equal(t, "83b215a6-4e26-4e44-9333-5c385eda6438", id.DeploymentID)
		assert.Equal(t, "2014-08-26T07:37:50.462Z", id.Version)
		id, err = endpoint.Client.DeleteApplication("no_such_app", force)
		assert.Error(t, err)
	}
}

func TestApplicationOK(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	ok, err := endpoint.Client.ApplicationOK(fakeAppName)
	assert.NoError(t, err)
	assert.True(t, ok)
	ok, err = endpoint.Client.ApplicationOK(fakeAppNameBroken)
	assert.NoError(t, err)
	assert.False(t, ok)
	ok, err = endpoint.Client.ApplicationOK(fakeAppNameUnhealthy)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func verifyApplication(application *Application, t *testing.T) {
	assert.NotNil(t, application)
	assert.Equal(t, application.ID, fakeAppName)
	assert.NotNil(t, application.HealthChecks)
	assert.NotNil(t, application.Tasks)
	assert.Equal(t, len(*application.HealthChecks), 1)
	assert.Equal(t, len(application.Tasks), 2)
}

func TestApplication(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	application, err := endpoint.Client.Application(fakeAppName)
	assert.NoError(t, err)
	verifyApplication(application, t)

	_, err = endpoint.Client.Application("no_such_app")
	assert.Error(t, err)
	apiErr, ok := err.(*APIError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, apiErr.ErrCode)

	config := NewDefaultConfig()
	config.URL = "http://non-existing-marathon-host.local:5555"
	// Reduce timeout to speed up test execution time.
	config.HTTPClient = &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	endpoint = newFakeMarathonEndpoint(t, &configContainer{
		client: &config,
	})
	defer endpoint.Close()

	_, err = endpoint.Client.Application(fakeAppName)
	assert.Error(t, err)
	_, ok = err.(*APIError)
	assert.False(t, ok)
}

func TestApplicationConfiguration(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	application, err := endpoint.Client.ApplicationByVersion(fakeAppName, "2014-09-12T23:28:21.737Z")
	assert.NoError(t, err)
	verifyApplication(application, t)

	_, err = endpoint.Client.ApplicationByVersion(fakeAppName, "no_such_version")
	assert.Error(t, err)
	apiErr, ok := err.(*APIError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, apiErr.ErrCode)

	_, err = endpoint.Client.ApplicationByVersion("no_such_app", "latest")
	assert.Error(t, err)
	apiErr, ok = err.(*APIError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, apiErr.ErrCode)
}

func TestWaitOnApplication(t *testing.T) {
	waitTime := 100 * time.Millisecond

	tests := []struct {
		desc          string
		timeout       time.Duration
		appName       string
		testScope     string
		shouldSucceed bool
	}{
		{
			desc:          "initially existing app",
			timeout:       0,
			appName:       fakeAppName,
			shouldSucceed: true,
		},

		{
			desc:          "delayed existing app | timeout > ticker",
			timeout:       200 * time.Millisecond,
			appName:       fakeAppName,
			testScope:     "wait-on-app",
			shouldSucceed: true,
		},

		{
			desc:          "delayed existing app | timeout < ticker",
			timeout:       50 * time.Millisecond,
			appName:       fakeAppName,
			testScope:     "wait-on-app",
			shouldSucceed: false,
		},
		{
			desc:          "missing app | timeout > ticker",
			timeout:       200 * time.Millisecond,
			appName:       "no_such_app",
			shouldSucceed: false,
		},
		{
			desc:          "missing app | timeout < ticker",
			timeout:       50 * time.Millisecond,
			appName:       "no_such_app",
			shouldSucceed: false,
		},
	}

	for _, test := range tests {
		defaultConfig := NewDefaultConfig()
		defaultConfig.PollingWaitTime = waitTime
		configs := &configContainer{
			client: &defaultConfig,
			server: &serverConfig{
				scope: test.testScope,
			},
		}

		endpoint := newFakeMarathonEndpoint(t, configs)
		defer endpoint.Close()

		errCh := make(chan error)
		go func() {
			errCh <- endpoint.Client.WaitOnApplication(test.appName, test.timeout)
		}()

		select {
		case <-time.After(400 * time.Millisecond):
			assert.Fail(t, fmt.Sprintf("%s: WaitOnApplication did not complete in time", test.desc))
		case err := <-errCh:
			if test.shouldSucceed {
				assert.NoError(t, err, test.desc)
			} else {
				assert.IsType(t, err, ErrTimeoutError, test.desc)
			}
		}
	}
}

func TestAppExistAndRunning(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()
	client := endpoint.Client.(*marathonClient)
	assert.True(t, client.appExistAndRunning(fakeAppName))
	assert.False(t, client.appExistAndRunning("no_such_app"))
}

func TestSetIPPerTask(t *testing.T) {
	app := Application{}
	app.Ports = append(app.Ports, 10)
	app.AddPortDefinition(PortDefinition{})
	assert.Nil(t, app.IPAddressPerTask)
	assert.Equal(t, 1, len(app.Ports))
	assert.Equal(t, 1, len(*app.PortDefinitions))

	app.SetIPAddressPerTask(IPAddressPerTask{})
	assert.NotNil(t, app.IPAddressPerTask)
	assert.Equal(t, 0, len(app.Ports))
	assert.Equal(t, 0, len(*app.PortDefinitions))
}

func TestIPAddressPerTask(t *testing.T) {
	ipPerTask := IPAddressPerTask{}
	assert.Nil(t, ipPerTask.Groups)
	assert.Nil(t, ipPerTask.Labels)
	assert.Nil(t, ipPerTask.Discovery)

	ipPerTask.
		AddGroup("label").
		AddLabel("key", "value").
		SetDiscovery(Discovery{})

	assert.Equal(t, 1, len(*ipPerTask.Groups))
	assert.Equal(t, "label", (*ipPerTask.Groups)[0])
	assert.Equal(t, "value", (*ipPerTask.Labels)["key"])
	assert.NotEmpty(t, ipPerTask.Discovery)

	ipPerTask.EmptyGroups()
	assert.Equal(t, 0, len(*ipPerTask.Groups))

	ipPerTask.EmptyLabels()
	assert.Equal(t, 0, len(*ipPerTask.Labels))

}

func TestIPAddressPerTaskDiscovery(t *testing.T) {
	disc := Discovery{}
	assert.Nil(t, disc.Ports)

	disc.AddPort(Port{})
	assert.NotNil(t, disc.Ports)
	assert.Equal(t, 1, len(*disc.Ports))

	disc.EmptyPorts()
	assert.NotNil(t, disc.Ports)
	assert.Equal(t, 0, len(*disc.Ports))

}

func TestUpgradeStrategy(t *testing.T) {
	app := Application{}
	assert.Nil(t, app.UpgradeStrategy)
	app.SetUpgradeStrategy(UpgradeStrategy{}.SetMinimumHealthCapacity(1.0).SetMaximumOverCapacity(0.0))
	us := app.UpgradeStrategy
	assert.Equal(t, 1.0, *us.MinimumHealthCapacity)
	assert.Equal(t, 0.0, *us.MaximumOverCapacity)

	app.EmptyUpgradeStrategy()
	us = app.UpgradeStrategy
	assert.NotNil(t, us)
	assert.Nil(t, us.MinimumHealthCapacity)
	assert.Nil(t, us.MaximumOverCapacity)
}

