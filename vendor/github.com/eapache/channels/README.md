channels
========

[![Build Status](https://travis-ci.org/eapache/channels.svg?branch=master)](https://travis-ci.org/eapache/channels)
[![GoDoc](https://godoc.org/github.com/eapache/channels?status.png)](https://godoc.org/github.com/eapache/channels)
[![Code of Conduct](https://img.shields.io/badge/code%20of%20conduct-active-blue.svg)](https://eapache.github.io/conduct.html)

A collection of helper functions and special types for working with and
extending [Go](https://golang.org/)'s existing channels. Due to limitations
of Go's type system, importing this library directly is often not practical for
production code. It serves equally well, however, as a reference guide and
template for implementing many common idioms; if you use it in this way I would
appreciate the inclusion of some sort of credit in the resulting code.

See https://godoc.org/github.com/eapache/channels for full documentation or
https://gopkg.in/eapache/channels.v1 for a versioned import path.

Requires Go version 1.1 or later, as certain necessary elements of the `reflect`
package were not present in 1.0.

Most of the buffered channel types in this package are backed by a very fast
queue implementation that used to be built into this package but has now been
extracted into its own package at https://github.com/eapache/queue.

*Note:* Several types in this package provide so-called "infinite" buffers. Be
very careful using these, as no buffer is truly infinite. If such a buffer
grows too large your program will run out of memory and crash. Caveat emptor.
