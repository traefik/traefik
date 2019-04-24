# linodego

[![Build Status](https://travis-ci.org/linode/linodego.svg?branch=master)](https://travis-ci.org/linode/linodego)
[![GoDoc](https://godoc.org/github.com/linode/linodego?status.svg)](https://godoc.org/github.com/linode/linodego)
[![Go Report Card](https://goreportcard.com/badge/github.com/linode/linodego)](https://goreportcard.com/report/github.com/linode/linodego)
[![codecov](https://codecov.io/gh/linode/linodego/branch/master/graph/badge.svg)](https://codecov.io/gh/linode/linodego)

Go client for [Linode REST v4 API](https://developers.linode.com/v4/introduction)

## Installation

```sh
go get -u github.com/linode/linodego
```

## API Support

Check [API_SUPPORT.md](API_SUPPORT.md) for current support of the Linode `v4` API endpoints.

** Note: This project will change and break until we release a v1.0.0 tagged version. Breaking changes in v0.x.x will be denoted with a minor version bump (v0.2.4 -> v0.3.0) **

## Documentation

See [godoc](https://godoc.org/github.com/linode/linodego) for a complete reference.

The API generally follows the naming patterns prescribed in the [OpenAPIv3 document for Linode APIv4](https://developers.linode.com/api/v4).

Deviations in naming have been made to avoid using "Linode" and "Instance" redundantly or inconsistently.

A brief summary of the features offered in this API client are shown here.

## Examples

### General Usage

```go
package main

import (
  "context"
  "fmt"
  "log"
  "os"

  "github.com/linode/linodego"
  "golang.org/x/oauth2"
)

func main() {
  apiKey, ok := os.LookupEnv("LINODE_TOKEN")
  if !ok {
    log.Fatal("Could not find LINODE_TOKEN, please assert it is set.")
  }
  tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})

  oauth2Client := &http.Client{
    Transport: &oauth2.Transport{
      Source: tokenSource,
    },
  }

  linodeClient, err := linodego.NewClient(oauth2Client)
  if err != nil {
    log.Fatal(err)
  }
  linodeClient.SetDebug(true)
  res, err := linodeClient.GetInstance(context.Background(), 4090913)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%v", res)
}
```

### Pagination

#### Auto-Pagination Requests

```go
kernels, err := linodego.ListKernels(context.Background(), nil)
// len(kernels) == 218
```

Or, use a page value of "0":

```go
opts := NewListOptions(0,"")
kernels, err := linodego.ListKernels(context.Background(), opts)
// len(kernels) == 218
```

#### Single Page

```go
opts := NewListOptions(2,"")
// or opts := ListOptions{PageOptions: &PageOptions: {Page: 2 }}
kernels, err := linodego.ListKernels(context.Background(), opts)
// len(kernels) == 100
```

ListOptions are supplied as a pointer because the Pages and Results
values are set in the supplied ListOptions.

```go
// opts.Results == 218
```

#### Filtering

```go
opts := ListOptions{Filter: "{\"mine\":true}"}
// or opts := NewListOptions(0, "{\"mine\":true}")
stackscripts, err := linodego.ListStackscripts(context.Background(), opts)
```

### Error Handling

#### Getting Single Entities

```go
linode, err := linodego.GetLinode(context.Background(), 555) // any Linode ID that does not exist or is not yours
// linode == nil: true
// err.Error() == "[404] Not Found"
// err.Code == "404"
// err.Message == "Not Found"
```

#### Lists

For lists, the list is still returned as `[]`, but `err` works the same way as on the `Get` request.

```go
linodes, err := linodego.ListLinodes(context.Background(), NewListOptions(0, "{\"foo\":bar}"))
// linodes == []
// err.Error() == "[400] [X-Filter] Cannot filter on foo"
```

Otherwise sane requests beyond the last page do not trigger an error, just an empty result:

```go
linodes, err := linodego.ListLinodes(context.Background(), NewListOptions(9999, ""))
// linodes == []
// err = nil
```

### Writes

When performing a `POST` or `PUT` request, multiple field related errors will be returned as a single error, currently like:

```go
// err.Error() == "[400] [field1] foo problem; [field2] bar problem; [field3] baz problem"
```

## Tests

Run `make test` to run the unit tests.  This is the same as running `go test` except that `make test` will
execute the tests while playing back API response fixtures that were recorded during a previous development build.

`go test` can be used without the fixtures. Copy `env.sample` to `.env` and configure your persistent test
settings, including an API token.

`go test -short` can be used to run live API tests that do not require an account token.

This will be simplified in future versions.

To update the test fixtures, run `make fixtures`.  This will record the API responses into the `fixtures/` directory.
Be careful about committing any sensitive account details.  An attempt has been made to sanitize IP addresses and
dates, but no automated sanitization will be performed against `fixtures/*Account*.yaml`, for example.

To prevent disrupting unaffected fixtures, target fixture generation like so: `make ARGS="-run TestListVolumes" fixtures`.

## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

## License

[MIT License](LICENSE)
