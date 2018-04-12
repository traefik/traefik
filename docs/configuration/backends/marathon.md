# Marathon Backend

Træfik can be configured to use Marathon as a backend configuration.

See also [Marathon user guide](/user-guide/marathon).


## Configuration

```toml
################################################################
# Mesos/Marathon configuration backend
################################################################

# Enable Marathon configuration backend.
[marathon]

# Marathon server endpoint.
# You can also specify multiple endpoint for Marathon:
# endpoint = "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
#
# Required
# Default: "http://127.0.0.1:8080"
#
endpoint = "http://127.0.0.1:8080"

# Enable watch Marathon changes.
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
domain = "marathon.localhost"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "marathon.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = "2"

# Expose Marathon apps by default in Traefik.
#
# Optional
# Default: true
#
# exposedByDefault = false

# Convert Marathon groups to subdomains.
# Default behavior: /foo/bar/myapp => foo-bar-myapp.{defaultDomain}
# with groupsAsSubDomains enabled: /foo/bar/myapp => myapp.bar.foo.{defaultDomain}
#
# Optional
# Default: false
#
# groupsAsSubDomains = true

# Enable compatibility with marathon-lb labels.
#
# Optional
# Default: false
#
# marathonLBCompatibility = true

# Enable filtering using Marathon constraints..
# If enabled, Traefik will read Marathon constraints, as defined in https://mesosphere.github.io/marathon/docs/constraints.html
# Each individual constraint will be treated as a verbatim compounded tag.
# i.e. "rack_id:CLUSTER:rack-1", with all constraint groups concatenated together using ":"
#
# Optional
# Default: false
#
# filterMarathonConstraints = true

# Enable Marathon basic authentication.
#
# Optional
#
#    [marathon.basic]
#    httpBasicAuthUser = "foo"
#    httpBasicPassword = "bar"

# TLS client configuration. https://golang.org/pkg/crypto/tls/#Config
#
# Optional
#
#    [marathon.TLS]
#    CA = "/etc/ssl/ca.crt"
#    Cert = "/etc/ssl/marathon.cert"
#    Key = "/etc/ssl/marathon.key"
#    insecureSkipVerify = true

# DCOSToken for DCOS environment.
# This will override the Authorization header.
#
# Optional
#
# dcosToken = "xxxxxx"

# Override DialerTimeout.
# Amount of time to allow the Marathon provider to wait to open a TCP connection
# to a Marathon master.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits).
# If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "60s"
#
# dialerTimeout = "60s"

# Set the TCP Keep Alive interval for the Marathon HTTP Client.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits).
# If no units are provided, the value is parsed assuming seconds.
#
# Optional
# Default: "10s"
#
# keepAlive = "10s"

# By default, a task's IP address (as returned by the Marathon API) is used as
# backend server if an IP-per-task configuration can be found; otherwise, the
# name of the host running the task is used.
# The latter behavior can be enforced by enabling this switch.
#
# Optional
# Default: false
#
# forceTaskHostname = true

# Applications may define readiness checks which are probed by Marathon during
# deployments periodically and the results exposed via the API.
# Enabling the following parameter causes Traefik to filter out tasks
# whose readiness checks have not succeeded.
# Note that the checks are only valid at deployment times.
# See the Marathon guide for details.
#
# Optional
# Default: false
#
# respectReadinessChecks = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

## Labels: overriding default behavior

Marathon labels may be used to dynamically change the routing and forwarding behavior.

They may be specified on one of two levels: Application or service.

### Application Level

The following labels can be defined on Marathon applications. They adjust the behavior for the entire application.

| Label                                                      | Description                                                                                                                                                                                                            |
|------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.enable=false`                                     | Disable this container in Træfik                                                                                                                                                                                       |
| `traefik.port=80`                                          | Register this port. Useful when the container exposes multiples ports.                                                                                                                                                 |
| `traefik.portIndex=1`                                      | Register port by index in the application's ports array. Useful when the application exposes multiple ports.                                                                                                           |
| `traefik.protocol=https`                                   | Override the default `http` protocol                                                                                                                                                                                   |
| `traefik.weight=10`                                        | Assign this weight to the container                                                                                                                                                                                    |
| `traefik.backend=foo`                                      | Give the name `foo` to the generated backend for this container.                                                                                                                                                       |
| `traefik.backend.buffering.maxRequestBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.maxResponseBodyBytes=0`         | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.memRequestBodyBytes=0`          | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.memResponseBodyBytes=0`         | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.buffering.retryExpression=EXPR`           | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                            |
| `traefik.backend.circuitbreaker.expression=EXPR`           | Create a [circuit breaker](/basics/#backends) to be used against the backend                                                                                                                                           |
| `traefik.backend.healthcheck.path=/health`                 | Enable health check for the backend, hitting the container at `path`.                                                                                                                                                  |
| `traefik.backend.healthcheck.port=8080`                    | Allow to use a different port for the health check.                                                                                                                                                                    |
| `traefik.backend.healthcheck.interval=1s`                  | Define the health check interval. (Default: 30s)                                                                                                                                                                       |
| `traefik.backend.healthcheck.hostname=foobar.com`          | Define the health check hostname.                                                                                                                                                                                         |
| `traefik.backend.healthcheck.headers=EXPR`                 | Define the health check request headers <br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>    |
| `traefik.backend.loadbalancer.method=drr`                  | Override the default `wrr` load balancer algorithm                                                                                                                                                                     |
| `traefik.backend.loadbalancer.stickiness=true`             | Enable backend sticky sessions                                                                                                                                                                                         |
| `traefik.backend.loadbalancer.stickiness.cookieName=NAME`  | Manually set the cookie name for sticky sessions                                                                                                                                                                       |
| `traefik.backend.loadbalancer.sticky=true`                 | Enable backend sticky sessions (DEPRECATED)                                                                                                                                                                            |
| `traefik.backend.maxconn.amount=10`                        | Set a maximum number of connections to the backend.<br>Must be used in conjunction with the below label to take effect.                                                                                                |
| `traefik.backend.maxconn.extractorfunc=client.ip`          | Set the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect.                                  |
| `traefik.frontend.auth.basic=EXPR`                         | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`                                                                                                                                       |
| `traefik.frontend.entryPoints=http,https`                  | Assign this frontend to entry points `http` and `https`.<br>Overrides `defaultEntryPoints`                                                                                                                             |
| `traefik.frontend.errors.<name>.backend=NAME`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `traefik.frontend.errors.<name>.query=PATH`                | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `traefik.frontend.errors.<name>.status=RANGE`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                          |
| `traefik.frontend.passHostHeader=true`                     | Forward client `Host` header to the backend.                                                                                                                                                                           |
| `traefik.frontend.passTLSCert=true`                        | Forward TLS Client certificates to the backend.                                                                                                                                                                        |
| `traefik.frontend.priority=10`                             | Override default frontend priority                                                                                                                                                                                     |
| `traefik.frontend.rateLimit.extractorFunc=EXP`             | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `traefik.frontend.rateLimit.rateSet.<name>.period=6`       | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `traefik.frontend.rateLimit.rateSet.<name>.average=6`      | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `traefik.frontend.rateLimit.rateSet.<name>.burst=6`        | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                    |
| `traefik.frontend.redirect.entryPoint=https`               | Enables Redirect to another entryPoint for that frontend (e.g. HTTPS)                                                                                                                                                  |
| `traefik.frontend.redirect.regex=^http://localhost/(.*)`   | Redirect to another URL for that frontend.<br>Must be set with `traefik.frontend.redirect.replacement`.                                                                                                                |
| `traefik.frontend.redirect.replacement=http://mydomain/$1` | Redirect to another URL for that frontend.<br>Must be set with `traefik.frontend.redirect.regex`.                                                                                                                      |
| `traefik.frontend.redirect.permanent=true`                 | Return 301 instead of 302.                                                                                                                                                                                             |
| `traefik.frontend.rule=EXPR`                               | Override the default frontend rule. Default: `Host:{sub_domain}.{domain}`.                                                                                                                                             |
| `traefik.frontend.whiteList.sourceRange=RANGE`             | List of IP-Ranges which are allowed to access.<br>An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access. |
| `traefik.frontend.whiteList.useXForwardedFor=true`         | Use `X-Forwarded-For` header as valid source of IP for the white list.                                                                                                                                                 |

