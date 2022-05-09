---
title: "Traefik HTTP Retry Documentation"
description: "Configure Traefik Proxy's HTTP Retry middleware, so you can retry requests to a backend server until it succeeds. Read the technical documentation."
---

# Retry

Retrying until it Succeeds
{: .subtitle }

<!--
TODO: add schema
-->

The Retry middleware reissues requests a given number of times to a backend server if that server does not reply.
As soon as the server answers, the middleware stops retrying, regardless of the response status.
The Retry middleware has an optional configuration to enable an exponential backoff.

By default the middleware will not retry when the HTTP verb is non-idempotent, check the `retryNonIdempotent` option to disable this behaviour.

## Configuration Examples

```yaml tab="Docker"
# Retry 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
  - "traefik.http.middlewares.test-retry.retry.retryNonIdempotent=true"
```

```yaml tab="Kubernetes"
# Retry 4 times with exponential backoff
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-retry
spec:
  retry:
    attempts: 4
    initialInterval: 100ms
    retryNonIdempotent: true
```

```yaml tab="Consul Catalog"
# Retry 4 times with exponential backoff
- "traefik.http.middlewares.test-retry.retry.attempts=4"
- "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
- "traefik.http.middlewares.test-retry.retry.retryNonIdempotent=true"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-retry.retry.attempts": "4",
  "traefik.http.middlewares.test-retry.retry.initialinterval": "100ms",
  "traefik.http.middlewares.test-retry.retry.retryNonIdempotent": true
}
```

```yaml tab="Rancher"
# Retry 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
  - "traefik.http.middlewares.test-retry.retry.retryNonIdempotent=true"
```

```yaml tab="File (YAML)"
# Retry 4 times with exponential backoff
http:
  middlewares:
    test-retry:
      retry:
        attempts: 4
        initialInterval: 100ms
        retryNonIdempotent: true
```

```toml tab="File (TOML)"
# Retry 4 times with exponential backoff
[http.middlewares]
  [http.middlewares.test-retry.retry]
    attempts = 4
    initialInterval = "100ms"
    retryNonIdempotent = true
```

## Configuration Options

### `attempts`

_mandatory_

The `attempts` option defines how many times the request should be retried.

### `initialInterval`

The `initialInterval` option defines the first wait time in the exponential backoff series. The maximum interval is
calculated as twice the `initialInterval`. If unspecified, requests will be retried immediately.

The value of initialInterval should be provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

### `retryNonIdempotent`

The `retryNonIdempotent` option defines whether the retry should be run against non-idempotent HTTP methods (`POST` and `PATCH`). 
Retrying these requests may lead to issues in the backend depending on your implementation.


If not specified, `retryNonIdempotent` is assumed to be `false`, which means that by default non idempotent http methods will not be retried, even if `attempts` and `initialInterval` are defined