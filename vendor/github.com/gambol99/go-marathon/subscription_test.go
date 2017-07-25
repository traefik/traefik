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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	eventPublishTimeout time.Duration = 250 * time.Millisecond
	SSEConnectWaitTime  time.Duration = 250 * time.Millisecond
)

type testCaseList []testCase

func (l testCaseList) find(name string) *testCase {
	for _, testCase := range l {
		if testCase.name == name {
			return &testCase
		}
	}
	return nil
}

type testCase struct {
	name        string
	source      string
	expectation interface{}
}

var testCases = testCaseList{
	testCase{
		name: "status_update_event",
		source: `{
	"eventType": "status_update_event",
	"timestamp": "2014-03-01T23:29:30.158Z",
	"slaveId": "20140909-054127-177048842-5050-1494-0",
	"taskId": "my-app_0-1396592784349",
	"taskStatus": "TASK_RUNNING",
	"appId": "/my-app",
	"host": "slave-1234.acme.org",
	"ports": [31372],
	"version": "2014-04-04T06:26:23.051Z"
}`,
		expectation: &EventStatusUpdate{
			EventType:  "status_update_event",
			Timestamp:  "2014-03-01T23:29:30.158Z",
			SlaveID:    "20140909-054127-177048842-5050-1494-0",
			TaskID:     "my-app_0-1396592784349",
			TaskStatus: "TASK_RUNNING",
			AppID:      "/my-app",
			Host:       "slave-1234.acme.org",
			Ports:      []int{31372},
			Version:    "2014-04-04T06:26:23.051Z",
		},
	},
	testCase{
		name: "health_status_changed_event",
		source: `{
	"eventType": "health_status_changed_event",
	"timestamp": "2014-03-01T23:29:30.158Z",
	"appId": "/my-app",
	"taskId": "my-app_0-1396592784349",
	"version": "2014-04-04T06:26:23.051Z",
	"alive": true
}`,
		expectation: &EventHealthCheckChanged{
			EventType: "health_status_changed_event",
			Timestamp: "2014-03-01T23:29:30.158Z",
			AppID:     "/my-app",
			TaskID:    "my-app_0-1396592784349",
			Version:   "2014-04-04T06:26:23.051Z",
			Alive:     true,
		},
	},
	testCase{
		name: "failed_health_check_event",
		source: `{
	"eventType": "failed_health_check_event",
	"timestamp": "2014-03-01T23:29:30.158Z",
	"appId": "/my-app",
	"taskId": "my-app_0-1396592784349",
	"healthCheck": {
		"protocol": "HTTP",
		"path": "/health",
		"portIndex": 0,
		"gracePeriodSeconds": 5,
		"intervalSeconds": 10,
		"timeoutSeconds": 10,
		"maxConsecutiveFailures": 3
	}
}`,
		expectation: &EventFailedHealthCheck{
			EventType: "failed_health_check_event",
			Timestamp: "2014-03-01T23:29:30.158Z",
			AppID:     "/my-app",
			HealthCheck: struct {
				GracePeriodSeconds     float64 `json:"gracePeriodSeconds"`
				IntervalSeconds        float64 `json:"intervalSeconds"`
				MaxConsecutiveFailures float64 `json:"maxConsecutiveFailures"`
				Path                   string  `json:"path"`
				PortIndex              float64 `json:"portIndex"`
				Protocol               string  `json:"protocol"`
				TimeoutSeconds         float64 `json:"timeoutSeconds"`
			}{
				GracePeriodSeconds:     5,
				IntervalSeconds:        10,
				MaxConsecutiveFailures: 3,
				Path:           "/health",
				PortIndex:      0,
				Protocol:       "HTTP",
				TimeoutSeconds: 10,
			},
		},
	},
	// For Marathon 1.1.1 and before
	testCase{
		name: "deployment_info",
		source: `{
	"eventType": "deployment_info",
	"timestamp": "2016-07-29T08:03:52.542Z",
	"plan": {
		"id": "dcf63e4a-ef27-4816-e865-1730fcb26ac3",
		"version": "2016-07-29T08:03:52.542Z",
		"original": {},
		"target": {},
		"steps": [
			{
				"actions": [
					{
						"type": "ScaleApplication",
						"app": "/my-app"
					}
				]
			}
		]
	},
	"currentStep": {
		"actions": [
			{
				"type": "ScaleApplication",
				"app": "/my-app"
			}
		]
	}
}`,
		expectation: &EventDeploymentInfo{
			EventType: "deployment_info",
			Timestamp: "2016-07-29T08:03:52.542Z",
			Plan: &DeploymentPlan{
				ID:       "dcf63e4a-ef27-4816-e865-1730fcb26ac3",
				Version:  "2016-07-29T08:03:52.542Z",
				Original: &Group{},
				Target:   &Group{},
				Steps: []*StepActions{
					&StepActions{
						Actions: []struct {
							Action string `json:"action"`
							Type   string `json:"type"`
							App    string `json:"app"`
						}{
							{
								Type: "ScaleApplication",
								App:  "/my-app",
							},
						},
					},
				},
			},
			CurrentStep: &StepActions{
				Actions: []struct {
					Action string `json:"action"`
					Type   string `json:"type"`
					App    string `json:"app"`
				}{
					{
						Type: "ScaleApplication",
						App:  "/my-app",
					},
				},
			},
		},
	},
	// For Marathon 1.1.2 and after
	testCase{
		name: "deployment_step_success",
		source: `{
	"eventType": "deployment_step_success",
	"timestamp": "2016-07-29T08:03:52.542Z",
	"plan": {
		"id": "dcf63e4a-ef27-4816-e865-1730fcb26ac3",
		"version": "2016-07-29T08:03:52.542Z",
		"original": {},
		"target": {},
		"steps": [
			{
				"actions": [
					{
						"action": "ScaleApplication",
						"app": "/my-app"
					}
				]
			}
		]
	},
	"currentStep": {
		"actions": [
			{
				"action": "ScaleApplication",
				"app": "/my-app"
			}
		]
	}
}`,
		expectation: &EventDeploymentInfo{
			EventType: "deployment_info",
			Timestamp: "2016-07-29T08:03:52.542Z",
			Plan: &DeploymentPlan{
				ID:       "dcf63e4a-ef27-4816-e865-1730fcb26ac3",
				Version:  "2016-07-29T08:03:52.542Z",
				Original: &Group{},
				Target:   &Group{},
				Steps: []*StepActions{
					&StepActions{
						Actions: []struct {
							Action string `json:"action"`
							Type   string `json:"type"`
							App    string `json:"app"`
						}{
							{
								Action: "ScaleApplication",
								App:    "/my-app",
							},
						},
					},
				},
			},
			CurrentStep: &StepActions{
				Actions: []struct {
					Action string `json:"action"`
					Type   string `json:"type"`
					App    string `json:"app"`
				}{
					{
						Action: "ScaleApplication",
						App:    "/my-app",
					},
				},
			},
		},
	},
}

