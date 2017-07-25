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
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/donovanhide/eventsource"
	yaml "gopkg.in/yaml.v2"
)

const (
	fakeMarathonURL         = "http://127.0.0.1:3000,127.0.0.1:3000,127.0.0.1:3000"
	fakeMarathonURLWithPath = "http://127.0.0.1:3000/path,127.0.0.1:3000/path,127.0.0.1:3000/path"
	fakeGroupName           = "/test"
	fakeGroupName1          = "/qa/product/1"
	fakeAppName             = "/fake-app"
	fakeTaskID              = "fake-app.fake-task"
	fakeAppNameBroken       = "/fake-app-broken"
	fakeDeploymentID        = "867ed450-f6a8-4d33-9b0e-e11c5513990b"
	fakeAppNameUnhealthy    = "/no-health-check-results-app"
)

var (
	fakeResponses map[string][]indexedResponse
	once          sync.Once
)

type indexedResponse struct {
	Index   int    `yaml:"index,omitempty"`
	Content string `yaml:"content,omitempty"`
}

type responseIndices struct {
	sync.Mutex
	m map[string]int
}

func newResponseIndices() *responseIndices {
	return &responseIndices{m: map[string]int{}}
}

// restMethod represents an expected HTTP method and an associated fake response
type restMethod struct {
	// the uri of the method
	URI string `yaml:"uri,omitempty"`
	// the http method type (GET|PUT etc)
	Method string `yaml:"method,omitempty"`
	// the content i.e. response
	Content string `yaml:"content,omitempty"`
	// ContentSequence is a sequence of responses that are returned in order.
	ContentSequence []indexedResponse `yaml:"contentSequence,omitempty"`
	// the test scope
	Scope string `yaml:"scope,omitempty"`
}

// serverConfig holds the Marathon server configuration
type serverConfig struct {
	// Username for basic auth
	username string
	// Password for basic auth
	password string
	// Token for authorization in case of DCOS environment
	dcosToken string
	// scope is an arbitrary test scope to distinguish fake responses from
	// otherwise equal HTTP methods and query strings.
	scope string
}

// configContainer holds both server and client Marathon configuration
type configContainer struct {
	client *Config
	server *serverConfig
}

type fakeServer struct {
	io.Closer

	eventSrv        *eventsource.Server
	httpSrv         *httptest.Server
	fakeRespIndices *responseIndices
}

type endpoint struct {
	io.Closer

	Server fakeServer
	Client Marathon
	URL    string
}

type fakeEvent struct {
	data string
}

func getTestURL(urlString string) string {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL '%s': %s", urlString, err))
	}
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, strings.Join([]string{parsedURL.Host, parsedURL.Host, parsedURL.Host}, ","))
}

func newFakeMarathonEndpoint(t *testing.T, configs *configContainer) *endpoint {
	// step: read in the fake responses if required
	initFakeMarathonResponses(t)

	// step: create a fake SSE event service
	eventSrv := eventsource.NewServer()

	// step: fill in the default if required
	defaultConfig := NewDefaultConfig()
	if configs == nil {
		configs = &configContainer{}
	}
	if configs.client == nil {
		configs.client = &defaultConfig
	}
	if configs.server == nil {
		configs.server = &serverConfig{}
	}

	fakeRespIndices := newResponseIndices()

	// step: create the HTTP router
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/events", authMiddleware(configs.server, eventSrv.Handler("event")))
	mux.HandleFunc("/", authMiddleware(configs.server, func(writer http.ResponseWriter, reader *http.Request) {
		respKey := fakeResponseMapKey(reader.Method, reader.RequestURI, configs.server.scope)
		fakeRespIndices.Lock()
		fakeRespIndex := fakeRespIndices.m[respKey]
		fakeRespIndices.m[respKey]++
		responses, found := fakeResponses[respKey]
		fakeRespIndices.Unlock()
		if found {
			for _, response := range responses {
				// Index < 0 indicates a static response.
				if response.Index < 0 || response.Index == fakeRespIndex {
					writer.Header().Add("Content-Type", "application/json")
					writer.Write([]byte(response.Content))
					return
				}
			}
		}

		http.Error(writer, `{"message": "not found"}`, 404)
	}))

	// step: create HTTP test server
	httpSrv := httptest.NewServer(mux)

	if configs.client.URL == defaultConfig.URL {
		configs.client.URL = getTestURL(httpSrv.URL)
	}

	// step: create the client for the service
	client, err := NewClient(*configs.client)
	if err != nil {
		t.Fatalf("Failed to create the fake client, %s, error: %s", configs.client.URL, err)
	}

	return &endpoint{
		Server: fakeServer{
			eventSrv:        eventSrv,
			httpSrv:         httpSrv,
			fakeRespIndices: fakeRespIndices,
		},
		Client: client,
		URL:    configs.client.URL,
	}
}

