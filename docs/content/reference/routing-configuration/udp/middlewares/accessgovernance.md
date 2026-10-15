---
title: "Traefik AccessGovernance Middleware - UDP"
description: "Record UDP session governance and metering events in MySQL."
---

# AccessGovernance for UDP

UDP `accessGovernance` records one event when Traefik's client session ends,
including client/local addresses, duration, bytes received/sent, route identity,
tenant, policy, and billing dimensions. It shares the protocol-neutral network
schema and bounded asynchronous writer described by the
[HTTP middleware](../../http/middlewares/accessgovernance.md).

```yaml
udp:
  middlewares:
    governance:
      accessGovernance:
        defaultDatabase: default
        databases:
          default:
            dsnFile: /run/secrets/mysql-dsn
            networkTable: traefik_network_access_events
        queue:
          capacity: 4096
          workers: 2
          batchSize: 200
          flushInterval: 1s
        schema:
          autoCreate: true

  routers:
    dns:
      entryPoints: [dns]
      middlewares: [governance]
      service: dns
```

An entry-point default ensures that every UDP router is covered:

```yaml
entryPoints:
  dns:
    address: ":53/udp"
    udp:
      middlewares:
        - governance@file
```
