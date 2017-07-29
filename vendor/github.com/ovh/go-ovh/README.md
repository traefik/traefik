go-ovh
======

Lightweight Go wrapper around OVH's APIs. Handles all the hard work including credential creation and requests signing.

[![GoDoc](https://godoc.org/github.com/ovh/go-ovh/go-ovh?status.svg)](http://godoc.org/github.com/ovh/go-ovh/ovh)
[![Build Status](https://travis-ci.org/ovh/go-ovh.svg?branch=master)](https://travis-ci.org/ovh/go-ovh)
[![Coverage Status](https://coveralls.io/repos/github/ovh/go-ovh/badge.svg?branch=master)](https://coveralls.io/github/ovh/go-ovh?branch=master)
[![Go Report Card](http://goreportcard.com/badge/ovh/go-ovh)](http://goreportcard.com/report/ovh/go-ovh)

```go
package main

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
)

// PartialMe holds the first name of the currently logged-in user.
// Visit https://api.ovh.com/console/#/me#GET for the full definition
type PartialMe struct {
	Firstname string `json:"firstname"`
}

// Instantiate an OVH client and get the firstname of the currently logged-in user.
// Visit https://api.ovh.com/createToken/index.cgi?GET=/me to get your credentials.
func main() {
	var me PartialMe

	client, _ := ovh.NewClient(
		"ovh-eu",
		YOUR_APPLICATION_KEY,
		YOUR_APPLICATION_SECRET,
		YOUR_CONSUMER_KEY,
	)
	client.Get("/me", &me)
	fmt.Printf("Welcome %s!\n", me.Firstname)
}
```

## Installation

The Golang wrapper has been tested with Golang 1.5+. It may worker with older versions although it has not been tested.

To use it, just include it to your ``import`` and run ``go get``:

```go
import (
	...
	"github.com/ovh/go-ovh/ovh"
)
```

## Configuration

The straightforward way to use OVH's API keys is to embed them directly in the
application code. While this is very convenient, it lacks of elegance and
flexibility.

Alternatively it is suggested to use configuration files or environment
variables so that the same code may run seamlessly in multiple environments.
Production and development for instance.

This wrapper will first look for direct instanciation parameters then
``OVH_ENDPOINT``, ``OVH_APPLICATION_KEY``, ``OVH_APPLICATION_SECRET`` and
``OVH_CONSUMER_KEY`` environment variables. If either of these parameter is not
provided, it will look for a configuration file of the form:

```ini
[default]
; general configuration: default endpoint
endpoint=ovh-eu

[ovh-eu]
; configuration specific to 'ovh-eu' endpoint
application_key=my_app_key
application_secret=my_application_secret
consumer_key=my_consumer_key
```

Depending on the API you want to use, you may set the ``endpoint`` to:

* ``ovh-eu`` for OVH Europe API
* ``ovh-ca`` for OVH North-America API
* ``soyoustart-eu`` for So you Start Europe API
* ``soyoustart-ca`` for So you Start North America API
* ``kimsufi-eu`` for Kimsufi Europe API
* ``kimsufi-ca`` for Kimsufi North America API
* ``runabove-ca`` for RunAbove API
* Or any arbitrary URL to use in a test for example

The client will successively attempt to locate this configuration file in

1. Current working directory: ``./ovh.conf``
2. Current user's home directory ``~/.ovh.conf``
3. System wide configuration ``/etc/ovh.conf``

This lookup mechanism makes it easy to overload credentials for a specific
project or user.

## Register your app

OVH's API, like most modern APIs is designed to authenticate both an application and
a user, without requiring the user to provide a password. Your application will be
identified by its "application secret" and "application key" tokens.

Hence, to use the API, you must first register your application and then ask your
user to authenticate on a specific URL. Once authenticated, you'll have a valid
"consumer key" which will grant your application on specific APIs.

The user may choose the validity period of its authorization. The default period is
24h. He may also revoke an authorization at any time. Hence, your application should
be prepared to receive 403 HTTP errors and prompt the user to re-authenticated.

This process is detailed in the following section. Alternatively, you may only need
to build an application for a single user. In this case you may generate all
credentials at once. See below.

### Use the API on behalf of a user

Visit [https://eu.api.ovh.com/createApp](https://eu.api.ovh.com/createApp) and create your app
You'll get an application key and an application secret. To use the API you'll need a consumer key.

The consumer key has two types of restriction:

* path: eg. only the ```GET``` method on ```/me```
* time: eg. expire in 1 day


Then, get a consumer key. Here's an example on how to generate one.

First, create a 'ovh.conf' file in the current directory with the application key and
application secret. You can add the consumer key once generated. For alternate
configuration method, please see the [configuration section](#configuration).

```ini
[ovh-eu]
application_key=my_app_key
application_secret=my_application_secret
; consumer_key=my_consumer_key
```

Then, you may use a program like this example to create a consumer key for the application:

```go
package main

import (
	"fmt"

	"github.com/ovh/go-ovh/ovh"
)

func main() {
	// Create a client using credentials from config files or environment variables
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}
	ckReq := client.NewCkRequest()

	// Allow GET method on /me
	ckReq.AddRules(ovh.ReadOnly, "/me")

	// Allow GET method on /xdsl and all its sub routes
	ckReq.AddRecursiveRules(ovh.ReadOnly, "/xdsl")

	// Run the request
	response, err := ckReq.Do()
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	// Print the validation URL and the Consumer key
	fmt.Printf("Generated consumer key: %s\n", response.ConsumerKey)
	fmt.Printf("Please visit %s to validate it\n", response.ValidationURL)
}
```

### Use the API for a single user

Alternatively, you may generate all creadentials at once, including the consumer key. You will
typically want to do this when writing automation scripts for a single projects.

If this case, you may want to directly go to https://eu.api.ovh.com/createToken/ to generate
the 3 tokens at once. Make sure to save them in one of the 'ovh.conf' configuration file.
Please see the [configuration section](#configuration).

``ovh.conf`` should look like:

```ini
[ovh-eu]
application_key=my_app_key
application_secret=my_application_secret
consumer_key=my_consumer_key
```

## Use the lib

These examples assume valid credentials are available in the [configuration](#configuration).

### GET

```go
package main

import (
	"fmt"

	"github.com/ovh/go-ovh/ovh"
)

func main() {
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	// Get all the xdsl services
	xdslServices := []string{}
	if err := client.Get("/xdsl/", &xdslServices); err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	// xdslAccess represents a xdsl access returned by the API
	type xdslAccess struct {
		Name   string `json:"accessName"`
		Status string `json:"status"`
		Pairs  int	`json:"pairsNumber"`
		// Insert the other properties here
	}

	// Get the details of each service
	for i, serviceName := range xdslServices {
		access := xdslAccess{}
		url := "/xdsl/" + serviceName

		if err := client.Get(url, &access); err != nil {
			fmt.Printf("Error: %q\n", err)
			return
		}
		fmt.Printf("#%d : %+v\n", i+1, access)
	}
}
```

### PUT

```go
package main

import (
	"fmt"

	"github.com/ovh/go-ovh/ovh"
)

func main() {
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	// Params
	type AccessPutParams struct {
		Description string `json:"description"`
	}

	// Update the description of the service
	params := &AccessPutParams{Description: "My awesome access"}
	if err := client.Put("/xdsl/xdsl-yourservice", params, nil); err != nil {
		fmt.Printf("Error: %q\n", err)
		return
	}

	fmt.Println("Description updated")
}
```

## API Documentation

### Create a client

- Use ``ovh.NewClient()`` to have full controll over ther authentication
- Use ``ovh.NewEndpointClient()`` to create a client for a specific API and use credentials from config files or environment
- Use ``ovh.NewDefaultClient()`` to create a client unsing endpoint and credentials from config files or environment

### Query

Each HTTP verb has its own Client method. Some API methods supports unauthenticated calls. For
these methods, you may want to use the ``*UnAuth`` variant of the Client which will bypass
request signature.

Each helper accepts a ``method`` and ``resType`` argument. ``method`` is the full URI, including
the query string, and ``resType`` is a reference to an object in which the json response will
be unserialized.

Additionally, ``Post``, ``Put`` and their ``UnAuth`` variant accept a reqBody which is a
reference to a json serializable object or nil.

Alternatively, you may directly use the low level ``CallAPI`` method.

- Use ``client.Get()`` for GET requests
- Use ``client.Post()`` for POST requests
- Use ``client.Put()`` for PUT requests
- Use ``client.Delete()`` for DELETE requests

Or, for unautenticated requests:

- Use ``client.GetUnAuth()`` for GET requests
- Use ``client.PostUnAuth()`` for POST requests
- Use ``client.PutUnAuth()`` for PUT requests
- Use ``client.DeleteUnAuth()`` for DELETE requests

### Request consumer keys

Consumer keys may be restricted to a subset of the API. This allows to delegate the API to manage
only a specific server or domain name for example. This is called "scoping" a consumer key.

Rules are simple. They combine an HTTP verb (GET, POST, PUT or DELETE) with a pattern. A pattern
is a plain API method and may contain the '*' wilcard to match "anything". Just like glob on a
Unix machine.

While this is simple and may be managed directly with the API as-is, this can be cumbersome to do
and we recommend using the ``CkRequest`` helper. It basically manages the list of authorizations
for you and the actual request.

*example*: Grant on all /sms and identity
```go
client, err := ovh.NewEndpointClient("ovh-eu")
if err == nil {
    // Do something
}
req := client.NewCkRequest()
req.AddRules(ovh.ReadOnly, "/me")
req.AddRecursiveRulesRules(ovh.ReadWrite, "/sms")
pendingCk, err := req.Do()
```

This example will generate a request for:

- GET /me
- GET /sms
- GET /sms/*
- POST /sms
- POST /sms/*
- PUT /sms
- PUT /sms/*
- DELETE /sms
- DELETE /sms/*

Which would be tedious to do by hand...

*Create a ``CkRequest``*:

```go
req := client.NewCkRequest()
```

*Request access on a specific path and method* (advanced):
```go
// Use this method for fine-grain access control. In most case, you'll
// want to use the methods below.
req.AddRule("VERB", "PATTERN")
```

*Request access on specific path*:
```go
// This will generate all patterns for GET PATH
req.AddRules(ovh.ReadOnly, "/PATH")

// This will generate all patterns for PATH for all HTTP verbs
req.AddRules(ovh.ReadWrite, "/PATH")

// This will generate all patterns for PATH for all HTTP verbs, except DELETE
req.AddRules(ovh.ReadWriteSafe, "/PATH")
```

*Request access on path and all sub-path*:
```go
// This will generate all patterns for GET PATH
req.AddRecursiveRules(ovh.ReadOnly, "/PATH")

// This will generate all patterns for PATH for all HTTP verbs
req.AddRecursiveRules(ovh.ReadWrite, "/PATH")

// This will generate all patterns for PATH for all HTTP verbs, except DELETE
req.AddRecusriveRules(ovh.ReadWriteSafe, "/PATH")
```

*Create key*:

```go
pendingCk, err := req.Do()
```

This will initiate the consumer key validation process and return both a consumer key and
a validation URL. The consumer key is automatically added to the client which was used to
create the request. It may be used as soon as the user has authenticated the request on the
validation URL.


``pendingCk`` contains 3 fields:
- ``ValidationURL`` the URL the user needs to visit to activate the consumer key
- ``ConsumerKey`` the new consumer key. It won't be active until validation
- ``State`` the consumer key state. Always "pendingValidation" at this stage


## Hacking

This wrapper uses standard Go tools, so you should feel at home with it.
Here is a quick outline of what it may look like.

### Get the sources

```
go get github.com/ovh/go-ovh/ovh
cd $GOPATH/src/github.com/ovh/go-ovh/ovh
go get
```

You've developed a new cool feature ? Fixed an annoying bug ? We'd be happy
to hear from you ! See [CONTRIBUTING.md](https://github.com/ovh/go-ovh/blob/master/CONTRIBUTING.md)
for more informations

### Run the tests

Simply run ``go test``. Since we all love quality, please
note that we do not accept contributions lowering coverage.

```
# Run all tests, with coverage
go test -cover

# Validate code quality
golint ./...
go vet ./...
```

## Supported APIs

### OVH Europe

- **Documentation**: https://eu.api.ovh.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://eu.api.ovh.com/console
- **Create application credentials**: https://eu.api.ovh.com/createApp/
- **Create script credentials** (all keys at once): https://eu.api.ovh.com/createToken/

### OVH North America

- **Documentation**: https://ca.api.ovh.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://ca.api.ovh.com/console
- **Create application credentials**: https://ca.api.ovh.com/createApp/
- **Create script credentials** (all keys at once): https://ca.api.ovh.com/createToken/

### So you Start Europe

- **Documentation**: https://eu.api.soyoustart.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://eu.api.soyoustart.com/console/
- **Create application credentials**: https://eu.api.soyoustart.com/createApp/
- **Create script credentials** (all keys at once): https://eu.api.soyoustart.com/createToken/

### So you Start North America

- **Documentation**: https://ca.api.soyoustart.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://ca.api.soyoustart.com/console/
- **Create application credentials**: https://ca.api.soyoustart.com/createApp/
- **Create script credentials** (all keys at once): https://ca.api.soyoustart.com/createToken/

### Kimsufi Europe

- **Documentation**: https://eu.api.kimsufi.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://eu.api.kimsufi.com/console/
- **Create application credentials**: https://eu.api.kimsufi.com/createApp/
- **Create script credentials** (all keys at once): https://eu.api.kimsufi.com/createToken/

### Kimsufi North America

- **Documentation**: https://ca.api.kimsufi.com/
- **Community support**: api-subscribe@ml.ovh.net
- **Console**: https://ca.api.kimsufi.com/console/
- **Create application credentials**: https://ca.api.kimsufi.com/createApp/
- **Create script credentials** (all keys at once): https://ca.api.kimsufi.com/createToken/

### Runabove

- **Community support**: https://community.runabove.com/
- **Console**: https://api.runabove.com/console/
- **Create application credentials**: https://api.runabove.com/createApp/
- **High level SDK**: https://github.com/runabove/python-runabove

## License

3-Clause BSD

