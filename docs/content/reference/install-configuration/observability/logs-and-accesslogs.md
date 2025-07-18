---
title: "Traefik Logs Documentation"
description: "Logs are a key part of observability in Traefik Proxy. Read the technical documentation to learn their configurations, rotations, and time zones."
---

## Logs

Logs concern everything that happens to Traefik itself (startup, configuration, events, shutdown, and so on).

### Configuration Example

```yaml tab="File (YAML)"
log:
  filePath: "/path/to/log-file.log"
  format: json
  level: INFO
```

```toml tab="File (TOML)"
[log]
  filePath = "/path/to/log-file.log"
  format = "json"
  level = "INFO"
```

```sh tab="CLI"
--log.filePath=/path/to/log-file.log
--log.format=json
--log.level=INFO
```

### Configuration Options

The section below describe how to configure Traefik logs using the static configuration.

| Field      | Description  | Default | Required |
|:-----------|:----------------------------|:--------|:---------|
| `log.filePath` | By default, the logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option.| - | No      |
| `log.format` | Log format (`common`or `json`).<br /> The fields displayed with the format `common` cannot be customized. | "common" | No      |
| `log.level` | Log level (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, and `PANIC`)| ERROR | No      |
| `log.noColor` | When using the format `common`, disables the colorized output. | false      | No      |
| `log.maxSize` | Maximum size in megabytes of the log file before it gets rotated. | 100MB      | No      |
| `log.maxAge` | Maximum number of days to retain old log files based on the timestamp encoded in their filename.<br /> A day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc.<br />By default files are not removed based on their age.  |   0   | No      |
| `log.maxBackups` | Maximum number of old log files to retain.<br />The default is to retain all old log files. |  0  | No      |
| `log.compress` | Compress log files in gzip after rotation. | false | No      |

### OpenTelemetry

Traefik supports OpenTelemetry for logging. To enable OpenTelemetry, you need to set the following in the static configuration:

```yaml tab="File (YAML)"
experimental:
  otlpLogs: true
```

```toml tab="File (TOML)"
[experimental]
  otlpLogs = true
```

```sh tab="CLI"
--experimental.otlpLogs=true
```

!!! warning
    This is an experimental feature.

#### Configuration Example

```yaml tab="File (YAML)"
experimental:
  otlpLogs: true

log:
  otlp:
    http:
      endpoint: https://collector:4318/v1/logs
      headers:
        Authorization: Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL
```

```toml tab="File (TOML)"
[experimental]
  otlpLogs = true

[log.otlp]
  http.endpoint = "https://collector:4318/v1/logs"
  http.headers.Authorization = "Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL"
```

```sh tab="CLI"
--experimental.otlpLogs=true
--log.otlp.http.endpoint=https://collector:4318/v1/logs
--log.otlp.http.headers.Authorization=Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL
```

#### Configuration Options

| Field                                  | Description                                                                                                                            | Default                          | Required |
|:---------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------|:---------|
| `log.otlp.serviceName`                 | Service name used in selected backend.                                                                                                 | "traefik"                        | No       |
| `log.otlp.resourceAttributes`          | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| `log.otlp.http`                        | This instructs the exporter to send logs to the OpenTelemetry Collector using HTTP.                                                    |                                  | No       |
| `log.otlp.http.endpoint`               | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| `log.otlp.http.headers`                | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| `log.otlp.http.tls`                    | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| `log.otlp.http.tls.ca`                 | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| `log.otlp.http.tls.cert`               | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| `log.otlp.http.tls.key`                | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| `log.otlp.http.tls.insecureSkipVerify` | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| `log.otlp.grpc`                        | This instructs the exporter to send logs to the OpenTelemetry Collector using gRPC.                                                    |                                  | No       |
| `log.otlp.grpc.endpoint`               | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| `log.otlp.grpc.headers`                | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| `log.otlp.grpc.insecure`               | Instructs the exporter to send logs to the OpenTelemetry Collector using an insecure protocol.                                         | false                            | No       |
| `log.otlp.grpc.tls`                    | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| `log.otlp.grpc.tls.ca`                 | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| `log.otlp.grpc.tls.cert`               | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| `log.otlp.grpc.tls.key`                | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| `log.otlp.grpc.tls.insecureSkipVerify` | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

## AccessLogs

Access logs concern everything that happens to the requests handled by Traefik.

### Configuration Example

```yaml tab="File (YAML)"
accessLog:
  # JSON format
  format: json
  # Filter on status codes, retry attempts and minimal duration
  filters:
    statusCodes:
      - "200"
      - "300-302"
    retryAttempts: true
    minDuration: "10ms"
  fields:
    # Keep all the fields by default
    defaultMode: keep
    names:
      # Drop the Field ClientUserName
      ClientUsername: drop
    headers:
      # Keep all the headers by default
      defaultMode: keep
      names:
        # Redact the User-Agent header value
        User-Agent: redact
        # Drop the Authorization header value
        Authorization: drop
```

