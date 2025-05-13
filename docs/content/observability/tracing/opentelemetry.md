---
title: "Traefik OpenTelemetry Documentation"
description: "Traefik supports several tracing backends, including OpenTelemetry. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# OpenTelemetry

Traefik Proxy follows [official OpenTelemetry semantic conventions v1.26.0](https://github.com/open-telemetry/semantic-conventions/blob/v1.26.0/docs/http/http-spans.md).

To enable the OpenTelemetry tracer:

```yaml tab="File (YAML)"
tracing:
  otlp: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp]
```

```bash tab="CLI"
--tracing.otlp=true
```

!!! info "Default protocol"

    The OpenTelemetry trace exporter will export traces to the collector using HTTPS by default to https://localhost:4318/v1/traces, see the [gRPC Section](#grpc-configuration) to use gRPC.

!!! info "Trace sampling"

	By default, the OpenTelemetry trace exporter will sample 100% of traces.  
	See [OpenTelemetry's SDK configuration](https://opentelemetry.io/docs/reference/specification/sdk-environment-variables/#general-sdk-configuration) to customize the sampling strategy.

!!! info "Propagation"
    
    Traefik supports the `OTEL_PROPAGATORS` env variable to set up the propragators. The supported propagators are:

    - tracecontext (default)
    - baggage (default)
    - b3
    - b3multi
    - jaeger
    - xray
    - ottrace

    Example of configuration:

        OTEL_PROPAGATORS=b3,jaeger


### HTTP configuration

_Optional_

This instructs the exporter to send spans to the OpenTelemetry Collector using HTTP.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.http]
```

```bash tab="CLI"
--tracing.otlp.http=true
```

#### `endpoint`

_Optional, Default="https://localhost:4318/v1/traces", Format="`<scheme>://<host>:<port><path>`"_

URL of the OpenTelemetry Collector to send spans to.

!!! info "Insecure mode"

    To disable TLS, use `http://` instead of `https://` in the `endpoint` configuration.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      endpoint: https://collector:4318/v1/traces
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.http]
    endpoint = "https://collector:4318/v1/traces"
```

```bash tab="CLI"
--tracing.otlp.http.endpoint=https://collector:4318/v1/traces
```

#### `headers`

_Optional, Default={}_

Additional headers sent with traces by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.http.headers]
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--tracing.otlp.http.headers.foo=bar --tracing.otlp.http.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send spans to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[tracing.otlp.http.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--tracing.otlp.http.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.otlp.http.tls.cert=path/to/foo.cert
--tracing.otlp.http.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.otlp.http.tls.cert=path/to/foo.cert
--tracing.otlp.http.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
tracing:
  otlp:
    http:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[tracing.otlp.http.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--tracing.otlp.http.tls.insecureSkipVerify=true
```

### gRPC configuration

_Optional_

This instructs the exporter to send spans to the OpenTelemetry Collector using gRPC.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.grpc]
```

```bash tab="CLI"
--tracing.otlp.grpc=true
```

#### `endpoint`

_Required, Default="localhost:4317", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send spans to.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      endpoint: localhost:4317
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.grpc]
    endpoint = "localhost:4317"
```

```bash tab="CLI"
--tracing.otlp.grpc.endpoint=localhost:4317
```
#### `insecure`

_Optional, Default=false_

Allows exporter to send spans to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.grpc]
    insecure = true
```

```bash tab="CLI"
--tracing.otlp.grpc.insecure=true
```

#### `headers`

_Optional, Default={}_

Additional headers sent with traces by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[tracing]
  [tracing.otlp.grpc.headers]
    foo = "bar"
    baz = "buz"
```

```bash tab="CLI"
--tracing.otlp.grpc.headers.foo=bar --tracing.otlp.grpc.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send spans to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[tracing.otlp.grpc.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--tracing.otlp.grpc.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.otlp.grpc.tls.cert=path/to/foo.cert
--tracing.otlp.grpc.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[tracing.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--tracing.otlp.grpc.tls.cert=path/to/foo.cert
--tracing.otlp.grpc.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
tracing:
  otlp:
    grpc:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[tracing.otlp.grpc.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--tracing.otlp.grpc.tls.insecureSkipVerify=true
```
