# Consul Catalog backend

Træfik can be configured to use service discovery catalog of Consul as a backend configuration.

```toml
################################################################
# Consul Catalog configuration backend
################################################################

# Enable Consul Catalog configuration backend.
[consulCatalog]

# Consul server endpoint.
#
# Required
# Default: "127.0.0.1:8500"
#
endpoint = "127.0.0.1:8500"

# Expose Consul catalog services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Default domain used.
#
# Optional
#
domain = "consul.localhost"

# Prefix for Consul catalog tags.
#
# Optional
# Default: "traefik"
#
prefix = "traefik"

# Default frontEnd Rule for Consul services.
#
# The format is a Go Template with:
# - ".ServiceName", ".Domain" and ".Attributes" available
# - "getTag(name, tags, defaultValue)", "hasTag(name, tags)" and "getAttribute(name, tags, defaultValue)" functions are available
# - "getAttribute(...)" function uses prefixed tag names based on "prefix" value
#
# Optional
# Default: "Host:{{.ServiceName}}.{{.Domain}}"
#
#frontEndRule = "Host:{{.ServiceName}}.{{.Domain}}"

# Enable Consul catalog TLS connection.
#
# Optional
#
#    [consulCatalog.tls]
#    ca = "/etc/ssl/ca.crt"
#    cert = "/etc/ssl/consul.crt"
#    key = "/etc/ssl/consul.key"
#    insecureSkipVerify = true

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "consulcatalog.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = 2
```

This backend will create routes matching on hostname based on the service name used in Consul.

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

## Tags

Additional settings can be defined using Consul Catalog tags.

!!! note
    The default prefix is `traefik`.

