# Traefik & HTTP

Provide your [dynamic configuration](./overview.md) via an HTTP(s) endpoint and let Traefik do the rest!

## Routing Configuration

The HTTP provider uses the same configuration as the [File Provider](./file.md) in YAML or JSON format.

## Provider Configuration

### `endpoint`

_Required_

Defines the HTTP(s) endpoint to poll.

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

_Optional, Default="5s"_

Defines the polling interval.

```toml tab="File (TOML)"
[providers.http]
  pollInterval = "5s"
```

```yaml tab="File (YAML)"
providers:
  http:
    pollInterval: "5s"
```

```bash tab="CLI"
--providers.http.pollInterval=5s
```

### `pollTimeout`

_Optional, Default="5s"_

Defines the polling timeout when connecting to the configured endpoint.

```toml tab="File (TOML)"
[providers.http]
  pollTimeout = "5s"
```

```yaml tab="File (YAML)"
providers:
  http:
    pollTimeout: "5s"
```

```bash tab="CLI"
--providers.http.pollTimeout=5s
```

### `tls`

_Optional_

#### `tls.ca`

Certificate Authority used for the secured connection to the configured Endpoint.

```toml tab="File (TOML)"
[providers.http.tls]
  ca = "path/to/ca.crt"
```

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      ca: path/to/ca.crt
```

```bash tab="CLI"
--providers.http.tls.ca=path/to/ca.crt
```

#### `tls.caOptional`

Policy followed for the secured connection with TLS Client Authentication to the configured Endpoint.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

```toml tab="File (TOML)"
[providers.http.tls]
  caOptional = true
```

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      caOptional: true
```

```bash tab="CLI"
--providers.http.tls.caOptional=true
```

#### `tls.cert`

Public certificate used for the secured connection to the configured Endpoint.

```toml tab="File (TOML)"
[providers.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.http.tls.cert=path/to/foo.cert
--providers.http.tls.key=path/to/foo.key
```

#### `tls.key`

Private certificate used for the secured connection to the configured Endpoint.

```toml tab="File (TOML)"
[providers.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.http.tls.cert=path/to/foo.cert
--providers.http.tls.key=path/to/foo.key
```

#### `tls.insecureSkipVerify`

If `insecureSkipVerify` is `true`, TLS connection to the configured Endpoint accepts any certificate presented by the 
server and any host name in that certificate.

```toml tab="File (TOML)"
[providers.http.tls]
  insecureSkipVerify = true
```

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      insecureSkipVerify: true
```

```bash tab="CLI"
--providers.http.tls.insecureSkipVerify=true
```
