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

!!! info "The OpenTelemetry trace reporter will export traces to the collector using HTTP by default, see the [gRPC Section](#grpc-configuration) to use gRPC."

!!! info "Trace sampling"

	By default, the OpenTelemetry trace reporter will sample 100% of traces.
	See [OpenTelemetry's SDK configuration](https://opentelemetry.io/docs/reference/specification/sdk-environment-variables/#general-sdk-configuration) to customize the sampling strategy.

#### `address`

_Required, Default="localhost:4318", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send spans to.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    address: localhost:4318
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    address = "localhost:4318"
```

```bash tab="CLI"
--tracing.openTelemetry.address=localhost:4318
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
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--tracing.openTelemetry.headers.foo=bar --tracing.openTelemetry.headers.baz=buz
```

#### `insecure`

_Optional, Default=false_

Allows reporter to send spans to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    insecure: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    insecure = true
```

```bash tab="CLI"
--tracing.openTelemetry.insecure=true
```

#### `path`

_Required, Default="/v1/traces"_

Allows to override the default URL path used for sending traces.
This option has no effect when using gRPC transport.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    path: /foo/v1/traces
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    path = "/foo/v1/traces"
```

```bash tab="CLI"
--tracing.openTelemetry.path=/foo/v1/traces
```

#### `globalTags`

_Optional, Default=empty_

Applies a list of shared key:value tags on all spans.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    globalTags:
      tag1: foo
      tag2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    [tracing.openTelemetry.globalTags]
      tag1 = "foo"
      tag2 = "bar"
```

```bash tab="CLI"
--tracing.openTelemetry.globalTags.tag1=foo
--tracing.openTelemetry.globalTags.tag2=bar
```

#### `sampleRate`

_Optional, Default=1.0_

The proportion of requests to trace, specified between 0.0 and 1.0.

```yaml tab="File (YAML)"
tracing:
  openTelemetry:
    sampleRate: 0.2
```

```toml tab="File (TOML)"
[tracing]
  [tracing.openTelemetry]
    sampleRate = 0.2
```

```bash tab="CLI"
--tracing.openTelemetry.sampleRate=0.2
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

#### gRPC configuration

_Optional_

This instructs the reporter to send spans to the OpenTelemetry Collector using gRPC.

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
