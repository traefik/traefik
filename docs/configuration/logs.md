# Logs Definition

## Reference

### TOML

```toml
logLevel = "INFO"

[traefikLog]
  filePath = "/path/to/traefik.log"
  format   = "json"

[accessLog]
  filePath = "/path/to/access.log"
  format = "json"

  [accessLog.filters]
    statusCodes = ["200", "300-302"]
    retryAttempts = true
    minDuration = "10ms"

  [accessLog.fields]
    defaultMode = "keep"
    [accessLog.fields.names]
      "ClientUsername" = "drop"
      # ...

    [accessLog.fields.headers]
      defaultMode = "keep"
      [accessLog.fields.headers.names]
        "User-Agent" = "redact"
        "Authorization" = "drop"
        "Content-Type" = "keep"
        # ...
```

### CLI

For more information about the CLI, see the documentation about [Traefik command](/basics/#traefik).

```bash
--logLevel="DEBUG"
--traefikLog.filePath="/path/to/traefik.log"
--traefikLog.format="json"
--accessLog.filePath="/path/to/access.log"
--accessLog.format="json"
--accessLog.filters.statusCodes="200,300-302"
--accessLog.filters.retryAttempts="true"
--accessLog.filters.minDuration="10ms"
--accessLog.fields.defaultMode="keep"
--accessLog.fields.names="Username=drop Hostname=drop"
--accessLog.fields.headers.defaultMode="keep"
--accessLog.fields.headers.names="User-Agent=redact Authorization=drop Content-Type=keep"
```

## Traefik Logs

By default the Traefik log is written to stdout in text format.

To write the logs into a log file specify the `filePath`:

```toml
[traefikLog]
  filePath = "/path/to/traefik.log"
```

To switch to JSON format instead of standard format (`common`), specify `json` as the format:

```toml
[traefikLog]
  filePath = "/path/to/traefik.log"
  format   = "json"
```

Deprecated way (before 1.4):

!!! danger "DEPRECATED"
    `traefikLogsFile` is deprecated, use [traefikLog](/configuration/logs/#traefik-logs) instead.

```toml
# Traefik logs file
# If not defined, logs to stdout
#
# DEPRECATED - see [traefikLog] lower down
# In case both traefikLogsFile and traefikLog.filePath are specified, the latter will take precedence.
# Optional
#
traefikLogsFile = "log/traefik.log"
```

To customize the log level:

```toml
# Log level
#
# Optional
# Default: "ERROR"
#
# Accepted values, in order of severity: "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"
# Messages at and above the selected level will be logged.
#
logLevel = "ERROR"
```

## Access Logs

Access logs are written when the entry `[accessLog]` is defined (or the command line flag `--accesslog`).
By default it writes to stdout and produces logs in the textual [Common Log Format (CLF)](#clf-common-log-format), extended with additional fields.

To enable access logs using the default settings, add the `[accessLog]` entry in your `traefik.toml` configuration file:

```toml
[accessLog]
```

To write the logs into a log file specify the `filePath`:

```toml
[accessLog]
filePath = "/path/to/access.log"
```

To switch to JSON format instead of [Common Log Format (CLF)](#clf-common-log-format), specify `json` as the format:

```toml
[accessLog]
filePath = "/path/to/access.log"
format = "json"  # Default: "common"
```

To write the logs in async, specify `bufferingSize` as the format (must be >0):

```toml
[accessLog]
filePath = "/path/to/access.log"
# Buffering Size
#
# Optional
# Default: 0
#
# Number of access log lines to process in a buffered way.
#
bufferingSize = 100
```

To filter logs you can specify a set of filters which are logically "OR-connected". Thus, specifying multiple filters will keep more access logs than specifying only one:

```toml
[accessLog]
filePath = "/path/to/access.log"
format = "json"  # Default: "common"

  [accessLog.filters]

  # statusCodes: keep access logs with status codes in the specified range
  #
  # Optional
  # Default: []
  #
  statusCodes = ["200", "300-302"]

  # retryAttempts: keep access logs when at least one retry happened
  #
  # Optional
  # Default: false
  #
  retryAttempts = true

  # minDuration: keep access logs when request took longer than the specified duration
  #
  # Optional
  # Default: 0
  #
  minDuration = "10ms"
```

### CLF - Common Log Format

By default, Traefik use the CLF (`common`) as access log format.

```html
<remote_IP_address> - <client_user_name_if_available> [<timestamp>] "<request_method> <request_path> <request_protocol>" <origin_server_HTTP_status> <origin_server_content_size> "<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started> "<Traefik_frontend_name>" "<Traefik_backend_URL>" <request_duration_in_ms>ms
```

### Customize Fields

You can customize the fields written in the access logs.
The list of available fields is found below: [List of All Available Fields](#list-of-all-available-fields).

Each field has a "mode" which defines if it is written or not in the access log lines.
The possible values for the mode are:

* `keep`: the field and its value are written on the access log line. This is the default behavior.
* `drop`: the field is not written at all on the access log.

To customize the fields, you must:

* Switch to the JSON format (mandatory)
* Define the "default mode" for all fields (default is `keep`)
* OR Define the fields which does not follow the default mode

```toml
[accessLog]
# Access Log Format
#
# Optional
# Default: "common"
#
# Accepted values "common", "json"
#
format = "json"

  [accessLog.fields]

  # defaultMode
  #
  # Optional
  # Default: "keep"
  #
  # Accepted values "keep", "drop"
  #
  defaultMode = "keep"

  # Fields map which is used to override fields defaultMode
  [accessLog.fields.names]
    "ClientUsername" = "drop"
    # ...
```

### Customize Headers

Access logs prints the headers of each request, as fields of the access log line.
You can customize which and how the headers are printed, likewise the other fields (see ["Customize Fields" section](#customize-fields)).

Each header has a "mode" which defines how it is written in the access log lines.
The possible values for the mode are:

* `keep`: the header and its value are written on the access log line. This is the default behavior.
* `drop`: the header is not written at all on the access log.
* `redacted`: the header is written, but its value is redacted to avoid leaking sensitive information.

To customize the headers, you must:

* Switch to the JSON format (mandatory)
* Define the "default mode" for all headers (default is `keep`)
* OR Define the headers which does not follow the default mode

!!! important
    The headers are written with the prefix `request_` in the access log.
    This prefix must not be included when specifying a header in the TOML configuration.

    * Do: `"User-Agent" = "drop"`
    * Don't: `"redacted_User-Agent" = "drop"`

```toml
[accessLog]
# Access Log Format
#
# Optional
# Default: "common"
#
# Accepted values "common", "json"
#
format = "json"

  [accessLog.fields.headers]
    # defaultMode
    #
    # Optional
    # Default: "keep"
    #
    # Accepted values "keep", "drop", "redact"
    #
    defaultMode = "keep"
    # Fields map which is used to override headers defaultMode
    [accessLog.fields.headers.names]
      "User-Agent" = "redact"
      "Authorization" = "drop"
      "Content-Type" = "keep"
      # ...
```

### List of All Available Fields

| Field                   | Description                                                                                                                                                         |
|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `StartUTC`              | The time at which request processing started.                                                                                                                       |
| `StartLocal`            | The local time at which request processing started.                                                                                                                 |
| `Duration`              | The total time taken by processing the response, including the origin server's time but not the log writing time.                                                   |
| `FrontendName`          | The name of the Traefik frontend.                                                                                                                                   |
| `BackendName`           | The name of the Traefik backend.                                                                                                                                    |
| `BackendURL`            | The URL of the Traefik backend.                                                                                                                                     |
| `BackendAddr`           | The IP:port of the Traefik backend (extracted from `BackendURL`)                                                                                                    |
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
| `RequestLine`           | `RequestMethod` + `RequestPath` + `RequestProtocol`                                                                                                                 |
| `RequestContentSize`    | The number of bytes in the request entity (a.k.a. body) sent by the client.                                                                                         |
| `OriginDuration`        | The time taken by the origin server ('upstream') to return its response.                                                                                            |
| `OriginContentSize`     | The content length specified by the origin server, or 0 if unspecified.                                                                                             |
| `OriginStatus`          | The HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent.     |
| `OriginStatusLine`      | `OriginStatus` + Status code explanation                                                                                                                            |
| `DownstreamStatus`      | The HTTP status code returned to the client.                                                                                                                        |
| `DownstreamStatusLine`  | `DownstreamStatus` + Status code explanation                                                                                                                        |
| `DownstreamContentSize` | The number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
| `RequestCount`          | The number of requests received since the Traefik instance started.                                                                                                 |
| `GzipRatio`             | The response body compression ratio achieved.                                                                                                                       |
| `Overhead`              | The processing time overhead caused by Traefik.                                                                                                                     |
| `RetryAttempts`         | The amount of attempts the request was retried.                                                                                                                     |

### Depreciation Notice

Deprecated way (before 1.4):

!!! danger "DEPRECATED"
    `accessLogsFile` is deprecated, use [accessLog](/configuration/logs/#access-logs) instead.

```toml
# Access logs file
#
# DEPRECATED - see [accessLog]
#
accessLogsFile = "log/access.log"
```

## Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! note
    This does not work on Windows due to the lack of USR signals.

## Time Zones

The timestamp of each log line is in UTC time by default.

If you want to use local timezone, you need to ensure the 3 following elements:

1. Provide the timezone data into /usr/share/zoneinfo
2. Set the environement variable TZ to the timezone to be used
3. Specify the field StartLocal instead of StartUTC (works on default Common Log Format (CLF) as well as JSON)

Example using docker-compose:

```yml
version: '3'

services:
  traefik:
    image: traefik/traefik:[latest stable version]
    ports:
      - "80:80"
    environment:
      - "TZ=US/Alaska"
    command:
      - --docker
      - --accesslog
      - --accesslog.fields.names="StartUTC=drop"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "/usr/share/zoneinfo:/usr/share/zoneinfo:ro"
```
