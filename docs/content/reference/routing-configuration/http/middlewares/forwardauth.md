---
title: "Traefik ForwardAuth Documentation"
description: "In Traefik Proxy, the HTTP ForwardAuth middleware delegates authentication to an external Service. Read the technical documentation."
---

![AuthForward](../../../../assets/img/middleware/authforward.png)

The `forwardAuth` middleware delegates authentication to an external service.
If the service answers with a 2XX code, access is granted, and the original request is performed.
Otherwise, the response from the authentication server is returned.

## Configuration Example

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

```yaml tab="Docker & Swarm"
# Forward authentication to example.com
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```yaml tab="Consul Catalog"
# Forward authentication to example.com
- "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

</TabItem>

## Configuration Options

| Field      | Description     | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `address` | Authentication server address. | ""      | Yes      |
| `trustForwardHeader` | Trust all `X-Forwarded-*` headers. | false      | No      |
| `authResponseHeaders` | List of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers. | | No      |
| `authResponseHeadersRegex` | Regex to match by the headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.<br /> More information [here](#authresponseheadersregex). | ""      | No      |
| `authRequestHeaders` | List of the headers to copy from the request to the authentication server. <br /> It allows filtering headers that should not be passed to the authentication server. <br /> If not set or empty, then all request headers are passed. | | No      |
| `addAuthCookiesToResponse` | List of cookies to copy from the authentication server to the response, replacing any existing conflicting cookie from the forwarded response.<br /> Please note that all backend cookies matching the configured list will not be added to the response. || No      |
| `headerField` | Defines a header field to store the authenticated user. More information [here](#headerfield). || No      |
| `tls.caSecret` | Secret that contains the certificate authority used for the secured connection to the authentication server, it defaults to the system bundle. || No |
| `tls.certSecret` | Secret that contains both the private and public certificates used for the secure connection to the authentication server. || No |
| `tls.insecureSkipVerify` | During TLS connections, accepts any certificate presented by the server regardless of the host names it covers. | false | No |

### authResponseHeadersRegex

It allows partial matching of the regular expression against the header key.

The start of string (`^`) and end of string (`$`) anchors should be used to ensure a full match against the header key.

Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.

### headerField

You can define a header field to store the authenticated user using the `headerField` option.

```yaml tab="File (YAML)"
http:
  middlewares:
    test-auth:
      forwardAuth:
        # ...
        headerField: "X-WebAuth-User"
```

<TabItem value="Kubernetes">

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

## Forward-Request Headers

The following request properties are provided to the forward-auth target endpoint as `X-Forwarded-` headers.

| Property          | Forward-Request Header |
|-------------------|------------------------|
| HTTP Method       | X-Forwarded-Method     |
| Protocol          | X-Forwarded-Proto      |
| Host              | X-Forwarded-Host       |
| Request URI       | X-Forwarded-Uri        |
| Source IP-Address | X-Forwarded-For        |

{!traefik-for-business-applications.md!}
