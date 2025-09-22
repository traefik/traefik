---
title: "Traefik OpenTelemetry Documentation"
description: "Traefik supports several metrics backends, including OpenTelemetry. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# OpenTelemetry

To enable the OpenTelemetry metrics:

```yaml tab="File (YAML)"
metrics:
  otlp: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
```

```bash tab="CLI"
--metrics.otlp=true
```

!!! info "Default protocol"

    The OpenTelemetry exporter will export metrics to the collector using HTTPS by default to https://localhost:4318/v1/metrics, see the [gRPC Section](#grpc-configuration) to use gRPC.

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```yaml tab="File (YAML)"
metrics:
  otlp:
    addEntryPointsLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    addEntryPointsLabels = true
```

```bash tab="CLI"
--metrics.otlp.addEntryPointsLabels=true
```

#### `addRoutersLabels`

_Optional, Default=false_

Enable metrics on routers.

```yaml tab="File (YAML)"
metrics:
  otlp:
    addRoutersLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    addRoutersLabels = true
```

```bash tab="CLI"
--metrics.otlp.addRoutersLabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```yaml tab="File (YAML)"
metrics:
  otlp:
    addServicesLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    addServicesLabels = true
```

```bash tab="CLI"
--metrics.otlp.addServicesLabels=true
```

#### `explicitBoundaries`

_Optional, Default=".005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10"_

Explicit boundaries for Histogram data points.

```yaml tab="File (YAML)"
metrics:
  otlp:
    explicitBoundaries:
      - 0.1
      - 0.3
      - 1.2
      - 5.0
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    explicitBoundaries = [0.1,0.3,1.2,5.0]
```

```bash tab="CLI"
--metrics.otlp.explicitBoundaries=0.1,0.3,1.2,5.0
```

#### `pushInterval`

_Optional, Default=10s_

Interval at which metrics are sent to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  otlp:
    pushInterval: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    pushInterval = "10s"
```

```bash tab="CLI"
--metrics.otlp.pushInterval=10s
```

#### `serviceName`

_Optional, Default="traefik"_

Defines the service name resource attribute.

```yaml tab="File (YAML)"
metrics:
  otlp:
    serviceName: name
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
    serviceName = "name"
```

```bash tab="CLI"
--metrics.otlp.serviceName=name
```
#### `resourceAttributes`

_Optional, Default=empty_

Defines additional resource attributes to be sent to the collector.

```yaml tab="File (YAML)"
metrics:
  otlp:
    resourceAttributes:
      attr1: foo
      attr2: bar
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.resourceAttributes]
    attr1 = "foo"
    attr2 = "bar"
```

```bash tab="CLI"
--metrics.otlp.resourceAttributes.attr1=foo
--metrics.otlp.resourceAttributes.attr2=bar
```

### HTTP configuration

_Optional_

This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.http]
```

```bash tab="CLI"
--metrics.otlp.http=true
```

#### `endpoint`

_Optional, Default="https://localhost:4318/v1/metrics", Format="`<scheme>://<host>:<port><path>`"_

URL of the OpenTelemetry Collector to send metrics to.

!!! info "Insecure mode"

    To disable TLS, use `http://` instead of `https://` in the `endpoint` configuration.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      endpoint: https://collector:4318/v1/metrics
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.http]
    endpoint = "https://collector:4318/v1/metrics"
```

```bash tab="CLI"
--metrics.otlp.http.endpoint=https://collector:4318/v1/metrics
```

#### `headers`

_Optional, Default={}_

Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.http.headers]
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--metrics.otlp.http.headers.foo=bar --metrics.otlp.http.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send metrics to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[metrics.otlp.http.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--metrics.otlp.http.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.otlp.http.tls.cert=path/to/foo.cert
--metrics.otlp.http.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.otlp.http.tls.cert=path/to/foo.cert
--metrics.otlp.http.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
metrics:
  otlp:
    http:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[metrics.otlp.http.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--metrics.otlp.http.tls.insecureSkipVerify=true
```

### gRPC configuration

_Optional_

This instructs the exporter to send metrics to the OpenTelemetry Collector using gRPC.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.grpc]
```

```bash tab="CLI"
--metrics.otlp.grpc=true
```

#### `endpoint`

_Required, Default="localhost:4317", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send metrics to.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      endpoint: localhost:4317
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.grpc]
    endpoint = "localhost:4317"
```

```bash tab="CLI"
--metrics.otlp.grpc.endpoint=localhost:4317
```

#### `insecure`

_Optional, Default=false_

Allows exporter to send metrics to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.grpc]
    insecure = true
```

```bash tab="CLI"
--metrics.otlp.grpc.insecure=true
```

#### `headers`

_Optional, Default={}_

Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp.grpc.headers]
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--metrics.otlp.grpc.headers.foo=bar --metrics.otlp.grpc.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send metrics to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[metrics.otlp.grpc.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--metrics.otlp.grpc.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.otlp.grpc.tls.cert=path/to/foo.cert
--metrics.otlp.grpc.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[metrics.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--metrics.otlp.grpc.tls.cert=path/to/foo.cert
--metrics.otlp.grpc.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
metrics:
  otlp:
    grpc:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[metrics.otlp.grpc.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--metrics.otlp.grpc.tls.insecureSkipVerify=true
```
