---
title: "Traefik Proxy  HTTPS & TLS Overview |Traefik Docs"
description: "Traefik supports HTTPS & TLS, which concerns roughly two parts of the configuration: routers, and the TLS connection. Read the documentation to learn more."
---

# Introduction

## Configuring TLS Certificates

When configuring TLS in Traefik, there are two main components:

- The routers that handle TLS traffic,
- The TLS certificates that are served.

Routers in Traefik are responsible for directing incoming requests to the appropriate services.
To enable TLS on a router, you specify the TLS configuration within the router’s definition.
This setup tells Traefik that the router should handle requests using the TLS protocol, thus ensuring that the data is encrypted.

Providing TLS certificates is essential for establishing secure connections.
Traefik supports several methods for managing these certificates:

- **User-defined certificates** - Provided using the file provider or Kubernetes Secrets in the Traefik dynamic configuration.

    ```yaml tab="File (YAML)"
    # Dynamic configuration

    tls:
      certificates:
        - certFile: /path/to/domain.cert
          keyFile: /path/to/domain.key
        - certFile: /path/to/other-domain.cert
          keyFile: /path/to/other-domain.key
    ```

    ```toml tab="File (TOML)"
    # Dynamic configuration

    [[tls.certificates]]
      certFile = "/path/to/domain.cert"
      keyFile = "/path/to/domain.key"

    [[tls.certificates]]
      certFile = "/path/to/other-domain.cert"
      keyFile = "/path/to/other-domain.key"
    ```

- **Automated certificates** - Traefik supports the Automated Certificate Management Environment ([ACME](https://en.wikipedia.org/wiki/Automatic_Certificate_Management_Environment "Link to ACME Wikipedia page")).
ACME allows Hub API Gateway to automatically obtain and renew TLS certificates from Certificate Authorities like [Let's Encrypt](https://letsencrypt.org/ "Link to the official Let's Encrypt website").
This automation simplifies certificate management and ensures that certificates are always up-to-date.

## Managing TLS Certificates

Traefik stores TLS certificates together.
 
For each incoming connection, Traefik is serving the _best_ matching TLS certificate for the provided [Server Name Indication (SNI)](https://www.cloudflare.com/learning/ssl/what-is-sni/).

The TLS certificate selection process narrows down the list of TLS certificates matching the server name,
and then selects the last TLS certificate in this list after having ordered it by the identifier alphabetically.

While Traefik is serving the best matching TLS certificate for each incoming connection, the selection process cost for each incoming connection is avoided thanks to a cache mechanism.
Once a TLS certificate has been selected as the _best_ TLS certificate for a server name, it is cached for an hour, avoiding the selection process for further connections.
Nonetheless, when a new configuration is applied, the cache is reset.

!!! note Self-Signed Default TLS Certificate
    If no TLS certificate can be served, Traefik serves a self-signed certificate by default.

{!traefik-for-business-applications.md!}
