---
title: "Traefik HTTP Retry Documentation"
description: "Configure Traefik Proxy's HTTP Retry middleware, so you can retry requests to a backend server until it succeeds. Read the technical documentation."
---

The `retry` middleware retries requests a given number of times to a backend server if that server does not reply.  
As soon as the server answers, the middleware stops retrying, regardless of the response status.

The Retry middleware has an optional configuration to enable an exponential backoff.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Retry 4 times with exponential backoff
http:
  middlewares:
    test-retry:
      retry:
        attempts: 4
        initialInterval: 100ms
```

```toml tab="Structured (TOML)"
# Retry 4 times with exponential backoff
[http.middlewares]
  [http.middlewares.test-retry.retry]
    attempts = 4
    initialInterval = "100ms"
```

```yaml tab="Labels"
# Retry 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
```

```json tab="Tags"
// Retry 4 times with exponential backoff

{
  // ...
  "Tags" : [
    "traefik.http.middlewares.test-retry.retry.attempts=4",
    "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
  ]
}

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

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| `attempts` | number of times the request should be retried. |  | Yes |
| `initialInterval` | First wait time in the exponential backoff series. <br />The maximum interval is calculated as twice the `initialInterval`. <br /> If unspecified, requests will be retried immediately.<br /> Defined in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration). | 0 | No |
