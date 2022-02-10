# InfluxDB v2

To enable the InfluxDB2:

```yaml tab="File (YAML)"
metrics:
  influxDB2: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
```

```bash tab="CLI"
--metrics.influxdb2=true
```

#### `address`

_Required, Default="http://localhost:8086"_

Address of the InfluxDB v2 instance.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    address: http://localhost:8086
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    address = "http://localhost:8086"
```

```bash tab="CLI"
--metrics.influxdb2.address=http://localhost:8086
```

#### `token`

_Required, Default=""_

Token with which to connect to InfluxDB v2.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    token: secret
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    token = "secret"
```

```bash tab="CLI"
--metrics.influxdb2.token=secret
```

#### `org`

_Required, Default=""_

Organisation where metrics will be stored.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    org: my-org
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    org = "my-org"
```

```bash tab="CLI"
--metrics.influxdb2.org=my-org
```

#### `bucket`

_Required, Default=""_

Bucket where metrics will be stored.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    bucket: my-bucket
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    bucket = "my-bucket"
```

```bash tab="CLI"
--metrics.influxdb2.bucket=my-bucket
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    addEntryPointsLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    addEntryPointsLabels = true
```

```bash tab="CLI"
--metrics.influxdb2.addEntryPointsLabels=true
```

#### `addRoutersLabels`

_Optional, Default=false_

Enable metrics on routers.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    addRoutersLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    addRoutersLabels = true
```

```bash tab="CLI"
--metrics.influxdb2.addrouterslabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    addServicesLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    addServicesLabels = true
```

```bash tab="CLI"
--metrics.influxdb2.addServicesLabels=true
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to InfluxDB server.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    pushInterval: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    pushInterval = "10s"
```

```bash tab="CLI"
--metrics.influxdb2.pushInterval=10s
```

#### `additionalLabels`

_Optional, Default={}_

Additional labels (InfluxDB tags) on all metrics.

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    additionalLabels:
      host: example.com
      environment: production
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    [metrics.influxDB2.additionalLabels]
      host = "example.com"
      environment = "production"
```

```bash tab="CLI"
--metrics.influxdb2.additionallabels.host=example.com --metrics.influxdb2.additionallabels.environment=production
```