// basicAuthMiddleware handles basic auth
func basicAuthMiddleware(server *serverConfig, next http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	unauthorized := `{"message": "invalid username or password"}`

	return func(w http.ResponseWriter, r *http.Request) {
		// step: is authentication required?
		if server.username != "" && server.password != "" {
			u, p, found := r.BasicAuth()
			// step: if no auth found, error it
			if !found {
				http.Error(w, unauthorized, 401)
				return
			}
			// step: if username and password don't match, error it
			if server.username != u || server.password != p {
				http.Error(w, unauthorized, 401)
				return
			}
		}

		next(w, r)
	}
}

// authMiddleware handles basic auth and dcos_acs_token
func authMiddleware(server *serverConfig, next http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	unauthorized := `{"message": "invalid username or password"}`

	return func(w http.ResponseWriter, r *http.Request) {
		// step: is authentication required?

		if server.dcosToken != "" {
			headerValue := r.Header.Get("Authorization")
			// step: if no auth found, error it
			if headerValue == "" {
				http.Error(w, unauthorized, 401)
				return
			}

			s := strings.Split(headerValue, "=")

			if s[1] != server.dcosToken {
				http.Error(w, unauthorized, 401)
				return
			}
		} else if server.username != "" && server.password != "" {
			u, p, found := r.BasicAuth()
			// step: if no auth found, error it
			if !found {
				http.Error(w, unauthorized, 401)
				return
			}
			// step: if username and password don't match, error it
			if server.username != u || server.password != p {
				http.Error(w, unauthorized, 401)
				return
			}
		}

		next(w, r)
	}
}

// initFakeMarathonResponses reads in the marathon fake responses from the yaml file
func initFakeMarathonResponses(t *testing.T) {
	once.Do(func() {
		fakeResponses = make(map[string][]indexedResponse, 0)
		var methods []*restMethod

		// step: read in the test method specification
		methodSpec, err := ioutil.ReadFile("./tests/rest-api/methods.yml")
		if err != nil {
			t.Fatalf("failed to read in the fake yaml responses: %s", err)
		}

		if err = yaml.Unmarshal([]byte(methodSpec), &methods); err != nil {
			t.Fatalf("failed to unmarshal the response: %s", err)
		}
		for _, method := range methods {
			key := fakeResponseMapKey(method.Method, method.URI, method.Scope)
			switch {
			case method.Content != "" && len(method.ContentSequence) > 0:
				panic("content and contentSequence must not be provided simultaneously")
			case len(method.ContentSequence) > 0:
				fakeResponses[key] = method.ContentSequence
			default:
				// This combines the cases where static content was defined or not. The
				// latter models an empty response (via an empty content) that should
				// not result into a 404.
				fakeResponses[key] = []indexedResponse{
					indexedResponse{
						// Index -1 indicates a static response.
						Index:   -1,
						Content: method.Content,
					},
				}
			}
		}
	})
}

func fakeResponseMapKey(method, uri, scope string) string {
	return fmt.Sprintf("%s:%s:%s", method, uri, scope)
}

func (t fakeEvent) Id() string {
	return "0"
}

func (t fakeEvent) Event() string {
	return "MarathonEvent"
}

func (t fakeEvent) Data() string {
	return t.data
}

func (s *fakeServer) PublishEvent(event string) {
	s.eventSrv.Publish([]string{"event"}, fakeEvent{event})
}

func (s *fakeServer) Close() error {
	s.eventSrv.Close()
	s.httpSrv.Close()
	return nil
}

func (e *endpoint) Close() error {
	return e.Server.Close()
}
