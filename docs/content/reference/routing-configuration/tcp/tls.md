---
title: "Traefik TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection for TCP services in Traefik Proxy. Read the technical documentation."
---

## General

When a router is configured to handle HTTPS traffic, include a `tls` field in its definition. This field tells Traefik that the router should process only TLS requests and ignore non-TLS traffic.

By default, a router with a TLS field will terminate the TLS connections, meaning that it will send decrypted data to the services.

## Configuration Example

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  routers:
    Router-1:
      rule: "HostSNI(`foo-domain`)"
      service: service-id
       # Enable TLS termination for all requests.
      tls: {}
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.routers]
  [tcp.routers.Router-1]
    rule = "HostSNI(`foo-domain`)"
    service = "service-id"
     # Enable TLS termination for all requests.
    [tcp.routers.Router-1.tls]
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
|`passthrough`| Defines whether the requests should be forwarded "as is", keeping all data encrypted. More information [here](#passthrough) | false | No |
|`options`| enables fine-grained control of the TLS parameters. It refers to a [TLS Options](../http/tls/tls-certificates.md#tls-options) and will be applied only if a `HostSNI` rule is defined. | "" | No |
|`domains`| Defines a set of SANs (alternative domains) for each main domain. Every domain must have A/AAAA records pointing to Traefik. Each domain & SAN will lead to a certificate request.| [] | No |
|`certResolver`| If defined, Traefik will try to generate certificates based on routers `Host` & `HostSNI` rules. | "" | No |

### `passthrough`

A TLS router will terminate the TLS connection by default.
However, the `passthrough` option can be specified to set whether the requests should be forwarded "as is", keeping all data encrypted.

??? example "Configuring passthrough"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          tls:
            passthrough: true
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        [tcp.routers.Router-1.tls]
          passthrough = true
    ```

{!traefik-for-business-applications.md!}
