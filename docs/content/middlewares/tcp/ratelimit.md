# RateLimit

To Control the Number of Requests Going to a Service
{: .subtitle }

The RateLimit middleware ensures that services will receive a _fair_ amount of requests, and allows one to define what fair is.

## Configuration Example

```yaml tab="Docker"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=50"
```

```yaml tab="Kubernetes"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
apiVersion: traefik.containo.us/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    average: 100
    burst: 50
```

```yaml tab="Consul Catalog"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=50"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.average": "100",
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst": "50"
}
```

```yaml tab="Rancher"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=50"
```

```yaml tab="File (YAML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
tcp:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
        burst: 50
```

```toml tab="File (TOML)"
# Here, an average of 100 requests per second is allowed.
# In addition, a burst of 50 requests is allowed.
[tcp.middlewares]
  [tcp.middlewares.test-ratelimit.rateLimit]
    average = 100
    burst = 50
```

## Configuration Options

### `average`

`average` is the maximum rate, by default in requests per second, allowed from a given source.

It defaults to `0`, which means no rate limiting.

The rate is actually defined by dividing `average` by `period`.
So for a rate below 1 req/s, one needs to define a `period` larger than a second.

```yaml tab="Docker"
# 100 reqs/s
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
```

```yaml tab="Kubernetes"
# 100 reqs/s
apiVersion: traefik.containo.us/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    average: 100
```

```yaml tab="Consul Catalog"
# 100 reqs/s
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.average": "100",
}
```

```yaml tab="Rancher"
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=100"
```

```yaml tab="File (YAML)"
# 100 reqs/s
tcp:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 100
```

```toml tab="File (TOML)"
# 100 reqs/s
[tcp.middlewares]
  [tcp.middlewares.test-ratelimit.rateLimit]
    average = 100
```

### `period`

`period`, in combination with `average`, defines the actual maximum rate, such as:

```go
r = average / period
```

It defaults to `1` second.

```yaml tab="Docker"
# 6 reqs/minute
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=6"
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.period=1m"
```

```yaml tab="Kubernetes"
# 6 reqs/minute
apiVersion: traefik.containo.us/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    period: 1m
    average: 6
```

```yaml tab="Consul Catalog"
# 6 reqs/minute
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=6"
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.period=1m"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.average": "6",
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.period": "1m",
}
```

```yaml tab="Rancher"
# 6 reqs/minute
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.average=6"
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.period=1m"
```

```yaml tab="File (YAML)"
# 6 reqs/minute
tcp:
  middlewares:
    test-ratelimit:
      rateLimit:
        average: 6
        period: 1m
```

```toml tab="File (TOML)"
# 6 reqs/minute
[tcp.middlewares]
  [tcp.middlewares.test-ratelimit.rateLimit]
    average = 6
    period = "1m"
```

### `burst`

`burst` is the maximum number of requests allowed to go through in the same arbitrarily small period of time.

It defaults to `1`.

```yaml tab="Docker"
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=100"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: MiddlewareTCP
metadata:
  name: test-ratelimit
spec:
  rateLimit:
    burst: 100
```

```yaml tab="Consul Catalog"
- "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=100"
```

```json tab="Marathon"
"labels": {
  "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst": "100",
}
```

```yaml tab="Rancher"
labels:
  - "traefik.tcp.middlewares.test-ratelimit.ratelimit.burst=100"
```

```yaml tab="File (YAML)"
tcp:
  middlewares:
    test-ratelimit:
      rateLimit:
        burst: 100
```

```toml tab="File (TOML)"
[tcp.middlewares]
  [tcp.middlewares.test-ratelimit.rateLimit]
    burst = 100
```
