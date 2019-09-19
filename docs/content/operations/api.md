# API

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

## Configuration

If you enable the API, a new special `service` named `api@internal` is created and can then be referenced in a router.

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

And then you will be able to reference it like this:

```yaml tab="Docker"
  - "traefik.http.routers.api.rule=PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
  - "traefik.http.routers.api.service=api@internal"
  - "traefik.http.routers.api.middlewares=auth"
  - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```json tab="Marathon"
"labels": {
  "traefik.http.routers.api.rule": "PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
  "traefik.http.routers.api.service": "api@internal"
  "traefik.http.routers.api.middlewares": "auth"
  "traefik.http.middlewares.auth.basicauth.users": "test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
}
```

```yaml tab="Rancher"
# Declaring the user list
labels:
  - "traefik.http.routers.api.rule=PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
    - "traefik.http.routers.api.service=api@internal"
    - "traefik.http.routers.api.middlewares=auth"
    - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="File (TOML)"
[http.routers.my-api]
    rule="PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
    service="api@internal"
    middlewares=["auth"]

[http.middlewares.auth.basicAuth]
    users = [
        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
        "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
      ]
```

```yaml tab="File (YAML)"
http:
  routers:
    api:
      rule: PathPrefix(`/api`) || PathPrefix(`/dashboard`)
      service: api@internal
      middlewares:
        - auth
  middlewares:
    auth:
      basicAuth:
        users:
        - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/" 
        - "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

### `insecure`

Enable the API in `insecure` mode, which means that the API will be available directly on the entryPoint named `traefik`.

!!! Note
    If the entryPoint named `traefik` is not configured, it will be automatically created on port 8080.

```toml tab="File (TOML)"
[api]
  insecure = true
```

```yaml tab="File (YAML)"
api:
  insecure: true
```

```bash tab="CLI"
--api.insecure=true
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
