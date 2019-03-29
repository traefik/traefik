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
# This will install HTTP handlers to expose Go expvars under /debug/vars and
# pprof profiling data under /debug/pprof/.
# The log level will be set to DEBUG unless `logLevel` is specified.
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

# Tells traefik whether it should keep the trailing slashes in the paths (e.g. /paths/) or redirect to the no trailing slash paths instead (/paths).
#
# Optional
# Default: false
#
# keepTrailingSlash = false

# Providers throttle duration.
#
# Optional
# Default: "2s"
#
# providersThrottleDuration = "2s"

# Controls the maximum idle (keep-alive) connections to keep per-host.
#
# Optional
# Default: 200
#
# maxIdleConnsPerHost = 200

# If set to true invalid SSL certificates are accepted for backends.
# This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.
#
# Optional
# Default: false
#
# insecureSkipVerify = true

# Register Certificates in the rootCA.
#
# Optional
# Default: []
#
# rootCAs = [ "/mycert.cert" ]

# Entrypoints to be used by frontends that do not specify any entrypoint.
# Each frontend can specify its own entrypoints.
#
# Optional
# Default: ["http"]
#
# defaultEntryPoints = ["http", "https"]

# Allow the use of 0 as server weight.
# - false: a weight 0 means internally a weight of 1.
# - true: a weight 0 means internally a weight of 0 (a server with a weight of 0 is removed from the available servers).
#
# Optional
# Default: false
#
# AllowMinWeightZero = true
```

- `graceTimeOut`: Duration to give active requests a chance to finish before Traefik stops.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.  
**Note:** in this time frame no new requests are accepted.

- `providersThrottleDuration`: Providers throttle duration: minimum duration in seconds between 2 events from providers before applying a new configuration.
It avoids unnecessary reloads if multiples events are sent in a short amount of time.  
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

- `maxIdleConnsPerHost`: Controls the maximum idle (keep-alive) connections to keep per-host.  
If zero, `DefaultMaxIdleConnsPerHost` from the Go standard library net/http module is used.
If you encounter 'too many open files' errors, you can either increase this value or change the `ulimit`.

- `insecureSkipVerify` : If set to true invalid SSL certificates are accepted for backends.  
**Note:** This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.

- `rootCAs`: Register Certificates in the RootCA. This certificates will be use for backends calls.  
**Note** You can use file path or cert content directly

- `defaultEntryPoints`: Entrypoints to be used by frontends that do not specify any entrypoint.  
Each frontend can specify its own entrypoints.

- `keepTrailingSlash`: Tells Træfik whether it should keep the trailing slashes that might be present in the paths of incoming requests (true), or if it should redirect to the slashless version of the URL (default behavior: false) 

!!! note 
    Beware that the value of `keepTrailingSlash` can have a significant impact on the way your frontend rules are interpreted.
    The table below tries to sum up several behaviors depending on requests/configurations. 
    The current default behavior is deprecated and kept for compatibility reasons. 
    As a consequence, we encourage you to set `keepTrailingSlash` to true.
    
    | Incoming request     | keepTrailingSlash | Path:{value} | Behavior                              
    |----------------------|-------------------|--------------|----------------------------|
    | http://foo.com/path/ | false             | Path:/path/  | Proceeds with the request  |
    | http://foo.com/path/ | false             | Path:/path   | 301 to http://foo.com/path |           
    | http://foo.com/path  | false             | Path:/path/  | Proceeds with the request  |
    | http://foo.com/path  | false             | Path:/path   | Proceeds with the request  |
    | http://foo.com/path/ | true              | Path:/path/  | Proceeds with the request  |
    | http://foo.com/path/ | true              | Path:/path   | 404                        |
    | http://foo.com/path  | true              | Path:/path/  | 404                        |
    | http://foo.com/path  | true              | Path:/path   | Proceeds with the request  |


## Constraints

In a micro-service architecture, with a central service discovery, setting constraints limits Traefik scope to a smaller number of routes.

Traefik filters services according to service attributes/tags set in your providers.

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

### provider-specific

Supported Providers:

- Docker
- Consul K/V
- BoltDB
- Zookeeper
- ECS
- Etcd
- Consul Catalog
- Rancher
- Marathon
- Kubernetes (using a provider-specific mechanism based on label selectors)

```toml
# Provider-specific constraint
[consulCatalog]
# ...
constraints = ["tag==api"]

