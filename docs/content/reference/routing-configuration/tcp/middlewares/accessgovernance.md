---
title: "Traefik AccessGovernance Middleware - TCP"
description: "Record TCP connection governance and metering events in MySQL."
---

# AccessGovernance for TCP

TCP `accessGovernance` records one event when a proxied connection completes,
including client/local addresses, duration, bytes received/sent, route identity,
tenant, policy, and billing dimensions. It uses the same database routing,
bounded queue, batching, schema management, and network table as the
[HTTP middleware](../../http/middlewares/accessgovernance.md).

```yaml
tcp:
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
```

Attach it to an individual TCP router, or configure it as an entry-point
default:

```yaml
entryPoints:
  postgres:
    address: ":5432"
    tcp:
      middlewares:
        - governance@file
```
