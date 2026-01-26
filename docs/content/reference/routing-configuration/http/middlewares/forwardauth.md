---
title: "Traefik ForwardAuth Documentation"
description: "In Traefik Proxy, the HTTP ForwardAuth middleware delegates authentication to an external Service. Read the technical documentation."
---

The `forwardAuth` middleware delegates authentication to an external service.
If the service answers with a 2XX code, access is granted, and the original request is performed.
Otherwise, the response from the authentication server is returned.

## Configuration Example

```yaml tab="Structured (YAML)"
# Forward authentication to example.com
http:
  middlewares:
    test-auth:
      forwardAuth:
        address: "https://example.com/auth"
```

```toml tab="Structured (TOML)"
# Forward authentication to example.com
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://example.com/auth"
```

```yaml tab="Labels"
# Forward authentication to example.com
labels:
  - "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
```

```json tab="Tags"
// Forward authentication to example.com
{
  "Tags" : [
    "traefik.http.middlewares.test-auth.forwardauth.address=https://example.com/auth"
  ]
}
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

## Configuration Options

| Field                                                                                                                                          | Description                                                                                                                                                                                                                                                                 | Default | Required |
|:-----------------------------------------------------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-address" href="#opt-address" title="#opt-address">`address`</a> | Authentication server address.                                                                                                                                                                                                                                              | "" | Yes      |
| <a id="opt-trustForwardHeader" href="#opt-trustForwardHeader" title="#opt-trustForwardHeader">`trustForwardHeader`</a> | Trust all `X-Forwarded-*` headers.                                                                                                                                                                                                                                          | false | No      |
| <a id="opt-authResponseHeaders" href="#opt-authResponseHeaders" title="#opt-authResponseHeaders">`authResponseHeaders`</a> | List of headers to copy from the authentication server response and set on forwarded request, replacing any existing conflicting headers.                                                                                                                                   | [] | No      |
| <a id="opt-authResponseHeadersRegex" href="#opt-authResponseHeadersRegex" title="#opt-authResponseHeadersRegex">`authResponseHeadersRegex`</a> | Regex to match by the headers to copy from the authentication server response and set on forwarded request, after stripping all headers that match the regex.<br /> More information [here](#authresponseheadersregex).                                                     | "" | No      |
| <a id="opt-authRequestHeaders" href="#opt-authRequestHeaders" title="#opt-authRequestHeaders">`authRequestHeaders`</a> | List of the headers to copy from the request to the authentication server. <br /> It allows filtering headers that should not be passed to the authentication server. <br /> If not set or empty, then all request headers are passed.                                      | [] | No      |
| <a id="opt-addAuthCookiesToResponse" href="#opt-addAuthCookiesToResponse" title="#opt-addAuthCookiesToResponse">`addAuthCookiesToResponse`</a> | List of cookies to copy from the authentication server to the response, replacing any existing conflicting cookie from the forwarded response.<br /> Please note that all backend cookies matching the configured list will not be added to the response.                   | [] | No      |
| <a id="opt-forwardBody" href="#opt-forwardBody" title="#opt-forwardBody">`forwardBody`</a> | Sets the `forwardBody` option to `true` to send the Body. As body is read inside Traefik before forwarding, this breaks streaming.                                                                                                                                          | false | No      |
| <a id="opt-maxBodySize" href="#opt-maxBodySize" title="#opt-maxBodySize">`maxBodySize`</a> | Set the `maxBodySize` to limit the body size in bytes. If body is bigger than this, it returns a 401 (unauthorized). If left unset, the request body size is unrestricted which can have performance or security implications. < br/>More information [here](#maxbodysize). | -1 | No      |
| <a id="opt-headerField" href="#opt-headerField" title="#opt-headerField">`headerField`</a> | Defines a header field to store the authenticated user.                                                                                                                                                                                                                     | "" | No      |
| <a id="opt-preserveLocationHeader" href="#opt-preserveLocationHeader" title="#opt-preserveLocationHeader">`preserveLocationHeader`</a> | Defines whether to forward the Location header to the client as is or prefix it with the domain name of the authentication server.                                                                                                                                          | false | No      |
| <a id="opt-preserveRequestMethod" href="#opt-preserveRequestMethod" title="#opt-preserveRequestMethod">`preserveRequestMethod`</a> | Defines whether to preserve the original request method while forwarding the request to the authentication server.                                                                                                                                                          | false | No      |
| <a id="opt-authSigninURL" href="#opt-authSigninURL" title="#opt-authSigninURL">`authSigninURL`</a> | Specifies the URL to redirect to when the authentication server returns 401 Unauthorized.                                                                                                                                                                                   | "" | No      |
| <a id="opt-tls-ca" href="#opt-tls-ca" title="#opt-tls-ca">`tls.ca`</a> | Sets the path to the certificate authority used for the secured connection to the authentication server, it defaults to the system bundle.                                                                                                                                  | "" | No |
| <a id="opt-tls-cert" href="#opt-tls-cert" title="#opt-tls-cert">`tls.cert`</a> | Sets the path to the public certificate used for the secure connection to the authentication server. When using this option, setting the key option is required.                                                                                                            | "" | No |
| <a id="opt-tls-key" href="#opt-tls-key" title="#opt-tls-key">`tls.key`</a> | Sets the path to the private key used for the secure connection to the authentication server. When using this option, setting the `cert` option is required.                                                                                                                | "" | No |
| <a id="opt-tls-caSecret" href="#opt-tls-caSecret" title="#opt-tls-caSecret">`tls.caSecret`</a> | Defines the secret that contains the certificate authority used for the secured connection to the authentication server, it defaults to the system bundle. **This option is only available for the Kubernetes CRD**.                                                        | | No |
| <a id="opt-tls-certSecret" href="#opt-tls-certSecret" title="#opt-tls-certSecret">`tls.certSecret`</a> | Defines the secret that contains both the private and public certificates used for the secure connection to the authentication server. **This option is only available for the Kubernetes CRD**.                                                                            |  | No |
| <a id="opt-tls-insecureSkipVerify" href="#opt-tls-insecureSkipVerify" title="#opt-tls-insecureSkipVerify">`tls.insecureSkipVerify`</a> | During TLS connections, if this option is set to `true`, the authentication server will accept any certificate presented by the server regardless of the host names it covers.                                                                                              | false | No |

### authResponseHeadersRegex

It allows partial matching of the regular expression against the header key.

The start of string (`^`) and end of string (`$`) anchors should be used to ensure a full match against the header key.

Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).

### maxBodySize

The `maxBodySize` option controls the maximum size of request bodies that will be forwarded to the authentication server.

**⚠️ Important Security Consideration**

By default, `maxBodySize` is not set (value: -1), which means request body size is unlimited. This can have significant security and performance implications:

- **Security Risk**: Attackers can send extremely large request bodies, potentially causing DoS attacks or memory exhaustion
- **Performance Impact**: Large request bodies consume memory and processing resources, affecting overall system performance
- **Resource Consumption**: Unlimited body size can lead to unexpected resource usage patterns

**Recommended Configuration**

It is strongly recommended to set an appropriate `maxBodySize` value for your use case:

```yaml
# For most web applications (1MB limit)
maxBodySize: 1048576  # 1MB in bytes

