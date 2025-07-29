---
title: "HMAC"
description: "Traefik Hub API Gateway - The HMAC Middleware allows you secure your APIs using the HMAC mechanism."
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

This middleware validates a digital signature computed using the content of an HTTP request and a shared secret that is
sent to the proxy using the `Authorization` or `Proxy-Authorization` header.

It ensures:

- **The identity of the sender (Authentication)**: If the signature is validated by the proxy, it means that the sender
actually owns the shared secret. As a consequence, the sender's identity is considered to be proven.
- **The integrity of the request**: As the signature is based on a subset of the HTTP request, it means that if the
signature is validated by the proxy, the request used to generate the signature has not been modified between the sender
and the proxy. This middleware also allows validating the content integrity using the
[Digest header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Digest).

This middleware is based on the [HTTP Signature Draft](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12).

---

## Configuration Example

Below is an advanced configuration that enables the HMAC middleware, sets one secret, ensures that the digest sum of the
request body is validated and ensures that the given headers must be included in the computation of the signature of the
request.

```yaml tab="Middleware HMAC"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: hmac-auth
spec:
  plugin:
    hmac:
      keys:
        - id: secret-key
          key: secret
      validateDigest: true
      enforcedHeaders:
        - (request-target)
        - (created)
        - (expires)
        - host

```

## Configuration Options

| Field             | Description                | Default | Required |
|:------------------|:---------------------------------------------|:--------|:---------|
| `keys`            | A static set of secret keys to be used by HMAC middleware.    |         | Yes      |
| `validateDigest`  | Determines whether the middleware should validate the digest sum of the request body.      | true    | No  |
| `enforcedHeaders` | A set of headers that must be included in the computation of the signature of the request. |        | No    |

## Authentication Mechanism

The sender and proxy share a common secret identified by a `keyId`. How the sender gets the secret and keeps it safe is out of scope of this documentation.

### Crafting the Authorization Header

To authenticate a request, the sender must provide an `Authorization` or `ProxyAuthorization` header fulfilling the HMAC authentication scheme.

This header carries a set of parameters:

```bash
Authorization: Hmac keyId="secret-id-1",algorithm="hmac-sha256",headers="(request-target) (created) (expires) host x-example",signature="c29tZXNpZ25hdHVyZQ==",created="1584453022",expires="1584453032"
```

