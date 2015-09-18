# Exponential Backoff [![GoDoc][godoc image]][godoc] [![Build Status][travis image]][travis]

This is a Go port of the exponential backoff algorithm from [Google's HTTP Client Library for Java][google-http-java-client].

[Exponential backoff][exponential backoff wiki]
is an algorithm that uses feedback to multiplicatively decrease the rate of some process,
in order to gradually find an acceptable rate.
The retries exponentially increase and stop increasing when a certain threshold is met.

## How To

We define two functions, `Retry()` and `RetryNotify()`.
They receive an `Operation` to execute, a `BackOff` algorithm,
and an optional `Notify` error handler.

The operation will be executed, and will be retried on failure with delay
as given by the backoff algorithm. The backoff algorithm can also decide when to stop
retrying.
In addition, the notify error handler will be called after each failed attempt,
except for the last time, whose error should be handled by the caller.

```go
// An Operation is executing by Retry() or RetryNotify().
// The operation will be retried using a backoff policy if it returns an error.
type Operation func() error

// Notify is a notify-on-error function. It receives an operation error and
// backoff delay if the operation failed (with an error).
//
// NOTE that if the backoff policy stated to stop retrying,
// the notify function isn't called.
type Notify func(error, time.Duration)

func Retry(Operation, BackOff) error
func RetryNotify(Operation, BackOff, Notify)
```

## Examples

See more advanced examples in the [godoc][advanced example].

### Retry

Simple retry helper that uses the default exponential backoff algorithm:

```go
operation := func() error {
    // An operation that might fail.
    return nil // or return errors.New("some error")
}

err := Retry(operation, NewExponentialBackOff())
if err != nil {
    // Handle error.
    return err
}

// Operation is successful.
return nil
```

### Ticker

```go
operation := func() error {
    // An operation that might fail
    return nil // or return errors.New("some error")
}

b := NewExponentialBackOff()
ticker := NewTicker(b)

var err error

// Ticks will continue to arrive when the previous operation is still running,
// so operations that take a while to fail could run in quick succession.
for range ticker.C {
    if err = operation(); err != nil {
        log.Println(err, "will retry...")
        continue
    }

    ticker.Stop()
    break
}

if err != nil {
    // Operation has failed.
    return err
}

// Operation is successful.
return nil
```

## Getting Started

```bash
# install
$ go get github.com/cenkalti/backoff

# test
$ cd $GOPATH/src/github.com/cenkalti/backoff
$ go get -t ./...
$ go test -v -cover
```

[godoc]: https://godoc.org/github.com/cenkalti/backoff
[godoc image]: https://godoc.org/github.com/cenkalti/backoff?status.png
[travis]: https://travis-ci.org/cenkalti/backoff
[travis image]: https://travis-ci.org/cenkalti/backoff.png

[google-http-java-client]: https://github.com/google/google-http-java-client
[exponential backoff wiki]: http://en.wikipedia.org/wiki/Exponential_backoff

[advanced example]: https://godoc.org/github.com/cenkalti/backoff#example_
