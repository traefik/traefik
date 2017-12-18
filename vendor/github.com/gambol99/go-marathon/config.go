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
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const defaultPollingWaitTime = 500 * time.Millisecond

const defaultDCOSPath = "marathon"

// EventsTransport describes which transport should be used to deliver Marathon events
type EventsTransport int

// Config holds the settings and options for the client
type Config struct {
	// URL is the url for marathon
	URL string
	// EventsTransport is the events transport: EventsTransportCallback or EventsTransportSSE
	EventsTransport EventsTransport
	// EventsPort is the event handler port
	EventsPort int
	// the interface we should be listening on for events
	EventsInterface string
	// HTTPBasicAuthUser is the http basic auth
	HTTPBasicAuthUser string
	// HTTPBasicPassword is the http basic password
	HTTPBasicPassword string
	// CallbackURL custom callback url
	CallbackURL string
	// DCOSToken for DCOS environment, This will override the Authorization header
	DCOSToken string
	// LogOutput the output for debug log messages
	LogOutput io.Writer
	// HTTPClient is the HTTP client
	HTTPClient *http.Client
	// HTTPSSEClient is the HTTP client used for SSE subscriptions, can't have client.Timeout set
	HTTPSSEClient *http.Client
	// wait time (in milliseconds) between repetitive requests to the API during polling
	PollingWaitTime time.Duration
}

// NewDefaultConfig create a default client config
func NewDefaultConfig() Config {
	return Config{
		URL:             "http://127.0.0.1:8080",
		EventsTransport: EventsTransportCallback,
		EventsPort:      10001,
		EventsInterface: "eth0",
		LogOutput:       ioutil.Discard,
		PollingWaitTime: defaultPollingWaitTime,
	}
}
