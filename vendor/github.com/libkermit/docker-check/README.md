# Libkermit
[![GoDoc](https://godoc.org/github.com/libkermit/docker-check?status.png)](https://godoc.org/github.com/libkermit/docker-check)
[![Build Status](https://travis-ci.org/libkermit/docker-check.svg?branch=master)](https://travis-ci.org/libkermit/docker-check)
[![Go Report Card](https://goreportcard.com/badge/github.com/libkermit/docker-check)](https://goreportcard.com/report/github.com/libkermit/docker-check)
[![License](https://img.shields.io/github/license/libkermit/docker-check.svg)]()

When `libkermit` meet with `go-check`.

**Note: This is experimental and not even implemented yet. You are on your own right now**


## Package `docker`

This package map the `docker` package but takes a `*check.C` struct
on all methods. The idea is to write even less. Let's write the same
example as above.


```go
package yours

import (
    "testing"

	"github.com/go-check/check"
    docker "github.com/libkermit/docker-check"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type CheckSuite struct{}

var _ = check.Suite(&CheckSuite{})

func (s *CheckSuite) TestItMyFriend(c *check.C) {
    project := docker.NewProjectFromEnv(c)
    container := project.Start(c, "vdemeester/myawesomeimage")

    // Do your stuff
    // […]

    // Clean the containers managed by libkermit
    project.Clean(c)
}
```

