# Retry

Retrying until it Succeeds
{: .subtitle }

<!--
TODO: add schema
-->

The Retry middleware is in charge of reissuing a request a given number of times to a backend server if that server does not reply.
To be clear, as soon as the server answers, the middleware stops retrying, regardless of the response status.

## Configuration Examples

```yaml tab="Docker"
# Retry to send request 4 times
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
```

```yaml tab="Kubernetes"
# Retry to send request 4 times
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-retry
spec:
  retry:
    attempts: 4
```

```yaml tab="Consul Catalog"
# Retry to send request 4 times
- "traefik.http.middlewares.test-retry.retry.attempts=4"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-retry.retry.attempts": "4"
}
```

```yaml tab="Rancher"
# Retry to send request 4 times
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
```

```toml tab="File (TOML)"
# Retry to send request 4 times
[http.middlewares]
  [http.middlewares.test-retry.retry]
     attempts = 4
```

```yaml tab="File (YAML)"
# Retry to send request 4 times
http:
  middlewares:
    test-retry:
      retry:
       attempts: 4
```

## Configuration Options

### `attempts` 

_mandatory_

The `attempts` option defines how many times the request should be retried.
