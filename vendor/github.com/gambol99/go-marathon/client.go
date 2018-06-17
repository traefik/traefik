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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

	// -- PODS ---
	// whether this version of Marathon supports pods
	SupportsPods() (bool, error)

	// get pod status
	PodStatus(name string) (*PodStatus, error)
	// get all pod statuses
	PodStatuses() ([]*PodStatus, error)

	// get pod
	Pod(name string) (*Pod, error)
	// get all pods
	Pods() ([]Pod, error)
	// create pod
	CreatePod(pod *Pod) (*Pod, error)
	// update pod
	UpdatePod(pod *Pod, force bool) (*Pod, error)
	// delete pod
	DeletePod(name string, force bool) (*DeploymentID, error)
	// wait on pod to be deployed
	WaitOnPod(name string, timeout time.Duration) error
	// check if a pod is running
	PodIsRunning(name string) bool

	// get versions of a pod
	PodVersions(name string) ([]string, error)
	// get pod by version
	PodByVersion(name, version string) (*Pod, error)

	// delete instances of a pod
	DeletePodInstances(name string, instances []string) ([]*PodInstance, error)
	// delete pod instance
	DeletePodInstance(name, instance string) (*PodInstance, error)

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
	// ErrMarathonDown is thrown when all the marathon endpoints are down
	ErrMarathonDown = errors.New("all the Marathon hosts are presently down")
	// ErrTimeoutError is thrown when the operation has timed out
	ErrTimeoutError = errors.New("the operation has timed out")

	// Default HTTP client used for SSE subscription requests
	// It is invalid to set client.Timeout because it includes time to read response so
	// set dial, tls handshake and response header timeouts instead
	defaultHTTPSSEClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
		},
	}

	// Default HTTP client used for non SSE requests
	defaultHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}
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
	// the marathon hosts
	hosts *cluster
	// a map of service you wish to listen to
	listeners map[EventsChannel]EventsChannelContext
	// a custom log function for debug messages
	debugLog func(format string, v ...interface{})
	// the marathon HTTP client to ensure consistency in requests
	client *httpClient
}

type httpClient struct {
	// the configuration for the marathon HTTP client
	config Config
}

// newRequestError signals that creating a new http.Request failed
type newRequestError struct {
	error
}

