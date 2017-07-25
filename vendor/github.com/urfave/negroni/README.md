# Negroni
[![GoDoc](https://godoc.org/github.com/urfave/negroni?status.svg)](http://godoc.org/github.com/urfave/negroni)
[![Build Status](https://travis-ci.org/urfave/negroni.svg?branch=master)](https://travis-ci.org/urfave/negroni)
[![codebeat](https://codebeat.co/badges/47d320b1-209e-45e8-bd99-9094bc5111e2)](https://codebeat.co/projects/github-com-urfave-negroni)
[![codecov](https://codecov.io/gh/urfave/negroni/branch/master/graph/badge.svg)](https://codecov.io/gh/urfave/negroni)

**Notice:** This is the library formerly known as
`github.com/codegangsta/negroni` -- Github will automatically redirect requests
to this repository, but we recommend updating your references for clarity.

Negroni is an idiomatic approach to web middleware in Go. It is tiny,
non-intrusive, and encourages use of `net/http` Handlers.

If you like the idea of [Martini](https://github.com/go-martini/martini), but
you think it contains too much magic, then Negroni is a great fit.

Language Translations:
* [German (de_DE)](translations/README_de_de.md)
* [Português Brasileiro (pt_BR)](translations/README_pt_br.md)
* [简体中文 (zh_cn)](translations/README_zh_cn.md)
* [繁體中文 (zh_tw)](translations/README_zh_tw.md)
* [日本語 (ja_JP)](translations/README_ja_JP.md)

## Getting Started

After installing Go and setting up your
[GOPATH](http://golang.org/doc/code.html#GOPATH), create your first `.go` file.
We'll call it `server.go`.

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic() // Includes some default middlewares
  n.UseHandler(mux)

  http.ListenAndServe(":3000", n)
}
```

Then install the Negroni package (**NOTE**: &gt;= **go 1.1** is required):

```
go get github.com/urfave/negroni
```

Then run your server:

```
go run server.go
```

You will now have a Go `net/http` webserver running on `localhost:3000`.

### Packaging

If you are on Debian, `negroni` is also available as [a
package](https://packages.debian.org/sid/golang-github-urfave-negroni-dev) that
you can install via `apt install golang-github-urfave-negroni-dev` (at the time
of writing, it is in the `sid` repositories).

## Is Negroni a Framework?

Negroni is **not** a framework. It is a middleware-focused library that is
designed to work directly with `net/http`.

## Routing?

Negroni is BYOR (Bring your own Router). The Go community already has a number
of great http routers available, and Negroni tries to play well with all of them
by fully supporting `net/http`. For instance, integrating with [Gorilla Mux]
looks like so:

``` go
router := mux.NewRouter()
router.HandleFunc("/", HomeHandler)

n := negroni.New(Middleware1, Middleware2)
// Or use a middleware with the Use() function
n.Use(Middleware3)
// router goes last
n.UseHandler(router)

http.ListenAndServe(":3001", n)
```

## `negroni.Classic()`

`negroni.Classic()` provides some default middleware that is useful for most
applications:

* [`negroni.Recovery`](#recovery) - Panic Recovery Middleware.
* [`negroni.Logger`](#logger) - Request/Response Logger Middleware.
* [`negroni.Static`](#static) - Static File serving under the "public"
  directory.

This makes it really easy to get started with some useful features from Negroni.

## Handlers

Negroni provides a bidirectional middleware flow. This is done through the
`negroni.Handler` interface:

``` go
type Handler interface {
  ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
```

If a middleware hasn't already written to the `ResponseWriter`, it should call
the next `http.HandlerFunc` in the chain to yield to the next middleware
handler.  This can be used for great good:

``` go
func MyMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  // do some stuff before
  next(rw, r)
  // do some stuff after
}
```

And you can map it to the handler chain with the `Use` function:

``` go
n := negroni.New()
n.Use(negroni.HandlerFunc(MyMiddleware))
```

You can also map plain old `http.Handler`s:

``` go
n := negroni.New()

mux := http.NewServeMux()
// map your routes

n.UseHandler(mux)

http.ListenAndServe(":3000", n)
```

## `With()`

Negroni has a convenience function called `With`. `With` takes one or more
`Handler` instances and returns a new `Negroni` with the combination of the
receiver's handlers and the new handlers.

```go
// middleware we want to reuse
common := negroni.New()
common.Use(MyMiddleware1)
common.Use(MyMiddleware2)

// `specific` is a new negroni with the handlers from `common` combined with the
// the handlers passed in
specific := common.With(
	SpecificMiddleware1,
	SpecificMiddleware2
)
```

## `Run()`

Negroni has a convenience function called `Run`. `Run` takes an addr string
identical to [`http.ListenAndServe`](https://godoc.org/net/http#ListenAndServe).

<!-- { "interrupt": true } -->
``` go
package main

import (
  "github.com/urfave/negroni"
)

func main() {
  n := negroni.Classic()
  n.Run(":8080")
}
```

In general, you will want to use `net/http` methods and pass `negroni` as a
`Handler`, as this is more flexible, e.g.:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.Classic() // Includes some default middlewares
  n.UseHandler(mux)

  s := &http.Server{
    Addr:           ":8080",
    Handler:        n,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }
  log.Fatal(s.ListenAndServe())
}
```

## Route Specific Middleware

If you have a route group of routes that need specific middleware to be
executed, you can simply create a new Negroni instance and use it as your route
handler.

``` go
router := mux.NewRouter()
adminRoutes := mux.NewRouter()
// add admin routes here

// Create a new negroni for the admin middleware
router.PathPrefix("/admin").Handler(negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(adminRoutes),
))
```

If you are using [Gorilla Mux], here is an example using a subrouter:

``` go
router := mux.NewRouter()
subRouter := mux.NewRouter().PathPrefix("/subpath").Subrouter().StrictSlash(true)
subRouter.HandleFunc("/", someSubpathHandler) // "/subpath/"
subRouter.HandleFunc("/:id", someSubpathHandler) // "/subpath/:id"

// "/subpath" is necessary to ensure the subRouter and main router linkup
router.PathPrefix("/subpath").Handler(negroni.New(
  Middleware1,
  Middleware2,
  negroni.Wrap(subRouter),
))
```

`With()` can be used to eliminate redundancy for middlewares shared across
routes.

``` go
router := mux.NewRouter()
apiRoutes := mux.NewRouter()
// add api routes here
webRoutes := mux.NewRouter()
// add web routes here

// create common middleware to be shared across routes
common := negroni.New(
	Middleware1,
	Middleware2,
)

// create a new negroni for the api middleware
// using the common middleware as a base
router.PathPrefix("/api").Handler(common.With(
  APIMiddleware1,
  negroni.Wrap(apiRoutes),
))
// create a new negroni for the web middleware
// using the common middleware as a base
router.PathPrefix("/web").Handler(common.With(
  WebMiddleware1,
  negroni.Wrap(webRoutes),
))
```

## Bundled Middleware

### Static

This middleware will serve files on the filesystem. If the files do not exist,
it proxies the request to the next middleware. If you want the requests for
non-existent files to return a `404 File Not Found` to the user you should look
at using [http.FileServer](https://golang.org/pkg/net/http/#FileServer) as
a handler.

Example:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  // Example of using a http.FileServer if you want "server-like" rather than "middleware" behavior
  // mux.Handle("/public", http.FileServer(http.Dir("/home/public")))

  n := negroni.New()
  n.Use(negroni.NewStatic(http.Dir("/tmp")))
  n.UseHandler(mux)

  http.ListenAndServe(":3002", n)
}
```

Will serve files from the `/tmp` directory first, but proxy calls to the next
handler if the request does not match a file on the filesystem.

### Recovery

This middleware catches `panic`s and responds with a `500` response code. If
any other middleware has written a response code or body, this middleware will
fail to properly send a 500 to the client, as the client has already received
the HTTP response code. Additionally, an `ErrorHandlerFunc` can be attached
to report 500's to an error reporting service such as Sentry or Airbrake.

Example:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    panic("oh no")
  })

  n := negroni.New()
  n.Use(negroni.NewRecovery())
  n.UseHandler(mux)

  http.ListenAndServe(":3003", n)
}
```

Will return a `500 Internal Server Error` to each request. It will also log the
stack traces as well as print the stack trace to the requester if `PrintStack`
is set to `true` (the default).

Example with error handler:

``` go
package main

import (
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    panic("oh no")
  })

  n := negroni.New()
  recovery := negroni.NewRecovery()
  recovery.ErrorHandlerFunc = reportToSentry
  n.Use(recovery)
  n.UseHandler(mux)

  http.ListenAndServe(":3003", n)
}

func reportToSentry(error interface{}) {
    // write code here to report error to Sentry
}
```


## Logger

This middleware logs each incoming request and response.

Example:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/urfave/negroni"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := negroni.New()
  n.Use(negroni.NewLogger())
  n.UseHandler(mux)

  http.ListenAndServe(":3004", n)
}
```

Will print a log similar to:

```
[negroni] Started GET /
[negroni] Completed 200 OK in 145.446µs
```

on each request.

## Third Party Middleware

Here is a current list of Negroni compatible middlware. Feel free to put up a PR
linking your middleware if you have built one:

| Middleware | Author | Description |
| -----------|--------|-------------|
| [binding](https://github.com/mholt/binding) | [Matt Holt](https://github.com/mholt) | Data binding from HTTP requests into structs |
| [cloudwatch](https://github.com/cvillecsteele/negroni-cloudwatch) | [Colin Steele](https://github.com/cvillecsteele) | AWS cloudwatch metrics middleware |
| [cors](https://github.com/rs/cors) | [Olivier Poitrey](https://github.com/rs) | [Cross Origin Resource Sharing](http://www.w3.org/TR/cors/) (CORS) support |
| [csp](https://github.com/awakenetworks/csp) | [Awake Networks](https://github.com/awakenetworks) | [Content Security Policy](https://www.w3.org/TR/CSP2/) (CSP) support |
| [delay](https://github.com/jeffbmartinez/delay) | [Jeff Martinez](https://github.com/jeffbmartinez) | Add delays/latency to endpoints. Useful when testing effects of high latency |
| [New Relic Go Agent](https://github.com/yadvendar/negroni-newrelic-go-agent) | [Yadvendar Champawat](https://github.com/yadvendar) | Official [New Relic Go Agent](https://github.com/newrelic/go-agent) (currently in beta)  |
| [gorelic](https://github.com/jingweno/negroni-gorelic) | [Jingwen Owen Ou](https://github.com/jingweno) | New Relic agent for Go runtime |
| [Graceful](https://github.com/tylerb/graceful) | [Tyler Bunnell](https://github.com/tylerb) | Graceful HTTP Shutdown |
| [gzip](https://github.com/phyber/negroni-gzip) | [phyber](https://github.com/phyber) | GZIP response compression |
| [JWT Middleware](https://github.com/auth0/go-jwt-middleware) | [Auth0](https://github.com/auth0) | Middleware checks for a JWT on the `Authorization` header on incoming requests and decodes it|
| [logrus](https://github.com/meatballhat/negroni-logrus) | [Dan Buch](https://github.com/meatballhat) | Logrus-based logger |
| [oauth2](https://github.com/goincremental/negroni-oauth2) | [David Bochenski](https://github.com/bochenski) | oAuth2 middleware |
| [onthefly](https://github.com/xyproto/onthefly) | [Alexander Rødseth](https://github.com/xyproto) | Generate TinySVG, HTML and CSS on the fly |
| [permissions2](https://github.com/xyproto/permissions2) | [Alexander Rødseth](https://github.com/xyproto) | Cookies, users and permissions |
| [prometheus](https://github.com/zbindenren/negroni-prometheus) | [Rene Zbinden](https://github.com/zbindenren) | Easily create metrics endpoint for the [prometheus](http://prometheus.io) instrumentation tool |
| [render](https://github.com/unrolled/render) | [Cory Jacobsen](https://github.com/unrolled) | Render JSON, XML and HTML templates |
| [RestGate](https://github.com/pjebs/restgate) | [Prasanga Siripala](https://github.com/pjebs) | Secure authentication for REST API endpoints |
| [secure](https://github.com/unrolled/secure) | [Cory Jacobsen](https://github.com/unrolled) | Middleware that implements a few quick security wins |
| [sessions](https://github.com/goincremental/negroni-sessions) | [David Bochenski](https://github.com/bochenski) | Session Management |
| [stats](https://github.com/thoas/stats) | [Florent Messa](https://github.com/thoas) | Store information about your web application (response time, etc.) |
| [VanGoH](https://github.com/auroratechnologies/vangoh) | [Taylor Wrobel](https://github.com/twrobel3) | Configurable [AWS-Style](http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html) HMAC authentication middleware |
| [xrequestid](https://github.com/pilu/xrequestid) | [Andrea Franz](https://github.com/pilu) | Middleware that assigns a random X-Request-Id header to each request |
| [mgo session](https://github.com/joeljames/nigroni-mgo-session) | [Joel James](https://github.com/joeljames) | Middleware that handles creating and closing mgo sessions per request |
| [digits](https://github.com/bamarni/digits) | [Bilal Amarni](https://github.com/bamarni) | Middleware that handles [Twitter Digits](https://get.digits.com/) authentication |

## Examples

[Alexander Rødseth](https://github.com/xyproto) created
[mooseware](https://github.com/xyproto/mooseware), a skeleton for writing a
Negroni middleware handler.

## Live code reload?

[gin](https://github.com/codegangsta/gin) and
[fresh](https://github.com/pilu/fresh) both live reload negroni apps.

## Essential Reading for Beginners of Go & Negroni

* [Using a Context to pass information from middleware to end handler](http://elithrar.github.io/article/map-string-interface/)
* [Understanding middleware](https://mattstauffer.co/blog/laravel-5.0-middleware-filter-style)

## About

Negroni is obsessively designed by none other than the [Code
Gangsta](https://codegangsta.io/)

[Gorilla Mux]: https://github.com/gorilla/mux
[`http.FileSystem`]: https://godoc.org/net/http#FileSystem
