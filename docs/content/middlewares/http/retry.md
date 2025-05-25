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

The Retry middleware reissues requests a given number of times when it cannot contact the backend service. 
This applies at the transport level (TCP). 
If the service does not respond to the initial connection attempt, the middleware retries.
However, once the service responds, regardless of the HTTP status code, the middleware considers it operational and stops retrying.
This means that the retry mechanism does not handle HTTP errors; it only retries when there is no response at the TCP level.
The Retry middleware has an optional configuration to enable an exponential backoff.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Retry 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
```

```yaml tab="Kubernetes"
# Retry 4 times with exponential backoff
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-retry
spec:
  retry:
    attempts: 4
    initialInterval: 100ms
```

```yaml tab="Consul Catalog"
# Retry 4 times with exponential backoff
- "traefik.http.middlewares.test-retry.retry.attempts=4"
- "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
```

```yaml tab="File (YAML)"
# Retry 4 times with exponential backoff
http:
  middlewares:
    test-retry:
      retry:
        attempts: 4
        initialInterval: 100ms
```

```toml tab="File (TOML)"
# Retry 4 times with exponential backoff
[http.middlewares]
  [http.middlewares.test-retry.retry]
    attempts = 4
    initialInterval = "100ms"
```

## Configuration Options

### `attempts`

_mandatory_

The `attempts` option defines how many times the request should be retried.

### `initialInterval`

The `initialInterval` option defines the first wait time in the exponential backoff series. The maximum interval is
calculated as twice the `initialInterval`. If unspecified, requests will be retried immediately.

The value of initialInterval should be provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).
