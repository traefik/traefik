# API

Traefik exposes a number of information through an API handler, such as the configuration of all routers, services, middlewares, etc.

As with all features of Traefik, this handler can be enabled with the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).

## Security

Enabling the API in production is not recommended, because it will expose all configuration elements,
including sensitive data.

In production, it should be at least secured by authentication and authorizations.

A good sane default (non exhaustive) set of recommendations
would be to apply the following protection mechanisms:

* At the application level:  
  securing with middlewares such as [basic authentication](../middlewares/basicauth.md) or [white listing](../middlewares/ipwhitelist.md).

* At the transport level:  
  NOT publicly exposing the API's port,
  keeping it restricted to internal networks
  (as in the [principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege), applied to networks).

## Configuration

To enable the API handler:

```toml tab="File"
[api]
```

```bash tab="CLI"
--api
```

### `dashboard`

_Optional, Default=true_

Enable the dashboard. More about the dashboard features [here](./dashboard.md).

```toml tab="File"
[api]
  dashboard = true
```

```bash tab="CLI"
--api.dashboard
```

### `entrypoint`

_Optional, Default="traefik"_

The entry point that the API handler will be bound to.
The default ("traefik") is an internal entry point (which is always defined).

```toml tab="File"
[api]
  entrypoint = "web"
```

```bash tab="CLI"
--api.entrypoint="web"
```

### `middlewares`

_Optional, Default=empty_

The list of [middlewares](../middlewares/overview.md) applied to the API handler.

```toml tab="File"
[api]
  middlewares = ["api-auth", "api-prefix"]
```

```bash tab="CLI"
--api.middlewares="api-auth,api-prefix"
```

### `debug`

_Optional, Default=false_

Enable additional endpoints for debugging and profiling, served under `/debug/`.

```toml tab="File"
[api]
  debug = true
```

```bash tab="CLI"
--api.debug=true
```

## Endpoints

All the following endpoints must be accessed with a `GET` HTTP request.

| Path                           | Description                                                                               |
|--------------------------------|-------------------------------------------------------------------------------------------|
| `/api/http/routers`            | Lists all the HTTP routers information.                                                   |
| `/api/http/routers/{name}`     | Returns the information of the HTTP router specified by `name`.                           |
| `/api/http/services`           | Lists all the HTTP services information.                                                  |
| `/api/http/services/{name}`    | Returns the information of the HTTP service specified by `name`.                          |
| `/api/http/middlewares`        | Lists all the HTTP middlewares information.                                               |
| `/api/http/middlewares/{name}` | Returns the information of the HTTP middleware specified by `name`.                       |
| `/api/tcp/routers`             | Lists all the TCP routers information.                                                    |
| `/api/tcp/routers/{name}`      | Returns the information of the TCP router specified by `name`.                            |
| `/api/tcp/services`            | Lists all the TCP services information.                                                   |
| `/api/tcp/services/{name}`     | Returns the information of the TCP service specified by `name`.                           |
| `/api/entrypoints`             | Lists all the entry points information.                                                   |
| `/api/entrypoints/{name}`      | Returns the information of the entry point specified by `name`.                           |
| `/api/version`                 | Returns information about Traefik version.                                                |
| `/debug/vars`                  | See the [expvar](https://golang.org/pkg/expvar/) Go documentation.                        |
| `/debug/pprof/`                | See the [pprof Index](https://golang.org/pkg/net/http/pprof/#Index) Go documentation.     |
| `/debug/pprof/cmdline`         | See the [pprof Cmdline](https://golang.org/pkg/net/http/pprof/#Cmdline) Go documentation. |
| `/debug/pprof/profile`         | See the [pprof Profile](https://golang.org/pkg/net/http/pprof/#Profile) Go documentation. |
| `/debug/pprof/symbol`          | See the [pprof Symbol](https://golang.org/pkg/net/http/pprof/#Symbol) Go documentation.   |
| `/debug/pprof/trace`           | See the [pprof Trace](https://golang.org/pkg/net/http/pprof/#Trace) Go documentation.     |

## Common Configuration Use Cases

### Address / Port

You can define a custom address/port like this:

```toml
[entryPoints]
  [entryPoints.web]
    address = ":80"

  [entryPoints.foo]
    address = ":8082"

  [entryPoints.bar]
    address = ":8083"

[ping]
  entryPoint = "foo"

[api]
  entryPoint = "bar"
```

In the above example, you would access a service at /foo, an api endpoint, or the health-check as follows:

* Service: `http://hostname:80/foo`
* API: `http://hostname:8083/api/http/routers`
* Ping URL: `http://hostname:8082/ping`

### Authentication

To restrict access to the API handler, one can add authentication with the [basic auth middleware](../middlewares/basicauth.md).

```toml
[api]
  middlewares=["api-auth"]
```

```toml
[http.middlewares]
  [http.middlewares.api-auth.basicAuth]
    users = [
      "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
      "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
    ]
```
