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
| <a id="opt-providers-providersThrottleDuration" href="#opt-providers-providersThrottleDuration" title="#opt-providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="opt-providers-zooKeeper-endpoints" href="#opt-providers-zooKeeper-endpoints" title="#opt-providers-zooKeeper-endpoints">`providers.zooKeeper.endpoints`</a> | Defines the endpoint to access ZooKeeper. |  "127.0.0.1:2181"     | Yes   |
| <a id="opt-providers-zooKeeper-rootKey" href="#opt-providers-zooKeeper-rootKey" title="#opt-providers-zooKeeper-rootKey">`providers.zooKeeper.rootKey`</a> | Defines the root key for the configuration. |  "traefik"   | Yes   |
| <a id="opt-providers-zooKeeper-username" href="#opt-providers-zooKeeper-username" title="#opt-providers-zooKeeper-username">`providers.zooKeeper.username`</a> | Defines a username with which to connect to zooKeeper. |  ""   | No   |
| <a id="opt-providers-zooKeeper-password" href="#opt-providers-zooKeeper-password" title="#opt-providers-zooKeeper-password">`providers.zooKeeper.password`</a> | Defines a password for connecting to zooKeeper. |  ""    | No   |

## Routing Configuration

See the dedicated section in [routing](../../../../reference/routing-configuration/other-providers/kv.md).
