---
title: 'Traefik InFlightConn Middleware - TCP'
description: "Limiting the number of simultaneous connections."
---

To proactively prevent Services from being overwhelmed with high load, the number of allowed simultaneous connections by IP can be limited with the `inFlightConn` TCP middleware.

## Configuration Examples

```yaml tab="File (YAML)"
# Limiting to 10 simultaneous connections
tcp:
  middlewares:
    test-inflightconn:
      inFlightConn:
        amount: 10
```

```toml tab="File (TOML)"
# Limiting to 10 simultaneous connections
[tcp.middlewares]
  [tcp.middlewares.test-inflightconn.inFlightConn]
    amount = 10
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-inflightconn
spec:
  inFlightConn:
    amount: 10
```

```yaml tab="Docker & Swarm"
labels:
  - "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10"
```

```yaml tab="Consul Catalog"
# Limiting to 10 simultaneous connections
- "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10"
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|------------------|-------|
| `amount` | The `amount` option defines the maximum amount of allowed simultaneous connections. <br /> The middleware closes the connection if there are already `amount` connections opened. | "" | Yes |
