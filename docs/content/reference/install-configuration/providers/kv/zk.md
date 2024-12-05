---
title: "Traefik ZooKeeper Documentation"
description: "For configuration discovery in Traefik Proxy, you can store your configurations in ZooKeeper. Read the technical documentation."
---

# Traefik & ZooKeeper

## Configuration Example

You can enable the ZooKeeper provider as detailed below:

```yaml tab="File (YAML)"
providers:
  zooKeeper: {}
```

```toml tab="File (TOML)"
[providers.zooKeeper]
```

```bash tab="CLI"
--providers.zookeeper=true
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.zooKeeper.endpoints` | Defines the endpoint to access ZooKeeper. |  "127.0.0.1:2181"     | Yes   |
| `providers.zooKeeper.rootKey` | Defines the root key for the configuration. |  "traefik"   | Yes   |
| `providers.zooKeeper.username` | Defines a username with which to connect to zooKeeper. |  ""   | No   |
| `providers.zooKeeper.password` | Defines a password for connecting to zooKeeper. |  ""    | No   |
| `providers.zooKeeper.tls` | Defines the TLS configuration used for the secure connection to zooKeeper. |  -  | No   |
| `providers.zooKeeper.tls.ca` | Defines the path to the certificate authority used for the secure connection to zooKeeper, it defaults to the system bundle.  |  ""   | No   |
| `providers.zooKeeper.tls.cert` | Defines the path to the public certificate used for the secure connection to zooKeeper. When using this option, setting the `key` option is required. |  ""   | Yes   |
| `providers.zooKeeper.tls.key` | Defines the path to the private key used for the secure connection to zooKeeper. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| `providers.zooKeeper.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by etcd when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
