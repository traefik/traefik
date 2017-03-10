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
	DeleteApplication(name string) (*DeploymentID, error)
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
	// create a group deployment
	CreateGroup(group *Group) error
	// delete a group
	DeleteGroup(name string) (*DeploymentID, error)
	// update a groups
	UpdateGroup(id string, group *Group) (*DeploymentID, error)
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
	AddEventsListener(channel EventsChannel, filter int) error
	// remove a events listener
	RemoveEventsListener(channel EventsChannel)
	// remove our self from subscriptions
	Unsubscribe(string) error

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
	// ErrInvalidEndpoint is thrown when the marathon url specified was invalid
	ErrInvalidEndpoint = errors.New("invalid Marathon endpoint specified")
	// ErrInvalidResponse is thrown when marathon responds with invalid or error response
	ErrInvalidResponse = errors.New("invalid response from Marathon")
	// ErrMarathonDown is thrown when all the marathon endpoints are down
	ErrMarathonDown = errors.New("all the Marathon hosts are presently down")
	// ErrTimeoutError is thrown when the operation has timed out
	ErrTimeoutError = errors.New("the operation has timed out")
)

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
	// the marathon cluster
	cluster Cluster
	// a map of service you wish to listen to
	listeners map[EventsChannel]int
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
	// step: create a new cluster
	cluster, err := newCluster(config.HTTPClient, config.URL)
	if err != nil {
		return nil, err
	}

	debugLogOutput := config.LogOutput
	if debugLogOutput == nil {
		debugLogOutput = ioutil.Discard
	}

	return &marathonClient{
		config:     config,
		listeners:  make(map[EventsChannel]int, 0),
		cluster:    cluster,
		httpClient: config.HTTPClient,
		debugLog:   log.New(debugLogOutput, "", 0),
	}, nil
}

// GetMarathonURL retrieves the marathon url
func (r *marathonClient) GetMarathonURL() string {
	return r.cluster.URL()
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

	// Get a member from the cluster
	marathon, err := r.cluster.GetMember()
	if err != nil {
		return err
	}

	var url string

	if r.config.DCOSToken != "" {
		url = fmt.Sprintf("%s/%s", marathon+"/marathon", uri)
	} else {
		url = fmt.Sprintf("%s/%s", marathon, uri)
	}

	var jsonBody []byte
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	// step: create an API request
	request, err := r.apiRequest(method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	response, err := r.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(jsonBody) > 0 {
		r.debugLog.Printf("apiCall(): %v %v %s returned %v %s\n", request.Method, request.URL.String(), jsonBody, response.Status, oneLogLine(respBody))
	} else {
		r.debugLog.Printf("apiCall(): %v %v returned %v %s\n", request.Method, request.URL.String(), response.Status, oneLogLine(respBody))
	}

	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		if result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				r.debugLog.Printf("apiCall(): failed to unmarshall the response from marathon, error: %s\n", err)
				return ErrInvalidResponse
			}
		}
		return nil
	}
	return NewAPIError(response.StatusCode, respBody)
}

// apiRequest creates a default API request
func (r *marathonClient) apiRequest(method, url string, reader io.Reader) (*http.Request, error) {
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
