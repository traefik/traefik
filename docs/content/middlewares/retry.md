# Retry

Retrying until it Succeeds
{: .subtitle }

<!--
TODO: add schema
-->

The Retry middleware is in charge of reissuing a request a given number of times to a backend server if that server does not reply.
To be clear, as soon as the server answers, the middleware stops retrying, regardless of the response status.
The Retry middleware has an optional configuration for exponential backoff.

## Configuration Examples

```yaml tab="Docker"
# Retry to send request 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.backoff.initialInterval=500ms"
  - "traefik.http.middlewares.test-retry.retry.backoff.maxInterval=2s"
  - "traefik.http.middlewares.test-retry.retry.backoff.multiplier=1.5"
  - "traefik.http.middlewares.test-retry.retry.backoff.randomizationFactor=0.5"
```

```yaml tab="Kubernetes"
# Retry to send request 4 times with exponential backoff
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-retry
spec:
  retry:
    attempts: 4
    backoff:
      initialInterval: 500ms
      maxInterval: 2s
      multiplier: 1.5
      randomizationFactor: 0.5
```

```yaml tab="Consul Catalog"
# Retry to send request 4 times with exponential backoff
- "traefik.http.middlewares.test-retry.retry.attempts=4"
- "traefik.http.middlewares.test-retry.retry.backoff.initialInterval=500ms"
- "traefik.http.middlewares.test-retry.retry.backoff.maxInterval=2s"
- "traefik.http.middlewares.test-retry.retry.backoff.multiplier=1.5"
- "traefik.http.middlewares.test-retry.retry.backoff.randomizationFactor=0.5"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-retry.retry.attempts": "4",
  "traefik.http.middlewares.test-retry.retry.backoff.initialInterval": "500ms",
  "traefik.http.middlewares.test-retry.retry.backoff.maxInterval": "2s",
  "traefik.http.middlewares.test-retry.retry.backoff.multiplier": "1.5",
  "traefik.http.middlewares.test-retry.retry.backoff.randomizationFactor": "0.5"
}
```

```yaml tab="Rancher"
# Retry to send request 4 times with exponential backoff
labels:
  - "traefik.http.middlewares.test-retry.retry.attempts=4"
  - "traefik.http.middlewares.test-retry.retry.backoff.initialInterval=500ms"
  - "traefik.http.middlewares.test-retry.retry.backoff.maxInterval=2s"
  - "traefik.http.middlewares.test-retry.retry.backoff.multiplier=1.5"
  - "traefik.http.middlewares.test-retry.retry.backoff.randomizationFactor=0.5"
```

```toml tab="File (TOML)"
# Retry to send request 4 times
[http.middlewares]
  [http.middlewares.test-retry.retry]
     attempts = 4
    [http.middlewares.test-retry.retry.backoff]
      initialInterval = "500ms"
      maxInterval = "1500ms"
      multiplier = 2
```

```yaml tab="File (YAML)"
# Retry to send request 4 times with exponential backoff
http:
  middlewares:
    test-retry:
      retry:
        attempts: 4
        backoff:
          initialInterval: 500ms
          maxInterval: 2s
          multiplier: 1.5
          randomizationFactor: 0.5
```

## Configuration Options

### `attempts`

_mandatory_

The `attempts` option defines how many times the request should be retried.

(provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration))

### _backoff section_

The backoff functionality of the Retry middleware is activated if any one or more of the below options are set.

### `backoff.initialInterval`

The `backoff.initialInterval` option defines how long to wait before the first retry attempt (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)).

### `backoff.maxInterval`

The `backoff.maxInterval` option defines the limit for low long to wait between any two retry attempts (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)).

### `backoff.multiplier`

The `backoff.multiplier` option defines the factor to multiply by for calculating the next retry's waiting period from the current.

### `backoff.randomizationFactor`

The `backoff.randomizationFactor` option sets a random value in range [1 - RandomizationFactor, 1 + RandomizationFactor] to multiply the backoff duration by. This provides jitter to prevent overloading from coordinated retry.