# For API endpoints expecting larger payloads (10MB limit)  
maxBodySize: 10485760  # 10MB in bytes

# For file upload authentication (100MB limit)
maxBodySize: 104857600  # 100MB in bytes
```

**Guidelines for Setting maxBodySize**

- **Web Forms**: 1-5MB is typically sufficient for most form submissions
- **API Endpoints**: Consider your largest expected JSON/XML payload + buffer
- **File Uploads**: Set based on your maximum expected file size
- **High-Traffic Services**: Use smaller limits to prevent resource exhaustion

## Forward-Request Headers

The following request properties are provided to the forward-auth target endpoint as `X-Forwarded-` headers.

| Property          | Forward-Request Header |
|-------------------|------------------------|
| <a id="opt-HTTP-Method" href="#opt-HTTP-Method" title="#opt-HTTP-Method">HTTP Method</a> | `X-Forwarded-Method`     |
| <a id="opt-Protocol" href="#opt-Protocol" title="#opt-Protocol">Protocol</a> | `X-Forwarded-Proto`      |
| <a id="opt-Host" href="#opt-Host" title="#opt-Host">Host</a> | `X-Forwarded-Host`       |
| <a id="opt-Request-URI" href="#opt-Request-URI" title="#opt-Request-URI">Request URI</a> | `X-Forwarded-Uri`        |
| <a id="opt-Source-IP-Address" href="#opt-Source-IP-Address" title="#opt-Source-IP-Address">Source IP-Address</a> | `X-Forwarded-For`        |

{% include-markdown "includes/traefik-for-business-applications.md" %}
