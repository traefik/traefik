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

```yaml tab="File (YAML)"
providers:
  consul:
    endpoints:
      - "127.0.0.1:8500"
```

```toml tab="File (TOML)"
[providers.consul]
  endpoints = ["127.0.0.1:8500"]
```

```bash tab="CLI"
--providers.consul.endpoints=127.0.0.1:8500
```

### `rootKey`

_Required, Default="traefik"_

Defines the root key of the configuration.

```yaml tab="File (YAML)"
providers:
  consul:
    rootKey: "traefik"
```

```toml tab="File (TOML)"
[providers.consul]
  rootKey = "traefik"
```

```bash tab="CLI"
--providers.consul.rootkey=traefik
```

### `username`

_Optional, Default=""_

Defines a username to connect to Consul with.

```yaml tab="File (YAML)"
providers:
  consul:
    # ...
    username: "foo"
```

```toml tab="File (TOML)"
[providers.consul]
  # ...
  username = "foo"
```

```bash tab="CLI"
--providers.consul.username=foo
```

### `password`

_Optional, Default=""_

Defines a password with which to connect to Consul.

```yaml tab="File (YAML)"
providers:
  consul:
    # ...
    password: "bar"
```

```toml tab="File (TOML)"
[providers.consul]
  # ...
  password = "bar"
```

```bash tab="CLI"
--providers.consul.password=foo
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to Consul.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Consul,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.consul.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.consul.tls.ca=path/to/ca.crt
```

#### `caOptional`

_Optional_

The value of `caOptional` defines which policy should be used for the secure connection with TLS Client Authentication to Consul.

!!! warning ""

    If `ca` is undefined, this option will be ignored, and no client certificate will be requested during the handshake. Any provided certificate will thus never be verified.

When this option is set to `true`, a client certificate is requested during the handshake but is not required. If a certificate is sent, it is required to be valid.

When this option is set to `false`, a client certificate is requested during the handshake, and at least one valid certificate should be sent by the client.

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      caOptional: true
```

```toml tab="File (TOML)"
[providers.consul.tls]
  caOptional = true
```

```bash tab="CLI"
--providers.consul.tls.caOptional=true
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to Consul.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.consul.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.consul.tls.cert=path/to/foo.cert
--providers.consul.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to Consul.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.consul.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.consul.tls.cert=path/to/foo.cert
--providers.consul.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Consul accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  consul:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.consul.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.consul.tls.insecureSkipVerify=true
```
