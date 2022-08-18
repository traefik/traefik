---
title: "Traefik Proxy  HTTPS & TLS Overview |Traefik Docs"
description: "Traefik supports HTTPS & TLS, which concerns roughly two parts of the configuration: routers, and the TLS connection. Read the documentation to learn more."
---

# HTTPS & TLS

Overview
{: .subtitle }

Traefik supports HTTPS & TLS, which concerns roughly two parts of the configuration:
routers, and the TLS connection (and its underlying certificates).

When a router has to handle HTTPS traffic,
it should be specified with a `tls` field of the router definition.
See the TLS section of the [routers documentation](../routing/routers/index.md#tls).

The next sections of this documentation explain how to configure the TLS connection itself.
That is to say, how to obtain [TLS certificates](./tls.md#certificates-definition):
either through a definition in the dynamic configuration, or through [Let's Encrypt](./acme.md) (ACME).
And how to configure [TLS options](./tls.md#tls-options), and [certificates stores](./tls.md#certificates-stores).

!!! question "Using Traefik for Business Applications?"

    If you are using Traefik for commercial applications,
    consider the [Enterprise Edition](https://traefik.io/traefik-enterprise/).
    You can use it as your:

    - [Kubernetes Ingress Controller](https://traefik.io/solutions/kubernetes-ingress/)
    - [Load Balancer](https://traefik.io/solutions/docker-swarm-ingress/)
    - [API Gateway](https://traefik.io/solutions/api-gateway/)

    Traefik Enterprise enables centralized access management,
    distributed Let's Encrypt,
    and other advanced capabilities.
    Learn more in [this 15-minute technical walkthrough](https://info.traefik.io/watch-traefikee-demo).
