[![Build Status](https://travis-ci.org/gambol99/go-marathon.svg?branch=master)](https://travis-ci.org/gambol99/go-marathon)
[![GoDoc](http://godoc.org/github.com/gambol99/go-marathon?status.png)](http://godoc.org/github.com/gambol99/go-marathon)
[![Go Report Card](https://goreportcard.com/badge/github.com/katallaxie/go-marathon)](https://goreportcard.com/report/github.com/katallaxie/go-marathon)
[![Coverage Status](https://coveralls.io/repos/github/gambol99/go-marathon/badge.svg?branch=master)](https://coveralls.io/github/gambol99/go-marathon?branch=master)

# Go-Marathon

Go-marathon is a API library for working with [Marathon](https://mesosphere.github.io/marathon/).
It currently supports

- Application and group deployment
- Helper filters for pulling the status, configuration and tasks
- Multiple Endpoint support for HA deployments
- Marathon Event Subscriptions and Event Streams
- Pods

Note: the library is still under active development; users should expect frequent (possibly breaking) API changes for the time being.

It requires Go version 1.6 or higher.

## Code Examples

There is also an examples directory in the source which shows hints and snippets of code of how to use it —
which is probably the best place to start.

You can use `examples/docker-compose.yml` in order to start a test cluster.

### Creating a client

```go
import (
	marathon "github.com/gambol99/go-marathon"
)

marathonURL := "http://10.241.1.71:8080"
config := marathon.NewDefaultConfig()
config.URL = marathonURL
client, err := marathon.NewClient(config)
if err != nil {
	log.Fatalf("Failed to create a client for marathon, error: %s", err)
}

applications, err := client.Applications(nil)
...
```

Note, you can also specify multiple endpoint for Marathon (i.e. you have setup Marathon in HA mode and having multiple running)

```go
marathonURL := "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
```

The first one specified will be used, if that goes offline the member is marked as *"unavailable"* and a
background process will continue to ping the member until it's back online.

You can also pass a custom path to the URL, which is especially needed in case of DCOS:

```go
marathonURL := "http://10.241.1.71:8080/cluster,10.241.1.72:8080/cluster,10.241.1.73:8080/cluster"
```

If you specify a `DCOSToken` in the configuration file but do not pass a custom URL path, `/marathon` will be used.

### Customizing the HTTP Clients

HTTP clients with reasonable timeouts are used by default. It is possible to pass custom clients to the configuration though if the behavior should be customized (e.g., to bypass TLS verification, load root CAs, or change timeouts).

Two clients can be given independently of each other:

- `HTTPClient` used only for (non-SSE) HTTP API requests. By default, an http.Client with 10 seconds timeout for the entire request is used.
- `HTTPSSEClient` used only for SSE-based subscription requests. Note that `HTTPSSEClient` cannot have a response read timeout set as this breaks SSE communication; trying to do so will lead to an error during the SSE connection setup. By default, an http.Client with 5 seconds timeout for dial and TLS handshake, and 10 seconds timeout for response headers received is used.

If no `HTTPSSEClient` is given but an `HTTPClient` is, it will be used for SSE subscriptions as well (thereby overriding the default SSE HTTP client).

```go
marathonURL := "http://10.241.1.71:8080"
config := marathon.NewDefaultConfig()
config.URL = marathonURL
config.HTTPClient = &http.Client{
    Timeout: (time.Duration(10) * time.Second),
    Transport: &http.Transport{
        Dial: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 10 * time.Second,
        }).Dial,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    },
}
config.HTTPSSEClient = &http.Client{
    // Invalid to set Timeout as it contains timeout for reading a response body
    Transport: &http.Transport{
        Dial: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 10 * time.Second,
        }).Dial,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    },
}
```

### Listing the applications

```go
applications, err := client.Applications(nil)
if err != nil {
	log.Fatalf("Failed to list applications: %s", err)
}

log.Printf("Found %d application(s) running", len(applications.Apps))
for _, application := range applications.Apps {
	log.Printf("Application: %s", application)
	appID := application.ID

	details, err := client.Application(appID)
	if err != nil {
		log.Fatalf("Failed to get application %s: %s", appID, err)
	}
	if details.Tasks != nil {
		for _, task := range details.Tasks {
			log.Printf("application %s has task: %s", appID, task)
		}
	}
}
```

### Creating a new application

```go
log.Printf("Deploying a new application")
application := marathon.NewDockerApplication().
  Name(applicationName).
  CPU(0.1).
  Memory(64).
  Storage(0.0).
  Count(2).
  AddArgs("/usr/sbin/apache2ctl", "-D", "FOREGROUND").
  AddEnv("NAME", "frontend_http").
  AddEnv("SERVICE_80_NAME", "test_http").
  CheckHTTP("/health", 10, 5)

application.
  Container.Docker.Container("quay.io/gambol99/apache-php:latest").
  Bridged().
  Expose(80).
  Expose(443)

if _, err := client.CreateApplication(application); err != nil {
	log.Fatalf("Failed to create application: %s, error: %s", application, err)
} else {
	log.Printf("Created the application: %s", application)
}
```

Note: Applications may also be defined by means of initializing a `marathon.Application` struct instance directly. However, go-marathon's DSL as shown above provides a more concise way to achieve the same.

### Scaling application

Change the number of application instances to 4

```go
log.Printf("Scale to 4 instances")
if err := client.ScaleApplicationInstances(application.ID, 4); err != nil {
	log.Fatalf("Failed to delete the application: %s, error: %s", application, err)
} else {
	client.WaitOnApplication(application.ID, 30 * time.Second)
	log.Printf("Successfully scaled the application")
}
```

### Pods

Pods allow you to deploy groups of tasks as a unit. All tasks in a single instance of a pod share networking and storage. View the [Marathon documentation](https://mesosphere.github.io/marathon/docs/pods.html) for more details on this feature.

Examples of their usage can be seen in the `examples/pods` directory, and a smaller snippet is below.

```Go
// Initialize a single-container pod running nginx
pod := marathon.NewPod()

image := marathon.NewDockerPodContainerImage().SetID("nginx")

container := marathon.NewPodContainer().
	SetName("container", i).
	CPUs(0.1).
	Memory(128).
	SetImage(image)

pod.Name("mypod").AddContainer(container)

// Create it and wait for it to start up
pod, err := client.CreatePod(pod)
err = client.WaitOnPod(pod.ID, time.Minute*1)

// Scale it
pod.Count(5)
pod, err = client.UpdatePod(pod, true)

// Delete it
id, err := client.DeletePod(pod.ID, true)
```

### Subscription & Events

Request to listen to events related to applications — namely status updates, health checks
changes and failures. There are two different event transports controlled by `EventsTransport`
setting with the following possible values: `EventsTransportSSE` and `EventsTransportCallback` (default value).
See [Event Stream](https://mesosphere.github.io/marathon/docs/rest-api.html#event-stream) and
[Event Subscriptions](https://mesosphere.github.io/marathon/docs/rest-api.html#event-subscriptions) for details.

Event subscriptions can also be individually controlled with the `Subscribe` and `Unsubscribe` functions. See [Controlling subscriptions](#controlling-subscriptions) for more details.

#### Event Stream

Only available in Marathon >= 0.9.0. Does not require any special configuration or prerequisites.

```go
// Configure client
config := marathon.NewDefaultConfig()
config.URL = marathonURL
config.EventsTransport = marathon.EventsTransportSSE

client, err := marathon.NewClient(config)
if err != nil {
	log.Fatalf("Failed to create a client for marathon, error: %s", err)
}

// Register for events
events, err = client.AddEventsListener(marathon.EventIDApplications)
if err != nil {
	log.Fatalf("Failed to register for events, %s", err)
}

timer := time.After(60 * time.Second)
done := false

// Receive events from channel for 60 seconds
for {
	if done {
		break
	}
	select {
	case <-timer:
		log.Printf("Exiting the loop")
		done = true
	case event := <-events:
		log.Printf("Received event: %s", event)
	}
}

// Unsubscribe from Marathon events
client.RemoveEventsListener(events)
```

#### Event Subscriptions

Requires to start a built-in web server accessible by Marathon to connect and push events to. Consider the following
additional settings:

- `EventsInterface` — the interface we should be listening on for events. Default `"eth0"`.
- `EventsPort` — built-in web server port. Default `10001`.
- `CallbackURL` — custom callback URL. Default `""`.

```go
// Configure client
config := marathon.NewDefaultConfig()
config.URL = marathonURL
config.EventsInterface = marathonInterface
config.EventsPort = marathonPort

client, err := marathon.NewClient(config)
if err != nil {
	log.Fatalf("Failed to create a client for marathon, error: %s", err)
}

// Register for events
events, err = client.AddEventsListener(marathon.EventIDApplications)
if err != nil {
	log.Fatalf("Failed to register for events, %s", err)
}

timer := time.After(60 * time.Second)
done := false

// Receive events from channel for 60 seconds
for {
	if done {
		break
	}
	select {
	case <-timer:
		log.Printf("Exiting the loop")
		done = true
	case event := <-events:
		log.Printf("Received event: %s", event)
	}
}

// Unsubscribe from Marathon events
client.RemoveEventsListener(events)
```

See [events.go](events.go) for a full list of event IDs.

#### Controlling subscriptions
If you simply want to (de)register event subscribers (i.e. without starting an internal web server) you can use the `Subscribe` and `Unsubscribe` methods.

```go
// Configure client
config := marathon.NewDefaultConfig()
config.URL = marathonURL

client, err := marathon.NewClient(config)
if err != nil {
	log.Fatalf("Failed to create a client for marathon, error: %s", err)
}

// Register an event subscriber via a callback URL
callbackURL := "http://10.241.1.71:9494"
if err := client.Subscribe(callbackURL); err != nil {
	log.Fatalf("Unable to register the callbackURL [%s], error: %s", callbackURL, err)
}

// Deregister the same subscriber
if err := client.Unsubscribe(callbackURL); err != nil {
	log.Fatalf("Unable to deregister the callbackURL [%s], error: %s", callbackURL, err)
}
```

## Contributing

See the [contribution guidelines](CONTRIBUTING.md).

## Development

### Marathon Fake

go-marathon employs a [fake Marathon implementation](https://github.com/gambol99/go-marathon/blob/master/testing_utils_test.go) for testing purposes. It [maintains a YML-encoded list of HTTP response messages](https://github.com/gambol99/go-marathon/blob/master/tests/rest-api/methods.yml) which are returned upon a successful match based upon a number of attributes, the so-called _message identifier_:

- HTTP URI (without the protocol and the hostname, e.g., `/v2/apps`)
- HTTP method (e.g., `GET`)
- response content (i.e., the message returned)
- scope (see below)

#### Response Content

The response content can be provided in one of two forms:

- **static:** A pure response message returned on every match, including repeated queries.
- **index:** A list of response messages associated to a particular (indexed) sequence order. A message will be returned _iff_ it matches and its zero-based index equals the current request count.

An example for a trivial static response content is

```yaml
- uri: /v2/apps
  method: POST
  content: |
		{
		"app": {
		}
		}
```

which would be returned for every POST request targetting `/v2/apps`.

An indexed response content would look like:

```yaml
- uri: /v2/apps
  method: POST
  contentSequence:
		- index: 1
		- content: |
			{
			"app": {
				"id": "foo"
			}
			}
		- index: 3
		- content: |
			{
			"app": {
				"id": "bar"
			}
			}
```

What this means is that the first POST request to `/v2/apps` would yield a 404, the second one the _foo_ app, the third one 404 again, the fourth one _bar_, and every following request thereafter a 404 again. Indexed responses enable more flexible testing required by some use cases.

Trying to define both a static and indexed response content constitutes an error and leads to `panic`.

#### Scope

By default, all responses are defined globally: Every message can be queried by any request across all tests. This enables reusability and allows to keep the YML definition fairly short. For certain cases, however, it is desirable to define a set of responses that are delivered exclusively for a particular test. Scopes offer a means to do so by representing a concept similar to [namespaces](https://en.wikipedia.org/wiki/Namespace). Combined with indexed responses, they allow to return different responses for message identifiers already defined at the global level.

Scopes do not have a particular format -- they are just strings. A scope must be defined in two places: The message specification and the server configuration. They are pure strings without any particular structure. Given the messages specification

```yaml
- uri: /v2/apps
  method: GET
	# Note: no scope defined.
  content: |
		{
		"app": {
			"id": "foo"
		}
		}
- uri: /v2/apps
  method: GET
	scope: v1.1.1  # This one does have a scope.
  contentSequence:
		- index: 1
		- content: |
			{
			"app": {
				"id": "bar"
			}
			}
```

and the tests

```go
func TestFoo(t * testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)  // No custom configs given.
	defer endpoint.Close()
	app, err := endpoint.Client.Applications(nil)
	// Do something with "foo"
}

func TestFoo(t * testing.T) {
	endpoint := newFakeMarathonEndpoint(t, &configContainer{
		server: &serverConfig{
			scope: "v1.1.1",		// Matches the message spec's scope.
		},
	})
	defer endpoint.Close()
	app, err := endpoint.Client.Applications(nil)
	// Do something with "bar"
}
```

The "foo" response can be used by all tests using the default fake endpoint (such as `TestFoo`), while the "bar" response is only visible by tests that explicitly set the scope to `1.1.1` (as `TestBar` does) and query the endpoint twice.
