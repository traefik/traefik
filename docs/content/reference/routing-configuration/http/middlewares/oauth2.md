---
title: 'OAuth 2.0 Token Introspection Authentication'
description: 'Traefik Hub API Gateway - OAuth 2.0 Token Introspection allows to retrieve metadata about an access token from an OAuth 2.0 server.'
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

OAuth 2.0 Token Introspection allows Traefik Hub to retrieve metadata about an access token from an OAuth 2.0 server with the Token Introspection extension.

The metadata can be used to restrict the access to applications. For more information please refer to the [RFC](https://tools.ietf.org/html/rfc7662).

---

## Configuration Example

```yaml tab="Middleware OAuth Token Introspection"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-oauth-intro
spec:
  plugin:
    oAuthIntrospection:
      tokenSource:
        header: Authorization
        headerAuthScheme: Bearer
      clientConfig:
        url: "https://YOUR-KEYCLOAK-ADDRESS/realms/YOUR-REALM/protocol/openid-connect/token/introspect"
        headers:
          Authorization: Basic ZXhhbXBsZTpleGFtcGxl # echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64
      tokenTypeHint: access_token
      forwardHeaders:
        Group: grp
        Expires-At: exp
      claims: Equals(`grp`, `admin`)
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| `claims` | Defines the claims to validate in order to authorize the request. <br /> The `claims` option can only be used with JWT-formatted token.  (More information [here](#claims)) | ""      | No       |
| `clientConfig.url` | Defines the introspection endpoint URL. It must include the scheme and path. | ""      | Yes      |
| `clientConfig.headers` | Defines the headers to send in every introspection request. Values can be plain strings or a valid [Go template](https://pkg.go.dev/text/template). <br /> Currently, a variable of type [`Request`](https://pkg.go.dev/net/http#Request) corresponding to the request being introspected is accessible in templates. | ""      | No       |
| `clientConfig.tokenTypeHint` | Defines the type of token being introspected, sent as a hint to the introspection server. <br /> Please refer to the [official documentation](https://tools.ietf.org/html/rfc7662) for more details. | ""      | No       |
| `clientConfig.tls.ca` | PEM-encoded certificate bundle or a URN referencing a secret containing the certificate bundle used to establish a TLS connection with the authorization server  (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.cert` | PEM-encoded certificate or a URN referencing a secret containing the certificate used to establish a TLS connection with the Vault server (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.key`  | PEM-encoded key or a URN referencing a secret containing the key used to establish a TLS connection with the Vault server. (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.insecureSkipVerify` | Disables TLS certificate verification when communicating with the authorization server. <br /> Useful for testing purposes but strongly discouraged for production. (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.timeoutSeconds` | Defines the time before giving up requests to the authorization server.   | 5       | No       |
| `clientConfig.maxRetries` | Defines the number of retries for requests to authorization server that fail. | 3       | No       |
| `forwardAuthorization` | Defines whether the authorization header will be forwarded or stripped from a request after it has been approved by the middleware. | false   | No       |
| `forwardHeaders` | Defines the HTTP headers to add to requests and populates them with values extracted from the access token claims returned by the authorization server. <br /> Claims to be forwarded that are not found in the JWT result in empty headers. <br /> The `forwardHeaders` option can only be used with JWT-formatted token. | []      | No       |
| `tokenSource.header`           | Defines the header name containing the secret sent by the client.<br />At least one `tokenSource`option must be set.| ""      | No       |
| `tokenSource.headerAuthScheme` | Defines the scheme when using `Authorization` as header name. <br /> Check out the `Authorization` header [documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization#syntax).<br />At least one `tokenSource`option must be set. | ""      | No       |
| `tokenSource.query`            | Defines the query parameter name containing the secret sent by the client.<br />At least one `tokenSource`option must be set.| ""      | No       |
| `tokenSource.cookie`           | Defines the cookie name containing the secret sent by the client.<br />At least one `tokenSource`option must be set.| ""      | No       |
| `usernameClaim` | Defines the claim that will be evaluated to populate the `clientusername` in the access logs. <br /> The `usernameClaim` option can only be used with JWT-formatted token.| ""      | No       |

### claims

#### Syntax

The following functions are supported in `claims`:

| Function          | Description        | Example        |
|-------------------|--------------------|-----------------|
| Equals            | Validates the equality of the value in `key` with `value`.                     | Equals(\`grp\`, \`admin\`)                   |
| Prefix            | Validates the value in `key` has the prefix of `value`.                        | Prefix(\`referrer\`, \`http://example.com\`) |
| Contains (string) | Validates the value in `key` contains `value`.                                 | Contains(\`referrer\`, \`/foo/\`)            |
| Contains (array)  | Validates the `key` array contains the `value`.                                | Contains(\`areas\`, \`home\`)                |
| SplitContains     | Validates the value in `key` contains the `value` once split by the separator. | SplitContains(\`scope\`, \` \`, \`writer\`)  |
| OneOf             | Validates the `key` array contains one of the `values`.                        | OneOf(\`areas\`, \`office\`, \`lab\`)        |

All functions can be joined by boolean operands. The supported operands are:

| Operand | Description        | Example        |
|---------|--------------------|-----------------|
| &&      | Compares two functions and returns true only if both evaluate to true. | Equals(\`grp\`, \`admin\`) && Equals(\`active\`, \`true\`)   |
| \|\|    | Compares two functions and returns true if either evaluate to true.    | Equals(\`grp\`, \`admin\`) \|\| Equals(\`active\`, \`true\`) |
| !       | Returns false if the function is true, otherwise returns true.         | !Equals(\`grp\`, \`testers\`)                                |

All examples will return true for the following data structure:

```json tab="JSON"
{
  "active": true,
  "grp": "admin",
  "scope": "reader writer deploy",
  "referrer": "http://example.com/foo/bar",
  "areas": [
    "office",
    "home"
  ]
}
```

#### Nested Claims

Nested claims are supported by using a `.` between keys. For example:

```bash tab="Key"
user.name
```

```json  tab="Claims"
{
  "active": true,
  "grp": "admin",
  "scope": "reader writer deploy",
  "referrer": "http://example.com/foo/bar",
  "areas": [
    "office",
    "home"
  ],
  "user" {
    "name": "John Snow",
    "status": "undead"
  }
}
```

```bash tab="Result"
John Snow
```

!!! note "Handling keys that contain a '.'"

If the `key` contains a dot, the dot can be escaped using `\.`

!!! note "Handling a key that contains a '\'"

If the `key` contains a `\`, it needs to be doubled `\\`.

### clientConfig

Defines the configuration used to connect the API Gateway to a Third Party Software such as an Identity Provider.

#### `clientConfig.tls`

##### Storing secret values in Kubernetes secrets

When configuring the `tls.ca`, `tls.cert`, `tls.key`, it is possible to reference Kubernetes secrets defined in the same namespace as the Middleware.  
The reference to a Kubernetes secret takes the form of a URN:

```text
urn:k8s:secret:[name]:[valueKey]
```

```yaml tab="Middleware JWT"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-oauth-intro
spec:
  plugin:
    oAuthIntrospection:
      clientConfig:
        tls:
          ca: "urn:k8s:secret:tls:ca"
          cert: "urn:k8s:secret:tls:cert"
          key: "urn:k8s:secret:tls:key"
          insecureSkipVerify: true
```

```yaml tab="Kubernetes TLS Secret"
apiVersion: v1
kind: Secret
metadata:
  name: tls
stringData:
  ca: |-
    -----BEGIN CERTIFICATE-----
    MIIB9TCCAWACAQAwgbgxGTAXBgNVBAoMEFF1b1ZhZGlzIExpbWl0ZWQxHDAaBgNV
    BAsME0RvY3VtZW50IERlcGFydG1lbnQxOTA3BgNVBAMMMFdoeSBhcmUgeW91IGRl
    Y29kaW5nIG1lPyAgVGhpcyBpcyBvbmx5IGEgdGVzdCEhITERMA8GA1UEBwwISGFt
    aWx0b24xETAPBgNVBAgMCFBlbWJyb2tlMQswCQYDVQQGEwJCTTEPMA0GCSqGSIb3
    DQEJARYAMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCJ9WRanG/fUvcfKiGl
    EL4aRLjGt537mZ28UU9/3eiJeJznNSOuNLnF+hmabAu7H0LT4K7EdqfF+XUZW/2j
    RKRYcvOUDGF9A7OjW7UfKk1In3+6QDCi7X34RE161jqoaJjrm/T18TOKcgkkhRzE
    apQnIDm0Ea/HVzX/PiSOGuertwIDAQABMAsGCSqGSIb3DQEBBQOBgQBzMJdAV4QP
    Awel8LzGx5uMOshezF/KfP67wJ93UW+N7zXY6AwPgoLj4Kjw+WtU684JL8Dtr9FX
    ozakE+8p06BpxegR4BR3FMHf6p+0jQxUEAkAyb/mVgm66TyghDGC6/YkiKoZptXQ
    98TwDIK/39WEB/V607As+KoYazQG8drorw==
    -----END CERTIFICATE-----
  cert: |-
    -----BEGIN CERTIFICATE-----
    MIIB9TCCAWACAQAwgbgxGTAXBgNVBAoMEFF1b1ZhZGlzIExpbWl0ZWQxHDAaBgNV
    BAsME0RvY3VtZW50IERlcGFydG1lbnQxOTA3BgNVBAMMMFdoeSBhcmUgeW91IGRl
    Y29kaW5nIG1lPyAgVGhpcyBpcyBvbmx5IGEgdGVzdCEhITERMA8GA1UEBwwISGFt
    aWx0b24xETAPBgNVBAgMCFBlbWJyb2tlMQswCQYDVQQGEwJCTTEPMA0GCSqGSIb3
    DQEJARYAMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCJ9WRanG/fUvcfKiGl
    EL4aRLjGt537mZ28UU9/3eiJeJznNSOuNLnF+hmabAu7H0LT4K7EdqfF+XUZW/2j
    RKRYcvOUDGF9A7OjW7UfKk1In3+6QDCi7X34RE161jqoaJjrm/T18TOKcgkkhRzE
    apQnIDm0Ea/HVzX/PiSOGuertwIDAQABMAsGCSqGSIb3DQEBBQOBgQBzMJdAV4QP
    Awel8LzGx5uMOshezF/KfP67wJ93UW+N7zXY6AwPgoLj4Kjw+WtU684JL8Dtr9FX
    ozakE+8p06BpxegR4BR3FMHf6p+0jQxUEAkAyb/mVgm66TyghDGC6/YkiKoZptXQ
    98TwDIK/39WEB/V607As+KoYazQG8drorw==
    -----END CERTIFICATE-----
  key: |-
    -----BEGIN EC PRIVATE KEY-----
    MHcCAQEEIC8CsJ/B115S+JtR1/l3ZQwKA3XdXt9zLqusF1VXc/KloAoGCCqGSM49
    AwEHoUQDQgAEpwUmRIZHFt8CdDHYm1ikScCScd2q6QVYXxJu+G3fQZ78ScGtN7fu
    KXMnQqVjXVRAr8qUY8yipVKuMCepnPXScQ==
    -----END EC PRIVATE KEY-----
```

{!traefik-for-business-applications.md!}
