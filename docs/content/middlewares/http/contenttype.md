---
title: "Traefik ContentType Documentation"
description: "Traefik Proxy's HTTP middleware automatically sets the `Content-Type` header value when it is not set by the backend. Read the technical documentation."
---

# ContentType

Handling Content-Type auto-detection
{: .subtitle }

The Content-Type middleware sets the `Content-Type` header value to the media type detected from the response content,
when it is not set by the backend.

!!! info

    The scope of the Content-Type middleware is the MIME type detection done by the core of Traefik (the server part).
    Therefore, it has no effect against any other `Content-Type` header modifications (e.g.: in another middleware such as compress).

## Configuration Examples

```yaml tab="Docker & Swarm"
# Enable auto-detection
labels:
  - "traefik.http.middlewares.autodetect.contenttype=true"
```

```yaml tab="Kubernetes"
# Enable auto-detection
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: autodetect
spec:
  contentType: {}
```

```yaml tab="Consul Catalog"
# Enable auto-detection
- "traefik.http.middlewares.autodetect.contenttype=true"
```

```yaml tab="File (YAML)"
# Enable auto-detection
http:
  middlewares:
    autodetect:
      contentType: {}
```

```toml tab="File (TOML)"
# Enable auto-detection
[http.middlewares]
  [http.middlewares.autodetect.contentType]
```

## Configuration Options

### `autoDetect`

!!! warning

    `autoDetect` option is deprecated and should not be used.
    Moreover, it is redundant with an empty ContentType middleware declaration.

`autoDetect` specifies whether to let the `Content-Type` header,
if it has not been set by the backend,
be automatically set to a value derived from the contents of the response.
