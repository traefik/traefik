---
title: "Traefik FastProxy Experimental Configuration"
description: "This section of the Traefik Proxy documentation explains how to use the new FastProxy option."
---

# Traefik FastProxy Experimental Configuration

## Overview

This guide provides instructions on how to configure and use the new experimental `fastProxy` static configuration option in Traefik.
The `fastProxy` option introduces a high-performance reverse proxy designed to enhance the performance of routing.

!!! info "Limitations"

    Please note that the new fast proxy implementation does not work with HTTP/2.
    This means that when a H2C or HTTPS request with [HTTP2 enabled](../routing/services/index.md#disablehttp2) is sent to a backend, the fallback proxy is the regular one.

    Additionnaly, observability features like tracing and OTEL semconv metrics are not supported for the moment.

!!! warning "Experimental"
    
    The `fastProxy` option is currently experimental and subject to change in future releases. 
    Use with caution in production environments.

### Enabling FastProxy

The fastProxy option is a static configuration parameter.
To enable it, you need to configure it in your Traefik static configuration

```yaml tab="File (YAML)"
experimental:
  fastProxy: {}
```

```toml tab="File (TOML)"
[experimental.fastProxy]
```

```bash tab="CLI"
--experimental.fastProxy
```
