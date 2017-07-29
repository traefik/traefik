Oxy
=====

Oxy is a Go library with HTTP handlers that enhance HTTP standard library:

* [Stream](http://godoc.org/github.com/vulcand/oxy/stream) retries and buffers requests and responses 
* [Forward](http://godoc.org/github.com/vulcand/oxy/forward) forwards requests to remote location and rewrites headers 
* [Roundrobin](http://godoc.org/github.com/vulcand/oxy/roundrobin) is a round-robin load balancer 
* [Circuit Breaker](http://godoc.org/github.com/vulcand/oxy/cbreaker) Hystrix-style circuit breaker
* [Connlimit](http://godoc.org/github.com/vulcand/oxy/connlimit) Simultaneous connections limiter
* [Ratelimit](http://godoc.org/github.com/vulcand/oxy/ratelimit) Rate limiter (based on tokenbucket algo)
* [Trace](http://godoc.org/github.com/vulcand/oxy/trace) Structured request and response logger

It is designed to be fully compatible with http standard library, easy to customize and reuse.

Status
------

* Initial design is completed
* Covered by tests
* Used as a reverse proxy engine in [Vulcand](https://github.com/vulcand/vulcand)

Quickstart
-----------

Every handler is ``http.Handler``, so writing and plugging in a middleware is easy. Let us write a simple reverse proxy as an example:

Simple reverse proxy
====================

```go

import (
  "net/http"
  "github.com/vulcand/oxy/forward"
  "github.com/vulcand/oxy/testutils"
  )

// Forwards incoming requests to whatever location URL points to, adds proper forwarding headers
fwd, _ := forward.New()

redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    // let us forward this request to another server
		req.URL = testutils.ParseURI("http://localhost:63450")
		fwd.ServeHTTP(w, req)
})
	
// that's it! our reverse proxy is ready!
s := &http.Server{
	Addr:           ":8080",
	Handler:        redirect,
}
s.ListenAndServe()
```

As a next step, let us add a round robin load-balancer:


```go

import (
  "net/http"
  "github.com/vulcand/oxy/forward"
  "github.com/vulcand/oxy/roundrobin"
  )

// Forwards incoming requests to whatever location URL points to, adds proper forwarding headers
fwd, _ := forward.New()
lb, _ := roundrobin.New(fwd)

lb.UpsertServer(url1)
lb.UpsertServer(url2)

s := &http.Server{
	Addr:           ":8080",
	Handler:        lb,
}
s.ListenAndServe()
```

What if we want to handle retries and replay the request in case of errors? `stream` handler will help:


```go

import (
  "net/http"
  "github.com/vulcand/oxy/forward"
  "github.com/vulcand/oxy/roundrobin"
  )

// Forwards incoming requests to whatever location URL points to, adds proper forwarding headers

fwd, _ := forward.New()
lb, _ := roundrobin.New(fwd)

// stream will read the request body and will replay the request again in case if forward returned status
// corresponding to nework error (e.g. Gateway Timeout)
stream, _ := stream.New(lb, stream.Retry(`IsNetworkError() && Attempts() < 2`))

lb.UpsertServer(url1)
lb.UpsertServer(url2)

// that's it! our reverse proxy is ready!
s := &http.Server{
	Addr:           ":8080",
	Handler:        stream,
}
s.ListenAndServe()
```
