# ECS Provider

Træfik can be configured to use Amazon ECS as a provider.

## Configuration

```toml
################################################################
# ECS Provider
################################################################

# Enable ECS Provider.
[ecs]

# ECS Cluster Name.
#
# DEPRECATED - Please use `clusters`.
#
cluster = "default"

# ECS Clusters Name.
#
# Optional
# Default: ["default"]
#
clusters = ["default"]

# Enable watch ECS changes.
#
# Optional
# Default: true
#
watch = true

# Default domain used.
# Can be overridden by setting the "traefik.domain" label.
#
# Optional
# Default: ""
#
domain = "ecs.localhost"

# Enable auto discover ECS clusters.
#
# Optional
# Default: false
#
autoDiscoverClusters = false

# Polling interval (in seconds).
#
# Optional
# Default: 15
#
refreshSeconds = 15

# Expose ECS services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Region to use when connecting to AWS.
#
# Optional
#
region = "us-east-1"

# Access Key ID to use when connecting to AWS.
#
# Optional
#
accessKeyID = "abc"

# Secret Access Key to use when connecting to AWS.
#
# Optional
#
secretAccessKey = "123"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "ecs.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = 2
```

If `accessKeyID`/`secretAccessKey` is not given credentials will be resolved in the following order:

- From environment variables; `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.
- Shared credentials, determined by `AWS_PROFILE` and `AWS_SHARED_CREDENTIALS_FILE`, defaults to `default` and `~/.aws/credentials`.
- EC2 instance role or ECS task role

To enable constraints see [provider-specific constraints section](/configuration/commons/#provider-specific).

## Policy

Træfik needs the following policy to read ECS information:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "TraefikECSReadAccess",
            "Effect": "Allow",
            "Action": [
                "ecs:ListClusters",
                "ecs:DescribeClusters",
                "ecs:ListTasks",
                "ecs:DescribeTasks",
                "ecs:DescribeContainerInstances",
                "ecs:DescribeTaskDefinition",
                "ec2:DescribeInstances"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

## Labels: overriding default behaviour

Labels can be used on task containers to override default behaviour:

| Label                                                      | Description                                                                                                                                                                                                                   |
|------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.domain`                                           | Sets the default domain for frontend rules.                                                                                                                                                                                   |
| `traefik.enable=false`                                     | Disables this container in Træfik.                                                                                                                                                                                            |
| `traefik.port=80`                                          | Overrides the default `port` value. Overrides `NetworkBindings` from Docker Container                                                                                                                                         |
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
| `traefik.backend.loadbalancer.stickiness.cookieName=NAME`  | Sets the cookie manually  name for sticky sessions                                                                                                                                                                            |
| `traefik.backend.loadbalancer.sticky=true`                 | Enables backend sticky sessions (DEPRECATED)                                                                                                                                                                                  |
| `traefik.backend.maxconn.amount=10`                        | Sets a maximum number of connections to the backend.<br>Must be used in conjunction with the below label to take effect.                                                                                                      |
| `traefik.backend.maxconn.extractorfunc=client.ip`          | Sets the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect.                                        |
| `traefik.frontend.auth.basic=EXPR`                         | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash` (DEPRECATED).                                                                                                                                 |
| `traefik.frontend.auth.basic.removeHeader=true`            | If set to `true`, removes the `Authorization` header.                                                                                                                                                                         |
| `traefik.frontend.auth.basic.users=EXPR`                   | Sets basic authentication to this frontend in CSV format: `User:Hash,User:Hash`.                                                                                                                                              |
| `traefik.frontend.auth.basic.usersFile=/path/.htpasswd`    | Sets basic authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                         |
| `traefik.frontend.auth.digest.removeHeader=true`           | If set to `true`, removes the `Authorization` header.                                                                                                                                                                         |
| `traefik.frontend.auth.digest.users=EXPR`                  | Sets digest authentication to this frontend in CSV format: `User:Realm:Hash,User:Realm:Hash`.                                                                                                                                 |
| `traefik.frontend.auth.digest.usersFile=/path/.htdigest`   | Sets digest authentication with an external file; if users and usersFile are provided, both are merged, with external file contents having precedence.                                                                        |
| `traefik.frontend.auth.forward.address=https://example.com`| Sets the URL of the authentication server.                                                                                                                                                                                    |
| `traefik.frontend.auth.forward.tls.ca=/path/ca.pem`        | Sets the Certificate Authority (CA) for the TLS connection with the authentication server.                                                                                                                                    |
| `traefik.frontend.auth.forward.tls.caOptional=true`        | Checks the certificates if present but do not force to be signed by a specified Certificate Authority (CA).                                                                                                                   |
| `traefik.frontend.auth.forward.tls.cert=/path/server.pem`  | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `traefik.frontend.auth.forward.tls.insecureSkipVerify=true`| If set to true invalid SSL certificates are accepted.                                                                                                                                                                         |
| `traefik.frontend.auth.forward.tls.key=/path/server.key`   | Sets the Certificate for the TLS connection with the authentication server.                                                                                                                                                   |
| `traefik.frontend.auth.forward.trustForwardHeader=true`    | Trusts X-Forwarded-* headers.                                                                                                                                                                                                 |
| `traefik.frontend.auth.headerField=X-WebAuth-User`         | Sets the header used to pass the authenticated user to the application.                                                                                                                                                       |
| `traefik.frontend.auth.removeHeader=true`                  | If set to true, removes the Authorization header.                                                                                                                                                                             |
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
| `traefik.frontend.rule=EXPR`                               | Overrides the default frontend rule. Default: `Host:{instance_name}.{domain}`.                                                                                                                                                |
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
| `traefik.frontend.headers.publicKey=VALUE`               | Adds pinned HTST public key header.                                                                                                                                                                 |
| `traefik.frontend.headers.referrerPolicy=VALUE`          | Adds referrer policy  header.                                                                                                                                                                       |
| `traefik.frontend.headers.isDevelopment=false`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |
| `traefik.frontend.headers.SSLRedirect=true`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `traefik.frontend.headers.SSLTemporaryRedirect=true`     | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `traefik.frontend.headers.SSLHost=HOST`                  | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `traefik.frontend.headers.SSLForceHost=true`             | If `SSLForceHost` is `true` and `SSLHost` is set, requests will be forced to use `SSLHost` even the ones that are already using SSL. Default is false.                                              |
| `traefik.frontend.headers.SSLProxyHeaders=EXPR`          | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`).<br>Format:  <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                      |
| `traefik.frontend.headers.STSSeconds=315360000`          | Sets the max-age of the STS header.                                                                                                                                                                 |
| `traefik.frontend.headers.STSIncludeSubdomains=true`     | Adds the `IncludeSubdomains` section of the STS  header.                                                                                                                                            |
| `traefik.frontend.headers.STSPreload=true`               | Adds the preload flag to the STS  header.                                                                                                                                                           |
