# Prometheus

To enable the Prometheus:

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
```

```yaml tab="File (TOML)"
metrics:
  prometheus: {}
```

```bash tab="CLI"
--metrics
--metrics.prometheus
```

#### `buckets`

_Optional, Default="0.100000, 0.300000, 1.200000, 5.000000"_

Buckets for latency metrics.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    buckets = [0.1,0.3,1.2,5.0]
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    buckets:
    - 0.1
    - 0.3
    - 1.2
    - 5.0
```

```bash tab="CLI"
--metrics
--metrics.prometheus.buckets=0.100000, 0.300000, 1.200000, 5.000000
```

#### `entryPoint`

_Optional, Default=traefik_

Entry-point used by prometheus to expose metrics.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    entryPoint = traefik
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    entryPoint: traefik
```

```bash tab="CLI"
--metrics
--metrics.prometheus.entryPoint=traefik
```

#### `middlewares`

_Optional, Default=""_

Middlewares.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    middlewares = ["xxx", "yyy"]
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    middlewares:
    - xxx
    - yyy
```

```bash tab="CLI"
--metrics
--metrics.prometheus.middlewares="xxx,yyy"
```

#### `onEntryPoints`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    onEntryPoints = true
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    onEntryPoints: true
```

```bash tab="CLI"
--metrics
--metrics.prometheus.onEntryPoints=true
```

#### `onServices`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    onServices = true
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    onServices: true
```

```bash tab="CLI"
--metrics
--metrics.prometheus.onServices=true
```
