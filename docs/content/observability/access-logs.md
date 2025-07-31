---
title: "Traefik Access Logs Documentation"
description: "Access logs are a key part of observability in Traefik Proxy. Read the technical documentation to learn their configurations, rotations, and time zones."
---

# Access Logs

Who Calls Whom?
{.subtitle}

By default, logs are written to stdout, in text format.

## Configuration

To enable the access logs:

```yaml tab="File (YAML)"
accessLog: {}
```

```toml tab="File (TOML)"
[accessLog]
```

```bash tab="CLI"
--accesslog=true
```

### `addInternals`

_Optional, Default="false"_

Enables access logs for internal resources (e.g.: `ping@internal`).

```yaml tab="File (YAML)"
accesslog:
  addInternals: true
```

```toml tab="File (TOML)"
[accesslog]
  addInternals = true
```

```bash tab="CLI"
--accesslog.addinternals
```

### `filePath`

By default access logs are written to the standard output.
To write the logs into a log file, use the `filePath` option.

```yaml tab="File (YAML)"
accessLog:
  filePath: "/path/to/access.log"
```

```toml tab="File (TOML)"
[accessLog]
  filePath = "/path/to/access.log"
```

```bash tab="CLI"
--accesslog.filepath=/path/to/access.log
```

### `format`

_Optional, Default="common"_

By default, logs are written using the Common Log Format (CLF).
To write logs in JSON, use `json` in the `format` option.
If the given format is unsupported, the default (CLF) is used instead.

!!! info "Common Log Format"

    ```html
    <remote_IP_address> - <client_user_name_if_available> [<timestamp>] "<request_method> <request_path> <request_protocol>" <HTTP_status> <content-length> "<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started> "<Traefik_router_name>" "<Traefik_server_URL>" <request_duration_in_ms>ms
    ```

```yaml tab="File (YAML)"
accessLog:
  format: "json"
```

```toml tab="File (TOML)"
[accessLog]
  format = "json"
```

```bash tab="CLI"
--accesslog.format=json
```

### `bufferingSize`

To write the logs in an asynchronous fashion, specify a  `bufferingSize` option.
This option represents the number of log lines Traefik will keep in memory before writing them to the selected output.
In some cases, this option can greatly help performances.

```yaml tab="File (YAML)"
# Configuring a buffer of 100 lines
accessLog:
  filePath: "/path/to/access.log"
  bufferingSize: 100
```

```toml tab="File (TOML)"
# Configuring a buffer of 100 lines
[accessLog]
  filePath = "/path/to/access.log"
  bufferingSize = 100
```

```bash tab="CLI"
# Configuring a buffer of 100 lines
--accesslog.filepath=/path/to/access.log
--accesslog.bufferingsize=100
```

### Filtering

To filter logs, you can specify a set of filters which are logically "OR-connected".
Thus, specifying multiple filters will keep more access logs than specifying only one.

The available filters are:

- `statusCodes`, to limit the access logs to requests with a status codes in the specified range
- `retryAttempts`, to keep the access logs when at least one retry has happened
- `minDuration`, to keep access logs when requests take longer than the specified duration (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration))

```yaml tab="File (YAML)"
# Configuring Multiple Filters
accessLog:
  filePath: "/path/to/access.log"
  format: json
  filters:
    statusCodes:
      - "200"
      - "300-302"
    retryAttempts: true
    minDuration: "10ms"
```

```toml tab="File (TOML)"
# Configuring Multiple Filters
[accessLog]
  filePath = "/path/to/access.log"
  format = "json"

  [accessLog.filters]
    statusCodes = ["200", "300-302"]
    retryAttempts = true
    minDuration = "10ms"
```

```bash tab="CLI"
# Configuring Multiple Filters
--accesslog.filepath=/path/to/access.log
--accesslog.format=json
--accesslog.filters.statuscodes=200,300-302
--accesslog.filters.retryattempts
--accesslog.filters.minduration=10ms
```

### Limiting the Fields/Including Headers

You can decide to limit the logged fields/headers to a given list with the `fields.names` and `fields.headers` options.

Each field can be set to:

- `keep` to keep the value
- `drop` to drop the value

Header fields may also optionally be set to `redact` to replace the value with "REDACTED".

The `defaultMode` for `fields.names` is `keep`.

The `defaultMode` for `fields.headers` is `drop`.

```yaml tab="File (YAML)"
# Limiting the Logs to Specific Fields
accessLog:
  filePath: "/path/to/access.log"
  format: json
  fields:
    defaultMode: keep
    names:
      ClientUsername: drop
    headers:
      defaultMode: keep
      names:
        User-Agent: redact
        Authorization: drop
        Content-Type: keep
```

```toml tab="File (TOML)"
# Limiting the Logs to Specific Fields
[accessLog]
  filePath = "/path/to/access.log"
  format = "json"

  [accessLog.fields]
    defaultMode = "keep"

    [accessLog.fields.names]
      "ClientUsername" = "drop"

    [accessLog.fields.headers]
      defaultMode = "keep"

      [accessLog.fields.headers.names]
        "User-Agent" = "redact"
        "Authorization" = "drop"
        "Content-Type" = "keep"
```

```bash tab="CLI"
# Limiting the Logs to Specific Fields
--accesslog.filepath=/path/to/access.log
--accesslog.format=json
--accesslog.fields.defaultmode=keep
--accesslog.fields.names.ClientUsername=drop
--accesslog.fields.headers.defaultmode=keep
--accesslog.fields.headers.names.User-Agent=redact
--accesslog.fields.headers.names.Authorization=drop
--accesslog.fields.headers.names.Content-Type=keep
```

??? info "Available Fields"

    | Field                   | Description                                                                                                                                                         |
    |-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | `StartUTC`              | The time at which request processing started.                                                                                                                       |
    | `StartLocal`            | The local time at which request processing started.                                                                                                                 |
    | `Duration`              | The total time taken (in nanoseconds) by processing the response, including the origin server's time but not the log writing time.                                  |
    | `RouterName`            | The name of the Traefik  router.                                                                                                                                    |
    | `ServiceName`           | The name of the Traefik backend.                                                                                                                                    |
    | `ServiceURL`            | The URL of the Traefik backend.                                                                                                                                     |
    | `ServiceAddr`           | The IP:port of the Traefik backend (extracted from `ServiceURL`)                                                                                                    |
    | `ClientAddr`            | The remote address in its original form (usually IP:port).                                                                                                          |
    | `ClientHost`            | The remote IP address from which the client request was received.                                                                                                   |
    | `ClientPort`            | The remote TCP port from which the client request was received.                                                                                                     |
    | `ClientUsername`        | The username provided in the URL, if present.                                                                                                                       |
    | `RequestAddr`           | The HTTP Host header (usually IP:port). This is treated as not a header by the Go API.                                                                              |
    | `RequestHost`           | The HTTP Host server name (not including port).                                                                                                                     |
    | `RequestPort`           | The TCP port from the HTTP Host.                                                                                                                                    |
    | `RequestMethod`         | The HTTP method.                                                                                                                                                    |
    | `RequestPath`           | The HTTP request URI, not including the scheme, host or port.                                                                                                       |
    | `RequestProtocol`       | The version of HTTP requested.                                                                                                                                      |
    | `RequestScheme`         | The HTTP scheme requested `http` or `https`.                                                                                                                        |
    | `RequestLine`           | `RequestMethod` + `RequestPath` + `RequestProtocol`                                                                                                                 |
    | `RequestContentSize`    | The number of bytes in the request entity (a.k.a. body) sent by the client.                                                                                         |
    | `OriginDuration`        | The time taken (in nanoseconds) by the origin server ('upstream') to return its response.                                                                           |
    | `OriginContentSize`     | The content length specified by the origin server, or 0 if unspecified.                                                                                             |
    | `OriginStatus`          | The HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent (0). |
    | `DownstreamStatus`      | The HTTP status code returned to the client.                                                                                                                        |
    | `DownstreamContentSize` | The number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
    | `RequestCount`          | The number of requests received since the Traefik instance started.                                                                                                 |
    | `GzipRatio`             | The response body compression ratio achieved.                                                                                                                       |
    | `Overhead`              | The processing time overhead (in nanoseconds) caused by Traefik.                                                                                                    |
    | `RetryAttempts`         | The amount of attempts the request was retried.                                                                                                                     |
    | `TLSVersion`            | The TLS version used by the connection (e.g. `1.2`) (if connection is TLS).                                                                                         |
    | `TLSCipher`             | The TLS cipher used by the connection (e.g. `TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`) (if connection is TLS)                                                           |
    | `TLSClientSubject`      | The string representation of the TLS client certificate's Subject (e.g. `CN=username,O=organization`)                                                               |
    | `TraceId`               | A consistent identifier for tracking requests across services, including upstream ones managed by Traefik, shown as a 32-hex digit string                           |
    | `SpanId`                | A unique identifier for Traefik’s root span (EntryPoint) within a request trace, formatted as a 16-hex digit string.                                                |

## Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! warning
    This does not work on Windows due to the lack of USR signals.

## Time Zones

Traefik will timestamp each log line in UTC time by default.

It is possible to configure the Traefik to timestamp in a specific timezone by ensuring the following configuration has been made in your environment:

1. Provide time zone data to `/etc/localtime` or `/usr/share/zoneinfo` (based on your distribution) or set the environment variable TZ to the desired timezone
2. Specify the field `StartLocal` by dropping the field named `StartUTC` (available on the default Common Log Format (CLF) as well as JSON)

Example utilizing Docker Compose:

