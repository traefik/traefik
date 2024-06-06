---
title: "Traefik TCP Middlewares IPAllowList"
description: "Learn how to use IPAllowList in TCP middleware for limiting clients to specific IPs in Traefik Proxy. Read the technical documentation."
---

# IPAllowList

Limiting Clients to Specific IPs
{: .subtitle }

IPAllowList limits allowed requests based on the client IP.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Accepts connections from defined IP
labels:
  - "traefik.tcp.middlewares.test-ipallowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ipallowlist
spec:
  ipAllowList:
    sourceRange:
      - 127.0.0.1/32
      - 192.168.1.7
```

```yaml tab="Consul Catalog"
# Accepts request from defined IP
- "traefik.tcp.middlewares.test-ipallowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7"
```

```toml tab="File (TOML)"
# Accepts request from defined IP
[tcp.middlewares]
  [tcp.middlewares.test-ipallowlist.ipAllowList]
    sourceRange = ["127.0.0.1/32", "192.168.1.7"]
```

```yaml tab="File (YAML)"
# Accepts request from defined IP
tcp:
  middlewares:
    test-ipallowlist:
      ipAllowList:
        sourceRange:
          - "127.0.0.1/32"
          - "192.168.1.7"
```

## Configuration Options

### `sourceRange`

The `sourceRange` option sets the allowed IPs (or ranges of allowed IPs by using CIDR notation).
