---
title: "Traefik HTTP Documentation"
description: "Provide your dynamic configuration via an HTTP(S) endpoint and let Traefik Proxy do the rest. Read the technical documentation."
---

# Traefik & HTTP

Provide your [install configuration](../overview.md) via an HTTP(S) endpoint and let Traefik do the rest!

## Configuration Example

You can enable the HTTP provider as detailed below:

```yaml tab="File (YAML)"
providers:
  http:
    endpoint: "http://127.0.0.1:9000/api"
```

```toml tab="File (TOML)"
[providers.http]
  endpoint = "http://127.0.0.1:9000/api"
```

```bash tab="CLI"
--providers.http.endpoint=http://127.0.0.1:9000/api
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| <a id="providers-providersThrottleDuration" href="#providers-providersThrottleDuration" title="#providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| <a id="providers-http-endpoint" href="#providers-http-endpoint" title="#providers-http-endpoint">`providers.http.endpoint`</a> | Defines the HTTP(S) endpoint to poll. |  ""    | Yes   |
| <a id="providers-http-pollInterval" href="#providers-http-pollInterval" title="#providers-http-pollInterval">`providers.http.pollInterval`</a> | Defines the polling interval. |  5s    | No   |
| <a id="providers-http-pollTimeout" href="#providers-http-pollTimeout" title="#providers-http-pollTimeout">`providers.http.pollTimeout`</a> | Defines the polling timeout when connecting to the endpoint. |  5s    | No   |
| <a id="providers-http-headers" href="#providers-http-headers" title="#providers-http-headers">`providers.http.headers`</a> | Defines custom headers to be sent to the endpoint. |  ""    | No   |
| <a id="providers-http-tls-ca" href="#providers-http-tls-ca" title="#providers-http-tls-ca">`providers.http.tls.ca`</a> | Defines the path to the certificate authority used for the secure connection to the endpoint, it defaults to the system bundle.  |  ""   | No   |
| <a id="providers-http-tls-cert" href="#providers-http-tls-cert" title="#providers-http-tls-cert">`providers.http.tls.cert`</a> | Defines the path to the public certificate used for the secure connection to the endpoint. When using this option, setting the `key` option is required. |  ""   | Yes   |
| <a id="providers-http-tls-key" href="#providers-http-tls-key" title="#providers-http-tls-key">`providers.http.tls.key`</a> | Defines the path to the private key used for the secure connection to the endpoint. When using this option, setting the `cert` option is required. |  ""  | Yes   |
| <a id="providers-http-tls-insecureSkipVerify" href="#providers-http-tls-insecureSkipVerify" title="#providers-http-tls-insecureSkipVerify">`providers.http.tls.insecureSkipVerify`</a> | Instructs the provider to accept any certificate presented by endpoint when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

### headers

Defines custom headers to be sent to the endpoint.

```yaml tab="File (YAML)"
providers:
  http:
    headers:
      name: value
```

```toml tab="File (TOML)"
[providers.http.headers]
  name = "value"
```

```bash tab="CLI"
[providers.http.headers]
--providers.http.headers.name=value
```

## Routing Configuration

The HTTP provider uses the same configuration as the [File Provider](./file.md) in YAML or JSON format.
