# gotransip - TransIP API client for Go
[![Go Report Card](https://goreportcard.com/badge/github.com/transip/gotransip)](https://goreportcard.com/report/github.com/transip/gotransip) [![Documentation](https://godoc.org/github.com/transip/gotransip?status.svg)](http://godoc.org/github.com/transip/gotransip)

This is the Go client for the [TransIP API](https://api.transip.nl/). To use it you need an account with [TransIP](https://transip.nl/), enable API usage and setup a private API key.

**NOTE**: While the TransIP API's PHP client code is automatically generated, the Go client isn't. We try to follow the current PHP version as close as possible, but if something is not working 100% like you'd expect, please open an issue and of course: you're welcome to [contribute](CONTRIBUTING.md)!

## Example usage
To print a list of your account's VPSes:
```golang
package main

import (
  "fmt"

  "github.com/transip/gotransip"
  "github.com/transip/gotransip/vps"
)

func main() {
  // create new TransIP API SOAP client
  c, err := gotransip.NewSOAPClient(gotransip.ClientConfig{
    AccountName: "accountname",
    PrivateKeyPath:  "/path/to/api/private.key"
  })
  if err != nil {
    panic(err.Error())
  }

  // get list of VPSes
  list, err := vps.GetVpses(c)
  if err != nil {
    panic(err.Error())
  }

  // print name and description for each VPS
  for _, v := range list {
    fmt.Printf("vps: %s (%s)\n", v.Name, v.Description)
  }
}
```

## Documentation
For detailed descriptions of all functions, check out the [TransIP API documentation](https://api.transip.nl/). Details about the usage of the Go client can be found on [godoc.org](https://godoc.org/github.com/transip/gotransip).
