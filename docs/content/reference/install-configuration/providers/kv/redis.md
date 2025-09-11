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
| <a id="providers-providersThrottleDuration" href="#providers-providersThrottleDuration" title="#providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="providers-redis-endpoints" href="#providers-redis-endpoints" title="#providers-redis-endpoints">`providers.redis.endpoints`</a> | Defines the endpoint to access Redis. |  "127.0.0.1:6379"    | Yes   |
| <a id="providers-redis-rootKey" href="#providers-redis-rootKey" title="#providers-redis-rootKey">`providers.redis.rootKey`</a> | Defines the root key for the configuration. |  "traefik"     | Yes   |
| <a id="providers-redis-username" href="#providers-redis-username" title="#providers-redis-username">`providers.redis.username`</a> | Defines a username for connecting to Redis. |  ""    | No   |
| <a id="providers-redis-password" href="#providers-redis-password" title="#providers-redis-password">`providers.redis.password`</a> | Defines a password for connecting to Redis. |  ""    | No   |
| <a id="providers-redis-db" href="#providers-redis-db" title="#providers-redis-db">`providers.redis.db`</a> | Defines the database to be selected after connecting to the Redis. |  0    | No   |
| <a id="providers-redis-tls" href="#providers-redis-tls" title="#providers-redis-tls">`providers.redis.tls`</a> | Defines the TLS configuration used for the secure connection to Redis. |  -    | No   |
| <a id="providers-redis-tls-ca" href="#providers-redis-tls-ca" title="#providers-redis-tls-ca">`providers.redis.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to Redis, it defaults to the system bundle.  | "" | No   |
| <a id="providers-redis-tls-cert" href="#providers-redis-tls-cert" title="#providers-redis-tls-cert">`providers.redis.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to Redis. When using this option, setting the `key` option is required. |  ""   | Yes   |
| <a id="providers-redis-tls-key" href="#providers-redis-tls-key" title="#providers-redis-tls-key">`providers.redis.tls.key`</a> | Defines the path to the private key used for the secure connection to Redis. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| <a id="providers-redis-tls-insecureSkipVerify" href="#providers-redis-tls-insecureSkipVerify" title="#providers-redis-tls-insecureSkipVerify">`providers.redis.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by Redis when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |
| <a id="providers-redis-sentinel" href="#providers-redis-sentinel" title="#providers-redis-sentinel">`providers.redis.sentinel`</a> | Defines the Sentinel configuration used to interact with Redis Sentinel. | -   | No   |
| <a id="providers-redis-sentinel-masterName" href="#providers-redis-sentinel-masterName" title="#providers-redis-sentinel-masterName">`providers.redis.sentinel.masterName`</a> | Defines the name of the Sentinel master. |  ""  | Yes   |
| <a id="providers-redis-sentinel-username" href="#providers-redis-sentinel-username" title="#providers-redis-sentinel-username">`providers.redis.sentinel.username`</a> | Defines the username for Sentinel authentication. | "" | No   |
| <a id="providers-redis-sentinel-password" href="#providers-redis-sentinel-password" title="#providers-redis-sentinel-password">`providers.redis.sentinel.password`</a> | Defines the password for Sentinel authentication. | "" | No   |
| <a id="providers-redis-sentinel-latencyStrategy" href="#providers-redis-sentinel-latencyStrategy" title="#providers-redis-sentinel-latencyStrategy">`providers.redis.sentinel.latencyStrategy`</a> | Defines whether to route commands to the closest master or replica nodes (mutually exclusive with RandomStrategy and ReplicaStrategy). | false   | No   |
| <a id="providers-redis-sentinel-randomStrategy" href="#providers-redis-sentinel-randomStrategy" title="#providers-redis-sentinel-randomStrategy">`providers.redis.sentinel.randomStrategy`</a> | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| <a id="providers-redis-sentinel-replicaStrategy" href="#providers-redis-sentinel-replicaStrategy" title="#providers-redis-sentinel-replicaStrategy">`providers.redis.sentinel.replicaStrategy`</a> | Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy). | false   | No   |
| <a id="providers-redis-sentinel-useDisconnectedReplicas" href="#providers-redis-sentinel-useDisconnectedReplicas" title="#providers-redis-sentinel-useDisconnectedReplicas">`providers.redis.sentinel.useDisconnectedReplicas`</a> | Defines whether to use replicas disconnected with master when cannot get connected replicas. | false   | false   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