func TestSubscriptions(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	sub, err := endpoint.Client.Subscriptions()
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.NotNil(t, sub.CallbackURLs)
	assert.Equal(t, len(sub.CallbackURLs), 1)
}

func TestSubscribe(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	err := endpoint.Client.Subscribe("http://localhost:9292/callback")
	assert.NoError(t, err)
}

func TestUnsubscribe(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	err := endpoint.Client.Unsubscribe("http://localhost:9292/callback")
	assert.NoError(t, err)
}

func TestEventStreamEventsReceived(t *testing.T) {
	require.True(t, len(testCases) > 1, "must have at least 2 test cases to end prematurely")

	clientCfg := NewDefaultConfig()
	config := configContainer{
		client: &clientCfg,
	}
	config.client.EventsTransport = EventsTransportSSE
	endpoint := newFakeMarathonEndpoint(t, &config)
	defer endpoint.Close()

	events, err := endpoint.Client.AddEventsListener(EventIDApplications | EventIDDeploymentInfo | EventIDDeploymentStepSuccess)
	assert.NoError(t, err)

	almostAllTestCases := testCases[:len(testCases)-1]
	finalTestCase := testCases[len(testCases)-1]

	// Let a bit time pass so that the SSE subscription can connect
	time.Sleep(SSEConnectWaitTime)

	// Publish all but one test event.
	for _, testCase := range almostAllTestCases {
		endpoint.Server.PublishEvent(testCase.source)
	}

	// Receive test events.
	for i := 0; i < len(almostAllTestCases); i++ {
		select {
		case event := <-events:
			tc := testCases.find(event.Name)
			if !assert.NotNil(t, tc, "received unknown event: %s", event.Name) {
				continue
			}
			assert.Equal(t, tc.expectation, event.Event)
		case <-time.After(eventPublishTimeout):
			assert.Fail(t, "did not receive event in time")
		}
	}

	// Publish last test event that we do not intend to consume anymore.
	endpoint.Server.PublishEvent(finalTestCase.source)

	// Give event stream some time to buffer another event.
	time.Sleep(eventPublishTimeout)

	// Trigger done channel closure.
	endpoint.Client.RemoveEventsListener(events)

	// Give pending goroutine time to consume done signal.
	time.Sleep(eventPublishTimeout)

	// Validate that channel is closed.
	select {
	case _, more := <-events:
		assert.False(t, more, "should not have received another event")
	default:
		assert.Fail(t, "channel was not closed")
	}
}

