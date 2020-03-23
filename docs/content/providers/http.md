# Traefik & HTTP

Provide your configuration via an http endpoint and let Traefik do the rest!

## Routing Configuration

The HTTP provider uses the same configuration as the [File Provider](./file.md).

## Provider Configuration

### `endpoint`

_Required_

Configures the endpoint to access.

```toml tab="File (TOML)"
[providers.http]
  endpoint = "http://127.0.0.1:9000/api"
```

```yaml tab="File (YAML)"
providers:
  http:
    endpoint:
      - "http://127.0.0.1:9000/api"
```

```bash tab="CLI"
--providers.http.endpoint=http://127.0.0.1:9000/api
```

### `pollInterval`

Defines the interval with which to poll the endpoint.

_Optional, Default="15s"_

```toml tab="File (TOML)"
[providers.http]
  pollInterval = "15s"
```

```yaml tab="File (YAML)"
providers:
  http:
    pollInterval: "15s"
```

```bash tab="CLI"
--providers.http.pollinterval=15s
```

### `pollTimeout`

Defines a timeout when connecting to the endpoint.

_Optional, Default="15s"_

```toml tab="File (TOML)"
[providers.http]
  pollTimeout = "15s"
```

```yaml tab="File (YAML)"
providers:
  http:
    pollTimeout: "15s"
```

```bash tab="CLI"
--providers.http.polltimeout=15s
```
