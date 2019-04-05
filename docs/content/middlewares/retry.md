# Retry

Retrying until it Succeeds
{: .subtitle }

`TODO: add schema`

Retry to send request on attempt failure.

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

```toml tab="File"
# Retry to send request 4 times
[http.middlewares]
  [http.middlewares.test-retry.Retry]
     attempts = 4
```

## Configuration Options

### `attempts` 

(_mandatory_)

The `attempts` option defines how many times to try sending the request.