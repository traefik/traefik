# StatsD

To enable the Statsd:

```toml tab="File (TOML)"
[metrics]
  [metrics.statsd]
```

```yaml tab="File (TOML)"
metrics:
  statsd: {}
```

```bash tab="CLI"
--metrics
--metrics.statsd
```

#### `address`

_Required, Default="localhost:8125"_

Address instructs exporter to send metrics to statsd at this address.

```toml tab="File (TOML)"
[metrics]
  [metrics.statsd]
    address = "localhost:8125"
```

```yaml tab="File (TOML)"
metrics:
  statsd:
    address: localhost:8125
```

```bash tab="CLI"
--metrics
--metrics.statsd.address="localhost:8125"
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.statsd]
    addEntryPointsLabels = true
```

```yaml tab="File (TOML)"
metrics:
  statsd:
    addEntryPointsLabels: true
```

```bash tab="CLI"
--metrics
--metrics.statsd.addEntryPointsLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.statsd]
    addServicesLabels = true
```

```yaml tab="File (TOML)"
metrics:
  statsd:
    addServicesLabels: true
```

```bash tab="CLI"
--metrics
--metrics.statsd.addServicesLabels=true
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to statsD.

```toml tab="File (TOML)"
[metrics]
  [metrics.statsd]
    pushInterval = 10s
```

```yaml tab="File (TOML)"
metrics:
  statsd:
    pushInterval: 10s
```

```bash tab="CLI"
--metrics
--metrics.statsd.pushInterval=10s
```
