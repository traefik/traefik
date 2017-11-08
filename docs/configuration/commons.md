# Global Configuration

## Main Section

```toml
# DEPRECATED - for general usage instruction see [lifeCycle.graceTimeOut].
#
# If both the deprecated option and the new one are given, the deprecated one
# takes precedence.
# A value of zero is equivalent to omitting the parameter, causing
# [lifeCycle.graceTimeOut] to be effective. Pass zero to the new option in
# order to disable the grace period.
#
# Optional
# Default: "0s"
#
# graceTimeOut = "10s"

# Enable debug mode.
#
# Optional
# Default: false
#
# debug = true

# Periodically check if a new version has been released.
#
# Optional
# Default: true
#
# checkNewVersion = false

# Backends throttle duration.
#
# Optional
# Default: "2s"
#
# ProvidersThrottleDuration = "2s"

# Controls the maximum idle (keep-alive) connections to keep per-host.
#
# Optional
# Default: 200
#
# MaxIdleConnsPerHost = 200

# If set to true invalid SSL certificates are accepted for backends.
# This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.
#
# Optional
# Default: false
#
# InsecureSkipVerify = true

# Register Certificates in the RootCA.
#
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

- `graceTimeOut`: Duration to give active requests a chance to finish before Traefik stops.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.  
**Note:** in this time frame no new requests are accepted.

- `ProvidersThrottleDuration`: Backends throttle duration: minimum duration in seconds between 2 events from providers before applying a new configuration.
It avoids unnecessary reloads if multiples events are sent in a short amount of time.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

- `MaxIdleConnsPerHost`: Controls the maximum idle (keep-alive) connections to keep per-host.  
If zero, `DefaultMaxIdleConnsPerHost` from the Go standard library net/http module is used.
If you encounter 'too many open files' errors, you can either increase this value or change the `ulimit`.

- `InsecureSkipVerify` : If set to true invalid SSL certificates are accepted for backends.  
**Note:** This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.

- `RootCAs`: Register Certificates in the RootCA. This certificates will be use for backends calls.  
**Note** You can use file path or cert content directly

- `defaultEntryPoints`: Entrypoints to be used by frontends that do not specify any entrypoint.  
Each frontend can specify its own entrypoints.


## Constraints

In a micro-service architecture, with a central service discovery, setting constraints limits Træfik scope to a smaller number of routes.

Træfik filters services according to service attributes/tags set in your configuration backends.

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

```toml
# Backend-specific constraint
[consulCatalog]
# ...
constraints = ["tag==api"]

# Backend-specific constraint
[marathon]
# ...
constraints = ["tag==api", "tag!=v*-beta"]
```


## Logs Definition

### Traefik logs

```toml
# Traefik logs file
# If not defined, logs to stdout
#
# DEPRECATED - see [traefikLog] lower down
# In case both traefikLogsFile and traefikLog.filePath are specified, the latter will take precedence.
# Optional
#
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

## Traefik Logs

By default the Traefik log is written to stdout in text format.

To write the logs into a logfile specify the `filePath`.
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
    This does not work on Windows due to the lack of USR signals.


## Custom Error pages

Custom error pages can be returned, in lieu of the default, according to frontend-configured ranges of HTTP Status codes.

In the example below, if a 503 status is returned from the frontend "website", the custom error page at http://2.3.4.5/503.html is returned with the actual status code set in the HTTP header.

!!! note
    The `503.html` page itself is not hosted on Traefik, but some other infrastructure.

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

Custom error pages are easiest to implement using the file provider.
For dynamic providers, the corresponding template file needs to be customized accordingly and referenced in the Traefik configuration.


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

# Set the default health check interval.
#
# Optional
# Default: "30s"
#
# interval = "30s"
```

- `interval` set the default health check interval.  
Will only be effective if health check paths are defined.  
Given provider-specific support, the value may be overridden on a per-backend basis.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).  
If no units are provided, the value is parsed assuming seconds.

## Life Cycle

Controls the behavior of Traefik during the shutdown phase.

```toml
[lifeCycle]

# Duration to keep accepting requests prior to initiating the graceful
# termination period (as defined by the `graceTimeOut` option). This
# option is meant to give downstream load-balancers sufficient time to
# take Traefik out of rotation.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
# If no units are provided, the value is parsed assuming seconds.
# The zero duration disables the request accepting grace period, i.e.,
# Traefik will immediately proceed to the grace period.
#
# Optional
# Default: 0
#
# requestAcceptGraceTimeout = "10s"

# Duration to give active requests a chance to finish before Traefik stops.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
# If no units are provided, the value is parsed assuming seconds.
# Note: in this time frame no new requests are accepted.
#
# Optional
# Default: "10s"
#
# graceTimeOut = "10s"
```

## Timeouts

### Responding Timeouts

`respondingTimeouts` are timeouts for incoming requests to the Traefik instance.

```toml
[respondingTimeouts]

