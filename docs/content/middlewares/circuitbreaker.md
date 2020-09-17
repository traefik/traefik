# CircuitBreaker

Don't Waste Time Calling Unhealthy Services
{: .subtitle }

![CircuitBreaker](../assets/img/middleware/circuitbreaker.png) 

The circuit breaker protects your system from stacking requests to unhealthy services (resulting in cascading failures).

When your system is healthy, the circuit is closed (normal operations).
When your system becomes unhealthy, the circuit becomes open and the requests are no longer forwarded (but handled by a fallback mechanism).

To assess if your system is healthy, the circuit breaker constantly monitors the services. 

!!! note ""

    - The CircuitBreaker only analyses what happens _after_ it is positioned in the middleware chain. What happens _before_ has no impact on its state.
    - The CircuitBreaker only affects the routers that use it. Routers that don't use the CircuitBreaker won't be affected by its state.

!!! important

    Each router will eventually gets its own instance of a given circuit breaker.
    
    If two different routers refer to the same circuit breaker definition, they will get one instance each.
    It means that one circuit breaker can be open while the other stays closed: their state is not shared.
    
    This is the expected behavior, we want you to be able to define what makes a service healthy without having to declare a circuit breaker for each route.

## Configuration Examples

```yaml tab="Docker"
# Latency Check
labels:
  - "traefik.http.middlewares.latency-check.circuitbreaker.expression=LatencyAtQuantileMS(50.0) > 100"
```

```yaml tab="Kubernetes"
# Latency Check
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: latency-check
spec:
  circuitBreaker:
    expression: LatencyAtQuantileMS(50.0) > 100
```

```yaml tab="Consul Catalog"
# Latency Check
- "traefik.http.middlewares.latency-check.circuitbreaker.expression=LatencyAtQuantileMS(50.0) > 100"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.latency-check.circuitbreaker.expression": "LatencyAtQuantileMS(50.0) > 100"
}
```

```yaml tab="Rancher"
# Latency Check
labels:
  - "traefik.http.middlewares.latency-check.circuitbreaker.expression=LatencyAtQuantileMS(50.0) > 100"
```

```toml tab="File (TOML)"
# Latency Check
[http.middlewares]
  [http.middlewares.latency-check.circuitBreaker]
    expression = "LatencyAtQuantileMS(50.0) > 100"
```

```yaml tab="File (YAML)"
# Latency Check
http:
  middlewares:
    latency-check:
      circuitBreaker:
        expression: "LatencyAtQuantileMS(50.0) > 100"
```

## Possible States

There are three possible states for your circuit breaker:

- Closed (your service operates normally)
- Open (the fallback mechanism takes over your service)
- Recovering (the circuit breaker tries to resume normal operations by progressively sending requests to your service)

### Closed

While the circuit is closed, the circuit breaker only collects metrics to analyze the behavior of the requests.

At specified intervals (`checkPeriod`), it will evaluate `expression` to decide if its state must change. 

### Open

While open, the fallback mechanism takes over the normal service calls for a duration of `FallbackDuration`.
After this duration, it will enter the recovering state.

### Recovering

While recovering, the circuit breaker will progressively send requests to your service again (in a linear way, for `RecoveryDuration`).
If your service fails during recovery, the circuit breaker becomes open again.
If the service operates normally during the whole recovering duration, then the circuit breaker returns to close.

## Configuration Options

### Configuring the Trigger

You can specify an `expression` that, once matched, will trigger the circuit breaker (and apply the fallback mechanism instead of calling your services).

The `expression` can check three different metrics:

- The network error ratio (`NetworkErrorRatio`)
- The status code ratio (`ResponseCodeRatio`)
- The latency at quantile, in milliseconds (`LatencyAtQuantileMS`)

#### `NetworkErrorRatio`

If you want the circuit breaker to trigger at a 30% ratio of network errors, the expression will be `NetworkErrorRatio() > 0.30`

#### `ResponseCodeRatio`

You can trigger the circuit breaker based on the ratio of a given range of status codes.

The `ResponseCodeRatio` accepts four parameters, `from`, `to`, `dividedByFrom`, `dividedByTo`.

The operation that will be computed is sum(`to` -> `from`) / sum (`dividedByFrom` -> `dividedByTo`).

!!! note ""
    If sum (`dividedByFrom` -> `dividedByTo`) equals 0, then `ResponseCodeRatio` returns 0.
    
    `from`is inclusive, `to` is exclusive. 

For example, the expression `ResponseCodeRatio(500, 600, 0, 600) > 0.25` will trigger the circuit breaker if 25% of the requests returned a 5XX status (amongst the request that returned a status code from 0 to 5XX). 

#### `LatencyAtQuantileMS`

You can trigger the circuit breaker when a given proportion of your requests become too slow.

For example, the expression `LatencyAtQuantileMS(50.0) > 100` will trigger the circuit breaker when the median latency (quantile 50) reaches 100MS.

!!! note ""

    You must provide a float number (with the trailing .0) for the quantile value
 
#### Using multiple metrics

You can combine multiple metrics using operators in your expression.

Supported operators are:

- AND (`&&`)
- OR (`||`)

For example, `ResponseCodeRatio(500, 600, 0, 600) > 0.30 || NetworkErrorRatio() > 0.10` triggers the circuit breaker when 30% of the requests return a 5XX status code, or when the ratio of network errors reaches 10%. 

#### Operators

Here is the list of supported operators:

- Greater than (`>`)
- Greater or equal than (`>=`)
- Lesser than (`<`)
- Lesser or equal than (`<=`)
- Equal (`==`)
- Not Equal (`!=`)

### Fallback mechanism

The fallback mechanism returns a `HTTP 503 Service Unavailable` to the client (instead of calling the target service).
This behavior cannot be configured. 

### `CheckPeriod`

The interval used to evaluate `expression` and decide if the state of the circuit breaker must change.
By default, `CheckPeriod` is 100ms. This value cannot be configured.

### `FallbackDuration`

By default, `FallbackDuration` is 10 seconds. This value cannot be configured.

### `RecoveringDuration`

The duration of the recovering mode (recovering state). 

By default, `RecoveringDuration` is 10 seconds. This value cannot be configured.  
