# Libkermit
[![GoDoc](https://godoc.org/github.com/libkermit/compose?status.png)](https://godoc.org/github.com/libkermit/compose)
[![Build Status](https://travis-ci.org/libkermit/compose.svg?branch=master)](https://travis-ci.org/libkermit/compose)
[![Go Report Card](https://goreportcard.com/badge/github.com/libkermit/compose)](https://goreportcard.com/report/github.com/libkermit/compose)
[![License](https://img.shields.io/github/license/libkermit/compose.svg)]()
[![codecov](https://codecov.io/gh/libkermit/compose/branch/master/graph/badge.svg)](https://codecov.io/gh/libkermit/compose)

When `libermit` meet with `libcompose`.

**Note: This is experimental and not even implemented yet. You are on your own right now**


## Package `compose`

This package holds functions and structs to ease docker uses.

```go
package yours

import (
    "testing"

    "github.com/libkermit/docker/compose"
)

func TestItMyFriend(t *testing.T) {
    project, err := compose.CreateProject("simple", "./assets/simple.yml")
    if err != nil {
        t.Fatal(err)
    }
    err = project.Start()
	if err != nil {
		t.Fatal(err)
	}

    // Do your stuff

    err = project.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
```

### Package `compose/testing`

This package map the `compose` package but takes a `*testing.T` struct
on all methods. The idea is to write even less. Let's write the same
example as above.


```go
package yours

import (
    "testing"

    docker "github.com/libkermit/docker/compose/testing"
)

func TestItMyFriend(t *testing.T) {
    project := compose.CreateProject(t, "simple", "./assets/simple.yml")
    project.Start(t)

    // Do your stuff

    project.Stop(t)
}
```


## Other packages to come

- `suite` : functions and structs to setup tests suites.


