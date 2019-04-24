# goacmedns

A Go library to handle [acme-dns](https://github.com/joohoi/acme-dns) client
communication and persistent account storage.

[![Build Status](https://travis-ci.org/cpu/goacmedns.svg?branch=master)](https://travis-ci.org/cpu/goacmedns)
[![Coverage Status](https://coveralls.io/repos/github/cpu/goacmedns/badge.svg?branch=master)](https://coveralls.io/github/cpu/goacmedns?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/cpu/goacmedns)](https://goreportcard.com/report/github.com/cpu/goacmedns)

You may also be interested in a Python equivalent,
[pyacmedns](https://github.com/joohoi/pyacmedns/).

# Installation

Once you have [installed Go](https://golang.org/doc/install) 1.10+ you can
install `goacmedns` with `go get`:

     go get github.com/cpu/goacmedns/...

# Usage

The following is a short example of using the library to update a TXT record
served by an `acme-dns` instance.

```go
package main

import (
	"log"

	"github.com/cpu/goacmedns"
)

const (
	domain = "your.example.org"
)

var (
	whitelistedNetworks = []string{"192.168.11.0/24", "[::1]/128"}
)

func main() {
	// Initialize the client. Point it towards your acme-dns instance.
	client := goacmedns.NewClient("https://auth.acme-dns.io")
	// Initialize the storage. If the file does not exist, it will be
	// automatically created.
	storage := goacmedns.NewFileStorage("/tmp/storage.json", 0600)

	// Check if credentials were previously saved for your domain
	account, err := storage.Fetch(domain)
	if err != nil && err != goacmedns.ErrDomainNotFound {
		log.Fatal(err)
	} else if err == goacmedns.ErrDomainNotFound {
		// The account did not exist. Let's create a new one
		// The whitelisted networks parameter is optional and can be nil
		newAcct, err := client.RegisterAccount(whitelistedNetworks)
		if err != nil {
			log.Fatal(err)
		}
		// Save it
		err = storage.Put(domain, newAcct)
		if err != nil {
			log.Fatalf("Failed to put account in storage: %v", err)
		}
		err = storage.Save()
		if err != nil {
			log.Fatalf("Failed to save storage: %v", err)
		}
		account = newAcct
	}

	// Update the acme-dns TXT record
	err = client.UpdateTXTRecord(account, "___validation_token_recieved_from_the_ca___")
	if err != nil {
		log.Fatal(err)
	}
}
```

# Pre-Registration

When using `goacmedns` with an ACME client hook it may be desirable to do the
initial ACME-DNS account creation and CNAME delegation ahead of time  The
`goacmedns-register` command line utility provides an easy way to do this:

     go install github.com/cpu/goacmedns/...
     goacmedns-register -api http://10.0.0.1:4443 -domain example.com -storage /tmp/example.storage.json

This will register an account for `example.com` with the ACME-DNS server at
`http://10.0.0.1:4443`, saving the account details in
`/tmp/example.storage.json` and printing the required CNAME record for the
`example.com` DNS zone to stdout.
