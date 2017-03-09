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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

// Marathon is the interface to the marathon API
type Marathon interface {
	// -- APPLICATIONS ---

	// get a listing of the application ids
	ListApplications(url.Values) ([]string, error)
	// a list of application versions
	ApplicationVersions(name string) (*ApplicationVersions, error)
	// check a application version exists
	HasApplicationVersion(name, version string) (bool, error)
	// change an application to a different version
	SetApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error)
	// check if an application is ok
	ApplicationOK(name string) (bool, error)
	// create an application in marathon
	CreateApplication(application *Application) (*Application, error)
	// delete an application
	DeleteApplication(name string, force bool) (*DeploymentID, error)
	// update an application in marathon
	UpdateApplication(application *Application, force bool) (*DeploymentID, error)
	// a list of deployments on a application
	ApplicationDeployments(name string) ([]*DeploymentID, error)
	// scale a application
	ScaleApplicationInstances(name string, instances int, force bool) (*DeploymentID, error)
	// restart an application
	RestartApplication(name string, force bool) (*DeploymentID, error)
	// get a list of applications from marathon
	Applications(url.Values) (*Applications, error)
	// get an application by name
	Application(name string) (*Application, error)
	// get an application by options
	ApplicationBy(name string, opts *GetAppOpts) (*Application, error)
	// get an application by name and version
	ApplicationByVersion(name, version string) (*Application, error)
	// wait of application
	WaitOnApplication(name string, timeout time.Duration) error

	// -- TASKS ---

	// get a list of tasks for a specific application
	Tasks(application string) (*Tasks, error)
	// get a list of all tasks
	AllTasks(opts *AllTasksOpts) (*Tasks, error)
	// get the endpoints for a service on a application
	TaskEndpoints(name string, port int, healthCheck bool) ([]string, error)
	// kill all the tasks for any application
	KillApplicationTasks(applicationID string, opts *KillApplicationTasksOpts) (*Tasks, error)
	// kill a single task
	KillTask(taskID string, opts *KillTaskOpts) (*Task, error)
	// kill the given array of tasks
	KillTasks(taskIDs []string, opts *KillTaskOpts) error

	// --- GROUPS ---

	// list all the groups in the system
	Groups() (*Groups, error)
	// retrieve a specific group from marathon
	Group(name string) (*Group, error)
	// list all groups in marathon by options
	GroupsBy(opts *GetGroupOpts) (*Groups, error)
	// retrieve a specific group from marathon by options
	GroupBy(name string, opts *GetGroupOpts) (*Group, error)
	// create a group deployment
	CreateGroup(group *Group) error
	// delete a group
	DeleteGroup(name string, force bool) (*DeploymentID, error)
	// update a groups
	UpdateGroup(id string, group *Group, force bool) (*DeploymentID, error)
	// check if a group exists
	HasGroup(name string) (bool, error)
	// wait for an group to be deployed
	WaitOnGroup(name string, timeout time.Duration) error

	// --- DEPLOYMENTS ---

	// get a list of the deployments
	Deployments() ([]*Deployment, error)
	// delete a deployment
	DeleteDeployment(id string, force bool) (*DeploymentID, error)
	// check to see if a deployment exists
	HasDeployment(id string) (bool, error)
	// wait of a deployment to finish
	WaitOnDeployment(id string, timeout time.Duration) error

	// --- SUBSCRIPTIONS ---

	// a list of current subscriptions
	Subscriptions() (*Subscriptions, error)
	// add a events listener
	AddEventsListener(filter int) (EventsChannel, error)
	// remove a events listener
	RemoveEventsListener(channel EventsChannel)
	// Subscribe a callback URL
	Subscribe(string) error
	// Unsubscribe a callback URL
	Unsubscribe(string) error

	// --- QUEUE ---
	// get marathon launch queue
	Queue() (*Queue, error)
	// resets task launch delay of the specific application
	DeleteQueueDelay(appID string) error

	// --- MISC ---

	// get the marathon url
	GetMarathonURL() string
	// ping the marathon
	Ping() (bool, error)
	// grab the marathon server info
	Info() (*Info, error)
	// retrieve the leader info
	Leader() (string, error)
	// cause the current leader to abdicate
	AbdicateLeader() (string, error)
}

var (
	// ErrInvalidResponse is thrown when marathon responds with invalid or error response
	ErrInvalidResponse = errors.New("invalid response from Marathon")
	// ErrMarathonDown is thrown when all the marathon endpoints are down
	ErrMarathonDown = errors.New("all the Marathon hosts are presently down")
	// ErrTimeoutError is thrown when the operation has timed out
	ErrTimeoutError = errors.New("the operation has timed out")
)

// EventsChannelContext holds contextual data for an EventsChannel.
type EventsChannelContext struct {
	filter     int
	done       chan struct{}
	completion *sync.WaitGroup
}

type marathonClient struct {
	sync.RWMutex
	// the configuration for the client
	config Config
	// the flag used to prevent multiple SSE subscriptions
	subscribedToSSE bool
	// the ip address of the client
	ipAddress string
	// the http server
	eventsHTTP *http.Server
	// the http client use for making requests
	httpClient *http.Client
	// the marathon hosts
	hosts *cluster
	// a map of service you wish to listen to
	listeners map[EventsChannel]EventsChannelContext
	// a custom logger for debug log messages
	debugLog *log.Logger
}

