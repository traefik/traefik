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

!!! info ""

    The OpenTelemetry trace reporter will export traces to the collector by using HTTP by default,
    see the [GRPC Section](#grpc-configuration) to use GRPC.

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

#### `address`

_Required, Default="https://localhost:4318/v1/metrics"_

Address instructs exporter to send metrics to OpenTelemetry at this address.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    address: https://localhost:4318/v1/metrics
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    address = "https://localhost:4318/v1/metrics"
```

```bash tab="CLI"
--metrics.openTelemetry.address=https://localhost:4318/v1/metrics
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
--metrics.openTelemetry.addrouterslabels=true
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

#### `compress`

_Optional, Default=false_

Allows reporter to send metrics to the OpenTelemetry Collector using gzip compression.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    compress: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    compress = true
```

```bash tab="CLI"
--metrics.openTelemetry.compress=true
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
    foo = bar
    baz = buz
```

```bash tab="CLI"
--metrics.openTelemetry.headers.foo=bar --metrics.openTelemetry.headers.baz=buz
```

#### `pushInterval`

_Optional, Default=10s_

The interval used by the exporter to push metrics to OpenTelemetry.
The interval value must be greater than zero.

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

#### `pushTimeout`

_Optional, Default=10s_

Timeout defines how long to wait on an idle session before releasing the related resources
when pushing metrics to OpenTelemetry.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    pushTimeout: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    pushTimeout = "10s"
```

```bash tab="CLI"
--metrics.openTelemetry.pushTimeout=10s
```

#### `retry`

_Optional_

Enable retries when the reporter sends metrics to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    retry: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.retry]
```

```bash tab="CLI"
--metrics.openTelemetry.retry=true
```

##### `initialInterval`

_Optional, Default=5s_

The time to wait after the first failure before retrying.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    retry:
      initialInterval: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.retry]
    initialInterval = "10s"
```

```bash tab="CLI"
--metrics.openTelemetry.retry.initialInterval=10s
```

##### `maxElapsedTime`

_Optional, Default=1m_

The maximum amount of time (including retries) spent trying to send a request/batch.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    retry:
      maxElapsedTime: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.retry]
    maxElapsedTime = "10s"
```

```bash tab="CLI"
--metrics.openTelemetry.retry.maxElapsedTime=10s
```

##### `maxInterval`

_Optional, Default=30s_

The upper bound on backoff interval.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    retry:
      maxInterval: 10s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.retry]
    maxInterval = "10s"
```

```bash tab="CLI"
--metrics.openTelemetry.retry.maxInterval=10s
```

#### `timeout`

_Optional, Default="10s"_

The max waiting time for the backend to process each metrics batch.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    timeout: 3s
```

```toml tab="File (TOML)"
[metrics]
  [metric.openTelemetry]
    timeout = "3s"
```

```bash tab="CLI"
--metrics.openTelemetry.timeout=3s
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

##### `caOptional`

_Optional_

The value of `caOptional` defines which policy should be used for the secure connection with TLS Client Authentication the OpenTelemetry Collector.

!!! warning ""

    If `ca` is undefined, this option will be ignored,
    and no client certificate will be requested during the handshake.
    Any provided certificate will thus never be verified.

When this option is set to `true`, a client certificate is requested during the handshake but is not required.
If a certificate is sent, it is required to be valid.

When this option is set to `false`, a client certificate is requested during the handshake,
and at least one valid certificate should be sent by the client.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    tls:
      caOptional: true
```

```toml tab="File (TOML)"
[metrics.openTelemetry.tls]
  caOptional = true
```

```bash tab="CLI"
--metrics.openTelemetry.tls.caOptional=true
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


#### `withMemory`

_Optional, Default=false_

Controls whether the processor remembers metric instruments and label sets that were previously reported.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    withMemory: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    withMemory = true
```

```bash tab="CLI"
--metrics.openTelemetry.withMemory=true
```

#### GRPC configuration

This instructs the reporter to send metrics to the OpenTelemetry Collector using GRPC:

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

##### `insecure`

_Optional, Default=false_

Allows reporter to send metrics to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry.grpc]
    insecure = true
```

```bash tab="CLI"
--metrics.openTelemetry.grpc.insecure=true
```

##### `reconnectionPeriod`

_Optional_

The minimum amount of time between connection attempts to the target endpoint.

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    grpc:
      reconnectionPeriod: 30s
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    [metrics.openTelemetry.grpc]
      reconnectionPeriod = "30s"
```

```bash tab="CLI"
--metrics.openTelemetry.grpc.reconnectionPeriod=30s
```

##### `serviceConfig`

_Optional_

Defines the JSON representation of the default gRPC service config used.

For more information about service configurations,
see: [https://github.com/grpc/grpc/blob/master/doc/service_config.md](https://github.com/grpc/grpc/blob/master/doc/service_config.md)

```yaml tab="File (YAML)"
metrics:
  openTelemetry:
    grpc:
      serviceConfig: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.openTelemetry]
    [metrics.openTelemetry.grpc]
      serviceConfig = "{}"
```

```bash tab="CLI"
--metrics.openTelemetry.grpc.serviceConfig={}
```