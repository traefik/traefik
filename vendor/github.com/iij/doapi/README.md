# Golang binding for DO API

DO is IIJ DNS outsource service.

## Install

- go get -u github.com/iij/doapi

# Usage for Golang users

```go
package main

// Usage:
//   export IIJAPI_ACCESS_KEY=<YOUR ACCESSS KEY>
//   export IIJAPI_SECRET_KEY=<YOUR SECRET KEY>
//   export DOSERVICECODE=<YOUR DO CODE>

import (
	"log"
	"os"

	"github.com/iij/doapi"
	"github.com/iij/doapi/protocol"
)

func main() {
	api := doapi.NewAPI(os.Getenv("IIJAPI_ACCESS_KEY"), os.Getenv("IIJAPI_SECRET_KEY"))

    // List zones
	request := protocol.ZoneListGet{ DoServiceCode: os.Getenv("DOSERVICECODE"), }
	response := protocol.ZoneListGetResponse{}
	if err := doapi.Call(*api, request, &response); err == nil {
		for _, zone := range response.ZoneList { 
			log.Println("zone", zone)
		}
	}
}
```
