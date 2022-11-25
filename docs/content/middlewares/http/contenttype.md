---
title: "Traefik ContentType Documentation"
description: "Traefik Proxy's HTTP middleware automatically specifies the content-type header if it has not been defined by the backend. Read the technical documentation."
---

# ContentType

Handling Content-Type auto-detection
{: .subtitle }

The Content-Type middleware enables the `Content-Type` header auto-detection,
if it has not been defined by the backend.
The `Content-Type` header will be automatically set to a value derived from the content of the response.

!!! info

    The scope of the Content-Type middleware is the MIME type detection done by the core of Traefik (the server part).
    Therefore, it has no effect against any other `Content-Type` header modifications (e.g.: in another middleware such as compress).

## Configuration Examples

```yaml tab="Docker"
# Enable auto-detection
labels:
  - "traefik.http.middlewares.autodetect.contenttype=true"
```

```yaml tab="Kubernetes"
# Enable auto-detection
apiVersion: traefik.containo.us/v1alpha1
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

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.autodetect.contenttype": {}
}
```

```yaml tab="Rancher"
# Enable auto-detection
labels:
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