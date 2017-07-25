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

package main

import (
	"flag"
	"log"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

var marathonURL string

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080", "the url for the marathon endpoint")
}

func assert(err error) {
	if err != nil {
		log.Fatalf("Failed, error: %s", err)
	}
}

func waitOnDeployment(client marathon.Marathon, id *marathon.DeploymentID) {
	assert(client.WaitOnDeployment(id.DeploymentID, 0))
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	client, err := marathon.NewClient(config)
	assert(err)
	applications, err := client.Applications(nil)
	assert(err)

	log.Printf("Found %d application running", len(applications.Apps))
	for _, application := range applications.Apps {
		log.Printf("Application: %v", application)
		details, err := client.Application(application.ID)
		assert(err)
		if details.Tasks != nil && len(details.Tasks) > 0 {
			for _, task := range details.Tasks {
				log.Printf("task: %v", task)
			}
			health, err := client.ApplicationOK(details.ID)
			assert(err)
			log.Printf("Application: %s, healthy: %t", details.ID, health)
		}
	}

	applicationName := "/my/product"

	if _, err := client.Application(applicationName); err == nil {
		deployID, err := client.DeleteApplication(applicationName, false)
		assert(err)
		waitOnDeployment(client, deployID)
	}

	log.Printf("Deploying a new application")
	application := marathon.NewDockerApplication().
		Name(applicationName).
		CPU(0.1).
		Memory(64).
		Storage(0.0).
		Count(2).
		AddArgs("/usr/sbin/apache2ctl", "-D", "FOREGROUND").
		AddEnv("NAME", "frontend_http").
		AddEnv("SERVICE_80_NAME", "test_http")

	application.
		Container.Docker.Container("quay.io/gambol99/apache-php:latest").
		Bridged().
		Expose(80).
		Expose(443)

	*application.RequirePorts = true
	_, err = client.CreateApplication(application)
	assert(err)
	client.WaitOnApplication(application.ID, 30*time.Second)

	log.Printf("Scaling the application to 4 instances")
	deployID, err := client.ScaleApplicationInstances(application.ID, 4, false)
	assert(err)
	client.WaitOnApplication(application.ID, 30*time.Second)
	log.Printf("Successfully scaled the application, deployID: %s", deployID.DeploymentID)

	log.Printf("Deleting the application: %s", applicationName)
	deployID, err = client.DeleteApplication(application.ID, true)
	assert(err)
	waitOnDeployment(client, deployID)
	log.Printf("Successfully deleted the application")

	log.Printf("Starting the application again")
	_, err = client.CreateApplication(application)
	assert(err)
	log.Printf("Created the application: %s", application.ID)

	log.Printf("Delete all the tasks")
	_, err = client.KillApplicationTasks(application.ID, nil)
	assert(err)
}
