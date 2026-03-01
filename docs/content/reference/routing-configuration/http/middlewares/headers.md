---
title: "Traefik Headers Documentation"
description: "In Traefik Proxy, the HTTP headers middleware manages the headers of requests and responses. Read the technical documentation."
---

The Headers middleware manages the headers of requests and responses.

By default, the following headers are automatically added when proxying requests:

| Property                  | HTTP Header                |
|---------------------------|----------------------------|
| <a id="opt-Clients-IP" href="#opt-Clients-IP" title="#opt-Clients-IP">Client's IP</a> | `X-Forwarded-For`, `X-Real-Ip` |
| <a id="opt-Host" href="#opt-Host" title="#opt-Host">Host</a> | `X-Forwarded-Host`           |
| <a id="opt-Port" href="#opt-Port" title="#opt-Port">Port</a> | `X-Forwarded-Port`           |
| <a id="opt-Protocol" href="#opt-Protocol" title="#opt-Protocol">Protocol</a> | `X-Forwarded-Proto`          |
| <a id="opt-Proxy-Servers-Hostname" href="#opt-Proxy-Servers-Hostname" title="#opt-Proxy-Servers-Hostname">Proxy Server's Hostname</a> | `X-Forwarded-Server`         |

## Configuration Examples

### Adding Headers to the Request and the Response

The following example adds the `X-Script-Name` header to the proxied request and the `X-Custom-Response-Header` header to the response

```yaml tab="Structured (YAML)"
http:
  middlewares:
    testHeader:
      headers:
        customRequestHeaders:
          X-Script-Name: "test"
        customResponseHeaders:
          X-Custom-Response-Header: "value"
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.testHeader.headers]
    [http.middlewares.testHeader.headers.customRequestHeaders]
        X-Script-Name = "test"
    [http.middlewares.testHeader.headers.customResponseHeaders]
        X-Custom-Response-Header = "value"
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.testHeader.headers.customrequestheaders.X-Script-Name=test"
  - "traefik.http.middlewares.testHeader.headers.customresponseheaders.X-Custom-Response-Header=value"
```

```json tab="Tags"
{
  //...
  "Tags": [
    "traefik.http.middlewares.testheader.headers.customrequestheaders.X-Script-Name=test",
    "traefik.http.middlewares.testheader.headers.customresponseheaders.X-Custom-Response-Header=value"
  ]
}

```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-header
spec:
  headers:
    customRequestHeaders:
      X-Script-Name: "test"
    customResponseHeaders:
      X-Custom-Response-Header: "value"
```

### Adding and Removing Headers

In the following example, requests are proxied with an extra `X-Script-Name` header while their `X-Custom-Request-Header` header gets stripped,
and responses are stripped of their `X-Custom-Response-Header` header.

```yaml tab="Structured (YAML)"
http:
  middlewares:
    testHeader:
      headers:
        customRequestHeaders:
          X-Script-Name: "test" # Adds
          X-Custom-Request-Header: "" # Removes
        customResponseHeaders:
          X-Custom-Response-Header: "" # Removes
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.testHeader.headers]
    [http.middlewares.testHeader.headers.customRequestHeaders]
        X-Script-Name = "test" # Adds
        X-Custom-Request-Header = "" # Removes
    [http.middlewares.testHeader.headers.customResponseHeaders]
        X-Custom-Response-Header = "" # Removes
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.testheader.headers.customrequestheaders.X-Script-Name=test"
  - "traefik.http.middlewares.testheader.headers.customrequestheaders.X-Custom-Request-Header="
  - "traefik.http.middlewares.testheader.headers.customresponseheaders.X-Custom-Response-Header="
