# API

!!! important
    In the RC version, you can't configure middlewares (basic authentication or white listing) anymore, but as security is important, this will change before the GA version.

Traefik exposes a number of information through an API handler, such as the configuration of all routers, services, middlewares, etc.

As with all features of Traefik, this handler can be enabled with the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).

## Security

Enabling the API in production is not recommended, because it will expose all configuration elements,
including sensitive data.

In production, it should be at least secured by authentication and authorizations.

A good sane default (non exhaustive) set of recommendations
would be to apply the following protection mechanisms:

* At the transport level:  
  NOT publicly exposing the API's port,
  keeping it restricted to internal networks
  (as in the [principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege), applied to networks).

!!! important
    In the beta version, you can't configure middlewares (basic authentication or white listing) anymore, but as security is important, this will change before the RC version.

## Configuration

To enable the API handler:

```toml tab="File (TOML)"
[api]
```

```yaml tab="File (YAML)"
api: {}
```

```bash tab="CLI"
--api=true
```

### `dashboard`

_Optional, Default=true_

Enable the dashboard. More about the dashboard features [here](./dashboard.md).

```toml tab="File (TOML)"
[api]
  dashboard = true
```

```yaml tab="File (YAML)"
api:
  dashboard: true
```

```bash tab="CLI"
--api.dashboard=true
```

### `debug`

_Optional, Default=false_

Enable additional endpoints for debugging and profiling, served under `/debug/`.

```toml tab="File (TOML)"
[api]
  debug = true
```

```yaml tab="File (YAML)"
api:
  debug: true
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