// NewClient creates a new marathon client
//		config:			the configuration to use
func NewClient(config Config) (Marathon, error) {
	// step: if no http client, set to default
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	// step: if no polling wait time is set, default to 500 milliseconds.
	if config.PollingWaitTime == 0 {
		config.PollingWaitTime = defaultPollingWaitTime
	}

	// step: create a new cluster
	hosts, err := newCluster(config.HTTPClient, config.URL)
	if err != nil {
		return nil, err
	}

	debugLogOutput := config.LogOutput
	if debugLogOutput == nil {
		debugLogOutput = ioutil.Discard
	}

	return &marathonClient{
		config:     config,
		listeners:  make(map[EventsChannel]EventsChannelContext),
		hosts:      hosts,
		httpClient: config.HTTPClient,
		debugLog:   log.New(debugLogOutput, "", 0),
	}, nil
}

// GetMarathonURL retrieves the marathon url
func (r *marathonClient) GetMarathonURL() string {
	return r.config.URL
}

// Ping pings the current marathon endpoint (note, this is not a ICMP ping, but a rest api call)
func (r *marathonClient) Ping() (bool, error) {
	if err := r.apiGet(marathonAPIPing, nil, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (r *marathonClient) apiGet(uri string, post, result interface{}) error {
	return r.apiCall("GET", uri, post, result)
}

func (r *marathonClient) apiPut(uri string, post, result interface{}) error {
	return r.apiCall("PUT", uri, post, result)
}

func (r *marathonClient) apiPost(uri string, post, result interface{}) error {
	return r.apiCall("POST", uri, post, result)
}

func (r *marathonClient) apiDelete(uri string, post, result interface{}) error {
	return r.apiCall("DELETE", uri, post, result)
}

func (r *marathonClient) apiCall(method, uri string, body, result interface{}) error {
	for {
		// step: grab a member from the cluster and attempt to perform the request
		member, err := r.hosts.getMember()
		if err != nil {
			return ErrMarathonDown
		}

		// step: Create the endpoint url
		url := fmt.Sprintf("%s/%s", member, uri)
		if r.config.DCOSToken != "" {
			url = fmt.Sprintf("%s/%s", member+"/marathon", uri)
		}

		// step: marshall the request to json
		var requestBody []byte
		if body != nil {
			if requestBody, err = json.Marshal(body); err != nil {
				return err
			}
		}

		// step: create the api request
		request, err := r.buildAPIRequest(method, url, bytes.NewReader(requestBody))
		if err != nil {
			return err
		}
		response, err := r.httpClient.Do(request)
		if err != nil {
			r.hosts.markDown(member)
			// step: attempt the request on another member
			r.debugLog.Printf("apiCall(): request failed on host: %s, error: %s, trying another\n", member, err)
			continue
		}
		defer response.Body.Close()

		// step: read the response body
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if len(requestBody) > 0 {
			r.debugLog.Printf("apiCall(): %v %v %s returned %v %s\n", request.Method, request.URL.String(), requestBody, response.Status, oneLogLine(respBody))
		} else {
			r.debugLog.Printf("apiCall(): %v %v returned %v %s\n", request.Method, request.URL.String(), response.Status, oneLogLine(respBody))
		}

		// step: check for a successfull response
		if response.StatusCode >= 200 && response.StatusCode <= 299 {
			if result != nil {
				if err := json.Unmarshal(respBody, result); err != nil {
					r.debugLog.Printf("apiCall(): failed to unmarshall the response from marathon, error: %s\n", err)
					return ErrInvalidResponse
				}
			}
			return nil
		}

		// step: if the member node returns a >= 500 && <= 599 we should try another node?
		if response.StatusCode >= 500 && response.StatusCode <= 599 {
			// step: mark the host as down
			r.hosts.markDown(member)
			r.debugLog.Printf("apiCall(): request failed, host: %s, status: %d, trying another\n", member, response.StatusCode)
			continue
		}

		return NewAPIError(response.StatusCode, respBody)
	}
}

// buildAPIRequest creates a default API request
func (r *marathonClient) buildAPIRequest(method, url string, reader io.Reader) (*http.Request, error) {
	// Make the http request to Marathon
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	// Add any basic auth and the content headers
	if r.config.HTTPBasicAuthUser != "" && r.config.HTTPBasicPassword != "" {
		request.SetBasicAuth(r.config.HTTPBasicAuthUser, r.config.HTTPBasicPassword)
	}

	if r.config.DCOSToken != "" {
		request.Header.Add("Authorization", "token="+r.config.DCOSToken)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	return request, nil
}

var oneLogLineRegex = regexp.MustCompile(`(?m)^\s*`)

// oneLogLine removes indentation at the beginning of each line and
// escapes new line characters.
func oneLogLine(in []byte) []byte {
	return bytes.Replace(oneLogLineRegex.ReplaceAll(in, nil), []byte("\n"), []byte("\\n "), -1)
}
