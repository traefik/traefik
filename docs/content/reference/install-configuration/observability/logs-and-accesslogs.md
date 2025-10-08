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
| <a id="log-filePath" href="#log-filePath" title="#log-filePath">`log.filePath`</a> | By default, the logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option.| - | No      |
| <a id="log-format" href="#log-format" title="#log-format">`log.format`</a> | Log format (`common`or `json`).<br /> The fields displayed with the format `common` cannot be customized. | "common" | No      |
| <a id="log-level" href="#log-level" title="#log-level">`log.level`</a> | Log level (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, and `PANIC`)| ERROR | No      |
| <a id="log-noColor" href="#log-noColor" title="#log-noColor">`log.noColor`</a> | When using the format `common`, disables the colorized output. | false      | No      |
| <a id="log-maxSize" href="#log-maxSize" title="#log-maxSize">`log.maxSize`</a> | Maximum size in megabytes of the log file before it gets rotated. | 100MB      | No      |
| <a id="log-maxAge" href="#log-maxAge" title="#log-maxAge">`log.maxAge`</a> | Maximum number of days to retain old log files based on the timestamp encoded in their filename.<br /> A day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc.<br />By default files are not removed based on their age.  |   0   | No      |
| <a id="log-maxBackups" href="#log-maxBackups" title="#log-maxBackups">`log.maxBackups`</a> | Maximum number of old log files to retain.<br />The default is to retain all old log files. |  0  | No      |
| <a id="log-compress" href="#log-compress" title="#log-compress">`log.compress`</a> | Compress log files in gzip after rotation. | false | No      |

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
| <a id="log-otlp-serviceName" href="#log-otlp-serviceName" title="#log-otlp-serviceName">`log.otlp.serviceName`</a> | Service name used in selected backend.                                                                                                 | "traefik"                        | No       |
| <a id="log-otlp-resourceAttributes" href="#log-otlp-resourceAttributes" title="#log-otlp-resourceAttributes">`log.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| <a id="log-otlp-http" href="#log-otlp-http" title="#log-otlp-http">`log.otlp.http`</a> | This instructs the exporter to send logs to the OpenTelemetry Collector using HTTP.                                                    |                                  | No       |
| <a id="log-otlp-http-endpoint" href="#log-otlp-http-endpoint" title="#log-otlp-http-endpoint">`log.otlp.http.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| <a id="log-otlp-http-headers" href="#log-otlp-http-headers" title="#log-otlp-http-headers">`log.otlp.http.headers`</a> | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| <a id="log-otlp-http-tls" href="#log-otlp-http-tls" title="#log-otlp-http-tls">`log.otlp.http.tls`</a> | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| <a id="log-otlp-http-tls-ca" href="#log-otlp-http-tls-ca" title="#log-otlp-http-tls-ca">`log.otlp.http.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="log-otlp-http-tls-cert" href="#log-otlp-http-tls-cert" title="#log-otlp-http-tls-cert">`log.otlp.http.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="log-otlp-http-tls-key" href="#log-otlp-http-tls-key" title="#log-otlp-http-tls-key">`log.otlp.http.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="log-otlp-http-tls-insecureSkipVerify" href="#log-otlp-http-tls-insecureSkipVerify" title="#log-otlp-http-tls-insecureSkipVerify">`log.otlp.http.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| <a id="log-otlp-grpc" href="#log-otlp-grpc" title="#log-otlp-grpc">`log.otlp.grpc`</a> | This instructs the exporter to send logs to the OpenTelemetry Collector using gRPC.                                                    |                                  | No       |
| <a id="log-otlp-grpc-endpoint" href="#log-otlp-grpc-endpoint" title="#log-otlp-grpc-endpoint">`log.otlp.grpc.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| <a id="log-otlp-grpc-headers" href="#log-otlp-grpc-headers" title="#log-otlp-grpc-headers">`log.otlp.grpc.headers`</a> | Additional headers sent with logs by the exporter to the OpenTelemetry Collector.                                                      | [ ]                              | No       |
| <a id="log-otlp-grpc-insecure" href="#log-otlp-grpc-insecure" title="#log-otlp-grpc-insecure">`log.otlp.grpc.insecure`</a> | Instructs the exporter to send logs to the OpenTelemetry Collector using an insecure protocol.                                         | false                            | No       |
| <a id="log-otlp-grpc-tls" href="#log-otlp-grpc-tls" title="#log-otlp-grpc-tls">`log.otlp.grpc.tls`</a> | Defines the Client TLS configuration used by the exporter to send logs to the OpenTelemetry Collector.                                 |                                  | No       |
| <a id="log-otlp-grpc-tls-ca" href="#log-otlp-grpc-tls-ca" title="#log-otlp-grpc-tls-ca">`log.otlp.grpc.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="log-otlp-grpc-tls-cert" href="#log-otlp-grpc-tls-cert" title="#log-otlp-grpc-tls-cert">`log.otlp.grpc.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="log-otlp-grpc-tls-key" href="#log-otlp-grpc-tls-key" title="#log-otlp-grpc-tls-key">`log.otlp.grpc.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="log-otlp-grpc-tls-insecureSkipVerify" href="#log-otlp-grpc-tls-insecureSkipVerify" title="#log-otlp-grpc-tls-insecureSkipVerify">`log.otlp.grpc.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

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
| <a id="accesslog-filePath" href="#accesslog-filePath" title="#accesslog-filePath">`accesslog.filePath`</a> | By default, the access logs are written to the standard output.<br />You can configure a file path instead using the `filePath` option.|  | No      |
| <a id="accesslog-format" href="#accesslog-format" title="#accesslog-format">`accesslog.format`</a> | By default, logs are written using the Traefik Common Log Format (CLF).<br />Available formats: [`common`](#traefik-clf-format-fields) (Traefik extended CLF), [`genericCLF`](#generic-clf-format-fields) (standard CLF compatible with analyzers), or [`json`](#json-format-fields).<br />If the given format is unsupported, the default (`common`) is used instead. | "common" | No      |
| <a id="accesslog-bufferingSize" href="#accesslog-bufferingSize" title="#accesslog-bufferingSize">`accesslog.bufferingSize`</a> | To write the logs in an asynchronous fashion, specify a  `bufferingSize` option.<br />This option represents the number of log lines Traefik will keep in memory before writing them to the selected output.<br />In some cases, this option can greatly help performances.| 0 | No      |
| <a id="accesslog-addInternals" href="#accesslog-addInternals" title="#accesslog-addInternals">`accesslog.addInternals`</a> | Enables access logs for internal resources (e.g.: `ping@internal`). | false  | No      |
| <a id="accesslog-filters-statusCodes" href="#accesslog-filters-statusCodes" title="#accesslog-filters-statusCodes">`accesslog.filters.statusCodes`</a> | Limit the access logs to requests with a status codes in the specified range. | [ ]      | No      |
| <a id="accesslog-filters-retryAttempts" href="#accesslog-filters-retryAttempts" title="#accesslog-filters-retryAttempts">`accesslog.filters.retryAttempts`</a> | Keep the access logs when at least one retry has happened. | false      | No      |
| <a id="accesslog-filters-minDuration" href="#accesslog-filters-minDuration" title="#accesslog-filters-minDuration">`accesslog.filters.minDuration`</a> | Keep access logs when requests take longer than the specified duration (provided in seconds or as a valid duration format, see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)).  |  0   | No      |
| <a id="accesslog-fields-defaultMode" href="#accesslog-fields-defaultMode" title="#accesslog-fields-defaultMode">`accesslog.fields.defaultMode`</a> | Mode to apply by default to the access logs fields (`keep`, `redact` or `drop`). | keep | No      |
| <a id="accesslog-fields-names" href="#accesslog-fields-names" title="#accesslog-fields-names">`accesslog.fields.names`</a> | Set the fields list to display in the access logs (format `name:mode`).<br /> Available fields list [here](#json-format-fields). |  [ ]    | No      |
| <a id="accesslog-fields-headers-defaultMode" href="#accesslog-fields-headers-defaultMode" title="#accesslog-fields-headers-defaultMode">`accesslog.fields.headers.defaultMode`</a> | Mode to apply by default to the access logs headers (`keep`, `redact` or `drop`).  | drop | No      |
| <a id="accesslog-fields-headers-names" href="#accesslog-fields-headers-names" title="#accesslog-fields-headers-names">`accesslog.fields.headers.names`</a> | Set the headers list to display in the access logs (format `name:mode`). |   [ ]   | No      |

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
| <a id="accesslog-otlp-serviceName" href="#accesslog-otlp-serviceName" title="#accesslog-otlp-serviceName">`accesslog.otlp.serviceName`</a> | Defines the service name resource attribute.                                                                                           | "traefik"                        | No       |
| <a id="accesslog-otlp-resourceAttributes" href="#accesslog-otlp-resourceAttributes" title="#accesslog-otlp-resourceAttributes">`accesslog.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                    | []                               | No       |
| <a id="accesslog-otlp-http" href="#accesslog-otlp-http" title="#accesslog-otlp-http">`accesslog.otlp.http`</a> | This instructs the exporter to send access logs to the OpenTelemetry Collector using HTTP.                                             |                                  | No       |
| <a id="accesslog-otlp-http-endpoint" href="#accesslog-otlp-http-endpoint" title="#accesslog-otlp-http-endpoint">`accesslog.otlp.http.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<scheme>://<host>:<port><path>`)                                                 | `https://localhost:4318/v1/logs` | No       |
| <a id="accesslog-otlp-http-headers" href="#accesslog-otlp-http-headers" title="#accesslog-otlp-http-headers">`accesslog.otlp.http.headers`</a> | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| <a id="accesslog-otlp-http-tls" href="#accesslog-otlp-http-tls" title="#accesslog-otlp-http-tls">`accesslog.otlp.http.tls`</a> | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| <a id="accesslog-otlp-http-tls-ca" href="#accesslog-otlp-http-tls-ca" title="#accesslog-otlp-http-tls-ca">`accesslog.otlp.http.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="accesslog-otlp-http-tls-cert" href="#accesslog-otlp-http-tls-cert" title="#accesslog-otlp-http-tls-cert">`accesslog.otlp.http.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="accesslog-otlp-http-tls-key" href="#accesslog-otlp-http-tls-key" title="#accesslog-otlp-http-tls-key">`accesslog.otlp.http.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="accesslog-otlp-http-tls-insecureSkipVerify" href="#accesslog-otlp-http-tls-insecureSkipVerify" title="#accesslog-otlp-http-tls-insecureSkipVerify">`accesslog.otlp.http.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |
| <a id="accesslog-otlp-grpc" href="#accesslog-otlp-grpc" title="#accesslog-otlp-grpc">`accesslog.otlp.grpc`</a> | This instructs the exporter to send access logs to the OpenTelemetry Collector using gRPC.                                             |                                  | No       |
| <a id="accesslog-otlp-grpc-endpoint" href="#accesslog-otlp-grpc-endpoint" title="#accesslog-otlp-grpc-endpoint">`accesslog.otlp.grpc.endpoint`</a> | The endpoint of the OpenTelemetry Collector. (format=`<host>:<port>`)                                                                  | `localhost:4317`                 | No       |
| <a id="accesslog-otlp-grpc-headers" href="#accesslog-otlp-grpc-headers" title="#accesslog-otlp-grpc-headers">`accesslog.otlp.grpc.headers`</a> | Additional headers sent with access logs by the exporter to the OpenTelemetry Collector.                                               | [ ]                              | No       |
| <a id="accesslog-otlp-grpc-insecure" href="#accesslog-otlp-grpc-insecure" title="#accesslog-otlp-grpc-insecure">`accesslog.otlp.grpc.insecure`</a> | Instructs the exporter to send access logs to the OpenTelemetry Collector using an insecure protocol.                                  | false                            | No       |
| <a id="accesslog-otlp-grpc-tls" href="#accesslog-otlp-grpc-tls" title="#accesslog-otlp-grpc-tls">`accesslog.otlp.grpc.tls`</a> | Defines the Client TLS configuration used by the exporter to send access logs to the OpenTelemetry Collector.                          |                                  | No       |
| <a id="accesslog-otlp-grpc-tls-ca" href="#accesslog-otlp-grpc-tls-ca" title="#accesslog-otlp-grpc-tls-ca">`accesslog.otlp.grpc.tls.ca`</a> | The path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. |                                  | No       |
| <a id="accesslog-otlp-grpc-tls-cert" href="#accesslog-otlp-grpc-tls-cert" title="#accesslog-otlp-grpc-tls-cert">`accesslog.otlp.grpc.tls.cert`</a> | The path to the certificate to use for the OpenTelemetry Collector.                                                                    |                                  | No       |
| <a id="accesslog-otlp-grpc-tls-key" href="#accesslog-otlp-grpc-tls-key" title="#accesslog-otlp-grpc-tls-key">`accesslog.otlp.grpc.tls.key`</a> | The path to the key to use for the OpenTelemetry Collector.                                                                            |                                  | No       |
| <a id="accesslog-otlp-grpc-tls-insecureSkipVerify" href="#accesslog-otlp-grpc-tls-insecureSkipVerify" title="#accesslog-otlp-grpc-tls-insecureSkipVerify">`accesslog.otlp.grpc.tls.insecureSkipVerify`</a> | Instructs the OpenTelemetry Collector to accept any certificate presented by the server regardless of the hostname in the certificate. | false                            | No       |

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
| <a id="StartUTC" href="#StartUTC" title="#StartUTC">`StartUTC`</a> | The time at which request processing started.                                                                                                                       |
| <a id="StartLocal" href="#StartLocal" title="#StartLocal">`StartLocal`</a> | The local time at which request processing started.                                                                                                                 |
| <a id="Duration" href="#Duration" title="#Duration">`Duration`</a> | The total time taken (in nanoseconds) by processing the response, including the origin server's time but not the log writing time.                                  |
| <a id="RouterName" href="#RouterName" title="#RouterName">`RouterName`</a> | The name of the Traefik  router.                                                                                                                                    |
| <a id="ServiceName" href="#ServiceName" title="#ServiceName">`ServiceName`</a> | The name of the Traefik backend.          |
| <a id="ServiceURL" href="#ServiceURL" title="#ServiceURL">`ServiceURL`</a> | The URL of the Traefik backend.       |
| <a id="ServiceAddr" href="#ServiceAddr" title="#ServiceAddr">`ServiceAddr`</a> | The IP:port of the Traefik backend (extracted from `ServiceURL`). |
| <a id="ClientAddr" href="#ClientAddr" title="#ClientAddr">`ClientAddr`</a> | The remote address in its original form (usually IP:port).     |
| <a id="ClientHost" href="#ClientHost" title="#ClientHost">`ClientHost`</a> | The remote IP address from which the client request was received.     |
| <a id="ClientPort" href="#ClientPort" title="#ClientPort">`ClientPort`</a> | The remote TCP port from which the client request was received.   |
| <a id="ClientUsername" href="#ClientUsername" title="#ClientUsername">`ClientUsername`</a> | The username provided in the URL, if present.   |
| <a id="RequestAddr" href="#RequestAddr" title="#RequestAddr">`RequestAddr`</a> | The HTTP Host header (usually IP:port). This is treated as not a header by the Go API.   |
| <a id="RequestHost" href="#RequestHost" title="#RequestHost">`RequestHost`</a> | The HTTP Host server name (not including port).     |
| <a id="RequestPort" href="#RequestPort" title="#RequestPort">`RequestPort`</a> | The TCP port from the HTTP Host.    |
| <a id="RequestMethod" href="#RequestMethod" title="#RequestMethod">`RequestMethod`</a> | The HTTP method. |
| <a id="RequestPath" href="#RequestPath" title="#RequestPath">`RequestPath`</a> | The HTTP request URI, not including the scheme, host or port.   |
| <a id="RequestProtocol" href="#RequestProtocol" title="#RequestProtocol">`RequestProtocol`</a> | The version of HTTP requested.       |
| <a id="RequestScheme" href="#RequestScheme" title="#RequestScheme">`RequestScheme`</a> | The HTTP scheme requested `http` or `https`.   |
| <a id="RequestLine" href="#RequestLine" title="#RequestLine">`RequestLine`</a> | The `RequestMethod`, + `RequestPath` and `RequestProtocol`.   |
| <a id="RequestContentSize" href="#RequestContentSize" title="#RequestContentSize">`RequestContentSize`</a> | The number of bytes in the request entity (a.k.a. body) sent by the client.   |
| <a id="OriginDuration" href="#OriginDuration" title="#OriginDuration">`OriginDuration`</a> | The time taken (in nanoseconds) by the origin server ('upstream') to return its response. |
| <a id="OriginContentSize" href="#OriginContentSize" title="#OriginContentSize">`OriginContentSize`</a> | The content length specified by the origin server, or 0 if unspecified.    |
| <a id="OriginStatus" href="#OriginStatus" title="#OriginStatus">`OriginStatus`</a> | The HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent (0). |
| <a id="OriginStatusLine" href="#OriginStatusLine" title="#OriginStatusLine">`OriginStatusLine`</a> | `OriginStatus` + Status code explanation   |
| <a id="DownstreamStatus" href="#DownstreamStatus" title="#DownstreamStatus">`DownstreamStatus`</a> | The HTTP status code returned to the client.    |
| <a id="DownstreamStatusLine" href="#DownstreamStatusLine" title="#DownstreamStatusLine">`DownstreamStatusLine`</a> | The `DownstreamStatus` and status code explanation.     |
| <a id="DownstreamContentSize" href="#DownstreamContentSize" title="#DownstreamContentSize">`DownstreamContentSize`</a> | The number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
| <a id="RequestCount" href="#RequestCount" title="#RequestCount">`RequestCount`</a> | The number of requests received since the Traefik instance started.    |
| <a id="GzipRatio" href="#GzipRatio" title="#GzipRatio">`GzipRatio`</a> | The response body compression ratio achieved.   |
| <a id="Overhead" href="#Overhead" title="#Overhead">`Overhead`</a> | The processing time overhead (in nanoseconds) caused by Traefik.    |
| <a id="RetryAttempts" href="#RetryAttempts" title="#RetryAttempts">`RetryAttempts`</a> | The amount of attempts the request was retried.   |
| <a id="TLSVersion" href="#TLSVersion" title="#TLSVersion">`TLSVersion`</a> | The TLS version used by the connection (e.g. `1.2`) (if connection is TLS).   |
| <a id="TLSCipher" href="#TLSCipher" title="#TLSCipher">`TLSCipher`</a> | The TLS cipher used by the connection (e.g. `TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`) (if connection is TLS).      |
| <a id="TLSClientSubject" href="#TLSClientSubject" title="#TLSClientSubject">`TLSClientSubject`</a> | The string representation of the TLS client certificate's Subject (e.g. `CN=username,O=organization`).  |

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
