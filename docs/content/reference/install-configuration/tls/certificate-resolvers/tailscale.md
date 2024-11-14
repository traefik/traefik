---
title: "Traefik Tailscale Documentation"
description: "Learn how to configure Traefik Proxy to resolve TLS certificates for your Tailscale services. Read the technical documentation."
---

# Tailscale

Provision TLS certificates for your internal Tailscale services.
{: .subtitle }

To protect a service with TLS, a certificate from a public Certificate Authority is needed.
In addition to its vpn role, Tailscale can also [provide certificates](https://tailscale.com/kb/1153/enabling-https/) for the machines in your Tailscale network.

## Configuration Example

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

??? example "Domain from Router's Rule Example"

    ```yaml tab="Docker & Swarm"
    labels:
      - traefik.http.routers.blog.rule=Host(`monitoring.yak-bebop.ts.net`) && Path(`/metrics`)
      - traefik.http.routers.blog.tls.certresolver=myresolver
    ```

    ```yaml tab="Docker (Swarm)"
    deploy:
      labels:
        - traefik.http.routers.blog.rule=Host(`monitoring.yak-bebop.ts.net`) && Path(`/metrics`)
        - traefik.http.routers.blog.tls.certresolver=myresolver
    ```

    ```yaml tab="Kubernetes"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: blogtls
    spec:
      entryPoints:
        - websecure
      routes:
        - match: Host(`monitoring.yak-bebop.ts.net`) && Path(`/metrics`)
          kind: Rule
          services:
            - name: blog
              port: 8080
      tls:
        certResolver: myresolver
    ```

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        blog:
          rule: "Host(`monitoring.yak-bebop.ts.net`) && Path(`/metrics`)"
          tls:
            certResolver: myresolver
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.blog]
      rule = "Host(`monitoring.yak-bebop.ts.net`) && Path(`/metrics`)"
      [http.routers.blog.tls]
        certResolver = "myresolver"
    ```

??? example "Domain from Router's tls.domain Example"

    ```yaml tab="Docker & Swarm"
    labels:
      - traefik.http.routers.blog.rule=Path(`/metrics`)
      - traefik.http.routers.blog.tls.certresolver=myresolver
      - traefik.http.routers.blog.tls.domains[0].main=monitoring.yak-bebop.ts.net
    ```

    ```yaml tab="Docker (Swarm)"
    deploy:
      labels:
        - traefik.http.routers.blog.rule=Path(`/metrics`)
        - traefik.http.routers.blog.tls.certresolver=myresolver
        - traefik.http.routers.blog.tls.domains[0].main=monitoring.yak-bebop.ts.net
    ```

    ```yaml tab="Kubernetes"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: blogtls
    spec:
      entryPoints:
        - websecure
      routes:
        - match: Path(`/metrics`)
          kind: Rule
          services:
            - name: blog
              port: 8080
      tls:
        certResolver: myresolver
        domains:
          - main: monitoring.yak-bebop.ts.net
    ```

    ```yaml tab="File (YAML)"
    http:
      routers:
        blog:
          rule: "Path(`/metrics`)"
          tls:
            certResolver: myresolver
            domains:
              - main: "monitoring.yak-bebop.ts.net"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.blog]
        rule = "Path(`/metrics`)"
        [http.routers.blog.tls]
          certResolver = "myresolver"
          [[http.routers.blog.tls.domains]]
            main = "monitoring.yak-bebop.ts.net"
    ```

!!! info "Referencing a certificate resolver"

    Defining a certificate resolver does not imply that routers are going to use it automatically.
    Each router or entrypoint that is meant to use the resolver must explicitly [reference](./cert-resolvers.md#acme-certificates) it.

### Tailscale Certificates Renewal

Traefik automatically tracks the expiry date of each Tailscale certificate it fetches and starts to renew a certificate 14 days before its expiry to match the Tailscale daemon renewal policy.
