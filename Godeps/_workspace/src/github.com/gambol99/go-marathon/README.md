[![Build Status](https://travis-ci.org/gambol99/go-marathon.svg?branch=master)](https://travis-ci.org/gambol99/go-marathon)
[![GoDoc](http://godoc.org/github.com/gambol99/go-marathon?status.png)](http://godoc.org/github.com/gambol99/go-marathon)

#### **Go-Marathon**
-----

Go-marathon is a API library for working with [Marathon](https://mesosphere.github.io/marathon/). It currently supports

  > - Application and group deployment
  > - Helper filters for pulling the status, configuration and tasks
  > - Multiple Endpoint support for HA deployments
  > - Marathon Subscriptions and Event callbacks

 Note: the library still under active development; requires >= Go 1.3

#### **Code Examples**
 -------

There is also a examples directory in the source, which show hints and snippets of code of how to use it - which is probably the best place to start.

**Creating a client**

```Go
import (
    "flag"

    marathon "github.com/gambol99/go-marathon"
    "github.com/golang/glog"
    "time"
)

marathon_url := http://10.241.1.71:8080
  config := marathon.NewDefaultConfig()
  config.URL = marathon_url
  config.LogOutput = os.Stdout
  if client, err := marathon.NewClient(config); err != nil {
  	glog.Fatalf("Failed to create a client for marathon, error: %s", err)
  } else {
  	applications, err := client.Applications()
  	...
  ```

> Note, you can also specify multiple endpoint for Marathon (i.e. you have setup Marathon in HA mode and having multiple running)

```Go
marathon := "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
```

The first one specified will be used, if that goes offline the member is marked as *"unavailable"* and a background process will continue to ping the member until it's back online.

**Listing the applications**

```Go
if applications, err := client.Applications(); err != nil
	glog.Errorf("Failed to list applications")
} else {
	glog.Infof("Found %d application running", len(applications.Apps))
	for _, application := range applications.Apps {
		glog.Infof("Application: %s", application)
		details, err := client.Application(application.ID)
		Assert(err)
		if details.Tasks != nil && len(details.Tasks) > 0 {
			for _, task := range details.Tasks {
				glog.Infof("task: %s", task)
			}
			// check the health of the application
			health, err := client.ApplicationOK(details.ID)
			glog.Infof("Application: %s, healthy: %t", details.ID, health)
		}
	}
```

 **Creating a new application**

```Go
glog.Infof("Deploying a new application")
application := marathon.NewDockerApplication()
application.Name("/product/name/frontend")
application.CPU(0.1).Memory(64).Storage(0.0).Count(2)
application.Arg("/usr/sbin/apache2ctl").Arg("-D").Arg("FOREGROUND")
application.AddEnv("NAME", "frontend_http")
application.AddEnv("SERVICE_80_NAME", "test_http")
// add the docker container
application.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80).Expose(443)
application.CheckHTTP("/health", 10, 5)

if _, err := client.CreateApplication(application, true); err != nil {
	glog.Errorf("Failed to create application: %s, error: %s", application, err)
} else {
	glog.Infof("Created the application: %s", application)
}
```

**Scale Application**

Change the number of instance of the application to 4

```Go
glog.Infof("Scale to 4 instances")
if err := client.ScaleApplicationInstances(application.ID, 10); err != nil {
	glog.Errorf("Failed to delete the application: %s, error: %s", application, err)
} else {
	client.WaitOnApplication(application.ID, 0)
	glog.Infof("Successfully scaled the application")
}
```

**Subscription & Events**

Request to listen to events related to applications - namely status updates, health checks changes and failures

```Go
/* step: lets register for events */
update := make(marathon.EventsChannel,5)
if err := client.AddEventsListener(update, marathon.EVENTS_APPLICATIONS); err != nil {
	glog.Fatalf("Failed to register for subscriptions, %s", err)
} else {
	for {
	    event := <-update
	    glog.Infof("EVENT: %s", event )
	}
}

# A full list of the events

const (
    EVENT_API_REQUEST = 1 << iota
    EVENT_STATUS_UPDATE
    EVENT_FRAMEWORK_MESSAGE
    EVENT_SUBSCRIPTION
    EVENT_UNSUBSCRIBED
    EVENT_ADD_HEALTH_CHECK
    EVENT_REMOVE_HEALTH_CHECK
    EVENT_FAILED_HEALTH_CHECK
    EVENT_CHANGED_HEALTH_CHECK
    EVENT_GROUP_CHANGE_SUCCESS
    EVENT_GROUP_CHANGE_FAILED
    EVENT_DEPLOYMENT_SUCCESS
    EVENT_DEPLOYMENT_FAILED
    EVENT_DEPLOYMENT_INFO
    EVENT_DEPLOYMENT_STEP_SUCCESS
    EVENT_DEPLOYMENT_STEP_FAILED
)

const (
    EVENTS_APPLICATIONS  = EVENT_STATUS_UPDATE | EVENT_CHANGED_HEALTH_CHECK | EVENT_FAILED_HEALTH_CHECK
    EVENTS_SUBSCRIPTIONS = EVENT_SUBSCRIPTION | EVENT_UNSUBSCRIBED
)
```

----

#### **Contributing**

 - Fork it
 - Create your feature branch (git checkout -b my-new-feature)
 - Commit your changes (git commit -am 'Add some feature')
 - Push to the branch (git push origin my-new-feature)
 - Create new Pull Request
 - If applicable, update the README.md
