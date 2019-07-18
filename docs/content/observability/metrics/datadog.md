# DataDog

To enable the DataDog:

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
```

```bash tab="CLI"
--metrics
--metrics.datadog
```

#### `address`

_Required, Default="127.0.0.1:8125"_

Address instructs exporter to send metrics to datadog-agent at this address.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    address = "127.0.0.1:8125"
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    address: 127.0.0.1:8125
```

```bash tab="CLI"
--metrics
--metrics.datadog.address="127.0.0.1:8125"
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    addEntryPointsLabels = true
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    addEntryPointsLabels: true
```

```bash tab="CLI"
--metrics
--metrics.datadog.addEntryPointsLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    addServicesLabels = true
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    addServicesLabels: true
```

```bash tab="CLI"
--metrics
--metrics.datadog.addServicesLabels=true
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to datadog-agent.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    pushInterval = 10s
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    pushInterval: 10s
```

```bash tab="CLI"
--metrics
--metrics.datadog.pushInterval=10s
```

