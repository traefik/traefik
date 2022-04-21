---
title: "Traefik ZooKeeper Documentation"
description: "For configuration discovery in Traefik Proxy, you can store your configurations in ZooKeeper. Read the technical documentation."
---

# Traefik & ZooKeeper

A Story of KV Store & Containers
{: .subtitle }

Store your configuration in ZooKeeper and let Traefik do the rest!

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:2181"_

Defines how to access ZooKeeper.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    endpoints:
      - "127.0.0.1:2181"
```

```toml tab="File (TOML)"
[providers.zooKeeper]
  endpoints = ["127.0.0.1:2181"]
```

```bash tab="CLI"
--providers.zookeeper.endpoints=127.0.0.1:2181
```

### `rootKey`

_Required, Default="traefik"_

Defines the root key of the configuration.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    rootKey: "traefik"
```

```toml tab="File (TOML)"
[providers.zooKeeper]
  rootKey = "traefik"
```

```bash tab="CLI"
--providers.zookeeper.rootkey=traefik
```

### `username`

_Optional, Default=""_

Defines a username to connect with ZooKeeper.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    # ...
    username: "foo"
```

```toml tab="File (TOML)"
[providers.zooKeeper]
  # ...
  username = "foo"
```

```bash tab="CLI"
--providers.zookeeper.username=foo
```

### `password`

_Optional, Default=""_

Defines a password to connect with ZooKeeper.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    # ...
    password: "bar"
```

```toml tab="File (TOML)"
[providers.zooKeeper]
  # ...
  password = "bar"
```

```bash tab="CLI"
--providers.zookeeper.password=foo
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to ZooKeeper.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to ZooKeeper,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.zooKeeper.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.zookeeper.tls.ca=path/to/ca.crt
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to ZooKeeper.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.zooKeeper.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.zookeeper.tls.cert=path/to/foo.cert
--providers.zookeeper.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to ZooKeeper.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.zooKeeper.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.zookeeper.tls.cert=path/to/foo.cert
--providers.zookeeper.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Zookeeper accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  zooKeeper:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.zooKeeper.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.zookeeper.tls.insecureSkipVerify=true
```
