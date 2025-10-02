---
title: "Traefik Headers Documentation"
description: "In Traefik Proxy, the HTTP headers middleware manages the headers of requests and responses. Read the technical documentation."
---

The Headers middleware manages the headers of requests and responses.

By default, the following headers are automatically added when proxying requests:

| Property                  | HTTP Header                |
|---------------------------|----------------------------|
| <a id="Clients-IP" href="#Clients-IP" title="#Clients-IP">Client's IP</a> | `X-Forwarded-For`, `X-Real-Ip` |
| <a id="Host" href="#Host" title="#Host">Host</a> | `X-Forwarded-Host`           |
| <a id="Port" href="#Port" title="#Port">Port</a> | `X-Forwarded-Port`           |
| <a id="Protocol" href="#Protocol" title="#Protocol">Protocol</a> | `X-Forwarded-Proto`          |
| <a id="Proxy-Servers-Hostname" href="#Proxy-Servers-Hostname" title="#Proxy-Servers-Hostname">Proxy Server's Hostname</a> | `X-Forwarded-Server`         |

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
| <a id="customRequestHeaders" href="#customRequestHeaders" title="#customRequestHeaders">`customRequestHeaders`</a> | Lists the header names and values for requests.  | [] | No |
| <a id="customResponseHeaders" href="#customResponseHeaders" title="#customResponseHeaders">`customResponseHeaders`</a> | Lists the header names and values for responses. | [] | No |
| <a id="accessControlAllowCredentials" href="#accessControlAllowCredentials" title="#accessControlAllowCredentials">`accessControlAllowCredentials`</a> | Indicates if the request can include user credentials.| false     | No |
| <a id="accessControlAllowHeaders" href="#accessControlAllowHeaders" title="#accessControlAllowHeaders">`accessControlAllowHeaders`</a> | Specifies allowed request header names.          | [] | No |
| <a id="accessControlAllowMethods" href="#accessControlAllowMethods" title="#accessControlAllowMethods">`accessControlAllowMethods`</a> | Specifies allowed request methods.               | [] | No |
| <a id="accessControlAllowOriginList" href="#accessControlAllowOriginList" title="#accessControlAllowOriginList">`accessControlAllowOriginList`</a> | Specifies allowed origins. More information [here](#accesscontrolalloworiginlist)      | []      | No |
| <a id="accessControlAllowOriginListRegex" href="#accessControlAllowOriginListRegex" title="#accessControlAllowOriginListRegex">`accessControlAllowOriginListRegex`</a> | Allows origins matching regex. More information [here](#accesscontrolalloworiginlistregex)            | []      | No |
| <a id="accessControlExposeHeaders" href="#accessControlExposeHeaders" title="#accessControlExposeHeaders">`accessControlExposeHeaders`</a> | Specifies which headers are safe to expose to the API of a CORS API specification.       |  []    | No |
| <a id="accessControlMaxAge" href="#accessControlMaxAge" title="#accessControlMaxAge">`accessControlMaxAge`</a> | Time (in seconds) to cache preflight requests.   | 0         | No |
| <a id="addVaryHeader" href="#addVaryHeader" title="#addVaryHeader">`addVaryHeader`</a> | Used in conjunction with `accessControlAllowOriginList` to determine whether the `Vary` header should be added or modified to demonstrate that server responses can differ based on the value of the origin header. | false     | No |
| <a id="allowedHosts" href="#allowedHosts" title="#allowedHosts">`allowedHosts`</a> | Lists allowed domain names.                      | []      | No |
| <a id="hostsProxyHeaders" href="#hostsProxyHeaders" title="#hostsProxyHeaders">`hostsProxyHeaders`</a> | Specifies header keys for proxied hostname.      | []      | No |
| <a id="sslProxyHeaders" href="#sslProxyHeaders" title="#sslProxyHeaders">`sslProxyHeaders`</a> | Defines a set of header keys with associated values that would indicate a valid HTTPS request. It can be useful when using other proxies (example: `"X-Forwarded-Proto": "https"`).        |   {}   | No |
| <a id="stsSeconds" href="#stsSeconds" title="#stsSeconds">`stsSeconds`</a> | Max age for `Strict-Transport-Security` header.    | 0         | No |
| <a id="stsIncludeSubdomains" href="#stsIncludeSubdomains" title="#stsIncludeSubdomains">`stsIncludeSubdomains`</a> | If set to `true`, the `includeSubDomains` directive is appended to the `Strict-Transport-Security` header.    | false     | No |
| <a id="stsPreload" href="#stsPreload" title="#stsPreload">`stsPreload`</a> | Adds preload flag to STS header.                 | false     | No |
| <a id="forceSTSHeader" href="#forceSTSHeader" title="#forceSTSHeader">`forceSTSHeader`</a> | Adds STS header for HTTP connections.            | false     | No |
| <a id="frameDeny" href="#frameDeny" title="#frameDeny">`frameDeny`</a> | Set `frameDeny` to `true` to add the `X-Frame-Options` header with the value of `DENY`.                | false     | No |
| <a id="customFrameOptionsValue" href="#customFrameOptionsValue" title="#customFrameOptionsValue">`customFrameOptionsValue`</a> | allows the `X-Frame-Options` header value to be set with a custom value. This overrides the `FrameDeny` option.  |    ""  | No |
| <a id="contentTypeNosniff" href="#contentTypeNosniff" title="#contentTypeNosniff">`contentTypeNosniff`</a> | Set `contentTypeNosniff` to true to add the `X-Content-Type-Options` header with the value `nosniff`.  | false     | No |
| <a id="browserXssFilter" href="#browserXssFilter" title="#browserXssFilter">`browserXssFilter`</a> | Set `browserXssFilter` to true to add the `X-XSS-Protection` header with the value `1; mode=block`.  | false     | No |
| <a id="customBrowserXSSValue" href="#customBrowserXSSValue" title="#customBrowserXSSValue">`customBrowserXSSValue`</a> | allows the `X-XSS-Protection` header value to be set with a custom value. This overrides the `BrowserXssFilter` option.   | false | No |
| <a id="contentSecurityPolicy" href="#contentSecurityPolicy" title="#contentSecurityPolicy">`contentSecurityPolicy`</a> | allows the `Content-Security-Policy` header value to be set with a custom value.           | false | No |
| <a id="contentSecurityPolicyReportOnly" href="#contentSecurityPolicyReportOnly" title="#contentSecurityPolicyReportOnly">`contentSecurityPolicyReportOnly`</a> | allows the `Content-Security-Policy-Report-Only` header value to be set with a custom value.    |   ""  | No |
| <a id="publicKey" href="#publicKey" title="#publicKey">`publicKey`</a> | Implements HPKP for certificate pinning.         |  "" | No |
| <a id="referrerPolicy" href="#referrerPolicy" title="#referrerPolicy">`referrerPolicy`</a> | Controls forwarding of `Referer` header.         | "" | No |
| <a id="permissionsPolicy" href="#permissionsPolicy" title="#permissionsPolicy">`permissionsPolicy`</a> | allows sites to control browser features.                   | ""      | No |
| <a id="isDevelopment" href="#isDevelopment" title="#isDevelopment">`isDevelopment`</a> | Set `true` when developing to mitigate the unwanted effects of the `AllowedHosts`, SSL, and STS options. Usually testing takes place using HTTP, not HTTPS, and on `localhost`, not your production domain.    | false     | No |

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

{!traefik-for-business-applications.md!}
