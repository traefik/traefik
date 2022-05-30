---
title: "Traefik OpenTelemetry Documentation"
description: "Traefik supports several tracing backends, including OpenTelemetry. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# OpenTelemetry

To enable the OpenTelemetry tracer:

```yaml tab="File (YAML)"
tracing:
  openTelemetry: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
```

```bash tab="CLI"
--tracing.openTelemetry=true
```

!!! info ""

    The OpenTelemetry trace reporter will export traces to the collector by using HTTP by default,
    see the [GRPC Section](#grpc-configuration) to use GRPC.

#### `compress`

_Optional, Default=false_

Allows reporter to send span to the OpenTelemetry Collector using gzip compression.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    compress: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    compress = true
```

```bash tab="CLI"
--tracing.openTelemetry.compress=true
```

#### `endpoint`

_Required, Default="https://localhost:4318/v1/traces"_

This instructs the reporter to send spans to the OpenTelemetry Collector at this address (host:port).

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    endpoint: https://localhost:4318/v1/traces
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    endpoint = "https://localhost:4318/v1/traces"
```

```bash tab="CLI"
--tracing.openTelemetry.endpoint=https://localhost:4318/v1/traces
```

#### `headers`

_Optional, Default={}_

Additional headers sent with spans by the reporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    headers:
      foo: bar
      baz: buz
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.headers]
    foo = bar
    baz = buz
```

```bash tab="CLI"
--tracing.openTelemetry.headers.foo=bar --tracing.openTelemetry.headers.baz=buz
```

#### `retry`

_Optional_

Enable retries when the reporter sends spans to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    retry: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.retry]
```

```bash tab="CLI"
--tracing.openTelemetry.retry=true
```

##### `initialInterval`

_Optional, Default=5s_

The time to wait after the first failure before retrying.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    retry:
      initialInterval: 10s
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.retry]
    initialInterval = "10s"
```

```bash tab="CLI"
--tracing.openTelemetry.retry.initialInterval=10s
```

##### `maxElapsedTime`

_Optional, Default=1m_

The maximum amount of time (including retries) spent trying to send a request/batch.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    retry:
      maxElapsedTime: 10s
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.retry]
    maxElapsedTime = "10s"
```

```bash tab="CLI"
--tracing.openTelemetry.retry.maxElapsedTime=10s
```

##### `maxInterval`

_Optional, Default=30s_

The upper bound on backoff interval.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    retry:
      maxInterval: 10s
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.retry]
    maxInterval = "10s"
```

```bash tab="CLI"
--tracing.openTelemetry.retry.maxInterval=10s
```

#### `timeout`

_Optional, Default="10s"_

The max waiting time for the backend to process each spans batch.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    timeout: 3s
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    timeout = "3s"
```

```bash tab="CLI"
--tracing.openTelemetry.timeout=3s
```

#### `tls`

_Optional_

Defines the TLS configuration used by the reporter to send spans to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[tracing.openTelemetry.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--tracing.openTelemetry.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.openTelemetry.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.openTelemetry.tls.cert=path/to/foo.cert
--tracing.openTelemetry.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.openTelemetry.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.openTelemetry.tls.cert=path/to/foo.cert
--tracing.openTelemetry.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[tracing.openTelemetry.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--tracing.openTelemetry.tls.insecureSkipVerify=true
```

#### GRPC configuration

_Optional_

This instructs the reporter to send spans to the OpenTelemetry Collector using GRPC:

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    grpc: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.grpc]
```

```bash tab="CLI"
--tracing.openTelemetry.grpc=true
```

##### `insecure`

_Optional, Default=false_

Allows reporter to send span to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry.grpc]
    insecure = true
```

```bash tab="CLI"
--tracing.openTelemetry.grpc.insecure=true
```

##### `reconnectionPeriod`

_Optional_

The minimum amount of time between connection attempts to the target endpoint.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    grpc:
      reconnectionPeriod: 30s
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    [tracing.openTelemetry.grpc]
      reconnectionPeriod = "30s"
```

```bash tab="CLI"
--tracing.openTelemetry.grpc.reconnectionPeriod=30s
```

##### `serviceConfig`

_Optional_

Defines the JSON representation of the default gRPC service config used.

For more information about service configurations,
see: [https://github.com/grpc/grpc/blob/master/doc/service_config.md](https://github.com/grpc/grpc/blob/master/doc/service_config.md)

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    grpc:
      serviceConfig: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    [tracing.openTelemetry.grpc]
      serviceConfig = "{}"
```

```bash tab="CLI"
--tracing.openTelemetry.grpc.serviceConfig={}
```