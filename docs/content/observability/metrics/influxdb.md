# InfluxDB

To enable the InfluxDB:

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
```

```yaml tab="File (YAML)"
metrics:
  influxDB: {}
```

```bash tab="CLI"
--metrics.influxdb=true
```

#### `address`

_Required, Default="localhost:8089"_

Address instructs exporter to send metrics to influxdb at this address.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    address = "localhost:8089"
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    address: localhost:8089
```

```bash tab="CLI"
--metrics.influxdb.address="localhost:8089"
```

#### `protocol`

_Required, Default="udp"_

InfluxDB's address protocol (udp or http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    protocol = "udp"
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    protocol: udp
```

```bash tab="CLI"
--metrics.influxdb.protocol="udp"
```

#### `database`

_Optional, Default=""_

InfluxDB database used when protocol is http.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    database = ""
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    database: ""
```

```bash tab="CLI"
--metrics.influxdb.database=""
```

#### `retentionPolicy`

_Optional, Default=""_

InfluxDB retention policy used when protocol is http.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    retentionPolicy = ""
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    retentionPolicy: ""
```

```bash tab="CLI"
--metrics.influxdb.retentionPolicy=""
```

#### `username`

_Optional, Default=""_

InfluxDB username (only with http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    username = ""
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    username: ""
```

```bash tab="CLI"
--metrics.influxdb.username=""
```

#### `password`

_Optional, Default=""_

InfluxDB password (only with http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    password = ""
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    password: ""
```

```bash tab="CLI"
--metrics.influxdb.password=""
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    addEntryPointsLabels = true
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    addEntryPointsLabels: true
```

```bash tab="CLI"
--metrics.influxdb.addEntryPointsLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    addServicesLabels = true
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    addServicesLabels: true
```

```bash tab="CLI"
--metrics.influxdb.addServicesLabels=true
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to influxdb.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB]
    pushInterval = 10s
```

```yaml tab="File (YAML)"
metrics:
  influxDB:
    pushInterval: 10s
```

```bash tab="CLI"
--metrics.influxdb.pushInterval=10s
```
