# InfluxDB

To enable the InfluxDB:

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
```

```yaml tab="File (YAML)"
metrics:
  influxdb: {}
```

```bash tab="CLI"
--metrics
--metrics.influxdb
```

#### `address`

_Required, Default="localhost:8089"_

Address instructs exporter to send metrics to influxdb at this address.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    address = "localhost:8089"
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    address: localhost:8089
```

```bash tab="CLI"
--metrics
--metrics.influxdb.address="localhost:8089"
```

#### `protocol`

_Required, Default="udp"_

InfluxDB's address protocol (udp or http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    protocol = "upd"
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    protocol: udp
```

```bash tab="CLI"
--metrics
--metrics.influxdb.protocol="udp"
```

#### `database`

_Optional, Default=""_

InfluxDB database used when protocol is http.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    database = ""
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    database: ""
```

```bash tab="CLI"
--metrics
--metrics.influxdb.database=""
```

#### `retentionPolicy`

_Optional, Default=""_

InfluxDB retention policy used when protocol is http.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    retentionPolicy = ""
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    retentionPolicy: ""
```

```bash tab="CLI"
--metrics
--metrics.influxdb.retentionPolicy=""
```

#### `username`

_Optional, Default=""_

InfluxDB username (only with http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    username = ""
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    username: ""
```

```bash tab="CLI"
--metrics
--metrics.influxdb.username=""
```

#### `password`

_Optional, Default=""_

InfluxDB password (only with http).

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    password = ""
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    password: ""
```

```bash tab="CLI"
--metrics
--metrics.influxdb.password=""
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    addEntryPointsLabels = true
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    addEntryPointsLabels: true
```

```bash tab="CLI"
--metrics
--metrics.influxdb.addEntryPointsLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    addServicesLabels = true
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    addServicesLabels: true
```

```bash tab="CLI"
--metrics
--metrics.influxdb.addServicesLabels=true
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to influxdb.

```toml tab="File (TOML)"
[metrics]
  [metrics.influxdb]
    pushInterval = 10s
```

```yaml tab="File (YAML)"
metrics:
  influxdb:
    pushInterval: 10s
```

```bash tab="CLI"
--metrics
--metrics.influxdb.pushInterval=10s
```
