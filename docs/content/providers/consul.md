# Traefik & Consul

A Story of KV store & Containers
{: .subtitle }

Store your configuration in Consul and let Traefik do the rest!

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:8500"_

Defines how to access to Consul.

```toml tab="File (TOML)"
[providers.consul]
  endpoints = ["127.0.0.1:8500"]
```

```yaml tab="File (YAML)"
providers:
  consul:
    endpoints:
      - "127.0.0.1:8500"
```

```bash tab="CLI"
--providers.consul.endpoints=127.0.0.1:8500
```

### `rootKey`

Defines the root key of the configuration.

_Required, Default="traefik"_

```toml tab="File (TOML)"
[providers.consul]
  rootKey = "traefik"
```

```yaml tab="File (YAML)"
providers:
  consul:
    rootKey: "traefik"
```

```bash tab="CLI"
--providers.consul.rootkey=traefik
```

### `username`

Defines a username to connect with Consul.

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consul]
  # ...
  username = "foo"
```

```yaml tab="File (YAML)"
providers:
  consul:
    # ...
    usename: "foo"
```

```bash tab="CLI"
--providers.consul.username=foo
```

### `password`

_Optional, Default=""_

Defines a password to connect with Consul.

```toml tab="File (TOML)"
[providers.consul]
  # ...
  password = "bar"
```

```yaml tab="File (YAML)"
providers:
  consul:
    # ...
    password: "bar"
```

```bash tab="CLI"
--providers.consul.password=foo
```

### `tls`

_Optional_

#### `tls.ca`

Certificate Authority used for the secured connection to Consul.

```toml tab="File (TOML)"
[providers.consul.tls]
  ca = "path/to/ca.crt"
```

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      ca: path/to/ca.crt
```

```bash tab="CLI"
--providers.consul.tls.ca=path/to/ca.crt
```

#### `tls.caOptional`

Policy followed for the secured connection with TLS Client Authentication to Consul.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

```toml tab="File (TOML)"
[providers.consul.tls]
  caOptional = true
```

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      caOptional: true
```

```bash tab="CLI"
--providers.consul.tls.caOptional=true
```

#### `tls.cert`

Public certificate used for the secured connection to Consul.

```toml tab="File (TOML)"
[providers.consul.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.consul.tls.cert=path/to/foo.cert
--providers.consul.tls.key=path/to/foo.key
```

#### `tls.key`

Private certificate used for the secured connection to Consul.

```toml tab="File (TOML)"
[providers.consul.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.consul.tls.cert=path/to/foo.cert
--providers.consul.tls.key=path/to/foo.key
```

#### `tls.insecureSkipVerify`

If `insecureSkipVerify` is `true`, TLS for the connection to Consul accepts any certificate presented by the server and any host name in that certificate.

```toml tab="File (TOML)"
[providers.consul.tls]
  insecureSkipVerify = true
```

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      insecureSkipVerify: true
```

```bash tab="CLI"
--providers.consul.tls.insecureSkipVerify=true
```
