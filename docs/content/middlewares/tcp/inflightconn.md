# InFlightConn

Limiting the Number of Simultaneous connections.
{: .subtitle }

To proactively prevent services from being overwhelmed with high load, the number of allowed simultaneous connections by IP can be limited.

## Configuration Examples

```yaml tab="Docker"
labels:
  - "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10"
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

```yaml tab="Consul Catalog"
# Limiting to 10 simultaneous connections
- "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount": "10"
}
```

```yaml tab="Rancher"
# Limiting to 10 simultaneous connections.
labels:
  - "traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10"
```

```yaml tab="File (YAML)"
# Limiting to 10 simultaneous connections.
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

## Configuration Options

### `amount`

The `amount` option defines the maximum amount of allowed simultaneous connections.
The middleware closes the connection if there are already `amount` connections opened.
