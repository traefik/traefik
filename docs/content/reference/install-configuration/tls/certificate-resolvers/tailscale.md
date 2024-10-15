---
title: "Traefik Tailscale Documentation"
description: "Learn how to configure Traefik Proxy to resolve TLS certificates for your Tailscale services. Read the technical documentation."
---

# Tailscale

Provision TLS certificates for your internal Tailscale services.
{: .subtitle }

To protect a service with TLS, a certificate from a public Certificate Authority is needed.
In addition to its vpn role, Tailscale can also [provide certificates](https://tailscale.com/kb/1153/enabling-https/) for the machines in your Tailscale network.

## Certificate resolvers

To obtain a TLS certificate from the Tailscale daemon,
a Tailscale certificate resolver needs to be configured as below.

!!! example "Enabling Tailscale certificate resolution"

    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"

      websecure:
        address: ":443"

    certificatesResolvers:
      myresolver:
        tailscale: {}
    ```

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"

      [entryPoints.websecure]
        address = ":443"

    [certificatesResolvers.myresolver.tailscale]
    ```

    ```bash tab="CLI"
    --entrypoints.web.address=:80
    --entrypoints.websecure.address=:443
    # ...
    --certificatesresolvers.myresolver.tailscale=true
    ```

!!! info "Referencing a certificate resolver"

    Defining a certificate resolver does not imply that routers are going to use it automatically.
    Each router or entrypoint that is meant to use the resolver must explicitly [reference](./cert-resolvers.md#acme-certificates) it.

!!!info "Advanced Configuration"
    The options to set an advanced configuration are described in the [Cert Resolvers reference page](../certificate-resolvers/cert-resolvers.md).
