graceful [![GoDoc](https://godoc.org/github.com/tylerb/graceful?status.png)](http://godoc.org/github.com/tylerb/graceful) [![Build Status](https://drone.io/github.com/tylerb/graceful/status.png)](https://drone.io/github.com/tylerb/graceful/latest) [![Coverage Status](https://coveralls.io/repos/tylerb/graceful/badge.svg?branch=dronedebug)](https://coveralls.io/r/tylerb/graceful?branch=dronedebug) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tylerb/graceful?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
========

Graceful is a Go 1.3+ package enabling graceful shutdown of http.Handler servers.

## Installation

To install, simply execute:

```
go get gopkg.in/tylerb/graceful.v1
```

I am using [gopkg.in](http://labix.org/gopkg.in) to control releases.

## Usage

Using Graceful is easy. Simply create your http.Handler and pass it to the `Run` function:

```go
package main

import (
  "gopkg.in/tylerb/graceful.v1"
  "net/http"
  "fmt"
  "time"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  graceful.Run(":3001",10*time.Second,mux)
}
```

Another example, using [Negroni](https://github.com/codegangsta/negroni), functions in much the same manner:

```go
package main

import (
  "github.com/codegangsta/negroni"
  "gopkg.in/tylerb/graceful.v1"
  "net/http"
  "fmt"
  "time"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic()
  n.UseHandler(mux)
  //n.Run(":3000")
  graceful.Run(":3001",10*time.Second,n)
}
```

In addition to Run there are the http.Server counterparts ListenAndServe, ListenAndServeTLS and Serve, which allow you to configure HTTPS, custom timeouts and error handling.
Graceful may also be used by instantiating its Server type directly, which embeds an http.Server:

```go
mux := // ...

srv := &graceful.Server{
  Timeout: 10 * time.Second,

  Server: &http.Server{
    Addr: ":1234",
    Handler: mux,
  },
}

srv.ListenAndServe()
```

This form allows you to set the ConnState callback, which works in the same way as in http.Server:

```go
mux := // ...

srv := &graceful.Server{
  Timeout: 10 * time.Second,

  ConnState: func(conn net.Conn, state http.ConnState) {
    // conn has a new state
  },

  Server: &http.Server{
    Addr: ":1234",
    Handler: mux,
  },
}

srv.ListenAndServe()
```

## Behaviour

When Graceful is sent a SIGINT or SIGTERM (possibly from ^C or a kill command), it:

1. Disables keepalive connections.
2. Closes the listening socket, allowing another process to listen on that port immediately.
3. Starts a timer of `timeout` duration to give active requests a chance to finish.
4. When timeout expires, closes all active connections.
5. Closes the `stopChan`, waking up any blocking goroutines.
6. Returns from the function, allowing the server to terminate.

## Notes

If the `timeout` argument to `Run` is 0, the server never times out, allowing all active requests to complete.

If you wish to stop the server in some way other than an OS signal, you may call the `Stop()` function.
This function stops the server, gracefully, using the new timeout value you provide. The `StopChan()` function
returns a channel on which you can block while waiting for the server to stop. This channel will be closed when
the server is stopped, allowing your execution to proceed. Multiple goroutines can block on this channel at the
same time and all will be signalled when stopping is complete.

## Contributing

If you would like to contribute, please:

1. Create a GitHub issue regarding the contribution. Features and bugs should be discussed beforehand.
2. Fork the repository.
3. Create a pull request with your solution. This pull request should reference and close the issues (Fix #2).

All pull requests should:

1. Pass [gometalinter -t .](https://github.com/alecthomas/gometalinter) with no warnings.
2. Be `go fmt` formatted.
