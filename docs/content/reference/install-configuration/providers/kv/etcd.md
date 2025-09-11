---
title: "Traefik Etcd Documentation"
description: "Use etcd as a provider for configuration discovery in Traefik Proxy. Automate and store your configurations with Etcd. Read the technical documentation."
---

# Traefik & etcd

## Configuration Example

You can enable the etcd provider as detailed below:

```yaml tab="File (YAML)"
providers:
  etcd: {}
```

```toml tab="File (TOML)"
[providers.etcd]
```

```bash tab="CLI"
--providers.etcd=true
```

## Configuration Options 

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| <a id="providers-providersThrottleDuration" href="#providers-providersThrottleDuration" title="#providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="providers-etcd-endpoints" href="#providers-etcd-endpoints" title="#providers-etcd-endpoints">`providers.etcd.endpoints`</a> | Defines the endpoint to access etcd. |  "127.0.0.1:2379"     | Yes   |
| <a id="providers-etcd-rootKey" href="#providers-etcd-rootKey" title="#providers-etcd-rootKey">`providers.etcd.rootKey`</a> | Defines the root key for the configuration. |  "traefik"   | Yes   |
| <a id="providers-etcd-username" href="#providers-etcd-username" title="#providers-etcd-username">`providers.etcd.username`</a> | Defines a username with which to connect to etcd. |  ""   | No   |
| <a id="providers-etcd-password" href="#providers-etcd-password" title="#providers-etcd-password">`providers.etcd.password`</a> | Defines a password for connecting to etcd. |  ""    | No   |
| <a id="providers-etcd-tls" href="#providers-etcd-tls" title="#providers-etcd-tls">`providers.etcd.tls`</a> | Defines the TLS configuration used for the secure connection to etcd. |  -  | No   |
| <a id="providers-etcd-tls-ca" href="#providers-etcd-tls-ca" title="#providers-etcd-tls-ca">`providers.etcd.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to etcd, it defaults to the system bundle.  | "" | No   |
| <a id="providers-etcd-tls-cert" href="#providers-etcd-tls-cert" title="#providers-etcd-tls-cert">`providers.etcd.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to etcd. When using this option, setting the `key` option is required. | "" | Yes   |
| <a id="providers-etcd-tls-key" href="#providers-etcd-tls-key" title="#providers-etcd-tls-key">`providers.etcd.tls.key`</a> | Defines the path to the private key used for the secure connection to etcd. When using this option, setting the `cert` option is required. | ""  | Yes   |
| <a id="providers-etcd-tls-insecureSkipVerify" href="#providers-etcd-tls-insecureSkipVerify" title="#providers-etcd-tls-insecureSkipVerify">`providers.etcd.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by etcd when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