# readTimeout is the maximum duration for reading the entire request, including the body.
#
# Optional
# Default: "0s"
#
# readTimeout = "5s"

# writeTimeout is the maximum duration before timing out writes of the response.
#
# Optional
# Default: "0s"
#
# writeTimeout = "5s"

# idleTimeout is the maximum duration an idle (keep-alive) connection will remain idle before closing itself.
#
# Optional
# Default: "180s"
#
# idleTimeout = "360s"
```

- `readTimeout` is the maximum duration for reading the entire request, including the body.  
If zero, no timeout exists.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

- `writeTimeout` is the maximum duration before timing out writes of the response.  
It covers the time from the end of the request header read to the end of the response write.
If zero, no timeout exists.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

- `idleTimeout` is the maximum duration an idle (keep-alive) connection will remain idle before closing itself.  
If zero, no timeout exists.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

### Forwarding Timeouts

`forwardingTimeouts` are timeouts for requests forwarded to the backend servers.

```toml
[forwardingTimeouts]

# dialTimeout is the amount of time to wait until a connection to a backend server can be established.
#
# Optional
# Default: "30s"
#
# dialTimeout = "30s"

# responseHeaderTimeout is the amount of time to wait for a server's response headers after fully writing the request (including its body, if any).
#
# Optional
# Default: "0s"
#
# responseHeaderTimeout = "0s"
```

- `dialTimeout` is the amount of time to wait until a connection to a backend server can be established.  
If zero, no timeout exists.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

- `responseHeaderTimeout` is the amount of time to wait for a server's response headers after fully writing the request (including its body, if any).  
If zero, no timeout exists.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.


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


## Plugins

!!! warning
    For advanced users only.
    
Plugin support is currently provided via the use of [HashiCorp's Go Plugin System over g/RPC](https://github.com/hashicorp/go-plugin)

```toml
[[plugins]]

# path is the full system path to the plugin binary that implements the gRPC plugin interface
path = "/full/path/to/plugin/binary"

# type is the kind of plugin. Valid options are "grpc", "netrpc" and "go", where "go" is the natively compiled binary (not currently supported just yet)
type = "grpc"

# order is the place when the plugin should be executed. Valid options are "before", "after" and "around".  "go" plugin types always use "around" regardless of this setting 
order = "before"
```

### gRPC Remote Plugin

gRPC version of the plugin use the following schema definition to communicate with the remote process:

```proto
syntax = "proto3";
package proto;

message Request {
    // RequestUuid specifies the unique GUID for the given request.
    // It is useful for validating identity of a request when using
    // "around" plugin order to tie together "before" and "after"
    // parts of the remote plugin callback
    string request_uuid = 1;
    // Request contains the values of the HTTP Request made to the server
    HttpRequest request = 2;
    //map<string, ValueList> response_headers = 3;
}

message Response {
    // Request contains the modified HTTP Request and all the values
    // will be synced back into original "http.Request" struct
    HttpRequest request = 1;
    // Response specifies values like response status code, response headers
    // and content payload
    HttpResponse response = 2;
    // StopChain indicates that no further middleware handlers shall be called
    bool stopChain = 3;
    // RenderContent indicates that the value of "HttpResponse.Body" shoud be
    // rendered to the "http.ResponseWriter", which will also stop the chain,
    // similar to "StopChain" flag
    bool renderContent = 4;
    // Redirect indicates that the request should be forwarded to the URL set by
    // "HttpRequest.Url" field, which will also stop the chain, similar to
    // "StopChain" flag
    bool redirect = 5;
}