| Parameter   | Description  | Example                            |
|-------------|--------------------------------|------------------------------------|
| `keyId`     | Identifier of the key being used by the sender to build the signature     | `keyId="secret-key-1"`             |
| `algorithm` | Algorithm used to generate the signature.<br /> Supported values are `hmac-sha1`, `hmac-sha256`, `hmac-sha384` and `hmac-sha512`. | `algorithm="hmac-sha512"`          |
| `headers`   | List of headers to use in order to build the signature string.<br /> Each item **must** be lowercase.    | `headers="host content-type"`      |
| `signature` | Digital Signature of the request. See [computing the signature](#computing-the-signature).      | `signature="c29tZXNpZ25hdHVyZQ=="` |
| `created`   | Unix timestamp of the signature creation.    | `created="1574453022"`     |
| `expires`   | Unix timestamp of the signature expiration.     | `expires="1574453022"`             |

!!! danger "Time sensitivity"
    If the `created` timestamp is in the future or the `expires` timestamp is in the past, the middleware will refuse the request.
    This behaviour makes using this middleware sensitive to clock skew between the client and the server.

    Make sure that your client and your server have their clocks synchronized.

### Computing the Signature

The signature is the base64-encoded value of the result of an HMAC signature algorithm computed with a `signature string` and the sender's `secret key`.

For example:

```bash
signature=base64(HMAC(signatureString, secret))
```

### Crafting the Signature String

A signed HTTP request needs to be tolerant of some trivial alterations during transmission as it goes through gateways, proxies and other entities.
As a consequence, signing the whole request is not an option as a single header modification could result in a not valid signature.

To avoid this problem, this middleware builds the `signature string` from a subset of header values defined by the sender with the `headers` parameter of the authorization header.

To build the signature string, the client **must** take the values of each header specified by the `headers` parameter **in the order they appear**, then apply the following logic to each of them:

1. If the header is a special header, then evaluate its value according to [the special headers values section](#special-header-values)
2. If the header is not a special header, then append the lowercase header name followed with an ASCII colon `:`, an ASCII space \` \` and the header value.
If the header has multiple values then append those values separated by an ASCII comma `,` and an ASCII space \` \`
3. If value is not the last value then append an ASCII newline `\n`. The signature string MUST NOT include a trailing ASCII newline

!!! warning 
    All headers values are trimmed from their spaces."

#### Special Header Values

By design, all the information of an HTTP request is not available through headers. However, it makes sense to secure the request using them.

To allow this, the `headers` parameter accepts special header names that can be used.

| Value                 | Description     | Signature String Example         |
| --------------------- | ------------------------------------------------------------- |------------------------- |
| `(request-target)`    | Obtained by concatenating the lowercase `:method`, an ASCII space, and the `:path` pseudo-headers ([as specified in HTTP/2](https://tools.ietf.org/html/rfc7540#section-8.1.2.3)).    | `(request-target): get /api/V1/resource?query=foo` |
| `(created)`           | Value of the authorization header `created` parameter.         | `(created): 1584453022`      |
| `(expires)`           | Value of the authorization header `expires` parameter.      | `(expires): 1584453082`      |

Their evaluated value is obtained by appending the special header name with an ASCII colon `:` an ASCII space \` \` then the designated value.

```bash tab="Example"
(created): 1929494939
(request-target): get /foo/bar
```

#### Signature String Example

Here is an example with the authorization header parameters set:

- `headers="(request-target) (created) (expires) host x-example x-emptyheader cache-control"`
- `created="1584466921"`
- `expires="1584466931"`

```bash tab="Request"
GET /foo HTTP/1.1
Host: example.org
X-Example: Example header
    with some whitespace.
X-EmptyHeader:
X-NotIncluded: always
Cache-Control: max-age=60
Cache-Control: must-revalidate
```

```bash tab="Expected Signature String"
(request-target): get /foo
(created): 1584466921
(expires): 1584466931
host: example.org
x-example: Example header with some whitespace.
x-emptyheader:
cache-control: max-age=60, must-revalidate
```

#### Enforced Headers

It is possible to configure the middleware to enforce a minimum set of headers to create the signature string.
This means that any request that does not have the enforced headers in its signature is systematically refused.

This option also configures the headers list returned when [initiating the authentication](#initiating-the-authentication).

It defaults to `(request-target) (created) (expires)`.

!!! danger "Always enforce (created) and (expires)"
    The `created` and `expires` header parameters protect against replay attacks.
    To make sure that their value is not modified during transport, it is **highly recommended** to always include those parameters values in the signature using the `(created)` and `(expired)` special headers value.

    To do so, it is recommended to **always** configure the middleware to enforce `(created)` and `(expires)`.

### Initiating the Authentication

The authentication can be initiated by the proxy. A `401 Unauthorized` response is returned with a `WWW-Authenticate` header indicating to use the `Hmac` authentication scheme.

```bash
WWW-Authenticate: Hmac headers="(request-target) (created) (expires) host x-example"
```

This header indicates that the sender needs to provide an Authorization header that fulfills the `Hmac` authentication schemes.
It also indicates a list of headers that have to be included in the signature using the `headers` parameter.

!!! note "Enforced headers"
    The list of headers carried in the `WWW-Authenticate` header is the list of [enforced headers](#enforced-headers) indicated in the middleware configuration.

## Validating Request Body Integrity

It is possible to make sure that the body of the incoming request has not been altered during transmission by including the [digest header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Digest) in the signature string.

This middleware, by default, validates the digest sum of the body, if there is one.

Only SHA-256 and SHA-512 checksums are supported for checksum computation.

!!! warning "Potential CPU and memory usage"

    Validating the digest makes the middleware read the request body and computes a checksum for it.
    As a consequence it can cause high memory and CPU usage on proxies.

    To disable this feature and only perform authentication, set the `validateDigest` option to `false` in the middleware configuration.

{!traefik-for-business-applications.md!}