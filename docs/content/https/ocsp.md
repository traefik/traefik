---
title: "Traefik OCSP Documentation"
description: "Learn how to configure Traefik to use OCSP. Read the technical documentation."
---

# OCSP

Check certificate status and perform OCSP stapling.
{: .subtitle }

## Overview

### OCSP Stapling

When OCSP is enabled, Traefik checks the status of every certificate in the store that provides an OCSP responder URL,
including the default certificate, and staples the OCSP response to the TLS handshake.
The OCSP check is performed when the certificate is loaded,
and once every hour until it is successful at the halfway point before the update date.

### Caching

Traefik caches the OCSP response as long as the associated certificate is provided by the configuration.
When a certificate is no longer provided,
the OCSP response has a 24 hour TTL waiting to be provided again or eventually removed.
The OCSP response is cached in memory and is not persisted between Traefik restarts.

## Configuration

### General

Enabling OCSP is part of the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).
It can be defined by using a file (YAML or TOML) or CLI arguments:

```yaml tab="File (YAML)"
## Static configuration
ocsp: {}
```

```toml tab="File (TOML)"
## Static configuration
[ocsp]
```

```bash tab="CLI"
## Static configuration
--ocsp=true
```

### Responder Overrides

The `responderOverrides` option defines the OCSP responder URLs to use instead of the one provided by the certificate.
This is useful when you want to use a different OCSP responder.

```yaml tab="File (YAML)"
## Static configuration
ocsp:
  responderOverrides:
    foo: bar
```

```toml tab="File (TOML)"
## Static configuration
[ocsp]
  [ocsp.responderOverrides]
    foo = "bar"
```

```bash tab="CLI"
## Static configuration
-ocsp.responderoverrides.foo=bar
```
