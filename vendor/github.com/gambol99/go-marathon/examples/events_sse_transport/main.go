/*
Copyright 2015 Denis Parchenko All rights reserved.

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

package main

import (
	"flag"
	"log"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

var marathonURL string
var timeout int

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080", "the url for the Marathon endpoint")
	flag.IntVar(&timeout, "timeout", 60, "listen to events for x seconds")
}

func assert(err error) {
	if err != nil {
		log.Fatalf("Failed, error: %s", err)
	}
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	config.EventsTransport = marathon.EventsTransportSSE
	log.Printf("Creating a client, Marathon: %s", marathonURL)

	client, err := marathon.NewClient(config)
	assert(err)

	// Register for events
	events, err := client.AddEventsListener(marathon.EventIDApplications)
	assert(err)
	deployments, err := client.AddEventsListener(marathon.EventIDDeploymentStepSuccess)
	assert(err)

	// Listen for x seconds and then split
	timer := time.After(time.Duration(timeout) * time.Second)
	done := false
	for {
		if done {
			break
		}
		select {
		case <-timer:
			log.Printf("Exiting the loop")
			done = true
		case event := <-events:
			log.Printf("Recieved application event: %s", event)
		case event := <-deployments:
			log.Printf("Recieved deployment event: %v", event)
			var deployment *marathon.EventDeploymentStepSuccess
			deployment = event.Event.(*marathon.EventDeploymentStepSuccess)
			log.Printf("deployment step: %v", deployment.CurrentStep)
		}
	}

	log.Printf("Removing our subscription")
	client.RemoveEventsListener(events)
	client.RemoveEventsListener(deployments)
}
