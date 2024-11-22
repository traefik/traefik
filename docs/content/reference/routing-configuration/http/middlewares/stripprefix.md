---
title: "Traefik StripPrefix Documentation"
description: "In Traefik Proxy's HTTP middleware, StripPrefix removes prefixes from paths before forwarding requests. Read the technical documentation."
---

The `stripPrefix` middleware strips the matching path prefix and stores it in a `X-Forwarded-Prefix` header.

!!! tip

    Use a `StripPrefix` middleware if your backend listens on the root path (`/`) but should be exposed on a specific prefix

## Configuration Examples

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

```yaml tab="Docker & Swarm"
# Strip prefix /foobar and /fiibar
labels:
  - "traefik.http.middlewares.test-stripprefix.stripprefix.prefixes=/foobar,/fiibar"
```

```yaml tab="Consul Catalog"
# Strip prefix /foobar and /fiibar
- "traefik.http.middlewares.test-stripprefix.stripprefix.prefixes=/foobar,/fiibar"
```

## Configuration Options

| Field                        | Description           | Default | Required |
|:-----------------------------|:--------------------------------------------------------------|:--------|:---------|
| `prefixes` | List of prefixes to strip from the request URL.<br />If your backend is serving assets (for example, images or JavaScript files), it can use the `X-Forwarded-Prefix` header to construct relative URLs. | | No |
