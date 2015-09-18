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
	"os"
	"time"

	marathon "github.com/gambol99/go-marathon"

	"github.com/golang/glog"
)

var marathon_url string

func init() {
	flag.StringVar(&marathon_url, "url", "http://127.0.0.1:8080", "the url for the marathon endpoint")
}

func Assert(err error) {
	if err != nil {
		glog.Fatalf("Failed, error: %s", err)
	}
}

func waitOnDeployment(client marathon.Marathon, id *marathon.DeploymentID) {
	Assert(client.WaitOnDeployment(id.DeploymentID, 0))
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathon_url
	config.LogOutput = os.Stdout
	client, err := marathon.NewClient(config)
	Assert(err)
	applications, err := client.Applications(nil)
	Assert(err)

	glog.Infof("Found %d application running", len(applications.Apps))
	for _, application := range applications.Apps {
		glog.Infof("Application: %s", application)
		details, err := client.Application(application.ID)
		Assert(err)
		if details.Tasks != nil && len(details.Tasks) > 0 {
			for _, task := range details.Tasks {
				glog.Infof("task: %s", task)
			}
			health, err := client.ApplicationOK(details.ID)
			Assert(err)
			glog.Infof("Application: %s, healthy: %t", details.ID, health)
		}
	}

	APPLICATION_NAME := "/my/product"

	if found, _ := client.HasApplication(APPLICATION_NAME); found {
		deployId, err := client.DeleteApplication(APPLICATION_NAME)
		Assert(err)
		waitOnDeployment(client, deployId)
	}

	glog.Infof("Deploying a new application")
	application := marathon.NewDockerApplication()
	application.Name(APPLICATION_NAME)
	application.CPU(0.1).Memory(64).Storage(0.0).Count(2)
	application.Arg("/usr/sbin/apache2ctl").Arg("-D").Arg("FOREGROUND")
	application.AddEnv("NAME", "frontend_http")
	application.AddEnv("SERVICE_80_NAME", "test_http")
	application.RequirePorts = true
	application.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80).Expose(443)
	err, _ = client.CreateApplication(application, true)
	Assert(err)

	glog.Infof("Scaling the application to 4 instances")
	deployId, err := client.ScaleApplicationInstances(application.ID, 4)
	Assert(err)
	client.WaitOnApplication(application.ID, 0)
	glog.Infof("Successfully scaled the application, deployId: %s", deployId.DeploymentID)

	glog.Infof("Deleting the application: %s", APPLICATION_NAME)
	deployId, err = client.DeleteApplication(application.ID)
	Assert(err)
	time.Sleep(time.Duration(10) * time.Second)
	glog.Infof("Successfully deleted the application")

	glog.Infof("Starting the application again")
	err, _ = client.CreateApplication(application, true)
	Assert(err)
	glog.Infof("Created the application: %s", application.ID)

	glog.Infof("Delete all the tasks")
	_, err = client.KillApplicationTasks(application.ID, "", false)
	Assert(err)
}
