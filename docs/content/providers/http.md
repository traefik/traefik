# Traefik & HTTP

Provide your [dynamic configuration](./overview.md) via an HTTP(S) endpoint and let Traefik do the rest!

## Routing Configuration

The HTTP provider uses the same configuration as the [File Provider](./file.md) in YAML or JSON format.

## Provider Configuration

### `endpoint`

_Required_

Defines the HTTP(S) endpoint to poll.

```yaml tab="File (YAML)"
providers:
  http:
    endpoint:
      - "http://127.0.0.1:9000/api"
```

```toml tab="File (TOML)"
[providers.http]
  endpoint = "http://127.0.0.1:9000/api"
```

```bash tab="CLI"
--providers.http.endpoint=http://127.0.0.1:9000/api
```

### `pollInterval`

_Optional, Default="5s"_

Defines the polling interval.

```yaml tab="File (YAML)"
providers:
  http:
    pollInterval: "5s"
```

```toml tab="File (TOML)"
[providers.http]
  pollInterval = "5s"
```

```bash tab="CLI"
--providers.http.pollInterval=5s
```

### `pollTimeout`

_Optional, Default="5s"_

Defines the polling timeout when connecting to the endpoint.

```yaml tab="File (YAML)"
providers:
  http:
    pollTimeout: "5s"
```

```toml tab="File (TOML)"
[providers.http]
  pollTimeout = "5s"
```

```bash tab="CLI"
--providers.http.pollTimeout=5s
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to the endpoint.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the endpoint,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.http.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.http.tls.ca=path/to/ca.crt
```

#### `caOptional`

_Optional_

The value of `caOptional` defines which policy should be used for the secure connection with TLS Client Authentication to the endpoint.

!!! warning ""

    If `ca` is undefined, this option will be ignored, and no client certificate will be requested during the handshake. Any provided certificate will thus never be verified.

When this option is set to `true`, a client certificate is requested during the handshake but is not required. If a certificate is sent, it is required to be valid.

When this option is set to `false`, a client certificate is requested during the handshake, and at least one valid certificate should be sent by the client.

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      caOptional: true
```

```toml tab="File (TOML)"
[providers.http.tls]
  caOptional = true
```

```bash tab="CLI"
--providers.http.tls.caOptional=true
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the endpoint.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.http.tls.cert=path/to/foo.cert
--providers.http.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the endpoint.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.http.tls.cert=path/to/foo.cert
--providers.http.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to the endpoint accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  http:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.http.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.http.tls.insecureSkipVerify=true
```
