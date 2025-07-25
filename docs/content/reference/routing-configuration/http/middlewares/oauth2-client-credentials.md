---
title: 'OAuth 2.0 Client Credentials Authentication'
description: 'Traefik Hub API Gateway - The OAuth 2.0 Client Credentials Authentication middleware secures your applications using the client credentials flow'
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The OAuth 2.0 Client Credentials Authentication middleware allows Traefik Hub to secure routes using the OAuth 2.0 Client Credentials flow as described in the [RFC 6749](https://www.rfc-editor.org/rfc/rfc6749.html#section-4.4).
Access tokens can be cached using an external KV store.

The OAuth Client Credentials Authentication middleware allows using Redis (or Sentinel) as persistent KV store to authorization access tokens
while they are valid. This reduces latency and the number of calls made to the authorization server.

---

## Configuration Example

```yaml tab="Middleware OAuth Client Credentials"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-client-creds
spec:
  plugin:
    oAuthClientCredentials:
      url: https://tenant.auth0.com/oauth/token
      clientID: urn:k8s:my-secret:my-secret:clientID
      clientSecret: urn:k8s:my-secret:my-secret:clientSecret
      audience: https://api.example.com
      forwardHeaders:
        Group: grp
        Expires-At: exp
      claims: Equals(`grp`, `admin`)
```

```yaml tab="Kubernetes Secret"
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-secret
stringData:
  clientID: my-oauth-client-name
  clientSecret: mypasswd
```

## Configuration Options

| Field | Description                                                                                 | Default | Required |
|:------|:--------------------------------------------------------------------------------------------|:--------|:---------|
| `audience` | Defines the audience configured in your authorization server. <br /> The audience value is the base address of the resource being accessed, for example: https://api.example.com. | ""      | Yes      |
| `claims` | Defines the claims to validate in order to authorize the request. <br /> The `claims` option can only be used with JWT-formatted token.  (More information [here](#claims)) | ""      | No       |
| `clientConfig.tls.ca` | PEM-encoded certificate bundle or a URN referencing a secret containing the certificate bundle used to establish a TLS connection with the authorization server  (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.cert` | PEM-encoded certificate or a URN referencing a secret containing the certificate used to establish a TLS connection with the Vault server (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.key`  | PEM-encoded key or a URN referencing a secret containing the key used to establish a TLS connection with the Vault server. (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.tls.insecureSkipVerify` | Disables TLS certificate verification when communicating with the authorization server. <br /> Useful for testing purposes but strongly discouraged for production. (More information [here](#clientconfig)) | ""      | No       |
| `clientConfig.timeoutSeconds` | Defines the time before giving up requests to the authorization server.   | 5       | No       |
| `clientConfig.maxRetries` | Defines the number of retries for requests to authorization server that fail. | 3       | No       |
| `clientID`     | Defines the unique client identifier for an account on the OpenID Connect provider, must be set when the `clientSecret` option is set.<br />More information [here](#storing-secret-values-in-kubernetes-secrets). | ""      | Yes       |
| `clientSecret` | Defines the unique client secret for an account on the OpenID Connect provider, must be set when the `clientID` option is set.<br />More information [here](#storing-secret-values-in-kubernetes-secrets). | ""      | Yes       |
| `forwardHeaders` | Defines the HTTP headers to add to requests and populates them with values extracted from the access token claims returned by the authorization server. <br /> Claims to be forwarded that are not found in the JWT result in empty headers. <br /> The `forwardHeaders` option can only be used with JWT-formatted token. | []      | No       |
| `store.keyPrefix` | Defines the prefix of the key for the entries that store the sessions. | ""      | No       |
| `store.redis.endpoints`              | Endpoints of the Redis instances to connect to (example: `redis.traefik-hub.svc.cluster.local:6379`) | "" | Yes      |
| `store.redis.username`               | The username Traefik Hub will use to connect to Redis          | "" | No       |
| `store.redis.password`               | The password Traefik Hub will use to connect to Redis    | "" | No       |
| `store.redis.database`               | The database Traefik Hub will use to sore information (default: `0`)                                 | "" | No       |
| `store.redis.cluster`                | Enable Redis Cluster           | "" | No       |
| `store.redis.tls.caBundle`           | Custom CA bundle    | "" | No       |
| `store.redis.tls.cert`               | TLS certificate         | "" | No       |
| `store.redis.tls.key`                | TLS   | "" | No       |
| `store.redis.tls.insecureSkipVerify` | Allow skipping the TLS verification | "" | No       |
| `store.redis.sentinel.masterSet`     | Name of the set of main nodes to use for main selection. Required when using Sentinel.               | "" | No       |
| `store.redis.sentinel.username`      | Username to use for sentinel authentication (can be different from `username`)  | "" | No       |
| `store.redis.sentinel.password`      | Password to use for sentinel authentication (can be different from `password`)        | "" | No       |
| `url` | Defines the authorization server URL (for example: `https://tenant.auth0.com/oauth/token`). | ""      | Yes      |
| `usernameClaim` | Defines the claim that will be evaluated to populate the `clientusername` in the access logs. <br /> The `usernameClaim` option can only be used with JWT-formatted token.| ""      | No       |

### Storing secret values in Kubernetes secrets

It is possible to reference Kubernetes secrets defined in the same namespace as the Middleware.
The reference to a Kubernetes secret takes the form of a URN:

```text
urn:k8s:secret:[name]:[valueKey]
```

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
  name: test-client-creds
spec:
  plugin:
    oAuthClientCredentials:
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

### store.redis

Connection parameters to your [Redis](https://redis.io/ "Link to website of Redis") server are attached to your Middleware deployment.

The following Redis modes are supported:

- Single instance mode
- [Redis Cluster](https://redis.io/docs/management/scaling "Link to official Redis documentation about Redis Cluster mode")
- [Redis Sentinel](https://redis.io/docs/management/sentinel "Link to official Redis documentation about Redis Sentinel mode")

!!! info

    If you use Redis in single instance mode or Redis Sentinel, you can configure the `database` field.
    This value won't be taken into account if you use Redis Cluster (only database `0` is available).

    In this case, a warning is displayed, and the value is ignored.

For more information about Redis, we recommend the [official Redis documentation](https://redis.io/docs/ "Link to official Redis documentation").

{!traefik-for-business-applications.md!}
