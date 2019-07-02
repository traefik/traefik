# ForwardAuth

Using an External Service to Check for Credentials
{: .subtitle }

![AuthForward](../assets/img/middleware/authforward.png)

The ForwardAuth middleware delegate the authentication to an external service.
If the service response code is 2XX, access is granted and the original request is performed.
Otherwise, the response from the authentication server is returned.

## Configuration Examples

```yaml tab="Docker"
# Forward authentication to authserver.com
labels:
- "traefik.http.middlewares.test-auth.forwardauth.address=https://authserver.com/auth"
- "traefik.http.middlewares.test-auth.forwardauth.authResponseHeaders=X-Auth-User, X-Secret"
- "traefik.http.middlewares.test-auth.forwardauth.tls.ca=path/to/local.crt"
- "traefik.http.middlewares.test-auth.forwardauth.tls.caOptional=true"
- "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-auth.forwardauth.tls.insecureSkipVerify=true"
- "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
- "traefik.http.middlewares.test-auth.forwardauth.trustForwardHeader=true"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  forwardAuth:
    address: https://authserver.com/auth
    trustForwardHeader: true
    authResponseHeaders:
    - X-Auth-User
    - X-Secret
    tls:
      ca: path/to/local.crt
      caOptional: true
      cert: path/to/foo.cert
      key: path/to/foo.key  
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.test-auth.forwardauth.address": "https://authserver.com/auth",
  "traefik.http.middlewares.test-auth.forwardauth.authResponseHeaders": "X-Auth-User,X-Secret",
  "traefik.http.middlewares.test-auth.forwardauth.tls.ca": "path/to/local.crt",
  "traefik.http.middlewares.test-auth.forwardauth.tls.caOptional": "true",
  "traefik.http.middlewares.test-auth.forwardauth.tls.cert": "path/to/foo.cert",
  "traefik.http.middlewares.test-auth.forwardauth.tls.insecureSkipVerify": "true",
  "traefik.http.middlewares.test-auth.forwardauth.tls.key": "path/to/foo.key",
  "traefik.http.middlewares.test-auth.forwardauth.trustForwardHeader": "true"
}
```

```yaml tab="Rancher"
# Forward authentication to authserver.com
labels:
- "traefik.http.middlewares.test-auth.forwardauth.address=https://authserver.com/auth"
- "traefik.http.middlewares.test-auth.forwardauth.authResponseHeaders=X-Auth-User, X-Secret"
- "traefik.http.middlewares.test-auth.forwardauth.tls.ca=path/to/local.crt"
- "traefik.http.middlewares.test-auth.forwardauth.tls.caOptional=true"
- "traefik.http.middlewares.test-auth.forwardauth.tls.cert=path/to/foo.cert"
- "traefik.http.middlewares.test-auth.forwardauth.tls.InisecureSkipVerify=true"
- "traefik.http.middlewares.test-auth.forwardauth.tls.key=path/to/foo.key"
- "traefik.http.middlewares.test-auth.forwardauth.trustForwardHeader=true"
```

```toml tab="File"
# Forward authentication to authserver.com
[http.middlewares]
  [http.middlewares.test-auth.forwardAuth]
    address = "https://authserver.com/auth"
    trustForwardHeader = true
    authResponseHeaders = ["X-Auth-User", "X-Secret"]

    [http.middlewares.test-auth.forwardAuth.tls]
      ca = "path/to/local.crt"
      caOptional = true
      cert = "path/to/foo.cert"
      key = "path/to/foo.key"
```

## Configuration Options

### `address`

The `address` option defines the authentication server address.

### `trustForwardHeader`

Set the `trustForwardHeader` option to `true` to trust all the existing `X-Forwarded-*` headers.

### `authResponseHeaders`

The `authResponseHeaders` option is the list of the headers to copy from the authentication server to the request.

### `tls`

The `tls` option is the TLS configuration from Traefik to the authentication server.
