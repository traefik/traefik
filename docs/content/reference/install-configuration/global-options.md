---
title: "Traefik Global Configuration Documentation"
description: "Learn about global configuration options in Traefik Proxy, including control over X-Forwarded-For header behavior. Read the technical documentation."
---

# Global Configuration

Global Options
{: .subtitle }

Global configuration options affect Traefik's behavior across all entrypoints and services.

## Configuration Example

```yaml tab="File (YAML)"
## Static configuration
global:
  notAppendXForwardedFor: true
```

```toml tab="File (TOML)"
## Static configuration
[global]
  notAppendXForwardedFor = true
```

```bash tab="CLI"
## Static configuration
--global.notAppendXForwardedFor=true
```

## Available Options

### `notAppendXForwardedFor`

_Optional, Default=false_

By default, Traefik automatically appends the client's `RemoteAddr` (IP address and port) to the `X-Forwarded-For` header when forwarding requests to backend services.

When `notAppendXForwardedFor` is set to `true`, Traefik will **not** append the `RemoteAddr` to the `X-Forwarded-For` header.

#### Use Cases

This option is useful when:

- **Cloud Load Balancers**: Your application is behind a cloud provider's load balancer (AWS ALB/NLB, GCP Load Balancer, Azure Load Balancer) that already appends IPs to `X-Forwarded-For`, and you want to preserve only the original client IP without the load balancer's IP.

- **Application Compatibility**: Your backend application only reads the first IP from the `X-Forwarded-For` header and adding additional IPs would cause issues.

- **Migrating from Other Proxies**: When migrating from proxies like NGINX Ingress Controller with `compute-full-forwarded-for: false`, this option provides equivalent behavior.

#### Behavior

**With `notAppendXForwardedFor: false` (default)**:

If a request arrives with:
```
X-Forwarded-For: 203.0.113.1
```

Traefik forwards to the backend with:
```
X-Forwarded-For: 203.0.113.1, 10.0.1.99
```

Where `10.0.1.99` is the `RemoteAddr` (the IP Traefik sees as the request source).

**With `notAppendXForwardedFor: true`**:

If a request arrives with:
```
X-Forwarded-For: 203.0.113.1
```

Traefik forwards to the backend with:
```
X-Forwarded-For: 203.0.113.1
```

The header is preserved as-is without appending the `RemoteAddr`.

If a request arrives **without** an `X-Forwarded-For` header, Traefik will **not add one**.

#### Example Configuration

```yaml tab="File (YAML)"
## Static configuration
global:
  notAppendXForwardedFor: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"
```

```toml tab="File (TOML)"
## Static configuration
[global]
  notAppendXForwardedFor = true

[entryPoints]
  [entryPoints.web]
    address = ":80"
  [entryPoints.websecure]
    address = ":443"
```

```bash tab="CLI"
## Static configuration
--global.notAppendXForwardedFor=true
--entryPoints.web.address=:80
--entryPoints.websecure.address=:443
```