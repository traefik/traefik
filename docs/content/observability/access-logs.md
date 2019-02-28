# Access Logs

Who Calls Whom?
{.subtitle}

By default, logs are written to stdout, in text format.

## Configuration Examples

??? example "Enabling Access Logs"

    ```toml
    [accessLog]
    ```

## Configuration Options 

### filePath

By default access logs are written to the standard output.
To write the logs into a log file, use the `filePath` option.

in the Common Log Format (CLF), extended with additional fields.

### format
 
By default, logs are written using the Common Log Format (CLF).
To write logs in JSON, use `json` in the `format` option.

!!! note "Common Log Format"

#### CLF - Common Log Format

    ```html
    <remote_IP_address> - <client_user_name_if_available> [<timestamp>] "<request_method> <request_path> <request_protocol>" <origin_server_HTTP_status> <origin_server_content_size> "<request_referrer>" "<request_user_agent>" <number_of_requests_received_since_Traefik_started> "<Traefik_frontend_name>" "<Traefik_backend_URL>" <request_duration_in_ms>ms 
    ```

#### bufferingSize

To write the logs in an asynchronous fashion, specify a  `bufferingSize` option.
This option represents the number of log lines Traefik will keep in memory before writing them to the selected output.
In some cases, this option can greatly help performances.

??? example "Configuring a buffer of 100 lines"

    ```toml
    [accessLog]
    filePath = "/path/to/access.log"
    bufferingSize = 100
    ```

#### Filtering

To filter logs, you can specify a set of filters which are logically "OR-connected". 
Thus, specifying multiple filters will keep more access logs than specifying only one.

The available filters are: 

- `statusCodes`, to limit the access logs to requests with a status codes in the specified range
- `retryAttempts`, to keep the access logs when at least one retry has happened
- `minDuration`, to keep access logs when requests take longer than the specified duration

??? example "Configuring Multiple Filters"

    ```toml
    [accessLog]
    filePath = "/path/to/access.log"
    format = "json"
    
      [accessLog.filters]    
        statusCodes = ["200", "300-302"]
        retryAttempts = true
        minDuration = "10ms"
    ```

#### Limiting the Fields

You can decide to limit the logged fields/headers to a given list with the `fields.names` and `fields.header` options

Each field can be set to:

- `keep` to keep the value
- `drop` to drop the value
- `redact` to replace the value with "redacted"

??? example "Limiting the Logs to Specific Fields"

    ```toml
    [accessLog]
        filePath = "/path/to/access.log"
        format = "json"
        
        [accessLog.filters]
            statusCodes = ["200", "300-302"]
    
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
    
??? list "Available Fields"

    ```ini
    StartUTC
    StartLocal
    Duration
    FrontendName
    BackendName
    BackendURL
    BackendAddr
    ClientAddr
    ClientHost
    ClientPort
    ClientUsername
    RequestAddr
    RequestHost
    RequestPort
    RequestMethod
    RequestPath
    RequestProtocol
    RequestLine
    RequestContentSize
    OriginDuration
    OriginContentSize
    OriginStatus
    OriginStatusLine
    DownstreamStatus
    DownstreamStatusLine
    DownstreamContentSize
    RequestCount
    GzipRatio
    Overhead
    RetryAttempts
    ```

## Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! note
    This does not work on Windows due to the lack of USR signals.
