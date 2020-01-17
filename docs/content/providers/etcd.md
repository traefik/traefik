# Traefik & Etcd

A Story of KV store & Containers
{: .subtitle }

Store your configuration in Etcd and let Traefik do the rest!

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:2379"_

Defines how to access to Etcd.

```toml tab="File (TOML)"
[providers.etcd]
  endpoints = ["127.0.0.1:2379"]
```

```yaml tab="File (YAML)"
providers:
  etcd:
    endpoints:
      - "127.0.0.1:2379"
```

```bash tab="CLI"
--providers.etcd.endpoints=127.0.0.1:2379
```

### `rootKey`

Defines the root key of the configuration.

_Required, Default="traefik"_

```toml tab="File (TOML)"
[providers.etcd]
  rootKey = "traefik"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    rootKey: "traefik"
```

```bash tab="CLI"
--providers.etcd.rootkey=traefik
```

### `username`

Defines a username to connect with Etcd.

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.etcd]
  # ...
  username = "foo"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    # ...
    usename: "foo"
```

```bash tab="CLI"
--providers.etcd.username=foo
```

### `password`

_Optional, Default=""_

Defines a password to connect with Etcd.

```toml tab="File (TOML)"
[providers.etcd]
  # ...
  password = "bar"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    # ...
    password: "bar"
```

```bash tab="CLI"
--providers.etcd.password=foo
```

### `tls`

_Optional_

#### `tls.ca`

Certificate Authority used for the secured connection to Etcd.

```toml tab="File (TOML)"
[providers.etcd.tls]
  ca = "path/to/ca.crt"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      ca: path/to/ca.crt
```

```bash tab="CLI"
--providers.etcd.tls.ca=path/to/ca.crt
```

#### `tls.caOptional`

Policy followed for the secured connection with TLS Client Authentication to Etcd.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

```toml tab="File (TOML)"
[providers.etcd.tls]
  caOptional = true
```

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      caOptional: true
```

```bash tab="CLI"
--providers.etcd.tls.caOptional=true
```

#### `tls.cert`

Public certificate used for the secured connection to Etcd.

```toml tab="File (TOML)"
[providers.etcd.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.etcd.tls.cert=path/to/foo.cert
--providers.etcd.tls.key=path/to/foo.key
```

#### `tls.key`

Private certificate used for the secured connection to Etcd.

```toml tab="File (TOML)"
[providers.etcd.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```bash tab="CLI"
--providers.etcd.tls.cert=path/to/foo.cert
--providers.etcd.tls.key=path/to/foo.key
```

#### `tls.insecureSkipVerify`

If `insecureSkipVerify` is `true`, TLS for the connection to Etcd accepts any certificate presented by the server and any host name in that certificate.

```toml tab="File (TOML)"
[providers.etcd.tls]
  insecureSkipVerify = true
```

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      insecureSkipVerify: true
```

```bash tab="CLI"
--providers.etcd.tls.insecureSkipVerify=true
```
