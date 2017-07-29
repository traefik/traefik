# Contributing to DNSimple/Go

## Getting started

Clone the repository [in your workspace](https://golang.org/doc/code.html#Organization) and move into it:

```
$ mkdir -p $GOPATH/src/github.com/dnsimple && cd $_
$ git clone git@github.com:dnsimple/dnsimple-go.git
$ cd dnsimple-go
```

[Run the test suite](#testing) to check everything works as expected.


## Testing

To run the test suite:

```shell
$ go test ./... -v
```

### Live Testing

```shell
$ export DNSIMPLE_TOKEN="some-token"
$ go test ./... -v
```


## Tests

Submit unit tests for your changes. You can test your changes on your machine by [running the test suite](#testing).

When you submit a PR, tests will also be run on the continuous integration environment [through Travis](https://travis-ci.org/dnsimple/dnsimple-go).