#### Custom Headers

| Label                                                 | Description                                                                                                                                                                         |
|-------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.frontend.headers.customRequestHeaders=EXPR ` | Provides the container with custom request headers that will be appended to each request forwarded to the container.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code> |
| `traefik.frontend.headers.customResponseHeaders=EXPR` | Appends the headers to each response returned by the container, before forwarding the response to the client.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>        |
|

#### Security Headers

| Label                                                    | Description                                                                                                                                                                                         |
|----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.frontend.headers.allowedHosts=EXPR`             | Provides a list of allowed hosts that requests will be processed.<br>Format: `Host1,Host2`                                                                                                          |
| `traefik.frontend.headers.hostsProxyHeaders=EXPR `       | Provides a list of headers that the proxied hostname may be stored.<br>Format: `HEADER1,HEADER2`                                                                                                    |
| `traefik.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `traefik.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `traefik.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `traefik.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `traefik.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `traefik.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `traefik.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |
| `traefik.frontend.headers.forceSTSHeader=false`          | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `traefik.frontend.headers.frameDeny=false`               | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `traefik.frontend.headers.customFrameOptionsValue=VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `traefik.frontend.headers.contentTypeNosniff=true`       | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `traefik.frontend.headers.browserXSSFilter=true`         | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `traefik.frontend.headers.customBrowserXSSValue=VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `traefik.frontend.headers.contentSecurityPolicy=VALUE`   | Adds CSP Header with the custom value.                                                                                                                                                              |
| `traefik.frontend.headers.publicKey=VALUE`               | Adds pinned HTST public key header.                                                                                                                                                                 |
| `traefik.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `traefik.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |

### Applications with Multiple Ports (segment labels)

Segment labels are used to define routes to an application exposing multiple ports.
A segment is a group of labels that apply to a port exposed by an application.
You can define as many segments as ports exposed in an application.

Segment labels override the default behavior.

| Label                                                                     | Description                                                                                          |
|---------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| `traefik.<segment_name>.portIndex=1`                                      | Create a service binding with frontend/backend using this port index. Overrides `traefik.portIndex`. |
| `traefik.<segment_name>.port=PORT`                                        | Overrides `traefik.port`. If several ports need to be exposed, the service labels could be used.     |
| `traefik.<segment_name>.protocol=http`                                    | Overrides `traefik.protocol`.                                                                        |
| `traefik.<segment_name>.weight=10`                                        | Assign this service weight. Overrides `traefik.weight`.                                              |
| `traefik.<segment_name>.frontend.auth.basic=EXPR`                         | Sets a Basic Auth for that frontend                                                                  |
| `traefik.<segment_name>.frontend.backend=BACKEND`                         | Assign this service frontend to `BACKEND`. Default is to assign to the service backend.              |
| `traefik.<segment_name>.frontend.entryPoints=https`                       | Overrides `traefik.frontend.entrypoints`                                                             |
| `traefik.<segment_name>.frontend.errors.<name>.backend=NAME`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                        |
| `traefik.<segment_name>.frontend.errors.<name>.query=PATH`                | See [custom error pages](/configuration/commons/#custom-error-pages) section.                        |
| `traefik.<segment_name>.frontend.errors.<name>.status=RANGE`              | See [custom error pages](/configuration/commons/#custom-error-pages) section.                        |
| `traefik.<segment_name>.frontend.passHostHeader=true`                     | Overrides `traefik.frontend.passHostHeader`.                                                         |
| `traefik.<segment_name>.frontend.passTLSCert=true`                        | Overrides `traefik.frontend.passTLSCert`.                                                            |
| `traefik.<segment_name>.frontend.priority=10`                             | Overrides `traefik.frontend.priority`.                                                               |
| `traefik.<segment_name>.frontend.rateLimit.extractorFunc=EXP`             | See [rate limiting](/configuration/commons/#rate-limiting) section.                                  |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.period=6`       | See [rate limiting](/configuration/commons/#rate-limiting) section.                                  |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.average=6`      | See [rate limiting](/configuration/commons/#rate-limiting) section.                                  |
| `traefik.<segment_name>.frontend.rateLimit.rateSet.<name>.burst=6`        | See [rate limiting](/configuration/commons/#rate-limiting) section.                                  |
| `traefik.<segment_name>.frontend.redirect.entryPoint=https`               | Overrides `traefik.frontend.redirect.entryPoint`.                                                    |
| `traefik.<segment_name>.frontend.redirect.regex=^http://localhost/(.*)`   | Overrides `traefik.frontend.redirect.regex`.                                                         |
| `traefik.<segment_name>.frontend.redirect.replacement=http://mydomain/$1` | Overrides `traefik.frontend.redirect.replacement`.                                                   |
| `traefik.<segment_name>.frontend.redirect.permanent=true`                 | Return 301 instead of 302.                                                                           |
| `traefik.<segment_name>.frontend.rule=EXP`                                | Overrides `traefik.frontend.rule`. Default: `{service_name}.{sub_domain}.{domain}`                   |
| `traefik.<segment_name>.frontend.whitelistSourceRange=RANGE`              | Overrides `traefik.frontend.whitelistSourceRange`.                                                   |
| `traefik.<segment_name>.frontend.whiteList.sourceRange=RANGE`             | Overrides `traefik.frontend.whiteList.sourceRange`.                                                  |
| `traefik.<segment_name>.frontend.whiteList.useXForwardedFor=true`         | Use `X-Forwarded-For` header as valid source of IP for the white list.                               |

#### Custom Headers

| Label                                                                | Description                                                                                                                                                                         |
|----------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.<segment_name>.frontend.headers.customRequestHeaders=EXPR ` | Provides the container with custom request headers that will be appended to each request forwarded to the container.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code> |
| `traefik.<segment_name>.frontend.headers.customResponseHeaders=EXPR` | Appends the headers to each response returned by the container, before forwarding the response to the client.<br>Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>        |

#### Security Headers

| Label                                                                   | Description                                                                                                                                                                                         |
|-------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.<segment_name>.frontend.headers.allowedHosts=EXPR`             | Provides a list of allowed hosts that requests will be processed.<br>Format: `Host1,Host2`                                                                                                          |
| `traefik.<segment_name>.frontend.headers.hostsProxyHeaders=EXPR `       | Provides a list of headers that the proxied hostname may be stored.<br>Format: `HEADER1,HEADER2`                                                                                                    |
| `traefik.<segment_name>.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `traefik.<segment_name>.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `traefik.<segment_name>.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `traefik.<segment_name>.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `traefik.<segment_name>.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `traefik.<segment_name>.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `traefik.<segment_name>.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |
| `traefik.<segment_name>.frontend.headers.forceSTSHeader=false`          | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `traefik.<segment_name>.frontend.headers.frameDeny=false`               | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `traefik.<segment_name>.frontend.headers.customFrameOptionsValue=VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `traefik.<segment_name>.frontend.headers.contentTypeNosniff=true`       | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `traefik.<segment_name>.frontend.headers.browserXSSFilter=true`         | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `traefik.<segment_name>.frontend.headers.customBrowserXSSValue=VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `traefik.<segment_name>.frontend.headers.contentSecurityPolicy=VALUE`   | Adds CSP Header with the custom value.                                                                                                                                                              |
| `traefik.<segment_name>.frontend.headers.publicKey=VALUE`               | Adds pinned HTST public key header.                                                                                                                                                                 |
| `traefik.<segment_name>.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `traefik.<segment_name>.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |
