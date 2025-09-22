---
title: "Traefik Logs Documentation"
description: "Logs are a key part of observability in Traefik Proxy. Read the technical documentation to learn their configurations, rotations, and time zones."
---

# Logs

Reading What's Happening
{: .subtitle }

By default, logs are written to stdout, in text format.

## Configuration

### General

Traefik logs concern everything that happens to Traefik itself (startup, configuration, events, shutdown, and so on).

#### `filePath`

By default, the logs are written to the standard output.
You can configure a file path instead using the `filePath` option.

```yaml tab="File (YAML)"
# Writing Logs to a File
log:
  filePath: "/path/to/traefik.log"
```

```toml tab="File (TOML)"
# Writing Logs to a File
[log]
  filePath = "/path/to/traefik.log"
```

```bash tab="CLI"
# Writing Logs to a File
--log.filePath=/path/to/traefik.log
```

#### `format`

By default, the logs use a text format (`common`), but you can also ask for the `json` format in the `format` option.

```yaml tab="File (YAML)"
# Writing Logs to a File, in JSON
log:
  filePath: "/path/to/log-file.log"
  format: json
```

```toml tab="File (TOML)"
# Writing Logs to a File, in JSON
[log]
  filePath = "/path/to/log-file.log"
  format = "json"
```

```bash tab="CLI"
# Writing Logs to a File, in JSON
--log.filePath=/path/to/traefik.log
--log.format=json
```

#### `level`

By default, the `level` is set to `ERROR`.

Alternative logging levels are `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, and `PANIC`.

```yaml tab="File (YAML)"
log:
  level: DEBUG
```

```toml tab="File (TOML)"
[log]
  level = "DEBUG"
```

```bash tab="CLI"
--log.level=DEBUG
```

#### `noColor`

When using the 'common' format, disables the colorized output.

```yaml tab="File (YAML)"
log:
  noColor: true
```

```toml tab="File (TOML)"
[log]
  noColor = true
```

```bash tab="CLI"
--log.nocolor=true
```

## Log Rotation

The rotation of the log files can be configured with the following options.

### `maxSize`

`maxSize` is the maximum size in megabytes of the log file before it gets rotated.
It defaults to 100 megabytes.

```yaml tab="File (YAML)"
log:
  maxSize: 1
```

```toml tab="File (TOML)"
[log]
  maxSize = 1
```

```bash tab="CLI"
--log.maxsize=1
```

### `maxBackups`

`maxBackups` is the maximum number of old log files to retain.
The default is to retain all old log files (though `maxAge` may still cause them to get deleted).

```yaml tab="File (YAML)"
log:
  maxBackups: 3
```

```toml tab="File (TOML)"
[log]
  maxBackups = 3
```

```bash tab="CLI"
--log.maxbackups=3
```

### `maxAge`

`maxAge` is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc.
The default is not to remove old log files based on age.

```yaml tab="File (YAML)"
log:
  maxAge: 3
```

```toml tab="File (TOML)"
[log]
  maxAge = 3
```

```bash tab="CLI"
--log.maxage=3
```

### `compress`

`compress` determines if the rotated log files should be compressed using gzip.
The default is not to perform compression.

```yaml tab="File (YAML)"
log:
  compress: true
```

```toml tab="File (TOML)"
[log]
  compress = true
```

```bash tab="CLI"
--log.compress=true
```

## OpenTelemetry

!!! warning "Experimental Feature"
    
    The OpenTelemetry logs feature is currently experimental and must be explicitly enabled in the experimental section prior to use.
    
    ```yaml tab="File (YAML)"
    experimental:
      otlpLogs: true
    ```
    
    ```toml tab="File (TOML)"
    [experimental.otlpLogs]
    ```
    
    ```bash tab="CLI"
    --experimental.otlpLogs=true
    ```

To enable the OpenTelemetry Logger for logs:

```yaml tab="File (YAML)"
log:
  otlp: {}