message HttpRequest {
    // Method specifies the HTTP method (GET, POST, PUT, etc.).
    // For client requests an empty string means GET.
    string method = 1;
    // URL specifies either the URI being requested (for server
    // requests) or the URL to access (for client requests).
    //
    // For server requests the URL is parsed from the URI
    // supplied on the Request-Line as stored in RequestURI.  For
    // most requests, fields other than Path and RawQuery will be
    // empty. (See RFC 2616, Section 5.1.2)
    //
    // For client requests, the URL's Host specifies the server to
    // connect to, while the Request's Host field optionally
    // specifies the Host header value to send in the HTTP
    // request.
    string url = 2;
    // The protocol version for incoming server requests.
    //
    // For client requests these fields are ignored. The HTTP
    // client code always uses either HTTP/1.1 or HTTP/2.
    // See the docs on Transport for details.
    string proto = 3;
    int32 protoMajor = 4;
    int32 protoMinor = 5;
    // Header contains the request header fields either received
    // by the server or to be sent by the client.
    //
    // If a server received a request with header lines,
    //
    //	Host: example.com
    //	accept-encoding: gzip, deflate
    //	Accept-Language: en-us
    //	fOO: Bar
    //	foo: two
    //
    // then
    //
    //	Header = map[string]*proto.ValueList{
    //		"Accept-Encoding": &proto.ValueList{Value: {"gzip, deflate"}},
    //		"Accept-Language": &proto.ValueList{Value: {"en-us"}},
    //		"Foo": &proto.ValueList{Value: {"Bar", "two"}},
    //	}
    //
    // For incoming requests, the Host header is promoted to the
    // Request.Host field and removed from the Header map.
    //
    // HTTP defines that header names are case-insensitive. The
    // request parser implements this by using CanonicalHeaderKey,
    // making the first character and any characters following a
    // hyphen uppercase and the rest lowercase.
    //
    // For client requests, certain headers such as Content-Length
    // and Connection are automatically written when needed and
    // values in Header may be ignored. See the documentation
    // for the Request.Write method.
    map<string, ValueList> header = 6;
    // Body is the request's body.
    //
    // For client requests a nil body means the request has no
    // body, such as a GET request.
    bytes body = 7;
    // ContentLength records the length of the associated content.
    // The value -1 indicates that the length is unknown.
    // Values >= 0 indicate that the given number of bytes may
    // be read from Body.
    // For client requests, a value of 0 with a non-nil Body is
    // also treated as unknown.
    int64 contentLength = 8;
    // TransferEncoding lists the transfer encodings from outermost to
    // innermost. An empty list denotes the "identity" encoding.
    // TransferEncoding can usually be ignored; chunked encoding is
    // automatically added and removed as necessary when sending and
    // receiving requests.
    repeated string transferEncoding = 9;
    // Close indicates whether to close the connection after
    // replying to this request (for servers) or after sending this
    // request and reading its response (for clients).
    //
    // For server requests, the HTTP server handles this automatically
    // and this field is not needed by Handlers.
    //
    // For client requests, setting this field prevents re-use of
    // TCP connections between requests to the same hosts, as if
    // Transport.DisableKeepAlives were set.
    bool close = 10;
    // For server requests Host specifies the host on which the
    // URL is sought. Per RFC 2616, this is either the value of
    // the "Host" header or the host name given in the URL itself.
    // It may be of the form "host:port". For international domain
    // names, Host may be in Punycode or Unicode form. Use
    // golang.org/x/net/idna to convert it to either format if
    // needed.
    //
    // For client requests Host optionally overrides the Host
    // header to send. If empty, the Request.Write method uses
    // the value of URL.Host. Host may contain an international
    // domain name.
    string host = 11;
    // Form contains the parsed form data, including both the URL
    // field's query parameters and the POST or PUT form data.
    // This field is only available after ParseForm is called.
    // The HTTP client ignores Form and uses Body instead.
    map<string, ValueList> formValues = 12;
    // PostForm contains the parsed form data from POST, PATCH,
    // or PUT body parameters.
    //
    // This field is only available after ParseForm is called.
    // The HTTP client ignores PostForm and uses Body instead.
    map<string, ValueList> postFormValues = 13;
    //reserved for multipart.Form = 14

    // Trailer specifies additional headers that are sent after the request
    // body.
    //
    // For server requests the Trailer map initially contains only the
    // trailer keys, with nil values. (The client declares which trailers it
    // will later send.)  While the handler is reading from Body, it must
    // not reference Trailer. After reading from Body returns EOF, Trailer
    // can be read again and will contain non-nil values, if they were sent
    // by the client.
    //
    // For client requests Trailer must be initialized to a map containing
    // the trailer keys to later send. The values may be nil or their final
    // values. The ContentLength must be 0 or -1, to send a chunked request.
    // After the HTTP request is sent the map values can be updated while
    // the request body is read. Once the body returns EOF, the caller must
    // not mutate Trailer.
    //
    // Few HTTP clients, servers, or proxies support HTTP trailers.
    map<string, ValueList> trailer = 15;
    // RemoteAddr allows HTTP servers and other software to record
    // the network address that sent the request, usually for
    // logging. This field is not filled in by ReadRequest and
    // has no defined format. The HTTP server in this package
    // sets RemoteAddr to an "IP:port" address before invoking a
    // handler.
    // This field is ignored by the HTTP client.
    string remoteAddr = 16;
    // RequestURI is the unmodified Request-URI of the
    // Request-Line (RFC 2616, Section 5.1) as sent by the client
    // to a server. Usually the URL field should be used instead.
    // It is an error to set this field in an HTTP client request.
    string requestUri = 17;
}

message HttpResponse {
    int32 status_code = 1;
    map<string, ValueList> header = 2;
    bytes body = 3;
}

message ValueList {
    repeated string value = 1;
}

service Middleware {
    rpc ServeHTTP(Request) returns (Response);
}
```

### Examples

It is best to use the provided [golang library](https://github.com/hashicorp/go-plugin) to implement the plugin, using the following [examples](https://github.com/hashicorp/go-plugin/tree/master/examples/grpc).
Plugins written in Go will require less coding to get them configured and connected, however it is possible to use any other language that supports Google Protocol Buffers gRPC.  
The following [example](https://github.com/hashicorp/go-plugin/tree/master/examples/grpc/plugin-python) demonstrates how it can be done using Python 

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
