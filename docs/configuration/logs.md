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

```shell
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

To write JSON format logs, specify `json` as the format:

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

Access logs are written when `[accessLog]` is defined.
By default it will write to stdout and produce logs in the textual Common Log Format (CLF), extended with additional fields.

To enable access logs using the default settings just add the `[accessLog]` entry:

```toml
[accessLog]
```

To write the logs into a log file specify the `filePath`:

```toml
[accessLog]
filePath = "/path/to/access.log"
```

To write JSON format logs, specify `json` as the format:

```toml
[accessLog]
filePath = "/path/to/access.log"
format = "json"
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
format = "json"

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

To customize logs format:

```toml
[accessLog]
filePath = "/path/to/access.log"
format = "json"

  [accessLog.filters]

  # statusCodes keep only access logs with status codes in the specified range
  #
  # Optional
  # Default: []
  #
  statusCodes = ["200", "300-302"]

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


### List of all available fields

| Field                 | Description                                                                                                                                                                                                       |
|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| StartUTC              | StartUTC is the map key used for the time at which request processing started.                                                                                                                                    |
| StartLocal            | StartLocal is the map key used for the local time at which request processing started.                                                                                                                            |
| Duration              | Duration is the map key used for the total time taken by processing the response, including the origin server's time but not the log writing time.                                                                |
| FrontendName          | FrontendName is the map key used for the name of the Traefik frontend.                                                                                                                                            |
| BackendName           | BackendName is the map key used for the name of the Traefik backend.                                                                                                                                              |
| BackendURL            | BackendURL is the map key used for the URL of the Traefik backend.                                                                                                                                                |
| BackendAddr           | BackendAddr is the map key used for the IP:port of the Traefik backend (extracted from BackendURL)                                                                                                                |
| ClientAddr            | ClientAddr is the map key used for the remote address in its original form (usually IP:port).                                                                                                                     |
| ClientHost            | ClientHost is the map key used for the remote IP address from which the client request was received.                                                                                                              |
| ClientPort            | ClientPort is the map key used for the remote TCP port from which the client request was received.                                                                                                                |
| ClientUsername        | ClientUsername is the map key used for the username provided in the URL, if present.                                                                                                                              |
| RequestAddr           | RequestAddr is the map key used for the HTTP Host header (usually IP:port). This is treated as not a header by the Go API.                                                                                        |
| RequestHost           | RequestHost is the map key used for the HTTP Host server name (not including port).                                                                                                                               |
| RequestPort           | RequestPort is the map key used for the TCP port from the HTTP Host.                                                                                                                                              |
| RequestMethod         | RequestMethod is the map key used for the HTTP method.                                                                                                                                                            |
| RequestPath           | RequestPath is the map key used for the HTTP request URI, not including the scheme, host or port.                                                                                                                 |
| RequestProtocol       | RequestProtocol is the map key used for the version of HTTP requested.                                                                                                                                            |
| RequestLine           | RequestMethod + RequestPath + RequestProtocol                                                                                                                                                                     |
| RequestContentSize    | RequestContentSize is the map key used for the number of bytes in the request entity (a.k.a. body) sent by the client.                                                                                            |
| OriginDuration        | OriginDuration is the map key used for the time taken by the origin server ('upstream') to return its response.                                                                                                   |
| OriginContentSize     | OriginContentSize is the map key used for the content length specified by the origin server, or 0 if unspecified.                                                                                                 |
| OriginStatus          | OriginStatus is the map key used for the HTTP status code returned by the origin server. If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent.              |
| OriginStatusLine      | OriginStatus + Status code explanation                                                                                                                                                                            |
| DownstreamStatus      | DownstreamStatus is the map key used for the HTTP status code returned to the client.                                                                                                                             |
| DownstreamStatusLine  | DownstreamStatus + Status code explanation                                                                                                                                                                        |
| DownstreamContentSize | DownstreamContentSize is the map key used for the number of bytes in the response entity returned to the client. This is in addition to the "Content-Length" header, which may be present in the origin response. |
| RequestCount          | RequestCount is the map key used for the number of requests received since the Traefik instance started.                                                                                                          |
| GzipRatio             | GzipRatio is the map key used for the response body compression ratio achieved.                                                                                                                                   |
| Overhead              | Overhead is the map key used for the processing time overhead caused by Traefik.                                                                                                                                  |
| RetryAttempts         | RetryAttempts is the map key used for the amount of attempts the request was retried.                                                                                                                             |

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

### CLF - Common Log Format

By default, Traefik use the CLF (`common`) as access log format.

```html
<remote_IP_address> - <client_user_name_if_available> [<timestamp>] "<request_method> <request_path> <request_protocol>" <origin_server_HTTP_status> <origin_server_content_size> "<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started> "<Traefik_frontend_name>" "<Traefik_backend_URL>" <request_duration_in_ms>ms 
```


## Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! note
    This does not work on Windows due to the lack of USR signals.
