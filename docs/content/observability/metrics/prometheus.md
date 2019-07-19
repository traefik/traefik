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

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    addEntryPointsLabels = true
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    addEntryPointsLabels: true
```

```bash tab="CLI"
--metrics
--metrics.prometheus.addEntryPointsLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    addServicesLabels = true
```

```yaml tab="File (TOML)"
metrics:
  prometheus:
    addServicesLabels: true
```

```bash tab="CLI"
--metrics
--metrics.prometheus.addServicesLabels=true
```