| Label                                                       | Description                                                                                                                                                                                                            |
|-------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `<prefix>.enable=false`                                     | Disable this container in Træfik.                                                                                                                                                                                      |
| `<prefix>.port=80`                                          | Register this port. Useful when the container exposes multiples ports.                                                                                                                                                 |
| `<prefix>.protocol=https`                                   | Override the default `http` protocol.                                                                                                                                                                                  |
| `<prefix>.weight=10`                                        | Assign this weight to the container.                                                                                                                                                                                   |
| `traefik.backend.buffering.maxRequestBodyBytes=0`           | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.maxResponseBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.memRequestBodyBytes=0`           | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.memResponseBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.retryExpression=EXPR`            | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `<prefix>.backend.circuitbreaker.expression=EXPR`           | Create a [circuit breaker](/basics/#backends) to be used against the backend. ex: `NetworkErrorRatio() > 0.`                                                                                                           |
| `<prefix>.backend.healthcheck.path=/health`                 | Enable health check for the backend, hitting the container at `path`.                                                                                                                                                  |
| `<prefix>.backend.healthcheck.port=8080`                    | Allow to use a different port for the health check.                                                                                                                                                                    |
| `<prefix>.backend.healthcheck.interval=1s`                  | Define the health check interval.                                                                                                                                                                                      |
| `<prefix>.backend.loadbalancer.method=drr`                  | Override the default `wrr` load balancer algorithm.                                                                                                                                                                    |
| `<prefix>.backend.loadbalancer.stickiness=true`             | Enable backend sticky sessions.                                                                                                                                                                                        |
| `<prefix>.backend.loadbalancer.stickiness.cookieName=NAME`  | Manually set the cookie name for sticky sessions.                                                                                                                                                                      |
| `<prefix>.backend.loadbalancer.sticky=true`                 | Enable backend sticky sessions. (DEPRECATED)                                                                                                                                                                           |
| `<prefix>.backend.maxconn.amount=10`                        | Set a maximum number of connections to the backend.<br>Must be used in conjunction with the below label to take effect.                                                                                                |
| `<prefix>.backend.maxconn.extractorfunc=client.ip`          | Set the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect.                                  |
| `<prefix>.frontend.auth.basic=EXPR`                         | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`                                                                                                                                       |
| `<prefix>.frontend.entryPoints=http,https`                  | Assign this frontend to entry points `http` and `https`.<br>Overrides `defaultEntryPoints`                                                                                                                             |
| `<prefix>.frontend.errors.<name>.backend=NAME`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `<prefix>.frontend.errors.<name>.query=PATH`                | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `<prefix>.frontend.errors.<name>.status=RANGE`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `<prefix>.frontend.passHostHeader=true`                     | Forward client `Host` header to the backend.                                                                                                                                                                           |
| `<prefix>.frontend.passTLSCert=true`                        | Forward TLS Client certificates to the backend.                                                                                                                                                                        |
| `<prefix>.frontend.priority=10`                             | Override default frontend priority.                                                                                                                                                                                    |
| `<prefix>.frontend.rateLimit.extractorFunc=EXP`             | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `<prefix>.frontend.rateLimit.rateSet.<name>.period=6`       | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `<prefix>.frontend.rateLimit.rateSet.<name>.average=6`      | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `<prefix>.frontend.rateLimit.rateSet.<name>.burst=6`        | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `<prefix>.frontend.redirect.entryPoint=https`               | Enables Redirect to another entryPoint for that frontend (e.g. HTTPS).                                                                                                                                                 |
| `<prefix>.frontend.redirect.regex=^http://localhost/(.*)`   | Redirect to another URL for that frontend.<br>Must be set with `traefik.frontend.redirect.replacement`.                                                                                                                |
| `<prefix>.frontend.redirect.replacement=http://mydomain/$1` | Redirect to another URL for that frontend.<br>Must be set with `traefik.frontend.redirect.regex`.                                                                                                                      |
| `<prefix>.frontend.redirect.permanent=true`                 | Return 301 instead of 302.                                                                                                                                                                                             |
| `<prefix>.frontend.rule=EXPR`                               | Override the default frontend rule. Default: `Host:{{.ServiceName}}.{{.Domain}}`.                                                                                                                                      |
| `<prefix>.frontend.whiteList.sourceRange=RANGE`             | List of IP-Ranges which are allowed to access.<br>An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access. |
| `<prefix>.frontend.whiteList.useXForwardedFor=true`         | Use `X-Forwarded-For` header as valid source of IP for the white list.                                                                                                                                                 |

### Custom Headers

!!! note
    The default prefix is `traefik`.

| Label                                                  | Description                                                                                                                                                                         |
|--------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `<prefix>.frontend.headers.customRequestHeaders=EXPR ` | Provides the container with custom request headers that will be appended to each request forwarded to the container.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code> |
| `<prefix>.frontend.headers.customResponseHeaders=EXPR` | Appends the headers to each response returned by the container, before forwarding the response to the client.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>        |

### Security Headers

!!! note
    The default prefix is `traefik`.

| Label                                                     | Description                                                                                                                                                                                         |
|-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `<prefix>.frontend.headers.allowedHosts=EXPR`             | Provides a list of allowed hosts that requests will be processed.<br>Format: `Host1,Host2`                                                                                                          |
| `<prefix>.frontend.headers.hostsProxyHeaders=EXPR`        | Provides a list of headers that the proxied hostname may be stored.<br>Format: `HEADER1,HEADER2`                                                                                                    |
| `<prefix>.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `<prefix>.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `<prefix>.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `<prefix>.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `<prefix>.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `<prefix>.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `<prefix>.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |
| `<prefix>.frontend.headers.forceSTSHeader=false`          | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `<prefix>.frontend.headers.frameDeny=false`               | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `<prefix>.frontend.headers.customFrameOptionsValue=VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `<prefix>.frontend.headers.contentTypeNosniff=true`       | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `<prefix>.frontend.headers.browserXSSFilter=true`         | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `<prefix>.frontend.headers.customBrowserXSSValue=VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `<prefix>.frontend.headers.contentSecurityPolicy=VALUE`   | Adds CSP Header with the custom value.                                                                                                                                                              |
| `<prefix>.frontend.headers.publicKey=VALUE`               | Adds pinned HTST public key header.                                                                                                                                                                 |
| `<prefix>.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `<prefix>.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |

### Examples

If you want that Træfik uses Consul tags correctly you need to defined them like that:

```js
traefik.enable=true
traefik.tags=api
traefik.tags=external
```

If the prefix defined in Træfik configuration is `bla`, tags need to be defined like that:

```js
bla.enable=true
bla.tags=api
bla.tags=external
```