# Provider-specific constraint
[marathon]
# ...
constraints = ["tag==api", "tag!=v*-beta"]
```


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


## Rate limiting

Rate limiting can be configured per frontend.  
Multiple sets of rates can be added to each frontend, but the time periods must be unique.

```toml
[frontends]
    [frontends.frontend1]
      # ...
      [frontends.frontend1.ratelimit]
        extractorfunc = "client.ip"
          [frontends.frontend1.ratelimit.rateset.rateset1]
            period = "10s"
            average = 100
            burst = 200
          [frontends.frontend1.ratelimit.rateset.rateset2]
            period = "3s"
            average = 5
            burst = 10
```

In the above example, frontend1 is configured to limit requests by the client's ip address.  
An average of 100 requests every 10 seconds is allowed and an average of 5 requests every 3 seconds.  
These can "burst" up to 200 and 10 in each period respectively. 

Valid values for `extractorfunc` are:
  * `client.ip`
  * `request.host`
  * `request.header.<header name>`

## Buffering

In some cases request/buffering can be enabled for a specific backend.
By enabling this, Traefik will read the entire request into memory (possibly buffering large requests into disk) and will reject requests that are over a specified limit.
This may help services deal with large data (multipart/form-data for example) more efficiently and should minimise time spent when sending data to a backend server.

For more information please check [oxy/buffer](http://godoc.org/github.com/vulcand/oxy/buffer) documentation.

Example configuration:

```toml
[backends]
  [backends.backend1]
    [backends.backend1.buffering]
      maxRequestBodyBytes = 10485760  
      memRequestBodyBytes = 2097152  
      maxResponseBodyBytes = 10485760
      memResponseBodyBytes = 2097152
      retryExpression = "IsNetworkError() && Attempts() <= 2"
```

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

Use [respondingTimeouts](/configuration/commons/#responding-timeouts) instead of `idleTimeout`.
In the case both settings are configured, the deprecated option will be overwritten.

`idleTimeout` is the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
This is set to enforce closing of stale client connections.

Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
If no units are provided, the value is parsed assuming seconds.

```toml
# idleTimeout
#
# DEPRECATED - see [respondingTimeouts] section.
#
# Optional
# Default: "180s"
#
idleTimeout = "360s"
```

## Host Resolver

`hostResolver` are used for request host matching process.

```toml
[hostResolver]

# cnameFlattening is a trigger to flatten request host, assuming it is a CNAME record
#
# Optional
# Default : false
#
cnameFlattening = true

# resolvConf is dns resolving configuration file, the default is /etc/resolv.conf
#
# Optional
# Default : "/etc/resolv.conf"
#
# resolvConf = "/etc/resolv.conf"

# resolvDepth is the maximum CNAME recursive lookup
#
# Optional
# Default : 5
#
# resolvDepth = 5
```

- To allow serving secure https request and generate the SSL using ACME while `cnameFlattening` is active. 
The `acme` configuration for `HTTP-01` challenge and `onDemand` is mandatory. 
Refer to [ACME configuration](/configuration/acme) for more information.

## Override Default Configuration Template

!!! warning
    For advanced users only.

Supported by all providers except: File Provider, Web Provider and DynamoDB Provider.

```toml
[provider_name]

# Override default provider configuration template. For advanced users :)
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

## Pass TLS Client Cert

```toml
# Pass the escaped client cert infos selected below in a `X-Forwarded-Ssl-Client-Cert-Infos` header.
[frontends.frontend1.passTLSClientCert]
        pem = true
        [frontends.frontend1.passTLSClientCert.infos]
            notBefore = true
            notAfter = true
            [frontends.frontend1.passTLSClientCert.infos.subject]
                country = true
                domainComponent = true
                province = true
                locality = true
                organization = true
                commonName = true
                serialNumber = true
            [frontends.frontend1.passTLSClientCert.infos.issuer]
                country = true
                domainComponent = true
                province = true
                locality = true
                organization = true
                commonName = true
                serialNumber = true
```

Pass TLS Client Cert `pem` defines if the escaped pem is added to a `X-Forwarded-Ssl-Client-Cert` header.  
Pass TLS Client Cert `infos` defines how the certificate data are added to a `X-Forwarded-Ssl-Client-Cert-Infos` header.  

The following example shows an unescaped result that uses all the available fields:
If there are more than one certificate, they are separated by a `;`

```
Subject="DC=org,DC=cheese,C=FR,C=US,ST=Cheese org state,ST=Cheese com state,L=TOULOUSE,L=LYON,O=Cheese,O=Cheese 2,CN=*.cheese.com",Issuer="DC=org,DC=cheese,C=FR,C=US,ST=Signing State,ST=Signing State 2,L=TOULOUSE,L=LYON,O=Cheese,O=Cheese 2,CN=Simple Signing CA 2",NB=1544094616,NA=1607166616,SAN=*.cheese.org,*.cheese.net,*.cheese.com,test@cheese.org,test@cheese.net,10.0.1.0,10.0.1.2
```
