---
title: "Traefik TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection in Traefik Proxy. Read the technical documentation."
---

Traefik supports HTTPS & TLS, which concerns roughly two parts of the configuration: routers, and the TLS connection (and its underlying certificates).

When a router has to handle HTTPS traffic, it should be specified with a tls field of the router definition.

The next sections of this documentation explain how to configure TLS connections through a definition in the dynamic configuration and how to configure TLS options, and certificates stores.

{!traefik-for-business-applications.md!}
