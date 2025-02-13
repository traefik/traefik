---
title: "Traefik Buffering Documentation"
description: "The HTTP buffering middleware in Traefik Proxy limits the size of requests that can be forwarded to Services. Read the technical documentation."
---

![Buffering](../../../../assets/img/middleware/buffering.png)

The `buffering` middleware limits the size of requests that can be forwarded to services.

With buffering, Traefik reads the entire request into memory (possibly buffering large requests into disk), and rejects requests that are over a specified size limit.

This can help services avoid large amounts of data (`multipart/form-data` for example), and can minimize the time spent sending data to a Service

## Configuration Examples

```yaml tab="Structured (YAML)"
# Sets the maximum request body to 2MB
http:
  middlewares:
    limit:
      buffering:
        maxRequestBodyBytes: 2000000
```

```toml tab="Structured (TOML)"
# Sets the maximum request body to 2MB
[http.middlewares]
  [http.middlewares.limit.buffering]
    maxRequestBodyBytes = 2000000
```

```yaml tab="Labels"
# Sets the maximum request body to 2MB
labels:
  - "traefik.http.middlewares.limit.buffering.maxRequestBodyBytes=2000000"
```

```json tab="Tags"
// Sets the maximum request body to 2MB
{
  // ...
  "Tags": [
    "traefik.http.middlewares.test-auth.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
  ]
}
```

```yaml tab="Kubernetes"
# Sets the maximum request body to 2MB
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: limit
spec:
  buffering:
    maxRequestBodyBytes: 2000000
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| `maxRequestBodyBytes`  | Maximum allowed body size for the request (in bytes). <br /> If the request exceeds the allowed size, it is not forwarded to the Service, and the client gets a `413` (Request Entity Too Large) response. | 0 | No |
| `memRequestBodyBytes`  | Threshold (in bytes) from which the request will be buffered on disk instead of in memory with the `memRequestBodyBytes` option.| 1048576 | No |
| `maxResponseBodyBytes` | Maximum allowed response size from the Service (in bytes). <br /> If the response exceeds the allowed size, it is not forwarded to the client. The client gets a `500` (Internal Server Error) response instead. | 0 | No |
| `memResponseBodyBytes` | Threshold (in bytes) from which the response will be buffered on disk instead of in memory with the `memResponseBodyBytes` option.| 1048576 | No |
| `retryExpression`      | Replay the request using `retryExpression`.<br /> More information [here](#retryexpression). | "" | No |

### retryExpression

The retry expression is defined as a logical combination of the functions below with the operators AND (`&&`) and OR (`||`).  
At least one function is required:

- `Attempts()` number of attempts (the first one counts).
- `ResponseCode()` response code of the Service.
- `IsNetworkError()` whether the response code is related to networking error.