```yaml
services:
  traefik:
    image: traefik:v3.5
    environment:
      - TZ=US/Alaska
    command:
      - --accesslog.fields.names.StartUTC=drop
      - --providers.docker
    ports:
      - 80:80
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

## OpenTelemetry

!!! warning "Experimental Feature"

    The OpenTelemetry access logs feature is currently experimental and must be explicitly enabled in the experimental section prior to use.
    
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

To enable the OpenTelemetry Logger for access logs:

```yaml tab="File (YAML)"
accesslog:
  otlp: {}
```

```toml tab="File (TOML)"
[accesslog.otlp]
```

```bash tab="CLI"
--accesslog.otlp=true
```

!!! info "Default protocol"

    The OpenTelemetry Logger exporter will export access logs to the collector using HTTPS by default to https://localhost:4318/v1/logs, see the [gRPC Section](#grpc-configuration) to use gRPC.

### `serviceName`

_Optional, Default="traefik"_

Defines the service name resource attribute.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    serviceName: name
```

```toml tab="File (TOML)"
[accesslog]
  [accesslog.otlp]
    serviceName = "name"
```

```bash tab="CLI"
--accesslog.otlp.serviceName=name
```

### `resourceAttributes`

_Optional, Default=empty_

Defines additional resource attributes to be sent to the collector.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    resourceAttributes:
      attr1: foo
      attr2: bar
```

```toml tab="File (TOML)"
[accesslog]
  [accesslog.otlp.resourceAttributes]
    attr1 = "foo"
    attr2 = "bar"
```

```bash tab="CLI"
--accesslog.otlp.resourceAttributes.attr1=foo
--accesslog.otlp.resourceAttributes.attr2=bar
```

### HTTP configuration

_Optional_

This instructs the exporter to send access logs to the OpenTelemetry Collector using HTTP.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http: {}
```

```toml tab="File (TOML)"
[accesslog.otlp.http]
```

```bash tab="CLI"
--accesslog.otlp.http=true
```

#### `endpoint`

_Optional, Default="`https://localhost:4318/v1/logs`", Format="`<scheme>://<host>:<port><path>`"_

URL of the OpenTelemetry Collector to send access logs to.

!!! info "Insecure mode"

    To disable TLS, use `http://` instead of `https://` in the `endpoint` configuration.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      endpoint: https://collector:4318/v1/logs
```

```toml tab="File (TOML)"
[accesslog.otlp.http]
  endpoint = "https://collector:4318/v1/logs"
```

```bash tab="CLI"
--accesslog.otlp.http.endpoint=https://collector:4318/v1/logs
```

#### `headers`

_Optional, Default={}_

Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[accesslog.otlp.http.headers]
  foo = "bar"
  baz = "buz"
```

```bash tab="CLI"
--accesslog.otlp.http.headers.foo=bar --accesslog.otlp.http.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[accesslog.otlp.http.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--accesslog.otlp.http.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[accesslog.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--accesslog.otlp.http.tls.cert=path/to/foo.cert
--accesslog.otlp.http.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[accesslog.otlp.http.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--accesslog.otlp.http.tls.cert=path/to/foo.cert
--accesslog.otlp.http.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    http:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[accesslog.otlp.http.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--accesslog.otlp.http.tls.insecureSkipVerify=true
```

### gRPC configuration

_Optional_

This instructs the exporter to send access logs to the OpenTelemetry Collector using gRPC.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc: {}
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc]
```

```bash tab="CLI"
--accesslog.otlp.grpc=true
```

#### `endpoint`

_Required, Default="localhost:4317", Format="`<host>:<port>`"_

Address of the OpenTelemetry Collector to send access logs to.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      endpoint: localhost:4317
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc]
  endpoint = "localhost:4317"
```

```bash tab="CLI"
--accesslog.otlp.grpc.endpoint=localhost:4317
```

#### `insecure`

_Optional, Default=false_

Allows exporter to send access logs to the OpenTelemetry Collector without using a secured protocol.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      insecure: true
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc]
  insecure = true
```

```bash tab="CLI"
--accesslog.otlp.grpc.insecure=true
```

#### `headers`

_Optional, Default={}_

Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      headers:
        foo: bar
        baz: buz
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc.headers]
  foo = "bar"
  baz = "buz"
```

```bash tab="CLI"
--accesslog.otlp.grpc.headers.foo=bar --accesslog.otlp.grpc.headers.baz=buz
```

#### `tls`

_Optional_

Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to the OpenTelemetry Collector,
it defaults to the system bundle.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--accesslog.otlp.grpc.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--accesslog.otlp.grpc.tls.cert=path/to/foo.cert
--accesslog.otlp.grpc.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the OpenTelemetry Collector.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--accesslog.otlp.grpc.tls.cert=path/to/foo.cert
--accesslog.otlp.grpc.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`,
the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
accesslog:
  otlp:
    grpc:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[accesslog.otlp.grpc.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--accesslog.otlp.grpc.tls.insecureSkipVerify=true
```

{!traefik-for-business-applications.md!}
