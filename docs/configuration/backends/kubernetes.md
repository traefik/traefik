# Kubernetes Ingress Backend

Træfik can be configured to use Kubernetes Ingress as a backend configuration.

See also [Kubernetes user guide](/user-guide/kubernetes).


## Configuration

```toml
################################################################
# Kubernetes Ingress configuration backend
################################################################

# Enable Kubernetes Ingress configuration backend.
[kubernetes]

# Kubernetes server endpoint.
#
# Optional for in-cluster configuration, required otherwise.
# Default: empty
#
# endpoint = "http://localhost:8080"

# Bearer token used for the Kubernetes client configuration.
#
# Optional
# Default: empty
#
# token = "my token"

# Path to the certificate authority file.
# Used for the Kubernetes client configuration.
#
# Optional
# Default: empty
#
# certAuthFilePath = "/my/ca.crt"

# Array of namespaces to watch.
#
# Optional
# Default: all namespaces (empty array).
#
# namespaces = ["default", "production"]

# Ingress label selector to identify Ingress objects that should be processed.
#
# Optional
# Default: empty (process all Ingresses)
#
# labelselector = "A and not B"

# Disable PassHost Headers.
#
# Optional
# Default: false
#
# disablePassHostHeaders = true

# Enable PassTLSCert Headers.
#
# Optional
# Default: false
#
# enablePassTLSCert = true

# Override default configuration template.
#
# Optional
# Default: <built-in template>
#
# filename = "kubernetes.tmpl"
```

### `endpoint`

The Kubernetes server endpoint.

When deployed as a replication controller in Kubernetes, Traefik will use the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` to construct the endpoint.

Secure token will be found in `/var/run/secrets/kubernetes.io/serviceaccount/token` and SSL CA cert in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`

The endpoint may be given to override the environment variable values.

When the environment variables are not found, Traefik will try to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster from localhost.

### `labelselector`

Ingress label selector to identify Ingress objects that should be processed.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.


## Annotations

Annotations can be used on containers to override default behaviour for the whole Ingress resource:

- `traefik.frontend.rule.type: PathPrefixStrip`  
    Override the default frontend rule type. Default: `PathPrefix`.
- `traefik.frontend.priority: "3"`
    Override the default frontend rule priority.
- `traefik.frontend.redirect: https`:
    Enables Redirect to another entryPoint for that frontend (e.g. HTTPS).
- `traefik.frontend.entryPoints: http,https`  
    Override the default frontend endpoints.
- `traefik.frontend.passTLSCert: true`  
    Override the default frontend PassTLSCert value. Default: `false`.

Annotations can be used on the Kubernetes service to override default behaviour:

- `traefik.backend.loadbalancer.method=drr`  
    Override the default `wrr` load balancer algorithm
- `traefik.backend.loadbalancer.stickiness=true`      
    Enable backend sticky sessions
- `traefik.backend.loadbalancer.stickiness.cookieName=NAME`      
    Manually set the cookie name for sticky sessions
- `traefik.backend.loadbalancer.sticky=true`      
    Enable backend sticky sessions (DEPRECATED)

Additionally, an annotation can be used on Kubernetes services to set the [circuit breaker expression](/basics/#backends) for a backend.

- `traefik.backend.circuitbreaker: <expression>`  
    Set the circuit breaker expression for the backend. Default: `nil`.

As known from nginx when used as Kubernetes Ingress Controller, a list of IP-Ranges which are allowed to access can be configured by using an ingress annotation:

- `ingress.kubernetes.io/whitelist-source-range: "1.2.3.0/24, fe80::/16"`

An unset or empty list allows all Source-IPs to access.
If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access.

#### Security annotations

The following security annotations can be applied to the ingress object to add security features:

| Annotation                                               | Description                                                                                                                                                                                         |
|----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ingress.kubernetes.io/allowed-hosts:EXPR`               | Provides a list of allowed hosts that requests will be processed. Format: `Host1,Host2`                                                                                                             |
| `ingress.kubernetes.io/custom-request-headers:EXPR `     | Provides the container with custom request headers that will be appended to each request forwarded to the container. Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>                    |
| `ingress.kubernetes.io/custom-response-headers:EXPR`     | Appends the headers to each response returned by the container, before forwarding the response to the client. Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>                           |
| `ingress.kubernetes.io/proxy-headers:EXPR `              | Provides a list of headers that the proxied hostname may be stored. Format:  `HEADER1,HEADER2`                                                                                                      |
| `ingress.kubernetes.io/ssl-redirect:true`                | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `ingress.kubernetes.io/ssl-temporary-redirect:true`      | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `ingress.kubernetes.io/ssl-host:HOST`                    | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `ingress.kubernetes.io/ssl-proxy-headers:EXPR`           | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`). Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                          |
| `ingress.kubernetes.io/hsts-max-age:315360000`           | Sets the max-age of the HSTS header.                                                                                                                                                                |
| `ngress.kubernetes.io/hsts-include-subdomains:true`      | Adds the IncludeSubdomains section of the STS  header.                                                                                                                                              |
| `ingress.kubernetes.io/hsts-preload:true`                | Adds the preload flag to the HSTS  header.                                                                                                                                                          |
| `ingress.kubernetes.io/force-hsts:false`                 | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `ingress.kubernetes.io/frame-deny:false`                 | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `ingress.kubernetes.io/custom-frame-options-value:VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `ingress.kubernetes.io/content-type-nosniff:true`        | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `ingress.kubernetes.io/browser-xss-filter:true`          | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `ingress.kubernetes.io/content-security-policy:VALUE`    | Adds CSP Header with the custom value.                                                                                                                                                              |
| `ingress.kubernetes.io/public-key:VALUE`                 | Adds pinned HTST public key header.                                                                                                                                                                 |
| `ingress.kubernetes.io/referrer-policy:VALUE`            | Adds referrer policy  header.                                                                                                                                                                       |
| `ingress.kubernetes.io/is-development:false`             | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |

### Authentication

Is possible to add additional authentication annotations in the Ingress rule.
The source of the authentication is a secret that contains usernames and passwords inside the key auth.

- `ingress.kubernetes.io/auth-type`: `basic`
- `ingress.kubernetes.io/auth-secret`: `mysecret`  
    Contains the usernames and passwords with access to the paths defined in the Ingress Rule.

The secret must be created in the same namespace as the Ingress rule.

Limitations:

- Basic authentication only.
- Realm not configurable; only `traefik` default.
- Secret must contain only single file.