```toml tab="File (TOML)"
[accessLog]
  format = "json"

  [accessLog.filters]
    statusCodes = [ "200", "300-302" ]
    retryAttempts = true
    minDuration = "10ms"

  [accessLog.fields]
    defaultMode = "keep"

    [accessLog.fields.names]
      ClientUsername = "drop"

    [accessLog.fields.headers]
      defaultMode = "keep"

      [accessLog.fields.headers.names]
        User-Agent = "redact"
        Authorization = "drop"
```

```sh tab="CLI"
--accesslog=true
--accesslog.format=json
--accesslog.filters.statuscodes=200,300-302
--accesslog.filters.retryattempts
--accesslog.filters.minduration=10ms
--accesslog.fields.defaultmode=keep
--accesslog.fields.names.ClientUsername=drop
--accesslog.fields.headers.defaultmode=keep
--accesslog.fields.headers.names.User-Agent=redact
--accesslog.fields.headers.names.Authorization=drop
```


### Configuration Options

The section below describes how to configure Traefik access logs using the static configuration.

| Field      | Description    | Default | Required |
|:-----------|:--------------------------|:--------|:---------|
| `accesslog.filePath` | By default, the access logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option.|  | No      |
| `accesslog.format` | By default, logs are written using the Common Log Format (CLF).<br />To write logs in JSON, use `json` in the `format` option.<br />If the given format is unsupported, the default (CLF) is used instead.<br />More information about CLF fields [here](#clf-format-fields). | "common" | No      |
| `accesslog.bufferingSize` | To write the logs in an asynchronous fashion, specify a  `bufferingSize` option.<br />This option represents the number of log lines Traefik will keep in memory before writing them to the selected output.<br />In some cases, this option can greatly help performances.| 0 | No      |
| `accesslog.addInternals` | Enables access logs for internal resources (e.g.: `ping@internal`). | false  | No      |
| `accesslog.filters.statusCodes` | Limit the access logs to requests with a status codes in the specified range. | [ ]      | No      |
| `accesslog.filters.retryAttempts` | Keep the access logs when at least one retry has happened. | false      | No      |
| `accesslog.filters.minDuration` | Keep access logs when requests take longer than the specified duration (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)).  |  0   | No      |
| `accesslog.fields.defaultMode` | Mode to apply by default to the access logs fields (`keep`, `redact` or `drop`). | keep | No      |
| `accesslog.fields.names` | Set the fields list to display in the access logs (format `name:mode`).<br /> Available fields list [here](#available-fields). |  [ ]    | No      |
| `accesslog.headers.defaultMode` | Mode to apply by default to the access logs headers (`keep`, `redact` or `drop`).  | drop | No      |
| `accesslog.headers.names` | Set the headers list to display in the access logs (format `name:mode`). |   [ ]   | No      |

### OpenTelemetry

Traefik supports OpenTelemetry for access logs. To enable OpenTelemetry, you need to set the following in the static configuration:

```yaml tab="File (YAML)"
experimental:
  otlpLogs: true
```

```toml tab="File (TOML)"
[experimental]
  otlpLogs = true
```

```sh tab="CLI"
--experimental.otlpLogs=true
```

!!! warning
    This is an experimental feature.

#### Configuration Example

```yaml tab="File (YAML)"
experimental:
  otlpLogs: true

accesslog:
  otlp:
    http:
      endpoint: https://collector:4318/v1/logs
      headers:
        Authorization: Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL
```

```toml tab="File (TOML)"
[experimental]
  otlpLogs = true

[accesslog.otlp]
  http.endpoint = "https://collector:4318/v1/logs"
  http.headers.Authorization = "Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL"
```

```yaml tab="CLI"
--experimental.otlpLogs=true
--accesslog.otlp.http.endpoint=https://collector:4318/v1/logs
--accesslog.otlp.http.headers.Authorization=Bearer auth_asKXRhIMplM7El1JENjrotGouS1LYRdL
```

#### Configuration Options

| Field                                        | Description                                                                                                                            | Default                          | Required |
|:---------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------|:---------|
| `accesslog.otlp.serviceName`                 | Defines the service name resource attribute.                                                                                           | "traefik"                        | No       |
| `accesslog.otlp.resourceAttributes`          | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| `accesslog.otlp.http`                        | This instructs the exporter to send access logs to the OpenTelemetry Collector using HTTP.                                             |                                  | No       |
| `accesslog.otlp.http.endpoint`               | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| `accesslog.otlp.http.headers`                | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| `accesslog.otlp.http.tls`                    | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| `accesslog.otlp.http.tls.ca`                 | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| `accesslog.otlp.http.tls.cert`               | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| `accesslog.otlp.http.tls.key`                | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| `accesslog.otlp.http.tls.insecureSkipVerify` | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| `accesslog.otlp.grpc`                        | This instructs the exporter to send access logs to the OpenTelemetry Collector using gRPC.                                             |                                  | No       |
| `accesslog.otlp.grpc.endpoint`               | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| `accesslog.otlp.grpc.headers`                | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| `accesslog.otlp.grpc.insecure`               | Instructs the exporter to send access logs to the OpenTelemetry Collector using an insecure protocol.                                  | false                            | No       |
| `accesslog.otlp.grpc.tls`                    | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| `accesslog.otlp.grpc.tls.ca`                 | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| `accesslog.otlp.grpc.tls.cert`               | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| `accesslog.otlp.grpc.tls.key`                | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| `accesslog.otlp.grpc.tls.insecureSkipVerify` | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

### CLF format fields

Below the fields displayed with the CLF format:

```html
<remote_IP_address> - <client_user_name_if_available> [<timestamp>] 
"<request_method> <request_path> <request_protocol>" <HTTP_status> <content-length> 
"<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started>
"<Traefik_router_name>" "<Traefik_server_URL>" <request_duration_in_ms>ms
```

### Available Fields

| Field                   | Description   |
|-------------------------|------------------|
| `StartUTC`    | The time at which request processing started.                                                                                                                       |
| `StartLocal`  | The local time at which request processing started.                                                                                                                 |
| `Duration`    | The total time taken (in nanoseconds) by processing the response, including the origin server's time but not the log writing time.                                  |
| `RouterName`  | The name of the Traefik  router.                                                                                                                                    |
| `ServiceName`    | The name of the Traefik backend.          |
| `ServiceURL`   | The URL of the Traefik backend.       |
| `ServiceAddr`    | The IP:port of the Traefik backend (extracted from `ServiceURL`). |
| `ClientAddr`    | The remote address in its original form (usually IP:port).     |
| `ClientHost`   | The remote IP address from which the client request was received.     |
| `ClientPort`            | The remote TCP port from which the client request was received.   |
| `ClientUsername`        | The username provided in the URL, if present.   |
| `RequestAddr`           | The HTTP Host header (usually IP:port). This is treated as not a header by the Go API.   |
| `RequestHost`           | The HTTP Host server name (not including port).     |
| `RequestPort`           | The TCP port from the HTTP Host.    |
| `RequestMethod`         | The HTTP method. |
| `RequestPath`           | The HTTP request URI, not including the scheme, host or port.   |
| `RequestProtocol`       | The version of HTTP requested.       |
| `RequestScheme`         | The HTTP scheme requested `http` or `https`.   |
| `RequestLine`     | The `RequestMethod`, + `RequestPath` and `RequestProtocol`.   |
| `RequestContentSize`    | The number of bytes in the request entity (a.k.a. body) sent by the client.   |
| `OriginDuration`        | The time taken (in nanoseconds) by the origin server ('upstream') to return its response. |
| `OriginContentSize`     | The content length specified by the origin server, or 0 if unspecified.    |
| `OriginStatus`          | The HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent (0). |
| `OriginStatusLine`      | `OriginStatus` + Status code explanation   |
| `DownstreamStatus`      | The HTTP status code returned to the client.    |
| `DownstreamStatusLine`  | The `DownstreamStatus` and status code explanation.     |
| `DownstreamContentSize` | The number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
| `RequestCount`          | The number of requests received since the Traefik instance started.    |
| `GzipRatio`             | The response body compression ratio achieved.   |
| `Overhead`              | The processing time overhead (in nanoseconds) caused by Traefik.    |
| `RetryAttempts`         | The amount of attempts the request was retried.   |
| `TLSVersion`            | The TLS version used by the connection (e.g. `1.2`) (if connection is TLS).   |
| `TLSCipher`             | The TLS cipher used by the connection (e.g. `TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`) (if connection is TLS).      |
| `TLSClientSubject`      | The string representation of the TLS client certificate's Subject (e.g. `CN=username,O=organization`).  |

### Log Rotation

Traefik close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! warning
    This does not work on Windows due to the lack of USR signals.

### Time Zones

Traefik will timestamp each log line in UTC time by default.

It is possible to configure the Traefik to timestamp in a specific timezone by ensuring the following configuration has been made in your environment:

1. Provide time zone data to `/etc/localtime` or `/usr/share/zoneinfo` (based on your distribution) or set the environment variable TZ to the desired timezone.
2. Specify the field `StartLocal` by dropping the field named `StartUTC` (available on the default Common Log Format (CLF) as well as JSON): `accesslog.fields.names.StartUTC=drop`.

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

{!traefik-for-business-applications.md!}
