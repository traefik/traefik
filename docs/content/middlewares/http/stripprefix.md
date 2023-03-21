---
title: "Traefik StripPrefix Documentation"
description: "In Traefik Proxy's HTTP middleware, StripPrefix removes prefixes from paths before forwarding requests. Read the technical documentation."
---

# StripPrefix

Removing Prefixes From the Path Before Forwarding the Request
{: .subtitle }

<!--
TODO: add schema
-->

Remove the specified prefixes from the URL path.

## Configuration Examples

```yaml tab="Docker"
# Strip prefix /foobar and /fiibar
labels:
  - "traefik.http.middlewares.test-stripprefix.stripprefix.prefixes=/foobar,/fiibar"
```

```yaml tab="Kubernetes"
# Strip prefix /foobar and /fiibar
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-stripprefix
spec:
  stripPrefix:
    prefixes:
      - /foobar
      - /fiibar
```

```yaml tab="Consul Catalog"
# Strip prefix /foobar and /fiibar
- "traefik.http.middlewares.test-stripprefix.stripprefix.prefixes=/foobar,/fiibar"
```

```yaml tab="File (YAML)"
# Strip prefix /foobar and /fiibar
http:
  middlewares:
    test-stripprefix:
      stripPrefix:
        prefixes:
          - "/foobar"
          - "/fiibar"
```

```toml tab="File (TOML)"
# Strip prefix /foobar and /fiibar
[http.middlewares]
  [http.middlewares.test-stripprefix.stripPrefix]
    prefixes = ["/foobar", "/fiibar"]
```

## Configuration Options

### General

The StripPrefix middleware strips the matching path prefix and stores it in a `X-Forwarded-Prefix` header.

!!! tip

    Use a `StripPrefix` middleware if your backend listens on the root path (`/`) but should be exposed on a specific prefix.

### `prefixes`

The `prefixes` option defines the prefixes to strip from the request URL.

For instance, `/products` also matches `/products/shoes` and `/products/shirts`.

If your backend is serving assets (e.g., images or JavaScript files), it can use the `X-Forwarded-Prefix` header to properly construct relative URLs.
Using the previous example, the backend should return `/products/shoes/image.png` (and not `/image.png`, which Traefik would likely not be able to associate with the same backend).
