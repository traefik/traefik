# Consul Catalog Provider

Traefik can be configured to use service discovery catalog of Consul as a provider.

```toml
################################################################
# Consul Catalog Provider
################################################################

# Enable Consul Catalog Provider.
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

# Allow Consul server to serve the catalog reads regardless of whether it is the leader.
#
# Optional
# Default: false
#
stale = false

# Default base domain used for the frontend rules.
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

This provider will create routes matching on hostname based on the service name used in Consul.

To enable constraints see [provider-specific constraints section](/configuration/commons/#provider-specific).

## Tags

Additional settings can be defined using Consul Catalog tags.

!!! note
    The default prefix is `traefik`.

| Label                                                                    | Description                                                                                                                                                                                                                   |
|--------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `<prefix>.enable=false`                                                  | Disables this container in Traefik.                                                                                                                                                                                           |
| `<prefix>.protocol=https`                                                | Overrides the default `http` protocol.                                                                                                                                                                                        |
| `<prefix>.weight=10`                                                     | Assigns this weight to the container.                                                                                                                                                                                         |
| `traefik.backend.buffering.maxRequestBodyBytes=0`                        | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.maxResponseBodyBytes=0`                       | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.memRequestBodyBytes=0`                        | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.memResponseBodyBytes=0`                       | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `traefik.backend.buffering.retryExpression=EXPR`                         | See [buffering](/configuration/commons/#buffering) section.                                                                                                                                                                   |
| `<prefix>.backend.circuitbreaker.expression=EXPR`                        | Creates a [circuit breaker](/basics/#backends) to be used against the backend. ex: `NetworkErrorRatio() > 0.`                                                                                                                 |
| `<prefix>.backend.responseForwarding.flushInterval=10ms`                 | Defines the interval between two flushes when forwarding response from backend to client.                                                                                                                                     |
| `<prefix>.backend.healthcheck.path=/health`                              | Enables health check for the backend, hitting the container at `path`.                                                                                                                                                        |
| `<prefix>.backend.healthcheck.interval=1s`                               | Defines the health check interval.                                                                                                                                                                                            |
| `<prefix>.backend.healthcheck.port=8080`                                 | Sets a different port for the health check.                                                                                                                                                                                   |
| `traefik.backend.healthcheck.scheme=http`                                | Overrides the server URL scheme.                                                                                                                                                                                              |
| `<prefix>.backend.healthcheck.hostname=foobar.com`                       | Defines the health check hostname.                                                                                                                                                                                            |
| `<prefix>.backend.healthcheck.headers=EXPR`                              | Defines the health check request headers <br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                                                                                                     |
| `<prefix>.backend.loadbalancer.method=drr`                               | Overrides the default `wrr` load balancer algorithm.                                                                                                                                                                          |
| `<prefix>.backend.loadbalancer.stickiness=true`                          | Enables backend sticky sessions.                                                                                                                                                                                              |
| `<prefix>.backend.loadbalancer.stickiness.cookieName=NAME`               | Sets the cookie name manually for sticky sessions.                                                                                                                                                                            |
| `<prefix>.backend.loadbalancer.sticky=true`                              | Enables backend sticky sessions. (DEPRECATED)                                                                                                                                                                                 |
| `<prefix>.backend.maxconn.amount=10`                                     | Sets a maximum number of connections to the backend.<br>Must be used in conjunction with the below label to take effect.                                                                                                      |
| `<prefix>.backend.maxconn.extractorfunc=client.ip`                       | Sets the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect.                                        |
| `<prefix>.frontend.auth.basic=EXPR`                                      | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash` (DEPRECATED).                                                                                                                                 |
| `<prefix>.frontend.auth.basic.removeHeader=true`                         | If set to `true`, removes the `Authorization` header.                                                                                                                                                                         |
| `<prefix>.frontend.auth.basic.users=EXPR`                                | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash`.                                                                                                                                              |
| `<prefix>.frontend.auth.basic.usersfile=/path/.htpasswd`                 | Sets basic authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                         |
| `<prefix>.frontend.auth.digest.removeHeader=true`                        | If set to `true`, removes the `Authorization` header.                                                                                                                                                                         |
| `<prefix>.frontend.auth.digest.users=EXPR`                               | Sets digest authentication to this frontend in CSV format: `User:Realm:Hash,User:Realm:Hash`.                                                                                                                                 |
| `<prefix>.frontend.auth.digest.usersfile=/path/.htdigest`                | Sets digest authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                        |
| `<prefix>.frontend.auth.forward.address=https://example.com`             | Sets the URL of the authentication server.                                                                                                                                                                                    |
| `<prefix>.frontend.auth.forward.authResponseHeaders=EXPR`                | Sets the forward authentication authResponseHeaders in CSV format: `X-Auth-User,X-Auth-Header`                                                                                                                                |
| `<prefix>.frontend.auth.forward.tls.ca=/path/ca.pem`                     | Sets the Certificate Authority (CA) for the TLS connection with the authentication server.                                                                                                                                    |
| `<prefix>.frontend.auth.forward.tls.caOptional=true`                     | Checks the certificates if present but do not force to be signed by a specified Certificate Authority (CA).                                                                                                                   |
| `<prefix>.frontend.auth.forward.tls.cert=/path/server.pem`               | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `<prefix>.frontend.auth.forward.tls.insecureSkipVerify=true`             | If set to true invalid SSL certificates are accepted.                                                                                                                                                                         |
| `<prefix>.frontend.auth.forward.tls.key=/path/server.key`                | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `<prefix>.frontend.auth.forward.trustForwardHeader=true`                 | Trusts X-Forwarded-* headers.                                                                                                                                                                                                 |
| `<prefix>.frontend.auth.headerField=X-WebAuth-User`                      | Sets the header used to pass the authenticated user to the application.                                                                                                                                                       |
| `<prefix>.frontend.entryPoints=http,https`                               | Assigns this frontend to entry points `http` and `https`.<br>Overrides `defaultEntryPoints`                                                                                                                                   |
| `<prefix>.frontend.errors.<name>.backend=NAME`                           | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `<prefix>.frontend.errors.<name>.query=PATH`                             | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `<prefix>.frontend.errors.<name>.status=RANGE`                           | See [custom error pages](/configuration/commons/#custom-error-pages) section.                                                                                                                                                 |
| `<prefix>.frontend.passHostHeader=true`                                  | Forwards client `Host` header to the backend.                                                                                                                                                                                 |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.commonName=true`       | Add the issuer.commonName field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                  |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.country=true`          | Add the issuer.country field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                     |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.domainComponent=true`  | Add the issuer.domainComponent field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                             |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.locality=true`         | Add the issuer.locality field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                    |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.organization=true`     | Add the issuer.organization field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.province=true`         | Add the issuer.province field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                    |
| `<prefix>.frontend.passTLSClientCert.infos.issuer.serialNumber=true`     | Add the subject.serialNumber field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                               |
| `<prefix>.frontend.passTLSClientCert.infos.notAfter=true`                | Add the noAfter field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                            |
| `<prefix>.frontend.passTLSClientCert.infos.notBefore=true`               | Add the noBefore field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                           |
| `<prefix>.frontend.passTLSClientCert.infos.sans=true`                    | Add the sans field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                               |
| `<prefix>.frontend.passTLSClientCert.infos.subject.commonName=true`      | Add the subject.commonName field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                 |
| `<prefix>.frontend.passTLSClientCert.infos.subject.country=true`         | Add the subject.country field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                    |
| `<prefix>.frontend.passTLSClientCert.infos.subject.domainComponent=true` | Add the subject.domainComponent field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                            |
| `<prefix>.frontend.passTLSClientCert.infos.subject.locality=true`        | Add the subject.locality field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                   |
| `<prefix>.frontend.passTLSClientCert.infos.subject.organization=true`    | Add the subject.organization field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                               |
| `<prefix>.frontend.passTLSClientCert.infos.subject.province=true`        | Add the subject.province field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                                   |
| `<prefix>.frontend.passTLSClientCert.infos.subject.serialNumber=true`    | Add the subject.serialNumber field in a escaped client infos in the `X-Forwarded-Ssl-Client-Cert-Infos` header.                                                                                                               |
| `<prefix>.frontend.passTLSClientCert.pem=true`                           | Pass the escaped pem in the `X-Forwarded-Ssl-Client-Cert` header.                                                                                                                                                             |
| `<prefix>.frontend.passTLSCert=true`                                     | Forwards TLS Client certificates to the backend.                                                                                                                                                                              |
| `<prefix>.frontend.priority=10`                                          | Overrides default frontend priority.                                                                                                                                                                                          |
| `<prefix>.frontend.rateLimit.extractorFunc=EXP`                          | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `<prefix>.frontend.rateLimit.rateSet.<name>.period=6`                    | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `<prefix>.frontend.rateLimit.rateSet.<name>.average=6`                   | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `<prefix>.frontend.rateLimit.rateSet.<name>.burst=6`                     | See [rate limiting](/configuration/commons/#rate-limiting) section.                                                                                                                                                           |
| `<prefix>.frontend.redirect.entryPoint=https`                            | Enables Redirect to another entryPoint to this frontend (e.g. HTTPS).                                                                                                                                                         |
| `<prefix>.frontend.redirect.regex=^http://localhost/(.*)`                | Redirects to another URL to this frontend.<br>Must be set with `traefik.frontend.redirect.replacement`.                                                                                                                       |
| `<prefix>.frontend.redirect.replacement=http://mydomain/$1`              | Redirects to another URL to this frontend.<br>Must be set with `traefik.frontend.redirect.regex`.                                                                                                                             |
| `<prefix>.frontend.redirect.permanent=true`                              | Returns 301 instead of 302.                                                                                                                                                                                                   |
| `<prefix>.frontend.rule=EXPR`                                            | Overrides the default frontend rule. Default: `Host:{{.ServiceName}}.{{.Domain}}`.                                                                                                                                            |
| `<prefix>.frontend.whiteList.sourceRange=RANGE`                          | Sets a list of IP-Ranges which are allowed to access.<br>An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access. |
| `<prefix>.frontend.whiteList.useXForwardedFor=true`                      | Uses `X-Forwarded-For` header as valid source of IP for the white list.                                                                                                                                                       |

### Multiple frontends for a single service

If you need to support multiple frontends for a service, for example when having multiple `rules` that can't be combined, specify them as follows:

```
<prefix>.frontends.A.rule=Host:A:PathPrefix:/A
<prefix>.frontends.B.rule=Host:B:PathPrefix:/
```

`A` and `B` here are just arbitrary names, they can be anything. You can use any setting that applies to `<prefix>.frontend` from the table above.

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
| `<prefix>.frontend.headers.browserXSSFilter=true`         | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `<prefix>.frontend.headers.contentSecurityPolicy=VALUE`   | Adds CSP Header with the custom value.                                                                                                                                                              |
| `<prefix>.frontend.headers.contentTypeNosniff=true`       | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `<prefix>.frontend.headers.customBrowserXSSValue=VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `<prefix>.frontend.headers.customFrameOptionsValue=VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `<prefix>.frontend.headers.forceSTSHeader=false`          | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `<prefix>.frontend.headers.frameDeny=false`               | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `<prefix>.frontend.headers.hostsProxyHeaders=EXPR`        | Provides a list of headers that the proxied hostname may be stored.<br>Format: `HEADER1,HEADER2`                                                                                                    |
| `<prefix>.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |
| `<prefix>.frontend.headers.publicKey=VALUE`               | Adds HPKP header.                                                                                                                                                                                   |
| `<prefix>.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `<prefix>.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `<prefix>.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `<prefix>.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `<prefix>.frontend.headers.SSLForceHost=true`             | If `SSLForceHost` is `true` and `SSLHost` is set, requests will be forced to use `SSLHost` even the ones that are already using SSL. Default is false.                                              |
| `<prefix>.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `<prefix>.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `<prefix>.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `<prefix>.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |


### Examples

If you want that Traefik uses Consul tags correctly you need to defined them like that:

```js
traefik.enable=true
traefik.tags=api
traefik.tags=external
```

If the prefix defined in Traefik configuration is `bla`, tags need to be defined like that:

```js
bla.enable=true
bla.tags=api
bla.tags=external
```
