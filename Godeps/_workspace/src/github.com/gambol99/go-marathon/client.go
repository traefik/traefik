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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	HTTP_GET    = "GET"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"
	HTTP_POST   = "POST"
)

type Marathon interface {
	/* -- APPLICATIONS --- */

	/* check it see if a application exists */
	HasApplication(name string) (bool, error)
	/* get a listing of the application ids */
	ListApplications(url.Values) ([]string, error)
	/* a list of application versions */
	ApplicationVersions(name string) (*ApplicationVersions, error)
	/* check a application version exists */
	HasApplicationVersion(name, version string) (bool, error)
	/* change an application to a different version */
	SetApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error)
	/* check if an application is ok */
	ApplicationOK(name string) (bool, error)
	/* create an application in marathon */
	CreateApplication(application *Application, wait_on_running bool) (*Application, error)
	/* delete an application */
	DeleteApplication(name string) (*DeploymentID, error)
	/* update an application in marathon */
	UpdateApplication(application *Application, wait_on_running bool) (*Application, error)
	/* a list of deployments on a application */
	ApplicationDeployments(name string) ([]*DeploymentID, error)
	/* scale a application */
	ScaleApplicationInstances(name string, instances int) (*DeploymentID, error)
	/* restart an application */
	RestartApplication(name string, force bool) (*DeploymentID, error)
	/* get a list of applications from marathon */
	Applications(url.Values) (*Applications, error)
	/* get a specific application */
	Application(name string) (*Application, error)
	/* wait of application */
	WaitOnApplication(name string, timeout time.Duration) error

	/* -- TASKS --- */

	/* get a list of tasks for a specific application */
	Tasks(application string) (*Tasks, error)
	/* get a list of all tasks */
	AllTasks() (*Tasks, error)
	/* get a listing of the task ids */
	ListTasks() ([]string, error)
	/* get the endpoints for a service on a application */
	TaskEndpoints(name string, port int, health_check bool) ([]string, error)
	/* kill all the tasks for any application */
	KillApplicationTasks(application_id, hostname string, scale bool) (*Tasks, error)
	/* kill a single task */
	KillTask(task_id string, scale bool) (*Task, error)
	/* kill the given array of tasks */
	KillTasks(task_ids []string, scale bool) error

	/* --- GROUPS --- */

	/* list all the groups in the system */
	Groups() (*Groups, error)
	/* retrieve a specific group from marathon */
	Group(name string) (*Group, error)
	/* create a group deployment */
	CreateGroup(group *Group, wait_on_running bool) error
	/* delete a group */
	DeleteGroup(name string) (*DeploymentID, error)
	/* update a groups */
	UpdateGroup(id string, group *Group) (*DeploymentID, error)
	/* check if a group exists */
	HasGroup(name string) (bool, error)
	/* wait for an group to be deployed */
	WaitOnGroup(name string, timeout time.Duration) error

	/* --- DEPLOYMENTS --- */

	/* get a list of the deployments */
	Deployments() ([]*Deployment, error)
	/* delete a deployment */
	DeleteDeployment(id string, force bool) (*DeploymentID, error)
	/* check to see if a deployment exists */
	HasDeployment(id string) (bool, error)
	/* wait of a deployment to finish */
	WaitOnDeployment(version string, timeout time.Duration) error

	/* --- SUBSCRIPTIONS --- */

	/* a list of current subscriptions */
	Subscriptions() (*Subscriptions, error)
	/* add a events listener */
	AddEventsListener(channel EventsChannel, filter int) error
	/* remove a events listener */
	RemoveEventsListener(channel EventsChannel)
	/* remove our self from subscriptions */
	UnSubscribe() error

	/* --- MISC --- */

	/* get the marathon url */
	GetMarathonURL() string
	/* ping the marathon */
	Ping() (bool, error)
	/* grab the marathon server info */
	Info() (*Info, error)
	/* retrieve the leader info */
	Leader() (string, error)
	/* cause the current leader to abdicate */
	AbdicateLeader() (string, error)
}

var (
	/* the url specified was invalid */
	ErrInvalidEndpoint = errors.New("Invalid Marathon endpoint specified")
	/* invalid or error response from marathon */
	ErrInvalidResponse = errors.New("Invalid response from Marathon")
	/* some resource does not exists */
	ErrDoesNotExist = errors.New("The resource does not exist")
	/* all the marathon endpoints are down */
	ErrMarathonDown = errors.New("All the Marathon hosts are presently down")
	/* unable to decode the response */
	ErrInvalidResult = errors.New("Unable to decode the response from Marathon")
	/* invalid argument */
	ErrInvalidArgument = errors.New("The argument passed is invalid")
	/* error return by marathon */
	ErrMarathonError = errors.New("Marathon error")
	/* the operation has timed out */
	ErrTimeoutError = errors.New("The operation has timed out")
)

type Client struct {
	sync.RWMutex
	/* the configuration for the client */
	config Config
	/* the ip addess of the client */
	ipaddress string
	/* the http server */
	events_http *http.Server
	/* the http client */
	http *http.Client
	/* the output for the logger */
	logger *log.Logger
	/* the marathon cluster */
	cluster Cluster
	/* a map of service you wish to listen to */
	listeners map[EventsChannel]int
}

type Message struct {
	Message string `json:"message"`
}

