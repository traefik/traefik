
# Examples

You will find here some configuration examples of Træfɪk.

## HTTP only

```
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
```

## HTTP + HTTPS (with SNI)

```
defaultEntryPoints = ["http", "https"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"
```
Note that we can either give path to certificate file or directly the file content itself ([like in this TOML example](/user-guide/kv-config/#upload-the-configuration-in-the-key-value-store)).

## HTTP redirect on HTTPS

```
defaultEntryPoints = ["http", "https"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "tests/traefik.crt"
      keyFile = "tests/traefik.key"
```

## Let's Encrypt support

```
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      # certs used as default certs
      [[entryPoints.https.tls.certificates]]
      certFile = "tests/traefik.crt"
      keyFile = "tests/traefik.key"
[acme]
email = "test@traefik.io"
storageFile = "acme.json"
onDemand = true
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"

[[acme.domains]]
  main = "local1.com"
  sans = ["test1.local1.com", "test2.local1.com"]
[[acme.domains]]
  main = "local2.com"
  sans = ["test1.local2.com", "test2x.local2.com"]
[[acme.domains]]
  main = "local3.com"
[[acme.domains]]
  main = "local4.com"
```

## Override entrypoints in frontends

```
[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
    rule = "Path:/test"
```

## Enable Basic authentication in an entrypoint

With two user/pass:

- `test`:`test`
- `test2`:`test2`

Passwords are encoded in MD5: you can use htpasswd to generate those ones.

```
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
```

## Pass Authenticated user to application via headers

Providing an authentication method as described above, it is possible to pass the user to the application
via a configurable header value

```
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth]
    headerField = "X-WebAuth-User"
    [entryPoints.http.auth.basic]
    users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
```


## Enable authentication forwarding

Authentication forwarding is a mechanism to allow for arbitrary authentication back ends to be used.
It works by forwarding any incoming request from traefik proxy to a specified endpoint (the setting key `entryPoints.http.auth.forward.address`).
This endpoint address is called "authentication back end".
The response of the authentication back end is then evaluated.
The authentication back end has to send a response with an HTTP status code 200 to allow access.
Any other status code will result in traefik returning the authentiucation back end response to the original request (so if your back end sends a response with a 503 header and any body, traefik will also send a 503 response including the body).

In addition to this, traefik provides basic functionality to modify the original request's query parameters before they are sent to the authentication back end.
This allows you to map a request parameter or header (in this case token) from e.g. `traefik.com/secret?token=foobar` to `authserver.com/auth?theToken=foobar`.
Use the keywords name and as in the `entryPoints.http.auth.forward.requestParameters` setting key as shown in the example.
As well you can set `entryPoints.http.auth.forward.forwardAllHeaders = true` to forward all request headers to the authentication server as they ware received by traefik.

Moreover, you can also replay certain information from the back end authentication response back to the original request received by traefik.
For this to work, your authentication back end must send a JSON response.
```
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.forward]
    address = "http://authserver.com/auth"
    forwardAllHeaders = true
    [entryPoints.http.auth.Forward.requestParameters.email]
      name = "email"
      as = "theEmail"
      in = "parameter"
    [entryPoints.http.auth.Forward.requestParameters.token]
      name = "token"
      as = "theToken"
      in = "header"
    [entryPoints.http.auth.Forward.responseReplayFields.userId]
      path = "user.id"
      as = "X-User-Id"
      in = "header"
    [entryPoints.http.auth.Forward.responseReplayFields.userName]
      path = "user.name"
      as = "" # No name transformation
      in = "parameter"
```
