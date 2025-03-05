---
title: "Traefik ContentType Documentation"
description: "Traefik Proxy's HTTP middleware automatically sets the `Content-Type` header value when it is not set by the backend. Read the technical documentation."
---

The `contentType` middleware sets the `Content-Type` header value to the media type detected from the response content,
when it is not set by the backend.

!!! info

    The `contentType` middleware only applies when Traefik detects the MIME type. If any middleware (such as Headers or Compress) sets the `contentType` header at any point in the chain, the `contentType` middleware has no effect.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Enable auto-detection
http:
  middlewares:
    autodetect:
      contentType: {}
```

```toml tab="Structured (TOML)"
# Enable auto-detection
[http.middlewares]
  [http.middlewares.autodetect.contentType]
```

```yaml tab="Labels"
# Enable auto-detection
labels:
  - "traefik.http.middlewares.autodetect.contenttype=true"
```

```json tab="Tags"
// Enable auto-detection
{
  // ...
  "Tags": [
    "traefik.http.middlewares.autodetect.contenttype=true"
  ]
}
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
