# Mesos Generic Provider

Træfik can be configured to use Mesos as a provider.

```toml
################################################################
# Mesos Provider
################################################################

# Enable Mesos Provider.
[mesos]

# Mesos server endpoint.
# You can also specify multiple endpoint for Mesos:
# endpoint = "192.168.35.40:5050,192.168.35.41:5050,192.168.35.42:5050"
# endpoint = "zk://192.168.35.20:2181,192.168.35.21:2181,192.168.35.22:2181/mesos"
#
# Required
# Default: "http://127.0.0.1:5050"
#
endpoint = "http://127.0.0.1:8080"

# Enable watch Mesos changes.
#
# Optional
# Default: true
#
watch = true

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on an application.
#
# Required
#
domain = "mesos.localhost"

# Expose Mesos apps by default in Traefik.
#
# Optional
# Default: true
#
# exposedByDefault = false

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "mesos.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = 2

# TLS client configuration. https://golang.org/pkg/crypto/tls/#Config
#
# Optional
#
# [mesos.TLS]
# insecureSkipVerify = true

# Zookeeper timeout (in seconds).
#
# Optional
# Default: 30
#
# zkDetectionTimeout = 30

# Polling interval (in seconds).
#
# Optional
# Default: 30
#
# refreshSeconds = 30

# IP sources (e.g. host, docker, mesos, netinfo).
#
# Optional
#
# ipSources = "host"

# HTTP Timeout (in seconds).
#
# Optional
# Default: 30
#
# stateTimeoutSecond = "30"

# Convert groups to subdomains.
# Default behavior: /foo/bar/myapp => foo-bar-myapp.{defaultDomain}
# with groupsAsSubDomains enabled: /foo/bar/myapp => myapp.bar.foo.{defaultDomain}
#
# Optional
# Default: false
#
# groupsAsSubDomains = true

```

## Labels: overriding default behavior

The following labels can be defined on Mesos tasks. They adjust the behavior for the entire application.

