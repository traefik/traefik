# Global Configuration

## Main Section

```toml
# Duration to give active requests a chance to finish before Traefik stops.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
# If no units are provided, the value is parsed assuming seconds.
# Note: in this time frame no new requests are accepted.
#
# Optional
# Default: "10s"
#
# graceTimeOut = "10s"

# Enable debug mode
#
# Optional
# Default: false
#
# debug = true

# Periodically check if a new version has been released
#
# Optional
# Default: true
#
# checkNewVersion = false

# Backends throttle duration: minimum duration in seconds between 2 events from providers
# before applying a new configuration. It avoids unnecessary reloads if multiples events
# are sent in a short amount of time.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming
# seconds.
#
# Optional
# Default: "2s"
#
# ProvidersThrottleDuration = "2s"

# Controls the maximum idle (keep-alive) connections to keep per-host. If zero, DefaultMaxIdleConnsPerHost
# from the Go standard library net/http module is used.
# If you encounter 'too many open files' errors, you can either increase this
# value or change the `ulimit`.
#
# Optional
# Default: 200
#
# MaxIdleConnsPerHost = 200

# If set to true invalid SSL certificates are accepted for backends.
# Note: This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.
#
# Optional
# Default: false
#
# InsecureSkipVerify = true

# Register Certificates in the RootCA. This certificates will be use for backends calls.
# Note: You can use file path or cert content directly
# Optional
# Default: []
#
# RootCAs = [ "/mycert.cert" ]

# Entrypoints to be used by frontends that do not specify any entrypoint.
# Each frontend can specify its own entrypoints.
#
# Optional
# Default: ["http"]
#
# defaultEntryPoints = ["http", "https"]
```


## Constraints

In a micro-service architecture, with a central service discovery, setting constraints limits Træfik scope to a smaller number of routes.

Træfik filters services according to service attributes/tags set in your configuration backends.

Supported backends:

- Docker
- Consul K/V
- BoltDB
- Zookeeper
- Etcd
- Consul Catalog
- Rancher
- Marathon
- Kubernetes (using a provider-specific mechanism based on label selectors)

Supported filters:

- `tag`

### Simple

```toml
# Simple matching constraint
constraints = ["tag==api"]

# Simple mismatching constraint
constraints = ["tag!=api"]

# Globbing
constraints = ["tag==us-*"]
```

### Multiple

```toml
# Multiple constraints
#   - "tag==" must match with at least one tag
#   - "tag!=" must match with none of tags
constraints = ["tag!=us-*", "tag!=asia-*"]
```

### Backend-specific

```toml
# Backend-specific constraint
[consulCatalog]
endpoint = "127.0.0.1:8500"
constraints = ["tag==api"]

[marathon]
endpoint = "127.0.0.1:8800"
constraints = ["tag==api", "tag!=v*-beta"]
```


## Logs Definition

### Traefik logs

```toml
# Traefik logs file
# If not defined, logs to stdout
traefikLogsFile = "log/traefik.log"

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

### Access Logs

Access logs are written when `[accessLog]` is defined. 
By default it will write to stdout and produce logs in the textual Common Log Format (CLF), extended with additional fields.

To enable access logs using the default settings just add the `[accessLog]` entry.
```toml
[accessLog]
```

To write the logs into a logfile specify the `filePath`.
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

Deprecated way (before 1.4):
```toml
# Access logs file
#
# DEPRECATED - see [accessLog] lower down
#
accessLogsFile = "log/access.log"
```

### Log Rotation

Traefik will close and reopen its log files, assuming they're configured, on receipt of a USR1 signal.
This allows the logs to be rotated and processed by an external program, such as `logrotate`.

!!! note
    that this does not work on Windows due to the lack of USR signals.


## Custom Error pages

Custom error pages can be returned, in lieu of the default, according to frontend-configured ranges of HTTP Status codes.
In the example below, if a 503 status is returned from the frontend "website", the custom error page at http://2.3.4.5/503.html is returned with the actual status code set in the HTTP header.
Note, the `503.html` page itself is not hosted on traefik, but some other infrastructure.   

```toml
[frontends]
  [frontends.website]
  backend = "website"
  [frontends.website.errors]
    [frontends.website.errors.network]
    status = ["500-599"]
    backend = "error"
    query = "/{status}.html"
  [frontends.website.routes.website]
  rule = "Host: website.mydomain.com"

