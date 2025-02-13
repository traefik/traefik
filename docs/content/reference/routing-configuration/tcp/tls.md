---
title: "Traefik TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection for TCP services in Traefik Proxy. Read the technical documentation."
---

## General

When a router is configured to handle HTTPS traffic, include a `tls` field in its definition. This field tells Traefik that the router should process only TLS requests and ignore non-TLS traffic.

By default, a router with a TLS field will terminate the TLS connections, meaning that it will send decrypted data to the services.

## Configuration Example

```yaml tab="Structured (YAML)"
tcp:
  routers:
    my-tls-router:
      rule: "HostSNI(`example.com`)"
      service: "my-tcp-service"
      tls:
        passthrough: true
        options: "my-tls-options"
        domains:
          - main: "example.com"
            sans:
              - "www.example.com"
              - "api.example.com"
        certResolver: "myresolver"
```

```toml tab="Structured (TOML)"
[tcp.routers.my-tls-router]
  rule = "HostSNI(`example.com`)"
  service = "my-tcp-service"

  [tcp.routers.my-tls-router.tls]
    passthrough = true
    options = "my-tls-options"
    certResolver = "myresolver"

    [[tcp.routers.my-tls-router.tls.domains]]
      main = "example.com"
      sans = ["www.example.com", "api.example.com"]
```

```yaml tab="Labels"
labels:
  - "traefik.tcp.routers.my-tls-router.tls=true"
  - "traefik.tcp.routers.my-tls-router.rule=HostSNI(`example.com`)"
  - "traefik.tcp.routers.my-tls-router.service=my-tcp-service"
  - "traefik.tcp.routers.my-tls-router.tls.passthrough=true"
  - "traefik.tcp.routers.my-tls-router.tls.options=my-tls-options"
  - "traefik.tcp.routers.my-tls-router.tls.certResolver=myresolver"
  - "traefik.tcp.routers.my-tls-router.tls.domains[0].main=example.com"
  - "traefik.tcp.routers.my-tls-router.tls.domains[0].sans=www.example.com,api.example.com"
```

```json tab="Tags"
{
  //...
  "Tags": [
    "traefik.tcp.routers.my-tls-router.tls=true"
    "traefik.tcp.routers.my-tls-router.rule=HostSNI(`example.com`)",
    "traefik.tcp.routers.my-tls-router.service=my-tcp-service",
    "traefik.tcp.routers.my-tls-router.tls.passthrough=true",
    "traefik.tcp.routers.my-tls-router.tls.options=my-tls-options",
    "traefik.tcp.routers.my-tls-router.tls.certResolver=myresolver",
    "traefik.tcp.routers.my-tls-router.tls.domains[0].main=example.com",
    "traefik.tcp.routers.my-tls-router.tls.domains[0].sans=www.example.com,api.example.com"
  ]
}
```

??? info "Postgres STARTTLS"

    Traefik supports the Postgres STARTTLS protocol,
    which allows TLS routing for Postgres connections.

    To do so, Traefik reads the first bytes sent by a Postgres client,
    identifies if they correspond to the message of a STARTTLS negotiation,
    and, if so, acknowledges and signals the client that it can start the TLS handshake.

    Please note/remember that there are subtleties inherent to STARTTLS in whether the connection ends up being a TLS one or not.
    These subtleties depend on the `sslmode` value in the client configuration (and on the server authentication rules).
    Therefore, it is recommended to use the `require` value for the `sslmode`.

    Afterwards, the TLS handshake, and routing based on TLS, can proceed as expected.

    !!! warning "Postgres STARTTLS with TCP TLS PassThrough routers"

        As mentioned above, the `sslmode` configuration parameter does have an impact on whether a STARTTLS session will succeed.
        In particular in the context of TCP TLS PassThrough, some of the values (such as `allow`) do not even make sense.
        Which is why, once more it is recommended to use the `require` value.

## Configuration Options

| Field   | Description  | Default    | Required |
|:------------------|:--------------------|:-----------------------------------------------|:---------|
|`passthrough`| Defines whether the requests should be forwarded "as is", keeping all data encrypted. | false | No |
|`options`| enables fine-grained control of the TLS parameters. It refers to a [TLS Options](../http/tls/tls-certificates.md#tls-options) and will be applied only if a `HostSNI` rule is defined. | "" | No |
|`domains`| Defines a set of SANs (alternative domains) for each main domain. Every domain must have A/AAAA records pointing to Traefik. Each domain & SAN will lead to a certificate request.| [] | No |
|`certResolver`| If defined, Traefik will try to generate certificates based on routers `Host` & `HostSNI` rules. | "" | No |

{!traefik-for-business-applications.md!}
