# Global Configuration

## Main Section

```toml
################################################################
# Global configuration
################################################################

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

# Traefik logs file
# If not defined, logs to stdout
#
# Optional
#
# traefikLogsFile = "log/traefik.log"

# Access logs file
#
# DEPRECATED - see [accessLog] lower down
# Optional
#
# accessLogsFile = "log/access.log"

# Log level
#
# Optional
# Default: "ERROR"
# Accepted values, in order of severity: "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"
# Messages at and above the selected level will be logged.
#
# logLevel = "ERROR"

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

# IdleTimeout
# 
# DEPRECATED - see [respondingTimeouts] section. In the case both settings are configured, the deprecated option will
# be overwritten.
#
# IdleTimeout is the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
# This is set to enforce closing of stale client connections.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "180s"
#
# IdleTimeout = "360s"

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

### Constraints

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

```toml
# Constraints definition
#
# Optional
#
# Simple matching constraint
# constraints = ["tag==api"]

# Simple mismatching constraint
# constraints = ["tag!=api"]

# Globbing
# constraints = ["tag==us-*"]

# Multiple constraints
#   - "tag==" must match with at least one tag
#   - "tag!=" must match with none of tags
# constraints = ["tag!=us-*", "tag!=asia-*"]

# Backend-specific constraint
# [consulCatalog]
#   endpoint = 127.0.0.1:8500
#   constraints = ["tag==api"]

# [consulCatalog]
#   endpoint = 127.0.0.1:8500
#   constraints = ["tag==api", "tag!=v*-beta"]
```

## Access Log Definition

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
  filePath   = "/path/to/access.log"
  format     = "json"
```

## Entry Points Definition

```toml
# Entrypoints definition
#
# Default:
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#
[entryPoints]
  [entryPoints.http]
  address = ":80"
```

### Redirect HTTP to HTTPS

```toml
# To redirect an http entrypoint to an https entrypoint (with SNI support):
#
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
      entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"
```

### Rewriting URL

```toml
# To redirect an entrypoint rewriting the URL:
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
      regex = "^http://localhost/(.*)"
      replacement = "http://mydomain/$1"
```

### TLS Mutual Authentication

```toml
# Only accept clients that present a certificate signed by a specified
# Certificate Authority (CA)
# ClientCAFiles can be configured with multiple CA:s in the same file or
# use multiple files containing one or several CA:s. The CA:s has to be in PEM format.
# All clients will be required to present a valid cert.
# The requirement will apply to all server certs in the entrypoint
# In the example below both snitest.com and snitest.org will require client certs
#
[entryPoints]
  [entryPoints.https]
  address = ":443"
  [entryPoints.https.tls]
  ClientCAFiles = ["tests/clientca1.crt", "tests/clientca2.crt"]
    [[entryPoints.https.tls.certificates]]
    CertFile = "integration/fixtures/https/snitest.com.cert"
    KeyFile = "integration/fixtures/https/snitest.com.key"
    [[entryPoints.https.tls.certificates]]
    CertFile = "integration/fixtures/https/snitest.org.cert"
    KeyFile = "integration/fixtures/https/snitest.org.key"
```

### Basic & Digest Authentication

```toml
# To enable basic auth on an entrypoint
#
# with 2 user/pass: test:test and test2:test2
# Passwords can be encoded in MD5, SHA1 and BCrypt: you can use htpasswd to generate those ones.
# Users can be specified directly in the toml file, or indirectly by referencing an external file;
# if both are provided, the two are merged, with external file contents having precedence.
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
  usersFile = "/path/to/.htpasswd"
```

```toml
# To enable digest auth on an entrypoint
#
# with 2 user/realm/pass: test:traefik:test and test2:traefik:test2
# You can use htdigest to generate those ones
# Users can be specified directly in the toml file, or indirectly by referencing an external file;
# if both are provided, the two are merged, with external file contents having precedence
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:traefik:a2688e031edb4be6a3797f3882655c05 ", "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"]
  usersFile = "/path/to/.htdigest"
```

### Specify Minimum TLS Version

```toml
# To specify an https entrypoint with a minimum TLS version, 
# and specifying an array of cipher suites (from crypto/tls):
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
    MinVersion = "VersionTLS12"
    CipherSuites = ["TLS_RSA_WITH_AES_256_GCM_SHA384"]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"
```

### Compression

```toml
# To enable compression support using gzip format:
[entryPoints]
  [entryPoints.http]
  address = ":80"
  compress = true
```

### Whitelisting

```toml
# To enable IP whitelisting at the entrypoint level:
[entryPoints]
  [entryPoints.http]
  address = ":80"
  whiteListSourceRange = ["127.0.0.1/32"]
```

### ProxyProtocol Support

```toml
# To enable ProxyProtocol support (https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt):
[entryPoints]
  [entryPoints.http]
  address = ":80"
  proxyprotocol = true
```

## Retry Configuration

```toml
# Enable retry sending request if network error
#
# Optional
#
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
#
# Optional
#
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

## Responding Timeouts

```toml
# respondingTimeouts are timeouts for incoming requests to the Traefik instance.
#
# Optional
# 
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

## Forwarding Timeouts

```toml
# forwardingTimeouts are timeouts for requests forwarded to the backend servers.
#
# Optional
# 
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