| Label                                                      | Description                                                                                                                                                                                                                   |
|------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.domain`                                           | Sets the default domain for the frontend rules.                                                                                                                                                                               |
| `traefik.enable=false`                                     | Disables this container in Træfik.                                                                                                                                                                                            |
| `traefik.port=80`                                          | Registers this port. Useful when the application exposes multiple ports.                                                                                                                                                      |
| `traefik.portName=web`                                     | Registers port by name in the application's ports array. Useful when the application exposes multiple ports.                                                                                                                  |
| `traefik.portIndex=1`                                      | Registers port by index in the application's ports array. Useful when the application exposes multiple ports.                                                                                                                 |
| `traefik.protocol=https`                                   | Overrides the default `http` protocol                                                                                                                                                                                         |
| `traefik.weight=10`                                        | Assigns this weight to the container                                                                                                                                                                                          |
| `traefik.backend=foo`                                      | Gives the name `foo` to the generated backend for this container.                                                                                                                                                             |
| `traefik.backend.buffering.maxRequestBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.maxResponseBodyBytes=0`         | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.memRequestBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.memResponseBodyBytes=0`         | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.retryExpression=EXPR`           | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.circuitbreaker.expression=EXPR`           | Creates a [circuit breaker](/basics/#backends) to be used against the backend                                                                                                                                                 |
| `traefik.backend.healthcheck.path=/health`                 | Enables health check for the backend, hitting the container at `path`.                                                                                                                                                        |
| `traefik.backend.healthcheck.interval=1s`                  | Defines the health check interval. (Default: 30s)                                                                                                                                                                             |
| `traefik.backend.healthcheck.scheme=http`                  | Overrides the server URL scheme.                                                                                                                                                                                              |
| `traefik.backend.healthcheck.port=8080`                    | Sets a different port for the health check.                                                                                                                                                                                   |
| `traefik.backend.healthcheck.hostname=foobar.com`          | Defines the health check hostname.                                                                                                                                                                                            |
| `traefik.backend.healthcheck.headers=EXPR`                 | Defines the health check request headers <br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                                                                                                     |
| `traefik.backend.loadbalancer.method=drr`                  | Overrides the default `wrr` load balancer algorithm                                                                                                                                                                           |
| `traefik.backend.loadbalancer.stickiness=true`             | Enables backend sticky sessions                                                                                                                                                                                               |
| `traefik.backend.loadbalancer.stickiness.cookieName=NAME`  | Sets the cookie manually name for sticky sessions                                                                                                                                                                             |
| `traefik.backend.maxconn.amount=10`                        | Sets a maximum number of connections to the backend.<br>Must be used in conjunction with the below label to take effect.                                                                                                      |
| `traefik.backend.maxconn.extractorfunc=client.ip`          | Sets the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect.                                        |
| `traefik.frontend.auth.basic=EXPR`                         | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash` (DEPRECATED).                                                                                                                                 |
| `traefik.frontend.auth.basic.users=EXPR`                   | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash`.                                                                                                                                              |
| `traefik.frontend.auth.basic.usersfile=/path/.htpasswd`    | Sets basic authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                         |
| `traefik.frontend.auth.digest.users=EXPR`                  | Sets digest authentication to this frontend in CSV format: `User:Realm:Hash,User:Realm:Hash`.                                                                                                                                 |
| `traefik.frontend.auth.digest.usersfile=/path/.htdigest`   | Sets digest authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                        |
| `traefik.frontend.auth.forward.address=https://example.com`| Sets the URL of the authentication server.                                                                                                                                                                                    |
| `traefik.frontend.auth.forward.tls.ca=/path/ca.pem`        | Sets the Certificate Authority (CA) for the TLS connection with the authentication server.                                                                                                                                    |
| `traefik.frontend.auth.forward.tls.caOptional=true`        | Checks the certificates if present but do not force to be signed by a specified Certificate Authority (CA).                                                                                                                   |
| `traefik.frontend.auth.forward.tls.cert=/path/server.pem`  | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `traefik.frontend.auth.forward.tls.insecureSkipVerify=true`| If set to true invalid SSL certificates are accepted.                                                                                                                                                                         |
| `traefik.frontend.auth.forward.tls.key=/path/server.key`   | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `traefik.frontend.auth.forward.trustForwardHeader=true`    | Trusts X-Forwarded-* headers.                                                                                                                                                                                                 |
| `traefik.frontend.auth.headerField=X-WebAuth-User`         | Sets the header used to pass the authenticated user to the application.                                                                                                                                                       |
| `traefik.frontend.entryPoints=http,https`                  | Assigns this frontend to entry points `http` and `https`.<br>Overrides `defaultEntryPoints`                                                                                                                                   |
| `traefik.frontend.errors.<name>.backend=NAME`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `traefik.frontend.errors.<name>.query=PATH`                | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `traefik.frontend.errors.<name>.status=RANGE`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `traefik.frontend.passHostHeader=true`                     | Forwards client `Host` header to the backend.                                                                                                                                                                                 |
| `traefik.frontend.passTLSCert=true`                        | Forwards TLS Client certificates to the backend.                                                                                                                                                                              |
| `traefik.frontend.priority=10`                             | Overrides default frontend priority                                                                                                                                                                                           |
| `traefik.frontend.rateLimit.extractorFunc=EXP`             | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `traefik.frontend.rateLimit.rateSet.<name>.period=6`       | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `traefik.frontend.rateLimit.rateSet.<name>.average=6`      | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `traefik.frontend.rateLimit.rateSet.<name>.burst=6`        | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `traefik.frontend.redirect.entryPoint=https`               | Enables Redirect to another entryPoint to this frontend (e.g. HTTPS)                                                                                                                                                          |
| `traefik.frontend.redirect.regex=^http://localhost/(.*)`   | Redirects to another URL to this frontend.<br>Must be set with `traefik.frontend.redirect.replacement`.                                                                                                                       |
| `traefik.frontend.redirect.replacement=http://mydomain/$1` | Redirects to another URL to this frontend.<br>Must be set with `traefik.frontend.redirect.regex`.                                                                                                                             |
| `traefik.frontend.redirect.permanent=true`                 | Returns 301 instead of 302.                                                                                                                                                                                                   |
| `traefik.frontend.rule=EXPR`                               | Overrides the default frontend rule. Default: `Host:{discovery_name}.{domain}`.                                                                                                                                               |
| `traefik.frontend.whiteList.sourceRange=RANGE`             | Sets a list of IP-Ranges which are allowed to access.<br>An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access. |
| `traefik.frontend.whiteList.useXForwardedFor=true`         | Uses `X-Forwarded-For` header as valid source of IP for the white list.                                                                                                                                                       |

### Custom Headers

| Label                                                 | Description                                                                                                                                                                         |
|-------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.frontend.headers.customRequestHeaders=EXPR ` | Provides the container with custom request headers that will be appended to each request forwarded to the container.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code> |
| `traefik.frontend.headers.customResponseHeaders=EXPR` | Appends the headers to each response returned by the container, before forwarding the response to the client.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>        |

### Security Headers

| Label                                                    | Description                                                                                                                                                                                         |
|----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.frontend.headers.allowedHosts=EXPR`             | Provides a list of allowed hosts that requests will be processed.<br>Format: `Host1,Host2`                                                                                                          |
| `traefik.frontend.headers.browserXSSFilter=true`         | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `traefik.frontend.headers.contentSecurityPolicy=VALUE`   | Adds CSP Header with the custom value.                                                                                                                                                              |
| `traefik.frontend.headers.contentTypeNosniff=true`       | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `traefik.frontend.headers.customBrowserXSSValue=VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `traefik.frontend.headers.customFrameOptionsValue=VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `traefik.frontend.headers.forceSTSHeader=false`          | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `traefik.frontend.headers.frameDeny=false`               | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `traefik.frontend.headers.hostsProxyHeaders=EXPR `       | Provides a list of headers that the proxied hostname may be stored.<br>Format: `HEADER1,HEADER2`                                                                                                    |
| `traefik.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |
| `traefik.frontend.headers.publicKey=VALUE`               | Adds pinned HTST public key header.                                                                                                                                                                 |
| `traefik.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `traefik.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `traefik.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `traefik.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `traefik.frontend.headers.SSLForceHost=true`             | If `SSLForceHost` is `true` and `SSLHost` is set, requests will be forced to use `SSLHost` even the ones that are already using SSL. Default is false.                                              |
| `traefik.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `traefik.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `traefik.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `traefik.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |

### Applications with Multiple Ports (segment labels)

Segment labels are used to define routes to an application exposing multiple ports.
A segment is a group of labels that apply to a port exposed by an application.
You can define as many segments as ports exposed in an application.

Additionally, if a segment name matches a named port, that port will be used unless `portIndex`, `portName`, or `port` labels are specified for that segment.

Segment labels override the default behavior.

| Label                                                                     | Description                                                 |
|---------------------------------------------------------------------------|-------------------------------------------------------------|
| `traefik.<segment_name>.backend=BACKEND`                                  | Same as `traefik.backend`                                   |
| `traefik.<segment_name>.domain=DOMAIN`                                    | Same as `traefik.domain`                                    |
| `traefik.<segment_name>.portIndex=1`                                      | Same as `traefik.portIndex`                                 |
| `traefik.<segment_name>.portName=web`                                     | Same as `traefik.portName`                                  |
| `traefik.<segment_name>.port=PORT`                                        | Same as `traefik.port`                                      |
| `traefik.<segment_name>.protocol=http`                                    | Same as `traefik.protocol`                                  |
| `traefik.<segment_name>.weight=10`                                        | Same as `traefik.weight`                                    |
| `traefik.<segment_name>.frontend.auth.basic=EXPR`                         | Same as `traefik.frontend.auth.basic`                       |
| `traefik.<segment_name>.frontend.entryPoints=https`                       | Same as `traefik.frontend.entryPoints`                      |
| `traefik.<segment_name>.frontend.errors.<name>.backend=NAME`              | Same as `traefik.frontend.errors.<name>.backend`            |
| `traefik.<segment_name>.frontend.errors.<name>.query=PATH`                | Same as `traefik.frontend.errors.<name>.query`              |
| `traefik.<segment_name>.frontend.errors.<name>.status=RANGE`              | Same as `traefik.frontend.errors.<name>.status`             |
| `traefik.<segment_name>.frontend.passHostHeader=true`                     | Same as `traefik.frontend.passHostHeader`                   |
| `traefik.<segment_name>.frontend.passTLSCert=true`                        | Same as `traefik.frontend.passTLSCert`                      |
| `traefik.<segment_name>.frontend.priority=10`                             | Same as `traefik.frontend.priority`                         |
| `traefik.<segment_name>.frontend.rateLimit.extractorFunc=EXP`             | Same as `traefik.frontend.rateLimit.extractorFunc`          |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.period=6`       | Same as `traefik.frontend.rateLimit.rateSet.<name>.period`  |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.average=6`      | Same as `traefik.frontend.rateLimit.rateSet.<name>.average` |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.burst=6`        | Same as `traefik.frontend.rateLimit.rateSet.<name>.burst`   |
| `traefik.<segment_name>.frontend.redirect.entryPoint=https`               | Same as `traefik.frontend.redirect.entryPoint`              |
| `traefik.<segment_name>.frontend.redirect.regex=^http://localhost/(.*)`   | Same as `traefik.frontend.redirect.regex`                   |
| `traefik.<segment_name>.frontend.redirect.replacement=http://mydomain/$1` | Same as `traefik.frontend.redirect.replacement`             |
| `traefik.<segment_name>.frontend.redirect.permanent=true`                 | Same as `traefik.frontend.redirect.permanent`               |
| `traefik.<segment_name>.frontend.rule=EXP`                                | Same as `traefik.frontend.rule`                             |
| `traefik.<segment_name>.frontend.whiteList.sourceRange=RANGE`             | Same as `traefik.frontend.whiteList.sourceRange`            |
| `traefik.<segment_name>.frontend.whiteList.useXForwardedFor=true`         | Same as `traefik.frontend.whiteList.useXForwardedFor`       |

#### Custom Headers

| Label                                                                | Description                                              |
|----------------------------------------------------------------------|----------------------------------------------------------|
| `traefik.<segment_name>.frontend.headers.customRequestHeaders=EXPR ` | Same as `traefik.frontend.headers.customRequestHeaders`  |
| `traefik.<segment_name>.frontend.headers.customResponseHeaders=EXPR` | Same as `traefik.frontend.headers.customResponseHeaders` |

#### Security Headers

| Label                                                                   | Description                                                  |
|-------------------------------------------------------------------------|--------------------------------------------------------------|
| `traefik.<segment_name>.frontend.headers.allowedHosts=EXPR`             | Same as `traefik.frontend.headers.allowedHosts`              |
| `traefik.<segment_name>.frontend.headers.browserXSSFilter=true`         | Same as `traefik.frontend.headers.browserXSSFilter`          |
| `traefik.<segment_name>.frontend.headers.contentSecurityPolicy=VALUE`   | Same as `traefik.frontend.headers.contentSecurityPolicy`     |
| `traefik.<segment_name>.frontend.headers.contentTypeNosniff=true`       | Same as `traefik.frontend.headers.contentTypeNosniff`        |
| `traefik.<segment_name>.frontend.headers.customBrowserXSSValue=VALUE`   | Same as `traefik.frontend.headers.customBrowserXSSValue`     |
| `traefik.<segment_name>.frontend.headers.customFrameOptionsValue=VALUE` | Same as `traefik.frontend.headers.customFrameOptionsValue`   |
| `traefik.<segment_name>.frontend.headers.forceSTSHeader=false`          | Same as `traefik.frontend.headers.forceSTSHeader`            |
| `traefik.<segment_name>.frontend.headers.frameDeny=false`               | Same as `traefik.frontend.headers.frameDeny`                 |
| `traefik.<segment_name>.frontend.headers.hostsProxyHeaders=EXPR`        | Same as `traefik.frontend.headers.hostsProxyHeaders`         |
| `traefik.<segment_name>.frontend.headers.isDevelopment=false`           | Same as `traefik.frontend.headers.isDevelopment`             |
| `traefik.<segment_name>.frontend.headers.publicKey=VALUE`               | Same as `traefik.frontend.headers.publicKey`                 |
| `traefik.<segment_name>.frontend.headers.referrerPolicy=VALUE`          | Same as `traefik.frontend.headers.referrerPolicy`            |
| `traefik.<segment_name>.frontend.headers.SSLRedirect=true`              | Same as `traefik.frontend.headers.SSLRedirect`               |
| `traefik.<segment_name>.frontend.headers.SSLTemporaryRedirect=true`     | Same as `traefik.frontend.headers.SSLTemporaryRedirect`      |
| `traefik.<segment_name>.frontend.headers.SSLHost=HOST`                  | Same as `traefik.frontend.headers.SSLHost`                   |
| `traefik.<segment_name>.frontend.headers.SSLForceHost=true`             | Same as `traefik.frontend.headers.SSLForceHost`              |
| `traefik.<segment_name>.frontend.headers.SSLProxyHeaders=EXPR`          | Same as `traefik.frontend.headers.SSLProxyHeaders=EXPR`      |
| `traefik.<segment_name>.frontend.headers.STSSeconds=315360000`          | Same as `traefik.frontend.headers.STSSeconds=315360000`      |
| `traefik.<segment_name>.frontend.headers.STSIncludeSubdomains=true`     | Same as `traefik.frontend.headers.STSIncludeSubdomains=true` |
| `traefik.<segment_name>.frontend.headers.STSPreload=true`               | Same as `traefik.frontend.headers.STSPreload=true`           |
