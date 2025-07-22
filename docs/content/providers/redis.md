---
title: "Traefik Redis Documentation"
description: "For configuration discovery in Traefik Proxy, you can store your configurations in Redis. Read the technical documentation."
---

# Traefik & Redis

A Story of KV store & Containers
{: .subtitle }

Store your configuration in Redis and let Traefik do the rest!

!!! tip "Dynamic configuration updates"

    Dynamic configuration updates require Redis [keyspace notifications](https://redis.io/docs/latest/develop/use/keyspace-notifications) to be enabled.
    Cloud-managed Redis services (e.g., GCP Memorystore, AWS ElastiCache) may disable this by default due to CPU performance issues.
    For more information, see the [Redis](https://redis.io/docs/latest/develop/use/keyspace-notifications/) documentation or refer to your cloud provider's documentation for specific configuration steps.

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kv.md).

## Provider Configuration

### `endpoints`

_Required, Default="127.0.0.1:6379"_

Defines how to access Redis.

```yaml tab="File (YAML)"
providers:
  redis:
    endpoints:
      - "127.0.0.1:6379"
```

```toml tab="File (TOML)"
[providers.redis]
  endpoints = ["127.0.0.1:6379"]
```

```bash tab="CLI"
--providers.redis.endpoints=127.0.0.1:6379
```

### `rootKey`

_Required, Default="traefik"_

Defines the root key of the configuration.

```yaml tab="File (YAML)"
providers:
  redis:
    rootKey: "traefik"
```

```toml tab="File (TOML)"
[providers.redis]
  rootKey = "traefik"
```

```bash tab="CLI"
--providers.redis.rootkey=traefik
```

### `username`

_Optional, Default=""_

Defines a username to connect with Redis.

```yaml tab="File (YAML)"
providers:
  redis:
    # ...
    username: "foo"
```

```toml tab="File (TOML)"
[providers.redis]
  # ...
  username = "foo"
```

```bash tab="CLI"
--providers.redis.username=foo
```

### `password`

_Optional, Default=""_

Defines a password to connect with Redis.

```yaml tab="File (YAML)"
providers:
  redis:
    # ...
    password: "bar"
```

```toml tab="File (TOML)"
[providers.redis]
  # ...
  password = "bar"
```

```bash tab="CLI"
--providers.redis.password=foo
```

### `db`

_Optional, Default=0_

Defines the database to be selected after connecting to the Redis.

```yaml tab="File (YAML)"
providers:
  redis:
    # ...
    db: 0
```

```toml tab="File (TOML)"
[providers.redis]
  db = 0
```

```bash tab="CLI"
--providers.redis.db=0
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to Redis.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Redis,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.redis.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.redis.tls.ca=path/to/ca.crt
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to Redis.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.redis.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.redis.tls.cert=path/to/foo.cert
--providers.redis.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to Redis.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.redis.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.redis.tls.cert=path/to/foo.cert
--providers.redis.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Redis accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  redis:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.redis.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.redis.tls.insecureSkipVerify=true
```

### `sentinel`

_Optional_

Defines the Sentinel configuration used to interact with Redis Sentinel.

#### `masterName`

_Required_

`masterName` is the name of the Sentinel master.

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      masterName: my-master
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
  masterName = "my-master"
```

```bash tab="CLI"
--providers.redis.sentinel.masterName=my-master
```

#### `username`

_Optional_

`username` is the username for Sentinel authentication.

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      username: user
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
  username = "user"
```

```bash tab="CLI"
--providers.redis.sentinel.username=user
```

#### `password`

_Optional_

`password` is the password for Sentinel authentication.

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      password: password
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
  password = "password"
```

```bash tab="CLI"
--providers.redis.sentinel.password=password
```

#### `latencyStrategy`

_Optional, Default=false_

`latencyStrategy` defines whether to route commands to the closest master or replica nodes 
(mutually exclusive with RandomStrategy and ReplicaStrategy).

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      latencyStrategy: true
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
latencyStrategy = true
```

```bash tab="CLI"
--providers.redis.sentinel.latencyStrategy=true
```

#### `randomStrategy`

_Optional, Default=false_

`randomStrategy` defines whether to route commands randomly to master or replica nodes 
(mutually exclusive with LatencyStrategy and ReplicaStrategy).

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      randomStrategy: true
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
randomStrategy = true
```

```bash tab="CLI"
--providers.redis.sentinel.randomStrategy=true
```

#### `replicaStrategy`

_Optional, Default=false_

`replicaStrategy` Defines whether to route all commands to replica nodes 
(mutually exclusive with LatencyStrategy and RandomStrategy).

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      replicaStrategy: true
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
replicaStrategy = true
```

```bash tab="CLI"
--providers.redis.sentinel.replicaStrategy=true
```

#### `useDisconnectedReplicas`

_Optional, Default=false_

`useDisconnectedReplicas` defines whether to use replicas disconnected with master when cannot get connected replicas.

```yaml tab="File (YAML)"
providers:
  redis:
    sentinel:
      useDisconnectedReplicas: true
```

```toml tab="File (TOML)"
[providers.redis.sentinel]
useDisconnectedReplicas = true
```

```bash tab="CLI"
--providers.redis.sentinel.useDisconnectedReplicas=true
```
