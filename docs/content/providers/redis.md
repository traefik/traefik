# Traefik & Redis

A Story of KV store & Containers
{: .subtitle }

Store your configuration in Redis and let Traefik do the rest!

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:6379"_

Defines how to access to Redis.

```toml tab="File (TOML)"
[providers.redis]
  endpoints = ["127.0.0.1:6379"]
```

```yaml tab="File (YAML)"
providers:
  redis:
    endpoints:
      - "127.0.0.1:6379"
```

```bash tab="CLI"
--providers.redis.endpoints=127.0.0.1:6379
```

### `rootKey`

Defines the root key of the configuration.

_Required, Default="traefik"_

```toml tab="File (TOML)"
[providers.redis]
  rootKey = "traefik"
```

```yaml tab="File (YAML)"
providers:
  redis:
    rootKey: "traefik"
```

```bash tab="CLI"
--providers.redis.rootkey=traefik
```

### `username`

Defines a username to connect with Redis.

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.redis]
  # ...
  username = "foo"
```

```yaml tab="File (YAML)"
providers:
  redis:
    # ...
    usename: "foo"
```

```bash tab="CLI"
--providers.redis.username=foo
```

### `password`

_Optional, Default=""_

Defines a password to connect with Redis.

```toml tab="File (TOML)"
[providers.redis]
  # ...
  password = "bar"
```

```yaml tab="File (YAML)"
providers:
  redis:
    # ...
    password: "bar"
```

```bash tab="CLI"
--providers.redis.password=foo
```

### `tls`

_Optional_

#### `tls.ca`

Certificate Authority used for the secured connection to Redis.

```toml tab="File (TOML)"
[providers.redis.tls]
  ca = "path/to/ca.crt"
```

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      ca: path/to/ca.crt
```

```bash tab="CLI"
--providers.redis.tls.ca=path/to/ca.crt
```

#### `tls.caOptional`

Policy followed for the secured connection with TLS Client Authentication to Redis.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

```toml tab="File (TOML)"
[providers.redis.tls]
  caOptional = true
```

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      caOptional: true
```

```bash tab="CLI"
--providers.redis.tls.caOptional=true
```

#### `tls.cert`

Public certificate used for the secured connection to Redis.

```toml tab="File (TOML)"
[providers.redis.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.redis.tls.cert=path/to/foo.cert
--providers.redis.tls.key=path/to/foo.key
```

#### `tls.key`

Private certificate used for the secured connection to Redis.

```toml tab="File (TOML)"
[providers.redis.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.redis.tls.cert=path/to/foo.cert
--providers.redis.tls.key=path/to/foo.key
```

#### `tls.insecureSkipVerify`

If `insecureSkipVerify` is `true`, TLS for the connection to Redis accepts any certificate presented by the server and any host name in that certificate.

```toml tab="File (TOML)"
[providers.redis.tls]
  insecureSkipVerify = true
```

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      insecureSkipVerify: true
```

```bash tab="CLI"
--providers.redis.tls.insecureSkipVerify=true
```
