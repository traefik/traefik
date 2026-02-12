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
        timeout: 60s
        maxRequestBodyBytes: 1024
        status: 500-599
        disableRetryOnNetworkError: true
        retryNonIdempotentMethod: true
```

```toml tab="Structured (TOML)"
# Retry 4 times with exponential backoff
[http.middlewares]
  [http.middlewares.test-retry.retry]
    attempts = 4
    initialInterval = "100ms"
    timeout = "60s"
    maxRequestBodyBytes = 1024
    status = "500-599"
    disableRetryOnNetworkError = true
    retryNonIdempotentMethod = true
```

```yaml tab="Labels"
# Retry 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.initialinterval=100ms"
  - "traefik.http.middlewares.test-retry.retry.timeout=60s"
  - "traefik.http.middlewares.test-retry.retry.maxrequestbodybytes=1024"
  - "traefik.http.middlewares.test-retry.retry.status=500-599"
  - "traefik.http.middlewares.test-retry.retry.disableretryonnetworkerror=true"
  - "traefik.http.middlewares.test-retry.retry.retrynonidempotentmethod=true"
```

```json tab="Tags"
// Retry 4 times with exponential backoff

{
  // ...
  "Tags" : [
    "traefik.http.middlewares.test-retry.retry.attempts=4",
    "traefik.http.middlewares.test-retry.retry.initialinterval=100ms",
    "traefik.http.middlewares.test-retry.retry.timeout=60s",
    "traefik.http.middlewares.test-retry.retry.maxrequestbodybytes=1024",
    "traefik.http.middlewares.test-retry.retry.status=500-599",
    "traefik.http.middlewares.test-retry.retry.disableretryonnetworkerror=true",
    "traefik.http.middlewares.test-retry.retry.retrynonidempotentmethod=true"
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
    timeout: 60s
    maxRequestBodyBytes: 1024
    status: 500-599
    disableRetryOnNetworkError: true
    retryNonIdempotentMethod: true
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-attempts" href="#opt-attempts" title="#opt-attempts">`attempts`</a> | number of times the request should be retried. |  | Yes |
| <a id="opt-initialInterval" href="#opt-initialInterval" title="#opt-initialInterval">`initialInterval`</a> | First wait time in the exponential backoff series. <br />The maximum interval is calculated as twice the `initialInterval`. <br /> If unspecified, requests will be retried immediately.<br /> Defined in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration). | 0 | No |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | How much time the middleware is allowed to retry the request. <br /> Defined in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration). | 0 | No |
| <a id="opt-maxRequestBodyBytes" href="#opt-maxRequestBodyBytes" title="#opt-maxRequestBodyBytes">`maxRequestBodyBytes`</a> | Defines the maximum size for the request body. Default is `-1`, which means no limit. | -1 | No |
| <a id="opt-status" href="#opt-status" title="#opt-status">`status`</a> | Defines the range of HTTP status codes to retry on. | "" | No |
| <a id="opt-disableRetryOnNetworkError" href="#opt-disableRetryOnNetworkError" title="#opt-disableRetryOnNetworkError">`disableRetryOnNetworkError`</a> | This option disables the retry if an error occurs when transmitting the request to the server. | false | No |
| <a id="opt-retryNonIdempotentMethod" href="#opt-retryNonIdempotentMethod" title="#opt-retryNonIdempotentMethod">`retryNonIdempotentMethod`</a> | Activates the retry for non-idempotent methods (`POST`, `LOCK`, `PATH`) | false | No |