func NewClient(config Config) (Marathon, error) {
	/* step: we parse the url and build a cluster */
	if cluster, err := NewMarathonCluster(config.URL); err != nil {
		return nil, err
	} else {
		// step: create the service marathon client
		service := new(Client)
		service.config = config
		// step: create a logger from the output
		if config.LogOutput == nil {
			config.LogOutput = ioutil.Discard
		}
		service.logger = log.New(config.LogOutput, "[debug]: ", 0)
		service.listeners = make(map[EventsChannel]int, 0)
		service.cluster = cluster
		service.http = &http.Client{
			Timeout: (time.Duration(config.RequestTimeout) * time.Second),
		}
		return service, nil
	}
}

func (client *Client) GetMarathonURL() string {
	return client.cluster.Url()
}

// Pings the current marathon endpoint (note, this is not a ICMP ping, but a rest api call)
func (client *Client) Ping() (bool, error) {
	if err := client.apiGet(MARATHON_API_PING, nil, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (client *Client) marshallJSON(data interface{}) (string, error) {
	if response, err := json.Marshal(data); err != nil {
		return "", err
	} else {
		return string(response), err
	}
}

func (client *Client) unMarshallDataToJson(stream io.Reader, result interface{}) error {
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(result); err != nil {
		return err
	}
	return nil
}

func (client *Client) unmarshallJsonArray(stream io.Reader, results []interface{}) error {
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(results); err != nil {
		return err
	}
	return nil
}

func (client *Client) apiPostData(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}
	content, err := client.marshallJSON(data)
	if err != nil {
		return "", err
	}
	return content, nil
}

func (client *Client) apiGet(uri string, post, result interface{}) error {
	if content, err := client.apiPostData(post); err != nil {
		return err
	} else {
		_, _, error := client.apiCall(HTTP_GET, uri, content, result)
		return error
	}
}

func (client *Client) apiPut(uri string, post, result interface{}) error {
	if content, err := client.apiPostData(post); err != nil {
		return err
	} else {
		_, _, error := client.apiCall(HTTP_PUT, uri, content, result)
		return error
	}
}

func (client *Client) apiPost(uri string, post, result interface{}) error {
	if content, err := client.apiPostData(post); err != nil {
		return err
	} else {
		_, _, error := client.apiCall(HTTP_POST, uri, content, result)
		return error
	}
}

func (client *Client) apiDelete(uri string, post, result interface{}) error {
	if content, err := client.apiPostData(post); err != nil {
		return err
	} else {
		_, _, error := client.apiCall(HTTP_DELETE, uri, content, result)
		return error
	}
}

func (client *Client) apiCall(method, uri, body string, result interface{}) (int, string, error) {
	client.log("apiCall() method: %s, uri: %s, body: %s", method, uri, body)
	if status, content, _, err := client.httpCall(method, uri, body); err != nil {
		return 0, "", err
	} else {
		client.log("apiCall() status: %d, content: %s\n", status, content)
		if status >= 200 && status <= 299 {
			if result != nil {
				if err := client.unMarshallDataToJson(strings.NewReader(content), result); err != nil {
					client.log("apiCall(): failed to unmarshall the response from marathon, error: %s", err)
					return status, content, ErrInvalidResponse
				}
			}
			client.log("apiCall() result: %V", result)
			return status, content, nil
		}
		switch status {
		case 500:
			return status, "", ErrInvalidResponse
		case 404:
			return status, "", ErrDoesNotExist
		}

		/* step: lets decode into a error message */
		var message Message
		if err := client.unMarshallDataToJson(strings.NewReader(content), &message); err != nil {
			return status, content, ErrInvalidResponse
		} else {
			errorMessage := "unknown error"
			if message.Message != "" {
				errorMessage = message.Message
			}
			return status, message.Message, errors.New(errorMessage)
		}
	}
}

func (client *Client) httpCall(method, uri, body string) (int, string, *http.Response, error) {
	/* step: get a member from the cluster */
	if marathon, err := client.cluster.GetMember(); err != nil {
		return 0, "", nil, err
	} else {
		url := fmt.Sprintf("%s/%s", marathon, uri)
		client.log("httpCall(): %s, uri: %s, url: %s", method, uri, url)

		if request, err := http.NewRequest(method, url, strings.NewReader(body)); err != nil {
			return 0, "", nil, err
		} else {
			if client.config.HttpBasicAuthUser != "" {
				request.SetBasicAuth(client.config.HttpBasicAuthUser, client.config.HttpBasicPassword)
			}
			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Accept", "application/json")
			var content string
			/* step: perform the request */
			if response, err := client.http.Do(request); err != nil {
				/* step: mark the endpoint as down */
				client.cluster.MarkDown()
				/* step: retry the request with another endpoint */
				return client.httpCall(method, uri, body)
			} else {
				/* step: lets read in any content */
				client.log("httpCall: %s, uri: %s, url: %s\n", method, uri, url)
				if response.ContentLength != 0 {
					/* step: read in the content from the request */
					response_content, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return response.StatusCode, "", response, err
					}
					content = string(response_content)
				}
				/* step: return the request */
				return response.StatusCode, content, response, nil
			}
		}
	}
	return 0, "", nil, errors.New("Unable to make call to marathon")
}

func (client *Client) log(message string, args ...interface{}) {
	client.logger.Printf(message+"\n", args...)
}
