---
title: "Traefik TLS Documentation"
description: "Learn how to configure the transport layer security (TLS) connection in Traefik Proxy. Read the technical documentation."
---

Traefik’s TLS configuration defines how to handle TLS negotiation for incoming connections. TLS settings are primarily configured on routers, where you specify certificates and security options for secure communication..

When a router has to handle HTTPS traffic, it should be specified with a `tls` field of the router definition.

The next section of this documentation explains how to configure TLS connections through a definition in the dynamic configuration and how to configure TLS options, and certificates stores.

{!traefik-for-business-applications.md!}
