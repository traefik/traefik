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
        status: ["400","500-599"]
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
    status = ["400","500-599"]
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
  - "traefik.http.middlewares.test-retry.retry.status=400,500-599"
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
    "traefik.http.middlewares.test-retry.retry.status=400,500-599",
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
    status: ["400","500-599"]
    disableRetryOnNetworkError: true
    retryNonIdempotentMethod: true
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-attempts" href="#opt-attempts" title="#opt-attempts">`attempts`</a> | number of times the request should be retried. |  | Yes |
| <a id="opt-initialInterval" href="#opt-initialInterval" title="#opt-initialInterval">`initialInterval`</a> | First wait time in the exponential backoff series. <br />The maximum interval is calculated as twice the `initialInterval`. <br /> If unspecified, requests will be retried immediately.<br /> Defined in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration). | 0 | No |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | How much time the middleware is allowed to retry the request. <br /> Defined in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration). | 0 | No |
| <a id="opt-maxRequestBodyBytes" href="#opt-maxRequestBodyBytes" title="#opt-maxRequestBodyBytes">`maxRequestBodyBytes`</a> | Defines the maximum size for the request body. Default is `-1`, which means no limit. <br/>More information [here](#maxrequestbodybytes). | -1 | No |
| <a id="opt-status" href="#opt-status" title="#opt-status">`status`</a> | Defines the range of HTTP status codes to retry on. <br/>More information [here](#disableretryonnetworkerror-and-status). | "" | No |
| <a id="opt-disableRetryOnNetworkError" href="#opt-disableRetryOnNetworkError" title="#opt-disableRetryOnNetworkError">`disableRetryOnNetworkError`</a> | This option disables the retry if an error occurs when transmitting the request to the server. <br/>More information [here](#disableretryonnetworkerror-and-status).  | false | No |
| <a id="opt-retryNonIdempotentMethod" href="#opt-retryNonIdempotentMethod" title="#opt-retryNonIdempotentMethod">`retryNonIdempotentMethod`</a> | Activates the retry for non-idempotent methods (`POST`, `LOCK`, `PATCH`) | false | No |

### maxRequestBodyBytes

The `maxRequestBodyBytes` option controls the maximum size of request bodies that will be sent to the server.

**⚠️ Important Security Consideration**

By default, `maxRequestBodyBytes` is not set (value: -1), which means request body size is unlimited. This can have significant security and performance implications:

- **Security Risk**: Attackers can send extremely large request bodies, potentially causing DoS attacks or memory exhaustion
- **Performance Impact**: Large request bodies consume memory and processing resources, affecting overall system performance
- **Resource Consumption**: Unlimited body size can lead to unexpected resource usage patterns

**Recommended Configuration**

It is strongly recommended to set an appropriate `maxRequestBodyBytes` value for your use case:

```yaml
# For most web applications (1MB limit)
maxRequestBodyBytes: 1048576  # 1MB in bytes

# For API endpoints expecting larger payloads (10MB limit)  
maxRequestBodyBytes: 10485760  # 10MB in bytes

# For file upload authentication (100MB limit)
maxRequestBodyBytes: 104857600  # 100MB in bytes
```

**Guidelines for Setting `maxRequestBodyBytes`**

- **Web Forms**: 1-5MB is typically sufficient for most form submissions
- **API Endpoints**: Consider your largest expected JSON/XML payload + buffer
- **File Uploads**: Set based on your maximum expected file size
- **High-Traffic Services**: Use smaller limits to prevent resource exhaustion

## disableRetryOnNetworkError and status

The `disableRetryOnNetworkError` option disables the retry if an error occurs when transmitting the request to the server, at the TCP layer.
However, if you want to retry only for specific HTTP status codes, you can configure the `status` option with the relevant status codes to retry on.

If `disableRetryOnNetworkError` is set to `true`, you must define the `status` option. Otherwise, the middleware will raise a configuration error.
