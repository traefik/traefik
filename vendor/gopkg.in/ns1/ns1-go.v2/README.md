[![Build Status](https://travis-ci.org/ns1/ns1-go.svg?branch=v2)](https://travis-ci.org/ns1/ns1-go) [![GoDoc](https://godoc.org/gopkg.in/ns1/ns1-go.v2?status.svg)](https://godoc.org/gopkg.in/ns1/ns1-go.v2)

# NS1 Golang SDK

The golang client for the NS1 API: https://ns1.com/api/

# Installing

```
$ go get gopkg.in/ns1/ns1-go.v2
```

Examples
========

[See more](https://github.com/ns1/ns1-go/tree/v2/rest/_examples)


```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	api "gopkg.in/ns1/ns1-go.v2/rest"
)

func main() {
	k := os.Getenv("NS1_APIKEY")
	if k == "" {
		fmt.Println("NS1_APIKEY environment variable is not set, giving up")
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(k))

	zones, _, err := client.Zones.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, z := range zones {
		fmt.Println(z.Zone)
	}

}
```

Contributing
============

Contributions, ideas and criticisms are all welcome.

# LICENSE

Apache2 - see the included LICENSE file for more information


