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
| <a id="opt-providers-providersThrottleDuration" href="#opt-providers-providersThrottleDuration" title="#opt-providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="opt-providers-redis-endpoints" href="#opt-providers-redis-endpoints" title="#opt-providers-redis-endpoints">`providers.redis.endpoints`</a> | Defines the endpoint to access Redis. |  "127.0.0.1:6379"    | Yes   |
| <a id="opt-providers-redis-rootKey" href="#opt-providers-redis-rootKey" title="#opt-providers-redis-rootKey">`providers.redis.rootKey`</a> | Defines the root key for the configuration. |  "traefik"     | Yes   |
| <a id="opt-providers-redis-username" href="#opt-providers-redis-username" title="#opt-providers-redis-username">`providers.redis.username`</a> | Defines a username for connecting to Redis. |  ""    | No   |
| <a id="opt-providers-redis-password" href="#opt-providers-redis-password" title="#opt-providers-redis-password">`providers.redis.password`</a> | Defines a password for connecting to Redis. |  ""    | No   |
| <a id="opt-providers-redis-db" href="#opt-providers-redis-db" title="#opt-providers-redis-db">`providers.redis.db`</a> | Defines the database to be selected after connecting to the Redis. |  0    | No   |
| <a id="opt-providers-redis-tls" href="#opt-providers-redis-tls" title="#opt-providers-redis-tls">`providers.redis.tls`</a> | Defines the TLS configuration used for the secure connection to Redis. |  -    | No   |
| <a id="opt-providers-redis-tls-ca" href="#opt-providers-redis-tls-ca" title="#opt-providers-redis-tls-ca">`providers.redis.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to Redis, it defaults to the system bundle.  | "" | No   |
| <a id="opt-providers-redis-tls-cert" href="#opt-providers-redis-tls-cert" title="#opt-providers-redis-tls-cert">`providers.redis.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to Redis. When using this option, setting the `key` option is required. |  ""   | Yes   |
| <a id="opt-providers-redis-tls-key" href="#opt-providers-redis-tls-key" title="#opt-providers-redis-tls-key">`providers.redis.tls.key`</a> | Defines the path to the private key used for the secure connection to Redis. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| <a id="opt-providers-redis-tls-insecureSkipVerify" href="#opt-providers-redis-tls-insecureSkipVerify" title="#opt-providers-redis-tls-insecureSkipVerify">`providers.redis.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by Redis when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |
| <a id="opt-providers-redis-sentinel" href="#opt-providers-redis-sentinel" title="#opt-providers-redis-sentinel">`providers.redis.sentinel`</a> | Defines the Sentinel configuration used to interact with Redis Sentinel. | -   | No   |
| <a id="opt-providers-redis-sentinel-masterName" href="#opt-providers-redis-sentinel-masterName" title="#opt-providers-redis-sentinel-masterName">`providers.redis.sentinel.masterName`</a> | Defines the name of the Sentinel master. |  ""  | Yes   |
| <a id="opt-providers-redis-sentinel-username" href="#opt-providers-redis-sentinel-username" title="#opt-providers-redis-sentinel-username">`providers.redis.sentinel.username`</a> | Defines the username for Sentinel authentication. | "" | No   |
| <a id="opt-providers-redis-sentinel-password" href="#opt-providers-redis-sentinel-password" title="#opt-providers-redis-sentinel-password">`providers.redis.sentinel.password`</a> | Defines the password for Sentinel authentication. | "" | No   |
| <a id="opt-providers-redis-sentinel-latencyStrategy" href="#opt-providers-redis-sentinel-latencyStrategy" title="#opt-providers-redis-sentinel-latencyStrategy">`providers.redis.sentinel.latencyStrategy`</a> | Defines whether to route commands to the closest master or replica nodes (mutually exclusive with RandomStrategy and ReplicaStrategy). | false   | No   |
| <a id="opt-providers-redis-sentinel-randomStrategy" href="#opt-providers-redis-sentinel-randomStrategy" title="#opt-providers-redis-sentinel-randomStrategy">`providers.redis.sentinel.randomStrategy`</a> | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| <a id="opt-providers-redis-sentinel-replicaStrategy" href="#opt-providers-redis-sentinel-replicaStrategy" title="#opt-providers-redis-sentinel-replicaStrategy">`providers.redis.sentinel.replicaStrategy`</a> | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| <a id="opt-providers-redis-sentinel-useDisconnectedReplicas" href="#opt-providers-redis-sentinel-useDisconnectedReplicas" title="#opt-providers-redis-sentinel-useDisconnectedReplicas">`providers.redis.sentinel.useDisconnectedReplicas`</a> | Defines whether to use replicas disconnected with master when cannot get connected replicas. | false   | false   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
