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

#### `onEntryPoints`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    onEntryPoints = true
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    onEntryPoints: true
```

```bash tab="CLI"
--metrics
--metrics.datadog.onEntryPoints=true
```

#### `onServices`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.dataDog]
    onServices = true
```

```yaml tab="File (TOML)"
metrics:
  dataDog:
    onServices: true
```

```bash tab="CLI"
--metrics
--metrics.datadog.onServices=true
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