func TestConnectToSSESuccess(t *testing.T) {
	clientCfg := NewDefaultConfig()
	// Use non-existent address as first cluster member
	clientCfg.URL = "http://127.0.0.1:11111"
	clientCfg.EventsTransport = EventsTransportSSE
	config := configContainer{client: &clientCfg}

	endpoint := newFakeMarathonEndpoint(t, &config)
	defer endpoint.Close()

	client := endpoint.Client.(*marathonClient)
	// Add real server as member to the cluster
	client.hosts.members = append(client.hosts.members, &member{endpoint: endpoint.Server.httpSrv.URL})

	// Connection should work as one of the Marathon members is up
	_, err := client.connectToSSE()
	assert.NoError(t, err, "expected no error in connectToSSE")
}

func TestConnectToSSEFailure(t *testing.T) {
	clientCfg := NewDefaultConfig()
	clientCfg.EventsTransport = EventsTransportSSE
	config := configContainer{client: &clientCfg}

	endpoint := newFakeMarathonEndpoint(t, &config)
	endpoint.Close()

	client := endpoint.Client.(*marathonClient)

	// No Marathon member is up, we should get an error
	_, err := client.connectToSSE()
	assert.Error(t, err, "expected error in connectToSSE when all cluster members are down")
}

func TestRegisterSEESubscriptionReconnectsStreamOnError(t *testing.T) {
	clientCfg := NewDefaultConfig()
	clientCfg.EventsTransport = EventsTransportSSE
	config := configContainer{client: &clientCfg}

	endpoint1 := newFakeMarathonEndpoint(t, &config)
	endpoint2 := newFakeMarathonEndpoint(t, &config)
	defer endpoint2.Close()

	client1 := endpoint1.Client.(*marathonClient)
	// Add the second server to the cluster members
	client1.hosts.members = append(client1.hosts.members, &member{endpoint: endpoint2.Server.httpSrv.URL})

	events, err := endpoint1.Client.AddEventsListener(EventIDApplications)
	require.NoError(t, err)

	// This should make the SSE subscription fail and reconnect to another cluster member
	endpoint1.Close()

	// Let a bit time so that subscription can reconnect
	time.Sleep(SSEConnectWaitTime)

	// Now that our SSE subscription failed over, we can publish on the second server and the message should be consumed
	endpoint2.Server.PublishEvent(testCases[0].source)

	select {
	case event := <-events:
		tc := testCases.find(event.Name)
		assert.NotNil(t, tc, "received unknown event: %s", event.Name)
	case <-time.After(eventPublishTimeout):
		assert.Fail(t, "did not receive event in time")
	}
}
