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
| <a id="opt-log-filePath" href="#opt-log-filePath" title="#opt-log-filePath">`log.filePath`</a> | By default, the logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option. When `filePath` is specified, Traefik will write logs only to that file (not to standard output).| - | No      |
| <a id="opt-log-format" href="#opt-log-format" title="#opt-log-format">`log.format`</a> | Log format (`common`or `json`).<br /> The fields displayed with the format `common` cannot be customized. | "common" | No      |
| <a id="opt-log-level" href="#opt-log-level" title="#opt-log-level">`log.level`</a> | Log level (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, and `PANIC`)| ERROR | No      |
| <a id="opt-log-noColor" href="#opt-log-noColor" title="#opt-log-noColor">`log.noColor`</a> | When using the format `common`, disables the colorized output. | false      | No      |
| <a id="opt-log-maxSize" href="#opt-log-maxSize" title="#opt-log-maxSize">`log.maxSize`</a> | Maximum size in megabytes of the log file before it gets rotated. | 100MB      | No      |
| <a id="opt-log-maxAge" href="#opt-log-maxAge" title="#opt-log-maxAge">`log.maxAge`</a> | Maximum number of days to retain old log files based on the timestamp encoded in their filename.<br /> A day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc.<br />By default files are not removed based on their age.  |   0   | No      |
| <a id="opt-log-maxBackups" href="#opt-log-maxBackups" title="#opt-log-maxBackups">`log.maxBackups`</a> | Maximum number of old log files to retain.<br />The default is to retain all old log files. |  0  | No      |
| <a id="opt-log-compress" href="#opt-log-compress" title="#opt-log-compress">`log.compress`</a> | Compress log files in gzip after rotation. | false | No      |

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
| <a id="opt-log-otlp-serviceName" href="#opt-log-otlp-serviceName" title="#opt-log-otlp-serviceName">`log.otlp.serviceName`</a> | Service name used in selected backend.                                                                                                 | "traefik"                        | No       |
| <a id="opt-log-otlp-resourceAttributes" href="#opt-log-otlp-resourceAttributes" title="#opt-log-otlp-resourceAttributes">`log.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| <a id="opt-log-otlp-http" href="#opt-log-otlp-http" title="#opt-log-otlp-http">`log.otlp.http`</a> | This instructs the exporter to send logs to the OpenTelemetry Collector using HTTP.                                                    |                                  | No       |
| <a id="opt-log-otlp-http-endpoint" href="#opt-log-otlp-http-endpoint" title="#opt-log-otlp-http-endpoint">`log.otlp.http.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| <a id="opt-log-otlp-http-headers" href="#opt-log-otlp-http-headers" title="#opt-log-otlp-http-headers">`log.otlp.http.headers`</a> | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| <a id="opt-log-otlp-http-tls" href="#opt-log-otlp-http-tls" title="#opt-log-otlp-http-tls">`log.otlp.http.tls`</a> | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| <a id="opt-log-otlp-http-tls-ca" href="#opt-log-otlp-http-tls-ca" title="#opt-log-otlp-http-tls-ca">`log.otlp.http.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="opt-log-otlp-http-tls-cert" href="#opt-log-otlp-http-tls-cert" title="#opt-log-otlp-http-tls-cert">`log.otlp.http.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="opt-log-otlp-http-tls-key" href="#opt-log-otlp-http-tls-key" title="#opt-log-otlp-http-tls-key">`log.otlp.http.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="opt-log-otlp-http-tls-insecureSkipVerify" href="#opt-log-otlp-http-tls-insecureSkipVerify" title="#opt-log-otlp-http-tls-insecureSkipVerify">`log.otlp.http.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| <a id="opt-log-otlp-grpc" href="#opt-log-otlp-grpc" title="#opt-log-otlp-grpc">`log.otlp.grpc`</a> | This instructs the exporter to send logs to the OpenTelemetry Collector using gRPC.                                                    |                                  | No       |
| <a id="opt-log-otlp-grpc-endpoint" href="#opt-log-otlp-grpc-endpoint" title="#opt-log-otlp-grpc-endpoint">`log.otlp.grpc.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| <a id="opt-log-otlp-grpc-headers" href="#opt-log-otlp-grpc-headers" title="#opt-log-otlp-grpc-headers">`log.otlp.grpc.headers`</a> | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| <a id="opt-log-otlp-grpc-insecure" href="#opt-log-otlp-grpc-insecure" title="#opt-log-otlp-grpc-insecure">`log.otlp.grpc.insecure`</a> | Instructs the exporter to send logs to the OpenTelemetry Collector using an insecure protocol.                                         | false                            | No       |
| <a id="opt-log-otlp-grpc-tls" href="#opt-log-otlp-grpc-tls" title="#opt-log-otlp-grpc-tls">`log.otlp.grpc.tls`</a> | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| <a id="opt-log-otlp-grpc-tls-ca" href="#opt-log-otlp-grpc-tls-ca" title="#opt-log-otlp-grpc-tls-ca">`log.otlp.grpc.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="opt-log-otlp-grpc-tls-cert" href="#opt-log-otlp-grpc-tls-cert" title="#opt-log-otlp-grpc-tls-cert">`log.otlp.grpc.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="opt-log-otlp-grpc-tls-key" href="#opt-log-otlp-grpc-tls-key" title="#opt-log-otlp-grpc-tls-key">`log.otlp.grpc.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="opt-log-otlp-grpc-tls-insecureSkipVerify" href="#opt-log-otlp-grpc-tls-insecureSkipVerify" title="#opt-log-otlp-grpc-tls-insecureSkipVerify">`log.otlp.grpc.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

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
| <a id="opt-accesslog-filePath" href="#opt-accesslog-filePath" title="#opt-accesslog-filePath">`accesslog.filePath`</a> | By default, the access logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option.|  | No      |
| <a id="opt-accesslog-format" href="#opt-accesslog-format" title="#opt-accesslog-format">`accesslog.format`</a> | By default, logs are written using the Traefik Common Log Format (CLF).<br />Available formats: [`common`](#traefik-clf-format-fields) (Traefik extended CLF), [`genericCLF`](#generic-clf-format-fields) (standard CLF compatible with analyzers), or [`json`](#json-format-fields).<br />If the given format is unsupported, the default (`common`) is used instead. | "common" | No      |
| <a id="opt-accesslog-bufferingSize" href="#opt-accesslog-bufferingSize" title="#opt-accesslog-bufferingSize">`accesslog.bufferingSize`</a> | To write the logs in an asynchronous fashion, specify a  `bufferingSize` option.<br />This option represents the number of log lines Traefik will keep in memory before writing them to the selected output.<br />In some cases, this option can greatly help performances.| 0 | No      |
| <a id="opt-accesslog-addInternals" href="#opt-accesslog-addInternals" title="#opt-accesslog-addInternals">`accesslog.addInternals`</a> | Enables access logs for internal resources (e.g.: `ping@internal`). | false  | No      |
| <a id="opt-accesslog-filters-statusCodes" href="#opt-accesslog-filters-statusCodes" title="#opt-accesslog-filters-statusCodes">`accesslog.filters.statusCodes`</a> | Limit the access logs to requests with a status codes in the specified range. | [ ]      | No      |
| <a id="opt-accesslog-filters-retryAttempts" href="#opt-accesslog-filters-retryAttempts" title="#opt-accesslog-filters-retryAttempts">`accesslog.filters.retryAttempts`</a> | Keep the access logs when at least one retry has happened. | false      | No      |
| <a id="opt-accesslog-filters-minDuration" href="#opt-accesslog-filters-minDuration" title="#opt-accesslog-filters-minDuration">`accesslog.filters.minDuration`</a> | Keep access logs when requests take longer than the specified duration (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)).  |  0   | No      |
| <a id="opt-accesslog-fields-defaultMode" href="#opt-accesslog-fields-defaultMode" title="#opt-accesslog-fields-defaultMode">`accesslog.fields.defaultMode`</a> | Mode to apply by default to the access logs fields (`keep`, `redact` or `drop`). | keep | No      |
| <a id="opt-accesslog-fields-names" href="#opt-accesslog-fields-names" title="#opt-accesslog-fields-names">`accesslog.fields.names`</a> | Set the fields list to display in the access logs (format `name:mode`).<br /> Available fields list [here](#json-format-fields). |  [ ]    | No      |
| <a id="opt-accesslog-fields-headers-defaultMode" href="#opt-accesslog-fields-headers-defaultMode" title="#opt-accesslog-fields-headers-defaultMode">`accesslog.fields.headers.defaultMode`</a> | Mode to apply by default to the access logs headers (`keep`, `redact` or `drop`).  | drop | No      |
| <a id="opt-accesslog-fields-headers-names" href="#opt-accesslog-fields-headers-names" title="#opt-accesslog-fields-headers-names">`accesslog.fields.headers.names`</a> | Set the headers list to display in the access logs (format `name:mode`). |   [ ]   | No      |

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
| <a id="opt-accesslog-otlp-serviceName" href="#opt-accesslog-otlp-serviceName" title="#opt-accesslog-otlp-serviceName">`accesslog.otlp.serviceName`</a> | Defines the service name resource attribute.                                                                                           | "traefik"                        | No       |
| <a id="opt-accesslog-otlp-resourceAttributes" href="#opt-accesslog-otlp-resourceAttributes" title="#opt-accesslog-otlp-resourceAttributes">`accesslog.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| <a id="opt-accesslog-otlp-http" href="#opt-accesslog-otlp-http" title="#opt-accesslog-otlp-http">`accesslog.otlp.http`</a> | This instructs the exporter to send access logs to the OpenTelemetry Collector using HTTP.                                             |                                  | No       |
| <a id="opt-accesslog-otlp-http-endpoint" href="#opt-accesslog-otlp-http-endpoint" title="#opt-accesslog-otlp-http-endpoint">`accesslog.otlp.http.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| <a id="opt-accesslog-otlp-http-headers" href="#opt-accesslog-otlp-http-headers" title="#opt-accesslog-otlp-http-headers">`accesslog.otlp.http.headers`</a> | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| <a id="opt-accesslog-otlp-http-tls" href="#opt-accesslog-otlp-http-tls" title="#opt-accesslog-otlp-http-tls">`accesslog.otlp.http.tls`</a> | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| <a id="opt-accesslog-otlp-http-tls-ca" href="#opt-accesslog-otlp-http-tls-ca" title="#opt-accesslog-otlp-http-tls-ca">`accesslog.otlp.http.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="opt-accesslog-otlp-http-tls-cert" href="#opt-accesslog-otlp-http-tls-cert" title="#opt-accesslog-otlp-http-tls-cert">`accesslog.otlp.http.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="opt-accesslog-otlp-http-tls-key" href="#opt-accesslog-otlp-http-tls-key" title="#opt-accesslog-otlp-http-tls-key">`accesslog.otlp.http.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="opt-accesslog-otlp-http-tls-insecureSkipVerify" href="#opt-accesslog-otlp-http-tls-insecureSkipVerify" title="#opt-accesslog-otlp-http-tls-insecureSkipVerify">`accesslog.otlp.http.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| <a id="opt-accesslog-otlp-grpc" href="#opt-accesslog-otlp-grpc" title="#opt-accesslog-otlp-grpc">`accesslog.otlp.grpc`</a> | This instructs the exporter to send access logs to the OpenTelemetry Collector using gRPC.                                             |                                  | No       |
| <a id="opt-accesslog-otlp-grpc-endpoint" href="#opt-accesslog-otlp-grpc-endpoint" title="#opt-accesslog-otlp-grpc-endpoint">`accesslog.otlp.grpc.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| <a id="opt-accesslog-otlp-grpc-headers" href="#opt-accesslog-otlp-grpc-headers" title="#opt-accesslog-otlp-grpc-headers">`accesslog.otlp.grpc.headers`</a> | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| <a id="opt-accesslog-otlp-grpc-insecure" href="#opt-accesslog-otlp-grpc-insecure" title="#opt-accesslog-otlp-grpc-insecure">`accesslog.otlp.grpc.insecure`</a> | Instructs the exporter to send access logs to the OpenTelemetry Collector using an insecure protocol.                                  | false                            | No       |
| <a id="opt-accesslog-otlp-grpc-tls" href="#opt-accesslog-otlp-grpc-tls" title="#opt-accesslog-otlp-grpc-tls">`accesslog.otlp.grpc.tls`</a> | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| <a id="opt-accesslog-otlp-grpc-tls-ca" href="#opt-accesslog-otlp-grpc-tls-ca" title="#opt-accesslog-otlp-grpc-tls-ca">`accesslog.otlp.grpc.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="opt-accesslog-otlp-grpc-tls-cert" href="#opt-accesslog-otlp-grpc-tls-cert" title="#opt-accesslog-otlp-grpc-tls-cert">`accesslog.otlp.grpc.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="opt-accesslog-otlp-grpc-tls-key" href="#opt-accesslog-otlp-grpc-tls-key" title="#opt-accesslog-otlp-grpc-tls-key">`accesslog.otlp.grpc.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="opt-accesslog-otlp-grpc-tls-insecureSkipVerify" href="#opt-accesslog-otlp-grpc-tls-insecureSkipVerify" title="#opt-accesslog-otlp-grpc-tls-insecureSkipVerify">`accesslog.otlp.grpc.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

