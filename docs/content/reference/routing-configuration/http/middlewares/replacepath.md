---
title: "Traefik ReplacePath Documentation"
description: "In Traefik Proxy's HTTP middleware, ReplacePath updates paths before forwarding requests. Read the technical documentation."
---

The `replacePath` middleware will:

- Replace the actual path with the specified one.
- Store the original path in a `X-Replaced-Path` header

## Configuration Examples

```yaml tab="Structured (YAML)"
# Replace the path with /foo
http:
  middlewares:
    test-replacepath:
      replacePath:
        path: "/foo"
```

```toml tab="Structured (TOML)"
# Replace the path with /foo
[http.middlewares]
  [http.middlewares.test-replacepath.replacePath]
    path = "/foo"
```

```yaml tab="Labels"
# Replace the path with /foo
labels:
  - "traefik.http.middlewares.test-replacepath.replacepath.path=/foo"
```

```json tab="Tags"
// Replace the path with /foo
{
  // ...
  "Tags" : [
    "traefik.http.middlewares.test-replacepath.replacepath.path=/foo"
  ]
} 
```

```yaml tab="Kubernetes"
# Replace the path with /foo
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-replacepath
spec:
  replacePath:
    path: "/foo"
```

## Configuration Options

| Field | Description |
|:------|:------------|
| `path` | The `path` option defines the path to use as replacement in the request URL. |
