---
title: "Traefik Etcd Documentation"
description: "Use Etcd as a provider for configuration discovery in Traefik Proxy. Automate and store your configurations with Etcd. Read the technical documentation."
---

# Traefik & Etcd

Store your configuration in etcd and let Traefik do the rest!

## Configuration Example

You can ensble the etcd provider with an endpoint as detailed below:

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

## Configuration Options 

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.etcd.endpoints` | Defines the endpoint to access etcd. |  `127.0.0.1:2379`     | Yes   |
| `providers.etcd.rootKey` | Defines the root key for the configuration. |  `traefik`     | Yes   |
| `providers.etcd.username` | Defines a username with which to connect to etcd. |  ""   | No   |
| `providers.etcd.password` | Defines a password for connecting to etcd. |  ""    | No   |
| `providers.etcd.tls` | Defines the TLS configuration used for the secure connection to etcd. |  N/A    | No   |
| `providers.etcd.tls.ca` | Defines the path to the certificate authority used for the secure connection to etcd, it defaults to the system bundle.  |  N/A   | No   |
| `providers.etcd.tls.cert` | Defines the path to the public certificate used for the secure connection to etcd. When using this option, setting the `key` option is required. |  N/A   | Yes   |
| `providers.etcd.tls.key` | Defines the path to the private key used for the secure connection to etcd. When using this option, setting the `cert` option is required. |  N/A   | Yes   |
| `providers.etcd.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by etcd when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