### Traefik CLF format fields

It's the default format provided by Traefik.
Below the fields displayed with the Traefik CLF format:

```html
<remote_IP_address> - <client_user_name_if_available> [<timestamp>] 
"<request_method> <request_path> <request_protocol>" <HTTP_status> <content-length> 
"<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started>
"<Traefik_router_name>" "<Traefik_server_URL>" <request_duration_in_ms>ms
```

### Generic CLF format fields

Below the fields displayed with the generic CLF format:

```html
<remote_IP_address> - <client_user_name_if_available> [<timestamp>] 
"<request_method> <request_path> <request_protocol>" <HTTP_status> <content-length> 
"<request_referrer>" "<request_user_agent>"
```

### JSON format fields

| Field                   | Description   |
|-------------------------|------------------|
| <a id="opt-StartUTC" href="#opt-StartUTC" title="#opt-StartUTC">`StartUTC`</a> | The time at which request processing started.                                                                                                                       |
| <a id="opt-StartLocal" href="#opt-StartLocal" title="#opt-StartLocal">`StartLocal`</a> | The local time at which request processing started.                                                                                                                 |
| <a id="opt-Duration" href="#opt-Duration" title="#opt-Duration">`Duration`</a> | The total time taken (in nanoseconds) by processing the response, including the origin server's time but not the log writing time.                                  |
| <a id="opt-RouterName" href="#opt-RouterName" title="#opt-RouterName">`RouterName`</a> | The name of the Traefik  router.                                                                                                                                    |
| <a id="opt-ServiceName" href="#opt-ServiceName" title="#opt-ServiceName">`ServiceName`</a> | The name of the Traefik backend.          |
| <a id="opt-ServiceURL" href="#opt-ServiceURL" title="#opt-ServiceURL">`ServiceURL`</a> | The URL of the Traefik backend.       |
| <a id="opt-ServiceAddr" href="#opt-ServiceAddr" title="#opt-ServiceAddr">`ServiceAddr`</a> | The IP:port of the Traefik backend (extracted from `ServiceURL`). |
| <a id="opt-ClientAddr" href="#opt-ClientAddr" title="#opt-ClientAddr">`ClientAddr`</a> | The remote address in its original form (usually IP:port).     |
| <a id="opt-ClientHost" href="#opt-ClientHost" title="#opt-ClientHost">`ClientHost`</a> | The remote IP address from which the client request was received.     |
| <a id="opt-ClientPort" href="#opt-ClientPort" title="#opt-ClientPort">`ClientPort`</a> | The remote TCP port from which the client request was received.   |
| <a id="opt-ClientUsername" href="#opt-ClientUsername" title="#opt-ClientUsername">`ClientUsername`</a> | The username provided in the URL, if present.   |
| <a id="opt-RequestAddr" href="#opt-RequestAddr" title="#opt-RequestAddr">`RequestAddr`</a> | The HTTP Host header (usually IP:port). This is treated as not a header by the Go API.   |
| <a id="opt-RequestHost" href="#opt-RequestHost" title="#opt-RequestHost">`RequestHost`</a> | The HTTP Host server name (not including port).     |
| <a id="opt-RequestPort" href="#opt-RequestPort" title="#opt-RequestPort">`RequestPort`</a> | The TCP port from the HTTP Host.    |
| <a id="opt-RequestMethod" href="#opt-RequestMethod" title="#opt-RequestMethod">`RequestMethod`</a> | The HTTP method. |
| <a id="opt-RequestPath" href="#opt-RequestPath" title="#opt-RequestPath">`RequestPath`</a> | The HTTP request URI, not including the scheme, host or port.   |
| <a id="opt-RequestProtocol" href="#opt-RequestProtocol" title="#opt-RequestProtocol">`RequestProtocol`</a> | The version of HTTP requested.       |
| <a id="opt-RequestScheme" href="#opt-RequestScheme" title="#opt-RequestScheme">`RequestScheme`</a> | The HTTP scheme requested `http` or `https`.   |
| <a id="opt-RequestLine" href="#opt-RequestLine" title="#opt-RequestLine">`RequestLine`</a> | The `RequestMethod`, + `RequestPath` and `RequestProtocol`.   |
| <a id="opt-RequestContentSize" href="#opt-RequestContentSize" title="#opt-RequestContentSize">`RequestContentSize`</a> | The number of bytes in the request entity (a.k.a. body) sent by the client.   |
| <a id="opt-OriginDuration" href="#opt-OriginDuration" title="#opt-OriginDuration">`OriginDuration`</a> | The time taken (in nanoseconds) by the origin server ('upstream') to return its response. |
| <a id="opt-OriginContentSize" href="#opt-OriginContentSize" title="#opt-OriginContentSize">`OriginContentSize`</a> | The content length specified by the origin server, or 0 if unspecified.    |
| <a id="opt-OriginStatus" href="#opt-OriginStatus" title="#opt-OriginStatus">`OriginStatus`</a> | The HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent (0). |
| <a id="opt-OriginStatusLine" href="#opt-OriginStatusLine" title="#opt-OriginStatusLine">`OriginStatusLine`</a> | `OriginStatus` + Status code explanation   |
| <a id="opt-DownstreamStatus" href="#opt-DownstreamStatus" title="#opt-DownstreamStatus">`DownstreamStatus`</a> | The HTTP status code returned to the client.    |
| <a id="opt-DownstreamStatusLine" href="#opt-DownstreamStatusLine" title="#opt-DownstreamStatusLine">`DownstreamStatusLine`</a> | The `DownstreamStatus` and status code explanation.     |
| <a id="opt-DownstreamContentSize" href="#opt-DownstreamContentSize" title="#opt-DownstreamContentSize">`DownstreamContentSize`</a> | The number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
| <a id="opt-RequestCount" href="#opt-RequestCount" title="#opt-RequestCount">`RequestCount`</a> | The number of requests received since the Traefik instance started.    |
| <a id="opt-GzipRatio" href="#opt-GzipRatio" title="#opt-GzipRatio">`GzipRatio`</a> | The response body compression ratio achieved.   |
| <a id="opt-Overhead" href="#opt-Overhead" title="#opt-Overhead">`Overhead`</a> | The processing time overhead (in nanoseconds) caused by Traefik.    |
| <a id="opt-RetryAttempts" href="#opt-RetryAttempts" title="#opt-RetryAttempts">`RetryAttempts`</a> | The amount of attempts the request was retried.   |
| <a id="opt-TLSVersion" href="#opt-TLSVersion" title="#opt-TLSVersion">`TLSVersion`</a> | The TLS version used by the connection (e.g. `1.2`) (if connection is TLS).   |
| <a id="opt-TLSCipher" href="#opt-TLSCipher" title="#opt-TLSCipher">`TLSCipher`</a> | The TLS cipher used by the connection (e.g. `TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`) (if connection is TLS).      |
| <a id="opt-TLSClientSubject" href="#opt-TLSClientSubject" title="#opt-TLSClientSubject">`TLSClientSubject`</a> | The string representation of the TLS client certificate's Subject (e.g. `CN=username,O=organization`).  |

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
