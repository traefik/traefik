---
title: "Traefik ForwardAuth Documentation"
description: "In Traefik Proxy, the HTTP ForwardAuth middleware delegates authentication to an external Service. Read the technical documentation."
---

# ForwardAuth

Using an External Service to Forward Authentication
{: .subtitle }

![AuthForward](../../assets/img/middleware/authforward.png)

The ForwardAuth middleware delegates authentication to an external service.
If the service answers with a 2XX code, access is granted, and the original request is performed.
Otherwise, the response from the authentication server is returned.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Forward authentication to example.com
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```yaml tab="Kubernetes"
# Forward authentication to example.com
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
```

```yaml tab="Consul Catalog"
# Forward authentication to example.com
- "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```yaml tab="File (YAML)"
# Forward authentication to example.com
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
```

```toml tab="File (TOML)"
# Forward authentication to example.com
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
```

## Forward-Request Headers

The following request properties are provided to the forward-auth target endpoint as `X-Forwarded-` headers.

| Property          | Forward-Request Header |
|-------------------|------------------------|
| HTTP Method       | X-Forwarded-Method     |
| Protocol          | X-Forwarded-Proto      |
| Host              | X-Forwarded-Host       |
| Request URI       | X-Forwarded-Uri        |
| Source IP-Address | X-Forwarded-For        |

## Configuration Options

### `address`

The `address` option defines the authentication server address.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
```

### `trustForwardHeader`

Set the `trustForwardHeader` option to `true` to trust all `X-Forwarded-*` headers.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.trustForwardHeader=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    trustForwardHeader: true
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.trustForwardHeader=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        trustForwardHeader: true
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    trustForwardHeader = true
```

### `authResponseHeaders`

The `authResponseHeaders` option is the list of headers to copy from the authentication server response and set on
forwarded request, replacing any existing conflicting headers.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.authResponseHeaders=X-Auth-User, X-Secret"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    authResponseHeaders:
      - X-Auth-User
      - X-Secret
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.authResponseHeaders=X-Auth-User, X-Secret"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        authResponseHeaders:
          - "X-Auth-User"
          - "X-Secret"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    authResponseHeaders = ["X-Auth-User", "X-Secret"]
```

### `authResponseHeadersRegex`

The `authResponseHeadersRegex` option is the regex to match headers to copy from the authentication server response and
set on forwarded request, after stripping all headers that match the regex.
It allows partial matching of the regular expression against the header key.
The start of string (`^`) and end of string (`$`) anchors should be used to ensure a full match against the header key.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.authResponseHeadersRegex=^X-"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    authResponseHeadersRegex: ^X-
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.authResponseHeadersRegex=^X-"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        authResponseHeadersRegex: "^X-"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    authResponseHeadersRegex = "^X-"
```

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.

### `authRequestHeaders`

The `authRequestHeaders` option is the list of the headers to copy from the request to the authentication server.
It allows filtering headers that should not be passed to the authentication server.
If not set or empty then all request headers are passed.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.authRequestHeaders=Accept,X-CustomHeader"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    authRequestHeaders:
      - "Accept"
      - "X-CustomHeader"
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.authRequestHeaders=Accept,X-CustomHeader"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        authRequestHeaders:
          - "Accept"
          - "X-CustomHeader"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    authRequestHeaders = "Accept,X-CustomHeader"
```

### `addAuthCookiesToResponse`

The `addAuthCookiesToResponse` option is the list of cookies to copy from the authentication server to the response, 
replacing any existing conflicting cookie from the forwarded response.

!!! info

    Please note that all backend cookies matching the configured list will not be added to the response.

```yaml tab="Docker"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.addAuthCookiesToResponse=Session-Cookie,State-Cookie"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    addAuthCookiesToResponse:
      - Session-Cookie
      - State-Cookie
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.addAuthCookiesToResponse=Session-Cookie,State-Cookie"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        addAuthCookiesToResponse:
          - "Session-Cookie"
          - "State-Cookie"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    addAuthCookiesToResponse = ["Session-Cookie", "State-Cookie"]
```

### `forwardBody`

_Optional, Default=false_

Set the `forwardBody` option to `true` to send Body.

