# zipkin-go-opentracing

[![Travis CI](https://travis-ci.org/openzipkin-contrib/zipkin-go-opentracing.svg?branch=master)](https://travis-ci.org/openzipkin-contrib/zipkin-go-opentracing)
[![GoDoc](https://godoc.org/github.com/openzipkin-contrib/zipkin-go-opentracing?status.svg)](https://godoc.org/github.com/openzipkin-contrib/zipkin-go-opentracing)
[![Go Report Card](https://goreportcard.com/badge/github.com/openzipkin-contrib/zipkin-go-opentracing)](https://goreportcard.com/report/github.com/openzipkin-contrib/zipkin-go-opentracing)
[![Sourcegraph](https://sourcegraph.com/github.com/openzipkin-contrib/zipkin-go-opentracing/-/badge.svg)](https://sourcegraph.com/github.com/openzipkin-contrib/zipkin-go-opentracing?badge)

[OpenTracing](http://opentracing.io) Tracer implementation for [Zipkin](http://zipkin.io) in Go.

### Notes

This package is a low level tracing "driver" to allow OpenTracing API consumers
to use Zipkin as their tracing backend. For details on how to work with spans
and traces we suggest looking at the documentation and README from the
[OpenTracing API](https://github.com/opentracing/opentracing-go).

For developers interested in adding Zipkin tracing to their Go services we
suggest looking at [Go kit](https://gokit.io) which is an excellent toolkit to
instrument your distributed system with Zipkin and much more with clean
separation of domains like transport, middleware / instrumentation and
business logic.

### Examples

For more information on zipkin-go-opentracing, please see the
[examples](https://github.com/openzipkin-contrib/zipkin-go-opentracing/tree/master/examples)
directory for usage examples as well as documentation at
[go doc](https://godoc.org/github.com/openzipkin-contrib/zipkin-go-opentracing).
