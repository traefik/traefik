---
title: "Traefik Redis Documentation"
description: "For configuration discovery in Traefik Proxy, you can store your configurations in Redis. Read the technical documentation."
---

# Traefik & Redis

## Configuration Example

You can enable the Redis provider as detailed below:

```yaml tab="File (YAML)"
providers:
  redis: {}
```

```toml tab="File (TOML)"
[providers.redis]
```

```bash tab="CLI"
--providers.redis.endpoints=true
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.redis.endpoints` | Defines the endpoint to access Redis. |  "127.0.0.1:6379"    | Yes   |
| `providers.redis.rootKey` | Defines the root key for the configuration. |  "traefik"     | Yes   |
| `providers.redis.username` | Defines a username for connecting to Redis. |  ""    | No   |
| `providers.redis.password` | Defines a password for connecting to Redis. |  ""    | No   |
| `providers.redis.db` | Defines the database to be selected after connecting to the Redis. |  0    | No   |
| `providers.redis.tls` | Defines the TLS configuration used for the secure connection to Redis. |  -    | No   |
| `providers.redis.tls.ca` | Defines the path to the certificate authority used for the secure connection to Redis, it defaults to the system bundle.  | "" | No   |
| `providers.redis.tls.cert` | Defines the path to the public certificate used for the secure connection to Redis. When using this option, setting the `key` option is required. |  ""   | Yes   |
| `providers.redis.tls.key` | Defines the path to the private key used for the secure connection to Redis. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| `providers.redis.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by Redis when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |
| `providers.redis.sentinel` | Defines the Sentinel configuration used to interact with Redis Sentinel. | -   | No   |
| `providers.redis.sentinel.masterName` | Defines the name of the Sentinel master. |  ""  | Yes   |
| `providers.redis.sentinel.username` | Defines the username for Sentinel authentication. | "" | No   |
| `providers.redis.sentinel.password` | Defines the password for Sentinel authentication. | "" | No   |
| `providers.redis.sentinel.latencyStrategy` | Defines whether to route commands to the closest master or replica nodes (mutually exclusive with RandomStrategy and ReplicaStrategy). | false   | No   |
| `providers.redis.sentinel.randomStrategy` | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| `providers.redis.sentinel.replicaStrategy` | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| `providers.redis.sentinel.useDisconnectedReplicas` | Defines whether to use replicas disconnected with master when cannot get connected replicas. | false   | false   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
