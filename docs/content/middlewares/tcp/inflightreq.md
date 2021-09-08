# InFlightReq

Limiting the Number of Simultaneous In-Flight Requests
{: .subtitle }

![InFlightReq](../../assets/img/middleware/inflightreq.png)

To proactively prevent services from being overwhelmed with high load, the number of allowed simultaneous in-flight requests can be limited.

## Configuration Examples

```yaml tab="Docker"
labels:
  - "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-inflightreq
spec:
  inFlightReq:
    amount: 10
```

```yaml tab="Consul Catalog"
# Limiting to 10 simultaneous connections
- "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount": "10"
}
```

```yaml tab="Rancher"
# Limiting to 10 simultaneous connections
labels:
  - "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```yaml tab="File (YAML)"
# Limiting to 10 simultaneous connections
tcp:
  middlewares:
    test-inflightreq:
      inFlightReq:
        amount: 10
```

```toml tab="File (TOML)"
# Limiting to 10 simultaneous connections
[tcp.middlewares]
  [tcp.middlewares.test-inflightreq.inFlightReq]
    amount = 10
```

## Configuration Options

### `amount`

The `amount` option defines the maximum amount of allowed simultaneous in-flight request.
The middleware closes the connection if there are already `amount` requests in progress.

```yaml tab="Docker"
labels:
  - "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-inflightreq
spec:
  inFlightReq:
    amount: 10
```

```yaml tab="Consul Catalog"
# Limiting to 10 simultaneous connections
- "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount": "10"
}
```

```yaml tab="Rancher"
# Limiting to 10 simultaneous connections
labels:
  - "traefik.tcp.middlewares.test-inflightreq.inflightreq.amount=10"
```

```yaml tab="File (YAML)"
# Limiting to 10 simultaneous connections
tcp:
  middlewares:
    test-inflightreq:
      inFlightReq:
        amount: 10
```

```toml tab="File (TOML)"
# Limiting to 10 simultaneous connections
[tcp.middlewares]
  [tcp.middlewares.test-inflightreq.inFlightReq]
    amount = 10
```
