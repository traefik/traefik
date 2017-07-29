Aurora DNS API client
=====================

[![Build Status](https://img.shields.io/travis/edeckers/auroradns_client/master.svg?maxAge=2592000&style=flat-square)](https://travis-ci.org/edeckers/auroradns_client)
[![License](https://img.shields.io/github/license/edeckers/auroradns_client.svg?maxAge=2592000&style=flat-square)](https://www.gnu.org/licenses/gpl-3.0.en.html)

An wrapper library for the Aurora DNS API, written in Go.

## Features

* List zones and records
* Add and remove records

## Requirements

* Go >= 1.6

## Build

```bash
make deps
make build
```

## Test

```bash
make test
```

## Basic usage

```go
client, _ := NewAuroraDNSClient(fakeAuroraEndpoint, fakeAuroraDNSUserId, fakeAuroraDNSKey)

zones, err := client.GetZones()
```

## License

`auroradns_client` is licensed under GPL - see the LICENSE file for details
