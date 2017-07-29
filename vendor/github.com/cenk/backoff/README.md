# Exponential Backoff [![GoDoc][godoc image]][godoc] [![Build Status][travis image]][travis] [![Coverage Status][coveralls image]][coveralls]

This is a Go port of the exponential backoff algorithm from [Google's HTTP Client Library for Java][google-http-java-client].

[Exponential backoff][exponential backoff wiki]
is an algorithm that uses feedback to multiplicatively decrease the rate of some process,
in order to gradually find an acceptable rate.
The retries exponentially increase and stop increasing when a certain threshold is met.

## Usage

See https://godoc.org/github.com/cenkalti/backoff#pkg-examples

## Contributing

* I would like to keep this library as small as possible.
* Please don't send a PR without opening an issue and discussing it first.
* If proposed change is not a common use case, I will probably not accept it.

[godoc]: https://godoc.org/github.com/cenkalti/backoff
[godoc image]: https://godoc.org/github.com/cenkalti/backoff?status.png
[travis]: https://travis-ci.org/cenkalti/backoff
[travis image]: https://travis-ci.org/cenkalti/backoff.png?branch=master
[coveralls]: https://coveralls.io/github/cenkalti/backoff?branch=master
[coveralls image]: https://coveralls.io/repos/github/cenkalti/backoff/badge.svg?branch=master

[google-http-java-client]: https://github.com/google/google-http-java-client
[exponential backoff wiki]: http://en.wikipedia.org/wiki/Exponential_backoff

[advanced example]: https://godoc.org/github.com/cenkalti/backoff#example_
