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
| <a id="providers-providersThrottleDuration" href="#providers-providersThrottleDuration" title="#providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="providers-consul-endpoints" href="#providers-consul-endpoints" title="#providers-consul-endpoints">`providers.consul.endpoints`</a> | Defines the endpoint to access Consul. |  "127.0.0.1:8500"     | yes   |
| <a id="providers-consul-rootKey" href="#providers-consul-rootKey" title="#providers-consul-rootKey">`providers.consul.rootKey`</a> | Defines the root key of the configuration. |  "traefik"     | yes   |
| <a id="providers-consul-namespaces" href="#providers-consul-namespaces" title="#providers-consul-namespaces">`providers.consul.namespaces`</a> | Defines the namespaces to query. See [here](#namespaces) for more information |  ""     | no   |
| <a id="providers-consul-username" href="#providers-consul-username" title="#providers-consul-username">`providers.consul.username`</a> | Defines a username to connect to Consul with. |  ""     | no   |
| <a id="providers-consul-password" href="#providers-consul-password" title="#providers-consul-password">`providers.consul.password`</a> | Defines a password with which to connect to Consul. |  ""     | no   |
| <a id="providers-consul-token" href="#providers-consul-token" title="#providers-consul-token">`providers.consul.token`</a> | Defines a token with which to connect to Consul. |  ""     | no   |
| <a id="providers-consul-tls" href="#providers-consul-tls" title="#providers-consul-tls">`providers.consul.tls`</a> | Defines the TLS configuration used for the secure connection to Consul  |  -   | No   |
| <a id="providers-consul-tls-ca" href="#providers-consul-tls-ca" title="#providers-consul-tls-ca">`providers.consul.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to Consul, it defaults to the system bundle.  |  -   | Yes   |
| <a id="providers-consul-tls-cert" href="#providers-consul-tls-cert" title="#providers-consul-tls-cert">`providers.consul.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to Consul. When using this option, setting the `key` option is required. |  -  | Yes   |
| <a id="providers-consul-tls-key" href="#providers-consul-tls-key" title="#providers-consul-tls-key">`providers.consul.tls.key`</a> | Defines the path to the private key used for the secure connection to Consul. When using this option, setting the `cert` option is required. |  -   | Yes   |
| <a id="providers-consul-tls-insecureSkipVerify" href="#providers-consul-tls-insecureSkipVerify" title="#providers-consul-tls-insecureSkipVerify">`providers.consul.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by Consul when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

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
