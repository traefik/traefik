# Kubernetes Ingress Provider

Traefik can be configured to use Kubernetes Ingress as a provider.

See also [Kubernetes user guide](/user-guide/kubernetes).

## Configuration

```toml
################################################################
# Kubernetes Ingress Provider
################################################################

# Enable Kubernetes Ingress Provider.
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

# Ingress label selector to filter Ingress objects that should be processed.
#
# Optional
# Default: empty (process all Ingresses)
#
# labelselector = "A and not B"

# Value of `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.
# If the parameter is non-empty, only Ingresses containing an annotation with the same value are processed.
# Otherwise, Ingresses missing the annotation, having an empty value, or the value `traefik` are processed.
#
# Optional
# Default: empty
#
# ingressClass = "traefik-internal"

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

# Enable IngressEndpoint configuration.
# This will allow Traefik to update the status section of ingress objects, if desired.
#
# Optional
#
# [kubernetes.ingressEndpoint]
#
# At least one must be configured.
# `publishedservice` will override the `hostname` and `ip` settings if configured.
#
# hostname = "localhost"
# ip = "127.0.0.1"
# publishedService = "namespace/servicename"
```

### `endpoint`

The Kubernetes server endpoint as URL.

When deployed into Kubernetes, Traefik will read the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` to construct the endpoint.

The access token will be looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are provided mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik will try to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

### `labelselector`

By default, Traefik processes all Ingress objects in the configured namespaces.
A label selector can be defined to filter on specific Ingress objects only.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

### `ingressEndpoint`

You can configure a static hostname or IP address that Traefik will add to the status section of Ingress objects that it manages.
If you prefer, you can provide a service, which traefik will copy the status spec from.
This will give more flexibility in cloud/dynamic environments.

### TLS communication between Traefik and backend pods

Traefik automatically requests endpoint information based on the service provided in the ingress spec.
Although traefik will connect directly to the endpoints (pods), it still checks the service port to see if TLS communication is required.

There are 3 ways to configure Traefik to use https to communicate with backend pods:

1. If the service port defined in the ingress spec is 443 (note that you can still use `targetPort` to use a different port on your pod).
2. If the service port defined in the ingress spec has a name that starts with `https` (such as `https-api`, `https-web` or just `https`).
3. If the ingress spec includes the annotation `ingress.kubernetes.io/protocol: https`.

If either of those configuration options exist, then the backend communication protocol is assumed to be TLS, and will connect via TLS automatically.