[backends]
  [backends.website]
    [backends.website.servers.website]
    url = "https://1.2.3.4"
  [backends.error]
    [backends.error.servers.error]
    url = "http://2.3.4.5"
```

In the above example, the error page rendered was based on the status code.
Instead, the query parameter can also be set to some generic error page like so: `query = "/500s.html"`

Now the `500s.html` error page is returned for the configured code range.
The configured status code ranges are inclusive; that is, in the above example, the `500s.html` page will be returned for status codes `500` through, and including, `599`.


## Retry Configuration

```toml
# Enable retry sending request if network error
[retry]

# Number of attempts
#
# Optional
# Default: (number servers in backend) -1
#
# attempts = 3
```


## Health Check Configuration

```toml
# Enable custom health check options.
[healthcheck]

# Set the default health check interval. Will only be effective if health check
# paths are defined. Given provider-specific support, the value may be
# overridden on a per-backend basis.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming
# seconds.
#
# Optional
# Default: "30s"
#
# interval = "30s"
```


## Timeouts

### Responding Timeouts

`respondingTimeouts` are timeouts for incoming requests to the Traefik instance.

```toml
[respondingTimeouts]

# readTimeout is the maximum duration for reading the entire request, including the body.
# If zero, no timeout exists.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
# 
# Optional
# Default: "0s"
# 
# readTimeout = "5s"

# writeTimeout is the maximum duration before timing out writes of the response. It covers the time from the end of 
# the request header read to the end of the response write.
# If zero, no timeout exists.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "0s"
# 
# writeTimeout = "5s"

# idleTimeout is the maximum duration an idle (keep-alive) connection will remain idle before closing itself.
# If zero, no timeout exists.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "180s"
#
# idleTimeout = "360s"
```

### Forwarding Timeouts

`forwardingTimeouts` are timeouts for requests forwarded to the backend servers.

```toml
[forwardingTimeouts]

# dialTimeout is the amount of time to wait until a connection to a backend server can be established. 
# If zero, no timeout exists.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
# 
# Optional
# Default: "30s"
# 
# dialTimeout = "30s"

# responseHeaderTimeout is the amount of time to wait for a server's response headers after fully writing the request (including its body, if any). 
# If zero, no timeout exists.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "0s"
# 
# responseHeaderTimeout = "0s"
```

### Idle Timeout (deprecated)

Use [respondingTimeouts](/configuration/commons/#responding-timeouts) instead of `IdleTimeout`. 
In the case both settings are configured, the deprecated option will be overwritten.

`IdleTimeout` is the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
This is set to enforce closing of stale client connections.

Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

```toml
# IdleTimeout
# 
# DEPRECATED - see [respondingTimeouts] section.
#
# Optional
# Default: "180s"
#
IdleTimeout = "360s"
```


## Override Default Configuration Template

!!! warning
    For advanced users only.

Supported by all backends except: File backend, Web backend and DynamoDB backend.

```toml
[backend_name]

# Override default configuration template. For advanced users :)
#
# Optional
# Default: ""
#
filename = "custom_config_template.tpml"

# Enable debug logging of generated configuration template.
#
# Optional
# Default: false
#
debugLogGeneratedTemplate = true
```

Example:

```toml
[marathon]
filename = "my_custom_config_template.tpml"
```

The template files can be written using functions provided by:

- [go template](https://golang.org/pkg/text/template/) 
- [sprig library](https://masterminds.github.io/sprig/)

Example:

```tmpl
[backends]
  [backends.backend1]
  url = "http://firstserver"
  [backends.backend2]
  url = "http://secondserver"

{{$frontends := dict "frontend1" "backend1" "frontend2" "backend2"}}
[frontends]
{{range $frontend, $backend := $frontends}}
  [frontends.{{$frontend}}]
  backend = "{{$backend}}"
{{end}}
```
