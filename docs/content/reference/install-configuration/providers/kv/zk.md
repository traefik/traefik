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
| <a id="opt-providers-zooKeeper-tls" href="#opt-providers-zooKeeper-tls" title="#opt-providers-zooKeeper-tls">`providers.zooKeeper.tls`</a> | Defines the TLS configuration used for the secure connection to zooKeeper. |  -  | No   |
| <a id="opt-providers-zooKeeper-tls-ca" href="#opt-providers-zooKeeper-tls-ca" title="#opt-providers-zooKeeper-tls-ca">`providers.zooKeeper.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to zooKeeper, it defaults to the system bundle.  |  ""   | No   |
| <a id="opt-providers-zooKeeper-tls-cert" href="#opt-providers-zooKeeper-tls-cert" title="#opt-providers-zooKeeper-tls-cert">`providers.zooKeeper.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to zooKeeper. When using this option, setting the `key` option is required. |  ""   | Yes   |
| <a id="opt-providers-zooKeeper-tls-key" href="#opt-providers-zooKeeper-tls-key" title="#opt-providers-zooKeeper-tls-key">`providers.zooKeeper.tls.key`</a> | Defines the path to the private key used for the secure connection to zooKeeper. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| <a id="opt-providers-zooKeeper-tls-insecureSkipVerify" href="#opt-providers-zooKeeper-tls-insecureSkipVerify" title="#opt-providers-zooKeeper-tls-insecureSkipVerify">`providers.zooKeeper.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by etcd when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
