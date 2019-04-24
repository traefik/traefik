# go-xerial-snappy

[![Build Status](https://travis-ci.org/eapache/go-xerial-snappy.svg?branch=master)](https://travis-ci.org/eapache/go-xerial-snappy)

Xerial-compatible Snappy framing support for golang.

Packages using Xerial for snappy encoding use a framing format incompatible with
basically everything else in existence. This package wraps Go's built-in snappy
package to support it.

Apps that use this format include Apache Kafka (see
https://github.com/dpkp/kafka-python/issues/126#issuecomment-35478921 for
details).