// NewClient creates a new marathon client
//		config:			the configuration to use
func NewClient(config Config) (Marathon, error) {
	// step: if the SSE HTTP client is missing, prefer a configured regular
	// client, and otherwise use the default SSE HTTP client.
	if config.HTTPSSEClient == nil {
		config.HTTPSSEClient = defaultHTTPSSEClient
		if config.HTTPClient != nil {
			config.HTTPSSEClient = config.HTTPClient
		}
	}

	// step: if a regular HTTP client is missing, use the default one.
	if config.HTTPClient == nil {
		config.HTTPClient = defaultHTTPClient
	}

	// step: if no polling wait time is set, default to 500 milliseconds.
	if config.PollingWaitTime == 0 {
		config.PollingWaitTime = defaultPollingWaitTime
	}

	// step: setup shared client
	client := &httpClient{config: config}

	// step: create a new cluster
	hosts, err := newCluster(client, config.URL, config.DCOSToken != "")
	if err != nil {
		return nil, err
	}

	debugLog := func(string, ...interface{}) {}
	if config.LogOutput != nil {
		logger := log.New(config.LogOutput, "", 0)
		debugLog = func(format string, v ...interface{}) {
			logger.Printf(format, v...)
		}
	}

	return &marathonClient{
		config:    config,
		listeners: make(map[EventsChannel]EventsChannelContext),
		hosts:     hosts,
		debugLog:  debugLog,
		client:    client,
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

func (r *marathonClient) apiHead(path string, result interface{}) error {
	return r.apiCall("HEAD", path, nil, result)
}

func (r *marathonClient) apiGet(path string, post, result interface{}) error {
	return r.apiCall("GET", path, post, result)
}

func (r *marathonClient) apiPut(path string, post, result interface{}) error {
	return r.apiCall("PUT", path, post, result)
}

func (r *marathonClient) apiPost(path string, post, result interface{}) error {
	return r.apiCall("POST", path, post, result)
}

func (r *marathonClient) apiDelete(path string, post, result interface{}) error {
	return r.apiCall("DELETE", path, post, result)
}

func (r *marathonClient) apiCall(method, path string, body, result interface{}) error {
	const deploymentHeader = "Marathon-Deployment-Id"

	for {
		// step: marshall the request to json
		var requestBody []byte
		var err error
		if body != nil {
			if requestBody, err = json.Marshal(body); err != nil {
				return err
			}
		}

		// step: create the API request
		request, member, err := r.buildAPIRequest(method, path, bytes.NewReader(requestBody))
		if err != nil {
			return err
		}

		// step: perform the API request
		response, err := r.client.Do(request)
		if err != nil {
			r.hosts.markDown(member)
			// step: attempt the request on another member
			r.debugLog("apiCall(): request failed on host: %s, error: %s, trying another", member, err)
			continue
		}
		defer response.Body.Close()

		// step: read the response body
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if len(requestBody) > 0 {
			r.debugLog("apiCall(): %v %v %s returned %v %s", request.Method, request.URL.String(), requestBody, response.Status, oneLogLine(respBody))
		} else {
			r.debugLog("apiCall(): %v %v returned %v %s", request.Method, request.URL.String(), response.Status, oneLogLine(respBody))
		}

		// step: check for a successful response
		if response.StatusCode >= 200 && response.StatusCode <= 299 {
			if result != nil {
				// If we have a deployment ID header and no response body, give them that
				// This specifically handles the use case of a DELETE on an app/pod
				// We need a way to retrieve the deployment ID
				deploymentID := response.Header.Get(deploymentHeader)
				if len(respBody) == 0 && deploymentID != "" {
					d := DeploymentID{
						DeploymentID: deploymentID,
					}
					if deployID, ok := result.(*DeploymentID); ok {
						*deployID = d
					}
				} else {
					if err := json.Unmarshal(respBody, result); err != nil {
						return fmt.Errorf("failed to unmarshal response from Marathon: %s", err)
					}
				}
			}
			return nil
		}

		// step: if the member node returns a >= 500 && <= 599 we should try another node?
		if response.StatusCode >= 500 && response.StatusCode <= 599 {
			// step: mark the host as down
			r.hosts.markDown(member)
			r.debugLog("apiCall(): request failed, host: %s, status: %d, trying another", member, response.StatusCode)
			continue
		}

		return NewAPIError(response.StatusCode, respBody)
	}
}

// wait waits until the provided function returns true (or times out)
func (r *marathonClient) wait(name string, timeout time.Duration, fn func(string) bool) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(r.config.PollingWaitTime)
	defer ticker.Stop()
	for {
		if fn(name) {
			return nil
		}

		select {
		case <-timer.C:
			return ErrTimeoutError
		case <-ticker.C:
			continue
		}
	}
}

// buildAPIRequest creates a default API request.
// It fails when there is no available member in the cluster anymore or when the request can not be built.
func (r *marathonClient) buildAPIRequest(method, path string, reader io.Reader) (request *http.Request, member string, err error) {
	// Grab a member from the cluster
	member, err = r.hosts.getMember()
	if err != nil {
		return nil, "", ErrMarathonDown
	}

	// Build the HTTP request to Marathon
	request, err = r.client.buildMarathonJSONRequest(method, member, path, reader)
	if err != nil {
		return nil, member, newRequestError{err}
	}
	return request, member, nil
}

// buildMarathonJSONRequest is like buildMarathonRequest but sets the
// Content-Type and Accept headers to application/json.
func (rc *httpClient) buildMarathonJSONRequest(method, member, path string, reader io.Reader) (request *http.Request, err error) {
	req, err := rc.buildMarathonRequest(method, member, path, reader)
	if err == nil {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
	}

	return req, err
}

// buildMarathonRequest creates a new HTTP request and configures it according to the *httpClient configuration.
// The path must not contain a leading "/", otherwise buildMarathonRequest will panic.
func (rc *httpClient) buildMarathonRequest(method, member, path string, reader io.Reader) (request *http.Request, err error) {
	if strings.HasPrefix(path, "/") {
		panic(fmt.Sprintf("Path '%s' must not start with a leading slash", path))
	}

	// Create the endpoint URL
	url := fmt.Sprintf("%s/%s", member, path)

	// Instantiate an HTTP request
	request, err = http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	// Add any basic auth and the content headers
	if rc.config.HTTPBasicAuthUser != "" && rc.config.HTTPBasicPassword != "" {
		request.SetBasicAuth(rc.config.HTTPBasicAuthUser, rc.config.HTTPBasicPassword)
	}

	if rc.config.DCOSToken != "" {
		request.Header.Add("Authorization", "token="+rc.config.DCOSToken)
	}

	return request, nil
}

func (rc *httpClient) Do(request *http.Request) (response *http.Response, err error) {
	return rc.config.HTTPClient.Do(request)
}

var oneLogLineRegex = regexp.MustCompile(`(?m)^\s*`)

// oneLogLine removes indentation at the beginning of each line and
// escapes new line characters.
func oneLogLine(in []byte) []byte {
	return bytes.Replace(oneLogLineRegex.ReplaceAll(in, nil), []byte("\n"), []byte("\\n "), -1)
}