!!! info

    As body is read inside Traefik before forwarding, this breaks streaming.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.forwardBody=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    forwardBody: true
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.forwardBody=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        forwardBody: true
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    forwardBody = true
```

### `maxBodySize`

_Optional, Default=-1_

Set the `maxBodySize` to limit the body size in bytes.
If body is bigger than this, it returns a 401 (unauthorized).
Default is `-1`, which means no limit.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.maxBodySize=1000"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    forwardBody: true
    maxBodySize: 1000
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.maxBodySize=1000"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        maxBodySize: 1000
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    forwardBody = true
    maxBodySize = 1000
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to the authentication server.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secured connection to the authentication server,
it defaults to the system bundle.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.tls.ca=path/to/local.crt"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    tls:
      caSecret: mycasercret

---
apiVersion: v1
kind: Secret
metadata:
  name: mycasercret
  namespace: default

data:
  # Must contain a certificate under either a `tls.ca` or a `ca.crt` key. 
  tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.tls.ca=path/to/local.crt"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        tls:
          ca: "path/to/local.crt"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    [http.middlewares.test-auth.forwardAuth.tls]
      ca = "path/to/local.crt"
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the authentication server.
When using this option, setting the `key` option is required.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
  - "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    tls:
      certSecret: mytlscert

---
apiVersion: v1
kind: Secret
metadata:
  name: mytlscert
  namespace: default

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        tls:
          cert: "path/to/foo.cert"
          key: "path/to/foo.key"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    [http.middlewares.test-auth.forwardAuth.tls]
      cert = "path/to/foo.cert"
      key = "path/to/foo.key"
```

!!! info

    For security reasons, the field does not exist for Kubernetes IngressRoute, and one should use the `secret` field instead.

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the authentication server.
When using this option, setting the `cert` option is required.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
  - "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    tls:
      certSecret: mytlscert

---
apiVersion: v1
kind: Secret
metadata:
  name: mytlscert
  namespace: default

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        tls:
          cert: "path/to/foo.cert"
          key: "path/to/foo.key"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    [http.middlewares.test-auth.forwardAuth.tls]
      cert = "path/to/foo.cert"
      key = "path/to/foo.key"
```

!!! info

    For security reasons, the field does not exist for Kubernetes IngressRoute, and one should use the `secret` field instead.

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to the authentication server accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.tls.insecureSkipVerify=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://example.com/auth
    tls:
      insecureSkipVerify: true
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.tls.InsecureSkipVerify=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
        tls:
          insecureSkipVerify: true
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
    [http.middlewares.test-auth.forwardAuth.tls]
      insecureSkipVerify: true
```

### `headerField`

_Optional_

You can define a header field to store the authenticated user using the `headerField`option.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.headerField=X-WebAuth-User"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    # ...
    headerField: X-WebAuth-User
```

```json tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.headerField=X-WebAuth-User"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        # ...
        headerField: "X-WebAuth-User"
```

```toml tab="File (TOML)"
[http.middlewares.test-auth.forwardAuth]
  # ...
  headerField = "X-WebAuth-User"
```

### `preserveLocationHeader`

_Optional, Default=false_

`preserveLocationHeader` defines whether to forward the `Location` header to the client as is or prefix it with the domain name of the authentication server.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.preserveLocationHeader=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    # ...
    preserveLocationHeader: true
```

```json tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.preserveLocationHeader=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        # ...
        preserveLocationHeader: true
```

```toml tab="File (TOML)"
[http.middlewares.test-auth.forwardAuth]
  # ...
  preserveLocationHeader = true
```

### `preserveRequestMethod`

_Optional, Default=false_

`preserveRequestMethod` defines whether to preserve the original request method while forwarding the request to the authentication server. By default, when this option is set to `false`, incoming requests are always forwarded as `GET` requests to the authentication server.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.preserveRequestMethod=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    # ...
    preserveRequestMethod: true
```

```json tab="Consul Catalog"
- "traefik.http.middlewares.test-auth.forwardauth.preserveRequestMethod=true"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        # ...
        preserveRequestMethod: true
```

```toml tab="File (TOML)"
[http.middlewares.test-auth.forwardAuth]
  # ...
  preserveRequestMethod = true
```

{!traefik-for-business-applications.md!}