```

```toml tab="File (TOML)"
[log.otlp]
```

```bash tab="CLI"
--log.otlp=true
```

!!! info "Default protocol"

    The OpenTelemetry Logger exporter will export logs to the collector using HTTPS by default to https://localhost:4318/v1/logs, see the [gRPC Section](#grpc-configuration) to use gRPC.

### `serviceName`

_Optional, Default="traefik"_

Defines the service name resource attribute.

```yaml tab="File (YAML)"
log:
  otlp:
    serviceName: name
```

```toml tab="File (TOML)"
[log]
  [log.otlp]
    serviceName = "name"
```

```bash tab="CLI"
--log.otlp.serviceName=name
```

### `resourceAttributes`

_Optional, Default=empty_

Defines additional resource attributes to be sent to the collector.

```yaml tab="File (YAML)"
log:
  otlp:
    resourceAttributes:
      attr1: foo
      attr2: bar
```

```toml tab="File (TOML)"
[log]
  [log.otlp.resourceAttributes]
    attr1 = "foo"
    attr2 = "bar"
```

```bash tab="CLI"
--log.otlp.resourceAttributes.attr1=foo
--log.otlp.resourceAttributes.attr2=bar
```

### HTTP configuration

_Optional_

This instructs the exporter to send logs to the OpenTelemetry Collector using HTTP.

```yaml tab="File (YAML)"
log:
  otlp:
    http: {}
```

```toml tab="File (TOML)"
[log.otlp.http]
```

```bash tab="CLI"
--log.otlp.http=true
```

#### `endpoint`

_Optional, Default="`https://localhost:4318/v1/logs`", Format="`<scheme>://<host>:<port><path>`"_

URL of the OpenTelemetry Collector to send logs to.

!!! info "Insecure mode"

    To disable TLS, use `http://` instead of `https://` in the `endpoint` configuration.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      endpoint: https://collector:4318/v1/logs
```

```toml tab="File (TOML)"
[log.otlp.http]
  endpoint = "https://collector:4318/v1/logs"
```

```bash tab="CLI"
--log.otlp.http.endpoint=https://collector:4318/v1/logs
```

#### `headers`

_Optional, Default={}_

Additional headers sent with logs by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[log.otlp.http.headers]
  foo = "bar"
  baz = "buz"
```

```bash tab="CLI"
--log.otlp.http.headers.foo=bar --log.otlp.http.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[log.otlp.http.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--log.otlp.http.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[log.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--log.otlp.http.tls.cert=path/to/foo.cert
--log.otlp.http.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[log.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--log.otlp.http.tls.cert=path/to/foo.cert
--log.otlp.http.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
log:
  otlp:
    http:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[log.otlp.http.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--log.otlp.http.tls.insecureSkipVerify=true
```

### gRPC configuration

_Optional_

This instructs the exporter to send logs to the OpenTelemetry Collector using gRPC.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc: {}
```

```toml tab="File (TOML)"
[log.otlp.grpc]
```

```bash tab="CLI"
--log.otlp.grpc=true
```

#### `endpoint`

_Required, Default="localhost:4317", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send logs to.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      endpoint: localhost:4317
```

```toml tab="File (TOML)"
[log.otlp.grpc]
  endpoint = "localhost:4317"
```

```bash tab="CLI"
--log.otlp.grpc.endpoint=localhost:4317
```

#### `insecure`

_Optional, Default=false_

Allows exporter to send logs to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[log.otlp.grpc]
  insecure = true
```

```bash tab="CLI"
--log.otlp.grpc.insecure=true
```

#### `headers`

_Optional, Default={}_

Additional headers sent with logs by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[log.otlp.grpc.headers]
  foo = "bar"
  baz = "buz"
```

```bash tab="CLI"
--log.otlp.grpc.headers.foo=bar --log.otlp.grpc.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[log.otlp.grpc.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--log.otlp.grpc.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[log.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--log.otlp.grpc.tls.cert=path/to/foo.cert
--log.otlp.grpc.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[log.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--log.otlp.grpc.tls.cert=path/to/foo.cert
--log.otlp.grpc.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
log:
  otlp:
    grpc:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[log.otlp.grpc.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--log.otlp.grpc.tls.insecureSkipVerify=true
```

{!traefik-for-business-applications.md!}
