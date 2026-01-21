---
title: "Traefik HTTP TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection for HTTP services in Traefik Proxy. Read the technical documentation."
---

## General

When an HTTP router is configured to handle HTTPS traffic, include a `tls` field in its definition.
This field tells Traefik that the router should process only TLS requests and ignore non-TLS traffic.

By default, an HTTP router with a TLS field will terminate the TLS connections,
meaning that it will send decrypted data to the services.
The TLS configuration provides several options for fine-tuning the TLS behavior,
including automatic certificate generation, custom TLS options, and explicit domain specification.

## Configuration Example

```yaml tab="Structured (YAML)"
http:
  routers:
    my-https-router:
      rule: "Host(`example.com`) && Path(`/api`)"
      service: "my-http-service"
      tls:
        certResolver: "letsencrypt"
        options: "modern-tls"
        domains:
          - main: "example.com"
            sans:
              - "www.example.com"
              - "api.example.com"
```

```toml tab="Structured (TOML)"
[http.routers.my-https-router]
  rule = "Host(`example.com`) && Path(`/api`)"
  service = "my-http-service"

  [http.routers.my-https-router.tls]
    certResolver = "letsencrypt"
    options = "modern-tls"

    [[http.routers.my-https-router.tls.domains]]
      main = "example.com"
      sans = ["www.example.com", "api.example.com"]
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.my-https-router.rule=Host(`example.com`) && Path(`/api`)"
  - "traefik.http.routers.my-https-router.service=my-http-service"
  - "traefik.http.routers.my-https-router.tls=true"
  - "traefik.http.routers.my-https-router.tls.certresolver=letsencrypt"
  - "traefik.http.routers.my-https-router.tls.options=modern-tls"
  - "traefik.http.routers.my-https-router.tls.domains[0].main=example.com"
  - "traefik.http.routers.my-https-router.tls.domains[0].sans=www.example.com,api.example.com"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.http.routers.my-https-router.rule=Host(`example.com`) && Path(`/api`)",
    "traefik.http.routers.my-https-router.service=my-http-service",
    "traefik.http.routers.my-https-router.tls=true",
    "traefik.http.routers.my-https-router.tls.certresolver=letsencrypt",
    "traefik.http.routers.my-https-router.tls.options=modern-tls",
    "traefik.http.routers.my-https-router.tls.domains[0].main=example.com",
    "traefik.http.routers.my-https-router.tls.domains[0].sans=www.example.com,api.example.com"
  ]
}
```

## Configuration Options

| Field                                                                              | Description                                                                                                                                                                                                    | Default   | Required |
|:-----------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|:---------|
| <a id="opt-options" href="#opt-options" title="#opt-options">`options`</a> | The name of the TLS options to use for configuring TLS parameters (cipher suites, min/max TLS version, client authentication, etc.). See [TLS Options](./tls-options.md) for detailed configuration.           | `default` | No       |
| <a id="opt-certResolver" href="#opt-certResolver" title="#opt-certResolver">`certResolver`</a> | The name of the certificate resolver to use for automatic certificate generation via ACME providers (such as Let's Encrypt). See the [Certificate Resolver](./#certificate-resolver) section for more details. | ""        | No       |
| <a id="opt-domains" href="#opt-domains" title="#opt-domains">`domains`</a> | List of domains and Subject Alternative Names (SANs) for explicit certificate domain specification. See the [Custom Domains](./#custom-domains) section for more details.                                      | []        | No       |

## Certificate Resolver

The `tls.certResolver` option allows you to specify a certificate resolver for automatic certificate generation via ACME providers (such as Let's Encrypt).

When a certificate resolver is configured for a router,
Traefik will automatically obtain and manage TLS certificates for the domains specified in the router's rule (in the `Host` matcher) or in the `tls.domains` configuration (with `tls.domains` taking precedence).

!!! important "Prerequisites"

    - Certificate resolvers must be defined in the [static configuration](../../../install-configuration/tls/certificate-resolvers/acme.md)
    - The router must have `tls` enabled
    - An ACME challenge type must be configured for the certificate resolver

## Custom Domains

When using ACME certificate resolvers, domains are automatically extracted from router rules,
but the `tls.domains` option allows you to explicitly specify the domains and Subject Alternative Names (SANs) for which certificates should be generated.

This provides fine-grained control over certificate generation and takes precedence over domains automatically extracted from router rules.

Every domain must have A/AAAA records pointing to Traefik.

{% include-markdown "includes/traefik-for-business-applications.md" %}
