---
title: "Traefik RedirectTrailingSlash Documentation"
description: "In Traefik Proxy's HTTP middleware, RedirectTrailingSlash redirects clients by adding or removing a trailing slash from the request path. Read the technical documentation."
---

The `RedirectTrailingSlash` middleware redirects requests to add or remove a trailing slash from the URL path.

The root path (`/`) is never redirected.
Query parameters are preserved during the redirect.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Add a trailing slash
http:
  middlewares:
    add-trailing-slash:
      redirectTrailingSlash:
        mode: add
        permanent: true
```

```toml tab="Structured (TOML)"
# Add a trailing slash
[http.middlewares]
  [http.middlewares.add-trailing-slash.redirectTrailingSlash]
    mode = "add"
    permanent = true
```

```yaml tab="Labels"
# Add a trailing slash
labels:
  - "traefik.http.middlewares.add-trailing-slash.redirecttrailingslash.mode=add"
  - "traefik.http.middlewares.add-trailing-slash.redirecttrailingslash.permanent=true"
```

```json tab="Tags"
// Add a trailing slash
{
  // ...
  "Tags": [
    "traefik.http.middlewares.add-trailing-slash.redirecttrailingslash.mode=add",
    "traefik.http.middlewares.add-trailing-slash.redirecttrailingslash.permanent=true"
  ]
}
```

```yaml tab="Kubernetes"
# Add a trailing slash
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: add-trailing-slash
spec:
  redirectTrailingSlash:
    mode: add
    permanent: true
```

## Behavior

| Mode     | Request path  | Redirected to |
|:---------|:--------------|:--------------|
| <a id="opt-add" href="#opt-add" title="#opt-add">`add`</a> | `/about`      | `/about/`     |
| <a id="opt-add-2" href="#opt-add-2" title="#opt-add-2">`add`</a> | `/about/`     | _(no redirect)_ |
| <a id="opt-add-3" href="#opt-add-3" title="#opt-add-3">`add`</a> | `/`           | _(no redirect)_ |
| <a id="opt-remove" href="#opt-remove" title="#opt-remove">`remove`</a> | `/about/`     | `/about`      |
| <a id="opt-remove-2" href="#opt-remove-2" title="#opt-remove-2">`remove`</a> | `/about`      | _(no redirect)_ |
| <a id="opt-remove-3" href="#opt-remove-3" title="#opt-remove-3">`remove`</a> | `/`           | _(no redirect)_ |

Query parameters are preserved: `/search?q=traefik` becomes `/search/?q=traefik` in `add` mode.

For non-GET requests (POST, PUT, etc.), the redirect status codes are **307** (temporary) and **308** (permanent) to preserve the request method.

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-mode" href="#opt-mode" title="#opt-mode">`mode`</a> | Controls whether a trailing slash is `add`ed or `remove`d. | `""` | Yes |
| <a id="opt-permanent" href="#opt-permanent" title="#opt-permanent">`permanent`</a> | Enable a permanent redirect (301/308) instead of temporary (302/307). | `false` | No |

### `mode`

The `mode` option accepts two values:

- `add`: redirects paths without a trailing slash to the same path with a trailing slash. `/about` → `/about/`
- `remove`: redirects paths with a trailing slash to the same path without a trailing slash. `/about/` → `/about`
