---
title: "Traefik RedirectScheme Documentation"
description: "In Traefik Proxy's HTTP middleware, RedirectScheme redirects clients to different schemes/ports. Read the technical documentation."
---

# RedirectScheme

Redirecting the Client to a Different Scheme/Port
{: .subtitle }

<!--
TODO: add schema
-->

The RedirectScheme middleware redirects the request if the request scheme is different from the configured scheme.
The middleware redirects for requests with `http` and `https` schemes only. 

The middleware gets the request scheme by analyzing the request. The `X-Forwarded-Proto` header set by a previous
network hop is used first, then the scheme from the URL is a fallback.

!!! warning "Security concerns using `X-Forwarded-Proto` header"
    When the redirection is based on the `X-Forwarded-Proto` header, the previous hop should be [trusted](../../routing/entrypoints.md#forwarded-headers).
    
    Taking into account a non-trusted hop header can expose to security vulnerabilities.
    It by-passes the redirection and forwards the request to the service defined in the router.

## Configuration Examples

```yaml tab="Docker"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```yaml tab="Kubernetes"
# Redirect to https
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-redirectscheme
spec:
  redirectScheme:
    scheme: https
    permanent: true
```

```yaml tab="Consul Catalog"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme": "https"
  "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent": "true"
}
```

```yaml tab="Rancher"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```yaml tab="File (YAML)"
# Redirect to https
http:
  middlewares:
    test-redirectscheme:
      redirectScheme:
        scheme: https
        permanent: true
```

```toml tab="File (TOML)"
# Redirect to https
[http.middlewares]
  [http.middlewares.test-redirectscheme.redirectScheme]
    scheme = "https"
    permanent = true
```

## Configuration Options

### `permanent`

Set the `permanent` option to `true` to apply a permanent redirection.

```yaml tab="Docker"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```yaml tab="Kubernetes"
# Redirect to https
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-redirectscheme
spec:
  redirectScheme:
    # ...
    permanent: true
```

```yaml tab="Consul Catalog"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```json tab="Marathon"
"labels": {

  "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent": "true"
}
```

```yaml tab="Rancher"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.permanent=true"
```

```yaml tab="File (YAML)"
# Redirect to https
http:
  middlewares:
    test-redirectscheme:
      redirectScheme:
        # ...
        permanent: true
```

```toml tab="File (TOML)"
# Redirect to https
[http.middlewares]
  [http.middlewares.test-redirectscheme.redirectScheme]
    # ...
    permanent = true
```

### `scheme`

The `scheme` option defines the scheme of the new URL.

```yaml tab="Docker"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
```

```yaml tab="Kubernetes"
# Redirect to https
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-redirectscheme
spec:
  redirectScheme:
    scheme: https
```

```yaml tab="Consul Catalog"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme": "https"
}
```

```yaml tab="Rancher"
# Redirect to https
labels:
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
```

```yaml tab="File (YAML)"
# Redirect to https
http:
  middlewares:
    test-redirectscheme:
      redirectScheme:
        scheme: https
```

```toml tab="File (TOML)"
# Redirect to https
[http.middlewares]
  [http.middlewares.test-redirectscheme.redirectScheme]
    scheme = "https"
```

### `port`

The `port` option defines the port of the new URL.

```yaml tab="Docker"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.port=443"
```

```yaml tab="Kubernetes"
# Redirect to https
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-redirectscheme
spec:
  redirectScheme:
    # ...
    port: "443"
```

```yaml tab="Consul Catalog"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.port=443"
```

```json tab="Marathon"
"labels": {

  "traefik.http.middlewares.test-redirectscheme.redirectscheme.port": "443"
}
```

```yaml tab="Rancher"
# Redirect to https
labels:
  # ...
  - "traefik.http.middlewares.test-redirectscheme.redirectscheme.port=443"
```

```yaml tab="File (YAML)"
# Redirect to https
http:
  middlewares:
    test-redirectscheme:
      redirectScheme:
        # ...
        port: "443"
```

```toml tab="File (TOML)"
# Redirect to https
[http.middlewares]
  [http.middlewares.test-redirectscheme.redirectScheme]
    # ...
    port = 443
```

!!! info "Port in this configuration is a string, not a numeric value."