!!! note
    Please note that by enabling TLS communication between traefik and your pods, you will have to have trusted certificates that have the proper trust chain and IP subject name.
    If this is not an option, you may need to skip TLS certificate verification.
    See the [insecureSkipVerify](/configuration/commons/#main-section) setting for more details.

## Annotations

### General annotations

The following general annotations are applicable on the Ingress object:

| Annotation                                                                      | Description                                                                                                                                                                                |
|---------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.ingress.kubernetes.io/app-root: "/index.html"`                         | Redirects all requests for `/` to the defined path. (1)                                                                                                                                    |
| `traefik.ingress.kubernetes.io/error-pages: <YML>`                              | See [custom error pages](/configuration/commons/#custom-error-pages) section. (2)                                                                                                          |
| `traefik.ingress.kubernetes.io/frontend-entry-points: http,https`               | Override the default frontend endpoints.                                                                                                                                                   |
| `traefik.ingress.kubernetes.io/pass-client-tls-cert: <YML>`                     | Forward the client certificate following the configuration in YAML. (3)                                                                                                                    |
| `traefik.ingress.kubernetes.io/pass-tls-cert: "true"`                           | Override the default frontend PassTLSCert value. Default: `false`.(DEPRECATED)                                                                                                             |
| `traefik.ingress.kubernetes.io/preserve-host: "true"`                           | Forward client `Host` header to the backend.                                                                                                                                               |
| `traefik.ingress.kubernetes.io/priority: "3"`                                   | Override the default frontend rule priority.                                                                                                                                               |
| `traefik.ingress.kubernetes.io/rate-limit: <YML>`                               | See [rate limiting](/configuration/commons/#rate-limiting) section. (4)                                                                                                                    |
| `traefik.ingress.kubernetes.io/redirect-entry-point: https`                     | Enables Redirect to another entryPoint for that frontend (e.g. HTTPS).                                                                                                                     |
| `traefik.ingress.kubernetes.io/redirect-permanent: "true"`                      | Return 301 instead of 302.                                                                                                                                                                 |
| `traefik.ingress.kubernetes.io/redirect-regex: ^http://localhost/(.*)`          | Redirect to another URL for that frontend. Must be set with `traefik.ingress.kubernetes.io/redirect-replacement`.                                                                          |
| `traefik.ingress.kubernetes.io/redirect-replacement: http://mydomain/$1`        | Redirect to another URL for that frontend. Must be set with `traefik.ingress.kubernetes.io/redirect-regex`.                                                                                |
| `traefik.ingress.kubernetes.io/request-modifier: AddPrefix: /users`             | Adds a [request modifier](/basics/#modifiers) to the backend request.                                                                                                                      |
| `traefik.ingress.kubernetes.io/rewrite-target: /users`                          | Replaces each matched Ingress path with the specified one, and adds the old path to the `X-Replaced-Path` header.                                                                          |
| `traefik.ingress.kubernetes.io/rule-type: PathPrefixStrip`                      | Overrides the default frontend rule type. Only path-related matchers can be specified [(`Path`, `PathPrefix`, `PathStrip`, `PathPrefixStrip`)](/basics/#path-matcher-usage-guidelines).(5) |
| `traefik.ingress.kubernetes.io/service-weights: <YML>`                          | Set ingress backend weights specified as percentage or decimal numbers in YAML. (6)                                                                                                        |
| `traefik.ingress.kubernetes.io/whitelist-source-range: "1.2.3.0/24, fe80::/16"` | A comma-separated list of IP ranges permitted for access (7).                                                                                                                              |
| `ingress.kubernetes.io/whitelist-x-forwarded-for: "true"`                       | Use `X-Forwarded-For` header as valid source of IP for the white list.                                                                                                                     |
| `ingress.kubernetes.io/protocol:<NAME>`                                | Set the protocol Traefik will use to communicate with pods. Acceptable protocols: http,https,h2c                                                                                                                        |

<1> `traefik.ingress.kubernetes.io/app-root`:
Non-root paths will not be affected by this annotation and handled normally.
This annotation may not be combined with other redirect annotations.
Trying to do so will result in the other redirects being ignored.
This annotation can be used in combination with `traefik.ingress.kubernetes.io/redirect-permanent` to configure whether the `app-root` redirect is a 301 or a 302.

<2> `traefik.ingress.kubernetes.io/error-pages` example:

```yaml
foo:
  status:
  - "404"
  backend: bar
  query: /bar
fii:
  status:
  - "503"
  - "500"
  backend: bar
  query: /bir
```

<3> `traefik.ingress.kubernetes.io/pass-client-tls-cert` example:

```yaml
# add escaped pem in the `X-Forwarded-Tls-Client-Cert` header
pem: true
# add escaped certificate following infos in the `X-Forwarded-Tls-Client-Cert-Infos` header
infos:
  notafter: true
  notbefore: true
  sans: true
  subject:
    country: true
    province: true
    locality: true
    organization: true
    commonname: true
    serialnumber: true
```

If `pem` is set, it will add a `X-Forwarded-Tls-Client-Cert` header that contains the escaped pem as value.  
If at least one flag of the `infos` part is set, it will add a `X-Forwarded-Tls-Client-Cert-Infos` header that contains an escaped string composed of the client certificate data selected by the infos flags.
This infos part is composed like the following example (not escaped):
```Subject="C=FR,ST=SomeState,L=Lyon,O=Cheese,CN=*.cheese.org",NB=1531900816,NA=1563436816,SAN=*.cheese.org,*.cheese.net,cheese.in,test@cheese.org,test@cheese.net,10.0.1.0,10.0.1.2```

<4> `traefik.ingress.kubernetes.io/rate-limit` example:

```yaml
extractorfunc: client.ip
rateset:
  bar:
    period: 3s
    average: 6
    burst: 9
  foo:
    period: 6s
    average: 12
    burst: 18
```

<5> `traefik.ingress.kubernetes.io/rule-type`
Note: `ReplacePath` is deprecated in this annotation, use the `traefik.ingress.kubernetes.io/request-modifier` annotation instead. Default: `PathPrefix`. 

<6> `traefik.ingress.kubernetes.io/service-weights`:
Service weights enable to split traffic across multiple backing services in a fine-grained manner.

Example:

```yaml
service_backend1: 12.50%
service_backend2: 12.50%
service_backend3: 75 # Same as 75%, the percentage sign is optional
```

A single service backend definition may be omitted; in this case, Traefik auto-completes that service backend to 100% automatically.
Conveniently, users need not bother to compute the percentage remainder for a main service backend.
For instance, in the example above `service_backend3` does not need to be specified to be assigned 75%.

!!! note
    For each service weight given, the Ingress specification must include a backend item with the corresponding `serviceName` and (if given) matching path.

Currently, 3 decimal places for the weight are supported.
An attempt to exceed the precision should be avoided as it may lead to percentage computation flaws and, in consequence, Ingress parsing errors.

For each path definition, this annotation will fail if:

- the sum of backend weights exceeds 100% or
- the sum of backend weights is less than 100% without one or more omitted backends

See also the [user guide section traffic splitting](/user-guide/kubernetes/#traffic-splitting).

<7> `traefik.ingress.kubernetes.io/whitelist-source-range`:
All source IPs are permitted if the list is empty or a single range is ill-formatted.
Please note, you may have to set `service.spec.externalTrafficPolicy` to the value `Local` to preserve the source IP of the request for filtering.
Please see [this link](https://kubernetes.io/docs/tutorials/services/source-ip/) for more information.


!!! note
    Please note that `traefik.ingress.kubernetes.io/redirect-regex` and `traefik.ingress.kubernetes.io/redirect-replacement` do not have to be set if `traefik.ingress.kubernetes.io/redirect-entry-point` is defined for the redirection (they will not be used in this case).

The following annotations are applicable on the Service object associated with a particular Ingress object:

| Annotation                                                               | Description                                                                                                                                                                           |
|--------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.ingress.kubernetes.io/buffering: <YML>`                         | (1) See the [buffering](/configuration/commons/#buffering) section.                                                                                                                   |
| `traefik.backend.loadbalancer.sticky: "true"`                            | Enable backend sticky sessions (DEPRECATED).                                                                                                                                          |
| `traefik.ingress.kubernetes.io/affinity: "true"`                         | Enable backend sticky sessions.                                                                                                                                                       |
| `traefik.ingress.kubernetes.io/circuit-breaker-expression: <expression>` | Set the circuit breaker expression for the backend.                                                                                                                                   |
| `traefik.ingress.kubernetes.io/responseforwarding-flushinterval: "10ms`  | Defines the interval between two flushes when forwarding response from backend to client.                                                                                             |
| `traefik.ingress.kubernetes.io/load-balancer-method: drr`                | Override the default `wrr` load balancer algorithm.                                                                                                                                   |
| `traefik.ingress.kubernetes.io/max-conn-amount: "10"`                    | Sets the maximum number of simultaneous connections to the backend.<br>Must be used in conjunction with the label below to take effect.                                               |
| `traefik.ingress.kubernetes.io/max-conn-extractor-func: client.ip`       | Set the function to be used against the request to determine what to limit maximum connections to the backend by.<br>Must be used in conjunction with the above label to take effect. |
| `traefik.ingress.kubernetes.io/session-cookie-name: <NAME>`              | Manually set the cookie name for sticky sessions.                                                                                                                                     |

<1> `traefik.ingress.kubernetes.io/buffering` example:

```yaml
maxrequestbodybytes: 10485760
memrequestbodybytes: 2097153
maxresponsebodybytes: 10485761
memresponsebodybytes: 2097152
retryexpression: IsNetworkError() && Attempts() <= 2
```

!!! note
    `traefik.ingress.kubernetes.io/` and `ingress.kubernetes.io/` are supported prefixes.

### Custom Headers Annotations

|                        Annotation                     |                                                                                             Description                                                                          |
| ------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ingress.kubernetes.io/custom-request-headers: EXPR`  | Provides the container with custom request headers that will be appended to each request forwarded to the container. Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code> |
| `ingress.kubernetes.io/custom-response-headers: EXPR` | Appends the headers to each response returned by the container, before forwarding the response to the client. Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>        |

### Security Headers Annotations

The following security annotations are applicable on the Ingress object:

|                        Annotation                         |                                                                                             Description                                                                                             |
| ----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ingress.kubernetes.io/allowed-hosts: EXPR`               | Provides a list of allowed hosts that requests will be processed. Format: `Host1,Host2`                                                                                                             |
| `ingress.kubernetes.io/browser-xss-filter: "true"`        | Adds the X-XSS-Protection header with the value `1; mode=block`.                                                                                                                                    |
| `ingress.kubernetes.io/content-security-policy: VALUE`    | Adds CSP Header with the custom value.                                                                                                                                                              |
| `ingress.kubernetes.io/content-type-nosniff: "true"`      | Adds the `X-Content-Type-Options` header with the value `nosniff`.                                                                                                                                  |
| `ingress.kubernetes.io/custom-browser-xss-value: VALUE`   | Set custom value for X-XSS-Protection header. This overrides the BrowserXssFilter option.                                                                                                           |
| `ingress.kubernetes.io/custom-frame-options-value: VALUE` | Overrides the `X-Frame-Options` header with the custom value.                                                                                                                                       |
| `ingress.kubernetes.io/force-hsts: "false"`               | Adds the STS  header to non-SSL requests.                                                                                                                                                           |
| `ingress.kubernetes.io/frame-deny: "true"`                | Adds the `X-Frame-Options` header with the value of `DENY`.                                                                                                                                         |
| `ingress.kubernetes.io/hsts-max-age: "315360000"`         | Sets the max-age of the HSTS header.                                                                                                                                                                |
| `ingress.kubernetes.io/hsts-include-subdomains: "true"`   | Adds the IncludeSubdomains section of the STS  header.                                                                                                                                              |
| `ingress.kubernetes.io/hsts-preload: "true"`              | Adds the preload flag to the HSTS  header.                                                                                                                                                          |
| `ingress.kubernetes.io/is-development: "false"`           | This will cause the `AllowedHosts`, `SSLRedirect`, and `STSSeconds`/`STSIncludeSubdomains` options to be ignored during development.<br>When deploying to production, be sure to set this to false. |
| `ingress.kubernetes.io/proxy-headers: EXPR`               | Provides a list of headers that the proxied hostname may be stored. Format:  `HEADER1,HEADER2`                                                                                                      |
| `ingress.kubernetes.io/public-key: VALUE`                 | Adds HPKP header.                                                                                                                                                                                   |
| `ingress.kubernetes.io/referrer-policy: VALUE`            | Adds referrer policy  header.                                                                                                                                                                       |
| `ingress.kubernetes.io/ssl-redirect: "true"`              | Forces the frontend to redirect to SSL if a non-SSL request is sent.                                                                                                                                |
| `ingress.kubernetes.io/ssl-temporary-redirect: "true"`    | Forces the frontend to redirect to SSL if a non-SSL request is sent, but by sending a 302 instead of a 301.                                                                                         |
| `ingress.kubernetes.io/ssl-host: HOST`                    | This setting configures the hostname that redirects will be based on. Default is "", which is the same host as the request.                                                                         |
| `ingress.kubernetes.io/ssl-force-host: "true"`            | If `SSLForceHost` is `true` and `SSLHost` is set, requests will be forced to use `SSLHost` even the ones that are already using SSL. Default is false.                                              |
| `ingress.kubernetes.io/ssl-proxy-headers: EXPR`           | Header combinations that would signify a proper SSL Request (Such as `X-Forwarded-For:https`). Format: <code>HEADER:value&vert;&vert;HEADER2:value2</code>                                          |

### Authentication

Additional authentication annotations can be added to the Ingress object.
The source of the authentication is a Secret object that contains the credentials.

| Annotation                                                           | basic | digest | forward | Description                                                                                                 |
|----------------------------------------------------------------------|-------|--------|---------|-------------------------------------------------------------------------------------------------------------|
| `ingress.kubernetes.io/auth-type: basic`                             |   x   |   x    |    x    | Contains the authentication type: `basic`, `digest`, `forward`.                                             |
| `ingress.kubernetes.io/auth-secret: mysecret`                        |   x   |   x    |         | Name of Secret containing the username and password with access to the paths defined in the Ingress object. |
| `ingress.kubernetes.io/auth-remove-header: true`                     |   x   |   x    |         | If set to `true` removes the `Authorization` header.                                                        |
| `ingress.kubernetes.io/auth-header-field: X-WebAuth-User`            |   x   |   x    |         | Pass Authenticated user to application via headers.                                                         |
| `ingress.kubernetes.io/auth-url: https://example.com`                |       |        |    x    | [The URL of the authentication server](/configuration/entrypoints/#forward-authentication).                 |
| `ingress.kubernetes.io/auth-trust-headers: false`                    |       |        |    x    | Trust `X-Forwarded-*` headers.                                                                              |
| `ingress.kubernetes.io/auth-response-headers: X-Auth-User, X-Secret` |       |        |    x    | Copy headers from the authentication server to the request.                                                 |
| `ingress.kubernetes.io/auth-tls-secret: secret`                      |       |        |    x    | Name of Secret containing the certificate and key for the forward auth.                                     |
| `ingress.kubernetes.io/auth-tls-insecure`                            |       |        |    x    | If set to `true` invalid SSL certificates are accepted.                                                     |

The secret must be created in the same namespace as the Ingress object.

The following limitations hold for basic/digest auth:

- The realm is not configurable; the only supported (and default) value is `traefik`.
- The Secret must contain a single file only.

### TLS certificates management

TLS certificates can be managed in Secrets objects.
More information are available in the  [User Guide](/user-guide/kubernetes/#add-a-tls-certificate-to-the-ingress).

!!! note
    Only TLS certificates provided by users can be stored in Kubernetes Secrets.
    [Let's Encrypt](https://letsencrypt.org) certificates cannot be managed in Kubernets Secrets yet.

### Global Default Backend Ingresses

Ingresses can be created that look like the following:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: cheese
spec:
  backend:
    serviceName: stilton
    servicePort: 80
```

This ingress follows the [Global Default Backend](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource) property of ingresses.
This will allow users to create a "default backend" that will match all unmatched requests.

!!! note
    Due to Traefik's use of priorities, you may have to set this ingress priority lower than other ingresses in your environment, to avoid this global ingress from satisfying requests that _could_ match other ingresses.
    To do this, use the `traefik.ingress.kubernetes.io/priority` annotation (as seen in [General Annotations](/configuration/backends/kubernetes/#general-annotations)) on your ingresses accordingly.
