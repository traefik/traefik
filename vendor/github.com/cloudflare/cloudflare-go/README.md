# cloudflare-go

[![GoDoc](https://img.shields.io/badge/godoc-reference-5673AF.svg?style=flat-square)](https://godoc.org/github.com/cloudflare/cloudflare-go)
[![Build Status](https://img.shields.io/travis/cloudflare/cloudflare-go/master.svg?style=flat-square)](https://travis-ci.org/cloudflare/cloudflare-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudflare/cloudflare-go?style=flat-square)](https://goreportcard.com/report/github.com/cloudflare/cloudflare-go)

> **Note**: This library is under active development as we expand it to cover
> our (expanding!) API. Consider the public API of this package a little
> unstable as we work towards a v1.0.

A Go library for interacting with
[Cloudflare's API v4](https://api.cloudflare.com/). This library allows you to:

* Manage and automate changes to your DNS records within Cloudflare
* Manage and automate changes to your zones (domains) on Cloudflare, including
  adding new zones to your account
* List and modify the status of WAF (Web Application Firewall) rules for your
  zones
* Fetch Cloudflare's IP ranges for automating your firewall whitelisting

A command-line client, [flarectl](cmd/flarectl), is also available as part of
this project.

## Features

The current feature list includes:

* [x] DNS Records
* [x] Zones
* [x] Web Application Firewall (WAF)
* [x] Cloudflare IPs
* [x] User Administration (partial)
* [x] Virtual DNS Management
* [x] Custom hostnames
* [x] Zone Lockdown and User-Agent Block rules
* [x] Cache purging
* [ ] Organization Administration
* [x] [Railgun](https://www.cloudflare.com/railgun/) administration
* [ ] [Keyless SSL](https://blog.cloudflare.com/keyless-ssl-the-nitty-gritty-technical-details/)
* [x] [Origin CA](https://blog.cloudflare.com/universal-ssl-encryption-all-the-way-to-the-origin-for-free/)
* [x] [Load Balancing](https://blog.cloudflare.com/introducing-load-balancing-intelligent-failover-with-cloudflare/)
* [x] Firewall (partial)
* [x] Rate Limiting

Pull Requests are welcome, but please open an issue (or comment in an existing
issue) to discuss any non-trivial changes before submitting code.

## Installation

You need a working Go environment.

```
go get github.com/cloudflare/cloudflare-go
```

## Getting Started

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	// Construct a new API object
	api, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	if err != nil {
		log.Fatal(err)
	}

	// Fetch user details on the account
	u, err := api.UserDetails()
	if err != nil {
		log.Fatal(err)
	}
	// Print user details
	fmt.Println(u)

	// Fetch the zone ID
	id, err := api.ZoneIDByName("example.com") // Assuming example.com exists in your Cloudflare account already
	if err != nil {
		log.Fatal(err)
	}

	// Fetch zone details
	zone, err := api.ZoneDetails(id)
	if err != nil {
		log.Fatal(err)
	}
	// Print zone details
	fmt.Println(zone)
}
```

Also refer to the
[API documentation](https://godoc.org/github.com/cloudflare/cloudflare-go) for
how to use this package in-depth.

# License

BSD licensed. See the [LICENSE](LICENSE) file for details.
