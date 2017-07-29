go-udp-testing
==============

[![Build Status](https://travis-ci.org/stvp/go-udp-testing.png?branch=master)](https://travis-ci.org/stvp/go-udp-testing)

Provides UDP socket test helpers for Go.

[Documentation](http://godoc.org/github.com/stvp/go-udp-testing)

Examples
--------

```go
package main

import (
  "github.com/stvp/go-udp-testing"
  "testing"
)

func TestStatsdReporting(t *testing.T) {
  udp.SetAddr(":8125")

  udp.ShouldReceiveOnly(t, "mystat:2|g", func() {
    statsd.Gauge("mystat", 2)
  })

  udp.ShouldNotReceiveOnly(t, "mystat:1|c", func() {
    statsd.Gauge("bukkit", 2)
  })

  udp.ShouldReceive(t, "bar:2|g", func() {
    statsd.Gauge("foo", 2)
    statsd.Gauge("bar", 2)
    statsd.Gauge("baz", 2)
  })

  udp.ShouldNotReceive(t, "bar:2|g", func() {
    statsd.Gauge("foo", 2)
    statsd.Gauge("baz", 2)
  })

  expected := []string{
    "bar:2|g",
    "baz:5|g",
  }
  udp.ShouldReceiveAll(t, expected, func() {
    statsd.Gauge("bar", 2)
    statsd.Gauge("baz", 2)
  })

  unexpected := []string{
    "bar",
    "baz",
  }
  udp.ShouldNotReceiveAny(t, unexpected, func() {
    statsd.Gauge("foo", 1)
  })

  expected := []string{ "" }
    "bar:2|g",
    "baz:5|g",
  }
  unexpected := []string{
    "foo",
  }
  udp.ShouldReceiveAllAndNotReceiveAny(t, expected, unexpected, func() {
    statsd.Gauge("bar", 2)
    statsd.Gauge("baz", 5)
  })
}
```

