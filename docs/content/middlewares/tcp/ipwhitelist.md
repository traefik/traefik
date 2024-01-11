---
title: "Traefik TCP Middlewares IPWhiteList"
description: "Learn how to use IPWhiteList in TCP middleware for limiting clients to specific IPs in Traefik Proxy. Read the technical documentation."
---

# IPWhiteList

Limiting Clients to Specific IPs
{: .subtitle }

IPWhiteList accepts / refuses connections based on the client IP.

!!! warning

    This middleware is deprecated, please use the [IPAllowList](./ipallowlist.md) middleware instead.

## Configuration Examples

```yaml tab="Docker"
# Accepts connections from defined IP
labels:
  - "traefik.tcp.middlewares.test-ipwhitelist.ipwhitelist.sourcerange=127.0.0.1/32, 192.168.1.7"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ipwhitelist
spec:
  ipWhiteList:
    sourceRange:
      - 127.0.0.1/32
      - 192.168.1.7
```

```yaml tab="Consul Catalog"
# Accepts request from defined IP
- "traefik.tcp.middlewares.test-ipwhitelist.ipwhitelist.sourcerange=127.0.0.1/32, 192.168.1.7"
```

```toml tab="File (TOML)"
# Accepts request from defined IP
[tcp.middlewares]
  [tcp.middlewares.test-ipwhitelist.ipWhiteList]
    sourceRange = ["127.0.0.1/32", "192.168.1.7"]
```

```yaml tab="File (YAML)"
# Accepts request from defined IP
tcp:
  middlewares:
    test-ipwhitelist:
      ipWhiteList:
        sourceRange:
          - "127.0.0.1/32"
          - "192.168.1.7"
```

## Configuration Options

### `sourceRange`

The `sourceRange` option sets the allowed IPs (or ranges of allowed IPs by using CIDR notation).
