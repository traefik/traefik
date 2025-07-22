---
title: "Traefik Consul Documentation"
description: "Use Consul as a provider for configuration discovery in Traefik Proxy. Automate and store your configurations with Consul. Read the technical documentation."
---

# Traefik & Consul

## Configuration Example

You can enable the Consul provider as detailed below:

```yaml tab="File (YAML)"
providers:
  consul: {}
```

```toml tab="File (TOML)"
[providers.consul]
```

```bash tab="CLI"
--providers.consul=true
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.consul.endpoints` | Defines the endpoint to access Consul. |  "127.0.0.1:8500"     | yes   |
| `providers.consul.rootKey` | Defines the root key of the configuration. |  "traefik"     | yes   |
| `providers.consul.namespaces` | Defines the namespaces to query. See [here](#namespaces) for more information |  ""     | no   |
| `providers.consul.username` | Defines a username to connect to Consul with. |  ""     | no   |
| `providers.consul.password` | Defines a password with which to connect to Consul. |  ""     | no   |
| `providers.consul.token` | Defines a token with which to connect to Consul. |  ""     | no   |
| `providers.consul.tls` | Defines the TLS configuration used for the secure connection to Consul  |  -   | No   |
| `providers.consul.tls.ca` | Defines the path to the certificate authority used for the secure connection to Consul, it defaults to the system bundle.  |  -   | Yes   |
| `providers.consul.tls.cert` | Defines the path to the public certificate used for the secure connection to Consul. When using this option, setting the `key` option is required. |  -  | Yes   |
| `providers.consul.tls.key` | Defines the path to the private key used for the secure connection to Consul. When using this option, setting the `cert` option is required. |  -   | Yes   |
| `providers.consul.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by Consul when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

### `namespaces`

The `namespaces` option defines the namespaces to query.
When using the `namespaces` option, the discovered configuration object names will be suffixed as shown below:

```text
<resource-name>@consul-<namespace>
```

!!! warning

    The namespaces option only works with [Consul Enterprise](https://www.consul.io/docs/enterprise),
    which provides the [Namespaces](https://www.consul.io/docs/enterprise/namespaces) feature.

!!! warning

    One should only define either the `namespaces` option or the `namespace` option.

```yaml tab="File (YAML)"
providers:
  consul:
    namespaces: 
      - "ns1"
      - "ns2"
    # ...
```

```toml tab="File (TOML)"
[providers.consul]
  namespaces = ["ns1", "ns2"]
  # ...
```

```bash tab="CLI"
--providers.consul.namespaces=ns1,ns2
# ...
```

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kv.md).
