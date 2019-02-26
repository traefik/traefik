# Global Configuration

## Main Section

```toml
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
```

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

- `keepTrailingSlash`: Tells Traefik whether it should keep the trailing slashes that might be present in the paths of incoming requests (true), or if it should redirect to the slashless version of the URL (default behavior: false) 

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

Supported by all providers except: File Provider, Rest Provider and DynamoDB Provider.

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

### Using ping for an external Load-balancer rotation health check

If you are running Traefik behind an external Load-balancer, and want to configure rotation health check on the Load-balancer to take a Traefik instance out of rotation gracefully, you can configure [lifecycle.requestAcceptGraceTimeout](/configuration/commons.md#life-cycle) and the ping endpoint will return `503` response on traefik server termination, so that the Load-balancer can take the terminating traefik instance out of rotation, before it stops responding.