```

```json tab="Tags"
{
  "Tags" : [
    "traefik.http.middlewares.testheader.headers.customrequestheaders.X-Script-Name=test",
    "traefik.http.middlewares.testheader.headers.customrequestheaders.X-Custom-Request-Header=",
    "traefik.http.middlewares.testheader.headers.customresponseheaders.X-Custom-Response-Header="
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-header
spec:
  headers:
    customRequestHeaders:
      X-Script-Name: "test" # Adds
      X-Custom-Request-Header: "" # Removes
    customResponseHeaders:
      X-Custom-Response-Header: "" # Removes
```

### Using Security Headers

Security-related headers (HSTS headers, Browser XSS filter, etc) can be managed similarly to custom headers as shown above.
This functionality makes it possible to easily use security features by adding headers.

```yaml tab="Structured (YAML)"
http:
  middlewares:
    testHeader:
      headers:
        frameDeny: true
        browserXssFilter: true
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.testHeader.headers]
    frameDeny = true
    browserXssFilter = true
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.testHeader.headers.framedeny=true"
  - "traefik.http.middlewares.testHeader.headers.browserxssfilter=true"
```

```json tab="Tags"
{
  "Tags" : [
    "traefik.http.middlewares.testheader.headers.framedeny=true",
    "traefik.http.middlewares.testheader.headers.browserxssfilter=true"
  ]
}

```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-header
spec:
  headers:
    frameDeny: true
    browserXssFilter: true
```

### CORS Headers

CORS (Cross-Origin Resource Sharing) headers can be added and configured in a manner similar to the custom headers above.
This functionality allows for more advanced security features to quickly be set.
If CORS headers are set, then the middleware does not pass preflight requests to any service,
instead the response will be generated and sent back to the client directly.  
Please note that the example below is by no means authoritative or exhaustive,
and should not be used as is for production.

```yaml tab="Structured (YAML)"
http:
  middlewares:
    testHeader:
      headers:
        accessControlAllowMethods:
          - GET
          - OPTIONS
          - PUT
        accessControlAllowHeaders: "*"
        accessControlAllowOriginList:
          - https://foo.bar.org
          - https://example.org
        accessControlMaxAge: 100
        addVaryHeader: true
```

```toml tab="Structured (TOML)"
[http.middlewares]
  [http.middlewares.testHeader.headers]
    accessControlAllowMethods = ["GET", "OPTIONS", "PUT"]
    accessControlAllowHeaders = [ "*" ]
    accessControlAllowOriginList = ["https://foo.bar.org","https://example.org"]
    accessControlMaxAge = 100
    addVaryHeader = true
```

```yaml tab="Labels"
labels:
  - "traefik.http.middlewares.testheader.headers.accesscontrolallowmethods=GET,OPTIONS,PUT"
  - "traefik.http.middlewares.testheader.headers.accesscontrolallowheaders=*"
  - "traefik.http.middlewares.testheader.headers.accesscontrolalloworiginlist=https://foo.bar.org,https://example.org"
  - "traefik.http.middlewares.testheader.headers.accesscontrolmaxage=100"
  - "traefik.http.middlewares.testheader.headers.addvaryheader=true"
```

```json tab="Tags"
{
  "Tags" : [
    "traefik.http.middlewares.testheader.headers.accesscontrolallowmethods=GET,OPTIONS,PUT",
     "traefik.http.middlewares.testheader.headers.accesscontrolallowheaders=*",
    "traefik.http.middlewares.testheader.headers.accesscontrolalloworiginlist=https://foo.bar.org,https://example.org",
    "traefik.http.middlewares.testheader.headers.accesscontrolmaxage=100",
    "traefik.http.middlewares.testheader.headers.addvaryheader=true"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-header
spec:
  headers:
    accessControlAllowMethods:
      - "GET"
      - "OPTIONS"
      - "PUT"
    accessControlAllowHeaders:
      - "*"
    accessControlAllowOriginList:
      - "https://foo.bar.org"
      - "https://example.org"
    accessControlMaxAge: 100
    addVaryHeader: true
```

## Configuration Options

!!! warning

    Custom headers will overwrite existing headers if they have identical names.

!!! note ""

    The detailed documentation for security headers can be found in [unrolled/secure](https://github.com/unrolled/secure#available-options).

| Field                         | Description                                       | Default   | Required |
| ----------------------------- | ------------------------------------------------- | --------- | -------- |
| <a id="opt-customRequestHeaders" href="#opt-customRequestHeaders" title="#opt-customRequestHeaders">`customRequestHeaders`</a> | Lists the header names and values for requests.  | [] | No |
| <a id="opt-customResponseHeaders" href="#opt-customResponseHeaders" title="#opt-customResponseHeaders">`customResponseHeaders`</a> | Lists the header names and values for responses. | [] | No |
| <a id="opt-accessControlAllowCredentials" href="#opt-accessControlAllowCredentials" title="#opt-accessControlAllowCredentials">`accessControlAllowCredentials`</a> | Indicates if the request can include user credentials.| false     | No |
| <a id="opt-accessControlAllowHeaders" href="#opt-accessControlAllowHeaders" title="#opt-accessControlAllowHeaders">`accessControlAllowHeaders`</a> | Specifies allowed request header names.          | [] | No |
| <a id="opt-accessControlAllowMethods" href="#opt-accessControlAllowMethods" title="#opt-accessControlAllowMethods">`accessControlAllowMethods`</a> | Specifies allowed request methods.               | [] | No |
| <a id="opt-accessControlAllowOriginList" href="#opt-accessControlAllowOriginList" title="#opt-accessControlAllowOriginList">`accessControlAllowOriginList`</a> | Specifies allowed origins. More information [here](#accesscontrolalloworiginlist)      | []      | No |
| <a id="opt-accessControlAllowOriginListRegex" href="#opt-accessControlAllowOriginListRegex" title="#opt-accessControlAllowOriginListRegex">`accessControlAllowOriginListRegex`</a> | Allows origins matching regex. More information [here](#accesscontrolalloworiginlistregex)            | []      | No |
| <a id="opt-accessControlExposeHeaders" href="#opt-accessControlExposeHeaders" title="#opt-accessControlExposeHeaders">`accessControlExposeHeaders`</a> | Specifies which headers are safe to expose to the API of a CORS API specification.       |  []    | No |
| <a id="opt-accessControlMaxAge" href="#opt-accessControlMaxAge" title="#opt-accessControlMaxAge">`accessControlMaxAge`</a> | Time (in seconds) to cache preflight requests.   | 0         | No |
| <a id="opt-addVaryHeader" href="#opt-addVaryHeader" title="#opt-addVaryHeader">`addVaryHeader`</a> | Used in conjunction with `accessControlAllowOriginList` to determine whether the `Vary` header should be added or modified to demonstrate that server responses can differ based on the value of the origin header. | false     | No |
| <a id="opt-allowedHosts" href="#opt-allowedHosts" title="#opt-allowedHosts">`allowedHosts`</a> | Lists allowed domain names.                      | []      | No |
| <a id="opt-hostsProxyHeaders" href="#opt-hostsProxyHeaders" title="#opt-hostsProxyHeaders">`hostsProxyHeaders`</a> | Specifies header keys for proxied hostname.      | []      | No |
| <a id="opt-sslProxyHeaders" href="#opt-sslProxyHeaders" title="#opt-sslProxyHeaders">`sslProxyHeaders`</a> | Defines a set of header keys with associated values that would indicate a valid HTTPS request. It can be useful when using other proxies (example: `"X-Forwarded-Proto": "https"`).        |   {}   | No |
| <a id="opt-stsSeconds" href="#opt-stsSeconds" title="#opt-stsSeconds">`stsSeconds`</a> | Max age for `Strict-Transport-Security` header.    | 0         | No |
| <a id="opt-stsIncludeSubdomains" href="#opt-stsIncludeSubdomains" title="#opt-stsIncludeSubdomains">`stsIncludeSubdomains`</a> | If set to `true`, the `includeSubDomains` directive is appended to the `Strict-Transport-Security` header.    | false     | No |
| <a id="opt-stsPreload" href="#opt-stsPreload" title="#opt-stsPreload">`stsPreload`</a> | Adds preload flag to STS header.                 | false     | No |
| <a id="opt-forceSTSHeader" href="#opt-forceSTSHeader" title="#opt-forceSTSHeader">`forceSTSHeader`</a> | Adds STS header for HTTP connections.            | false     | No |
| <a id="opt-frameDeny" href="#opt-frameDeny" title="#opt-frameDeny">`frameDeny`</a> | Set `frameDeny` to `true` to add the `X-Frame-Options` header with the value of `DENY`.                | false     | No |
| <a id="opt-customFrameOptionsValue" href="#opt-customFrameOptionsValue" title="#opt-customFrameOptionsValue">`customFrameOptionsValue`</a> | allows the `X-Frame-Options` header value to be set with a custom value. This overrides the `FrameDeny` option.  |    ""  | No |
| <a id="opt-contentTypeNosniff" href="#opt-contentTypeNosniff" title="#opt-contentTypeNosniff">`contentTypeNosniff`</a> | Set `contentTypeNosniff` to true to add the `X-Content-Type-Options` header with the value `nosniff`.  | false     | No |
| <a id="opt-browserXssFilter" href="#opt-browserXssFilter" title="#opt-browserXssFilter">`browserXssFilter`</a> | Set `browserXssFilter` to true to add the `X-XSS-Protection` header with the value `1; mode=block`.  | false     | No |
| <a id="opt-customBrowserXSSValue" href="#opt-customBrowserXSSValue" title="#opt-customBrowserXSSValue">`customBrowserXSSValue`</a> | allows the `X-XSS-Protection` header value to be set with a custom value. This overrides the `BrowserXssFilter` option.   | false | No |
| <a id="opt-contentSecurityPolicy" href="#opt-contentSecurityPolicy" title="#opt-contentSecurityPolicy">`contentSecurityPolicy`</a> | allows the `Content-Security-Policy` header value to be set with a custom value.           | false | No |
| <a id="opt-contentSecurityPolicyReportOnly" href="#opt-contentSecurityPolicyReportOnly" title="#opt-contentSecurityPolicyReportOnly">`contentSecurityPolicyReportOnly`</a> | allows the `Content-Security-Policy-Report-Only` header value to be set with a custom value.    |   ""  | No |
| <a id="opt-publicKey" href="#opt-publicKey" title="#opt-publicKey">`publicKey`</a> | Implements HPKP for certificate pinning.         |  "" | No |
| <a id="opt-referrerPolicy" href="#opt-referrerPolicy" title="#opt-referrerPolicy">`referrerPolicy`</a> | Controls forwarding of `Referer` header.         | "" | No |
| <a id="opt-permissionsPolicy" href="#opt-permissionsPolicy" title="#opt-permissionsPolicy">`permissionsPolicy`</a> | allows sites to control browser features.                   | ""      | No |
| <a id="opt-isDevelopment" href="#opt-isDevelopment" title="#opt-isDevelopment">`isDevelopment`</a> | Set `true` when developing to mitigate the unwanted effects of the `AllowedHosts`, SSL, and STS options. Usually testing takes place using HTTP, not HTTPS, and on `localhost`, not your production domain.    | false     | No |

### `accessControlAllowOriginList`

The `accessControlAllowOriginList` indicates whether a resource can be shared by returning different values.

A wildcard origin `*` can also be configured, and matches all requests.
If this value is set by a backend service, it will be overwritten by Traefik.

This value can contain a list of allowed origins.

More information including how to use the settings can be found at:

- [Mozilla.org](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin)
- [w3](https://fetch.spec.whatwg.org/#http-access-control-allow-origin)
- [IETF](https://tools.ietf.org/html/rfc6454#section-7.1)

Traefik no longer supports the `null` value, as it is [no longer recommended as a return value](https://w3c.github.io/webappsec-cors-for-developers/#avoid-returning-access-control-allow-origin-null).

### `accessControlAllowOriginListRegex`

The `accessControlAllowOriginListRegex` option is the counterpart of the `accessControlAllowOriginList` option with regular expressions instead of origin values.
It allows all origins that contain any match of a regular expression in the `accessControlAllowOriginList`.

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

    When defining a regular expression within YAML, any escaped character needs to be escaped twice: `example\.com` needs to be written as `example\\.com`.

{% include-markdown "includes/traefik-for-business-applications.md" %}
