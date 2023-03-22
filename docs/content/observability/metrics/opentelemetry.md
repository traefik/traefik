---
title: "Traefik OpenTelemetry Documentation"
description: "Traefik supports several metrics backends, including OpenTelemetry. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# OpenTelemetry

To enable the OpenTelemetry:

```yaml tab="File (YAML)"
metrics:
  openTelemetry: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
```

```bash tab="CLI"
--metrics.openTelemetry=true
```

!!! info "The OpenTelemetry exporter will export metrics to the collector by using HTTP by default, see the [gRPC Section](#grpc-configuration) to use gRPC."

#### `address`

_Required, Default="localhost:4318", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send metrics to.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    address: localhost:4318
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    address = "localhost:4318"
```

```bash tab="CLI"
--metrics.openTelemetry.address=localhost:4318
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    addEntryPointsLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    addEntryPointsLabels = true
```

```bash tab="CLI"
--metrics.openTelemetry.addEntryPointsLabels=true
```

#### `addRoutersLabels`

_Optional, Default=false_

Enable metrics on routers.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    addRoutersLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    addRoutersLabels = true
```

```bash tab="CLI"
--metrics.openTelemetry.addRoutersLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    addServicesLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    addServicesLabels = true
```

```bash tab="CLI"
--metrics.openTelemetry.addServicesLabels=true
```

#### `explicitBoundaries`

_Optional, Default=".005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10"_

Explicit boundaries for Histogram data points.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    explicitBoundaries:
      - 0.1
      - 0.3
      - 1.2
      - 5.0
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    explicitBoundaries = [0.1,0.3,1.2,5.0]
```

```bash tab="CLI"
--metrics.openTelemetry.explicitBoundaries=0.1,0.3,1.2,5.0
```

#### `headers`

_Optional, Default={}_

Additional headers sent with metrics by the reporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    headers:
      foo: bar
      baz: buz
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.headers]
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--metrics.openTelemetry.headers.foo=bar --metrics.openTelemetry.headers.baz=buz
```

#### `insecure`

_Optional, Default=false_

Allows reporter to send metrics to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    insecure: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    insecure = true
```

```bash tab="CLI"
--metrics.openTelemetry.insecure=true
```

#### `pushInterval`

_Optional, Default=10s_

Interval at which metrics are sent to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    pushInterval: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    pushInterval = "10s"
```

```bash tab="CLI"
--metrics.openTelemetry.pushInterval=10s
```

#### `path`

_Required, Default="/v1/metrics"_

Allows to override the default URL path used for sending metrics.
This option has no effect when using gRPC transport.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    path: /foo/v1/metrics
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    path = "/foo/v1/metrics"
```

```bash tab="CLI"
--metrics.openTelemetry.path=/foo/v1/metrics
```

#### `tls`

_Optional_

Defines the TLS configuration used by the reporter to send metrics to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[metrics.openTelemetry.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--metrics.openTelemetry.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.openTelemetry.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.openTelemetry.tls.cert=path/to/foo.cert
--metrics.openTelemetry.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.openTelemetry.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.openTelemetry.tls.cert=path/to/foo.cert
--metrics.openTelemetry.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[metrics.openTelemetry.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--metrics.openTelemetry.tls.insecureSkipVerify=true
```

#### gRPC configuration

This instructs the reporter to send metrics to the OpenTelemetry Collector using gRPC.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    grpc: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.grpc]
```

```bash tab="CLI"
--metrics.openTelemetry.grpc=true
```
