---
title: "Traefik Etcd Documentation"
description: "Use Etcd as a provider for configuration discovery in Traefik Proxy. Automate and store your configurations with Etcd. Read the technical documentation."
---

# Traefik & Etcd

A Story of KV store & Containers
{: .subtitle }

Store your configuration in etcd and let Traefik do the rest!

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:2379"_

Defines how to access etcd.

```yaml tab="File (YAML)"
providers:
  etcd:
    endpoints:
      - "127.0.0.1:2379"
```

```toml tab="File (TOML)"
[providers.etcd]
  endpoints = ["127.0.0.1:2379"]
```

```bash tab="CLI"
--providers.etcd.endpoints=127.0.0.1:2379
```

### `rootKey`

_Required, Default="traefik"_

Defines the root key of the configuration.

```yaml tab="File (YAML)"
providers:
  etcd:
    rootKey: "traefik"
```

```toml tab="File (TOML)"
[providers.etcd]
  rootKey = "traefik"
```

```bash tab="CLI"
--providers.etcd.rootkey=traefik
```

### `username`

_Optional, Default=""_

Defines a username with which to connect to etcd.

```yaml tab="File (YAML)"
providers:
  etcd:
    # ...
    username: "foo"
```

```toml tab="File (TOML)"
[providers.etcd]
  # ...
  username = "foo"
```

```bash tab="CLI"
--providers.etcd.username=foo
```

### `password`

_Optional, Default=""_

Defines a password with which to connect to etcd.

```yaml tab="File (YAML)"
providers:
  etcd:
    # ...
    password: "bar"
```

```toml tab="File (TOML)"
[providers.etcd]
  # ...
  password = "bar"
```

```bash tab="CLI"
--providers.etcd.password=foo
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to etcd.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to etcd,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.etcd.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.etcd.tls.ca=path/to/ca.crt
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to etcd.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.etcd.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.etcd.tls.cert=path/to/foo.cert
--providers.etcd.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to etcd.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.etcd.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.etcd.tls.cert=path/to/foo.cert
--providers.etcd.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to etcd accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  etcd:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.etcd.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.etcd.tls.insecureSkipVerify=true
```
