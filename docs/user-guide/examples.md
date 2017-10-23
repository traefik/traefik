# Examples

You will find here some configuration examples of Træfik.

## HTTP only

```toml
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
```

## HTTP + HTTPS (with SNI)

```toml
defaultEntryPoints = ["http", "https"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.com.cert"
      keyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.org.cert"
      keyFile = "integration/fixtures/https/snitest.org.key"
```
Note that we can either give path to certificate file or directly the file content itself ([like in this TOML example](/user-guide/kv-config/#upload-the-configuration-in-the-key-value-store)).

## HTTP redirect on HTTPS

```toml
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
      certFile = "examples/traefik.crt"
      keyFile = "examples/traefik.key"
```

!!! note
    Please note that `regex` and `replacement` do not have to be set in the `redirect` structure if an entrypoint is defined for the redirection (they will not be used in this case)

## Let's Encrypt support

### Basic example

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
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

This configuration allows generating Let's Encrypt certificates for the four domains `local[1-4].com` with described SANs.

Traefik generates these certificates when it starts and it needs to be restart if new domains are added.

### OnHostRule option

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
onHostRule = true
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

This configuration allows generating Let's Encrypt certificates for the four domains `local[1-4].com`.

Traefik generates these certificates when it starts.

If a backend is added with a `onHost` rule, Traefik will automatically generate the Let's Encrypt certificate for the new domain.

### OnDemand option

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
onDemand = true
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"
```

This configuration allows generating a Let's Encrypt certificate during the first HTTPS request on a new domain.


!!! note
    This option simplifies the configuration but :

    * TLS handshakes will be slow when requesting a hostname certificate for the first time, this can leads to DDoS attacks.
    * Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits

    That's why, it's better to use the `onHostRule` optin if possible.

### DNS challenge

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
dnsProvider = "digitalocean" # DNS Provider name (cloudflare, OVH, gandi...)
delayDontCheckDNS = 0
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

DNS challenge needs environment variables to be executed.
This variables have to be set on the machine/container which host Traefik.

These variables has described [in this section](/configuration/acme/#dnsprovider).

### OnHostRule option and provided certificates

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "examples/traefik.crt"
      keyFile = "examples/traefik.key"

[acme]
email = "test@traefik.io"
storage = "acme.json"
onHostRule = true
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"

```

Traefik will only try to generate a Let's encrypt certificate if the domain cannot be checked by the provided certificates.

### Cluster mode

#### Prerequisites

Before to use Let's Encrypt in a Traefik cluster, take a look to [the key-value store explanations](/user-guide/kv-config) and more precisely to [this section](/user-guide/kv-config/#store-configuration-in-key-value-store) in the way to know how to migrate from a acme local storage *(acme.json file)* to a key-value store configuration.

#### Configuration

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "traefik/acme/account"
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

[consul]
  endpoint = "127.0.0.1:8500"
  watch = true
  prefix = "traefik"
```

This configuration allows to use the key `traefik/acme/account` to get/set Let's Encrypt certificates content.
The `consul` provider contains the configuration.

!!! note
    It's possible to use others key-value store providers as described [here](/user-guide/kv-config/#key-value-store-configuration).

## Override entrypoints in frontends

```toml
[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  passTLSCert = true
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

```toml
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
```

## Pass Authenticated user to application via headers

Providing an authentication method as described above, it is possible to pass the user to the application
via a configurable header value.

```toml
defaultEntryPoints = ["http"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth]
    headerField = "X-WebAuth-User"
    [entryPoints.http.auth.basic]
    users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
```

## Override the Traefik HTTP server IdleTimeout and/or throttle configurations from re-loading too quickly

```toml
providersThrottleDuration = "5s"

[respondingTimeouts]
idleTimeout = "360s"
```

## Securing Ping Health Check

The `/ping` health-check URL is enabled together with the web admin panel, enabled with the command-line `--web` or config file option `[web]`.
Thus, if you have a regular path for `/foo` and an entrypoint on `:80`, you would access them as follows:

* Regular path: `http://hostname:80/foo`
* Admin panel: `http://hostname:8080/`
* Ping URL: `http://hostname:8080/ping`

However, for security reasons, you may want to be able to expose the `/ping` health-check URL to outside health-checkers, e.g. an Internet service or cloud load-balancer, _without_ exposing your admin panel's port.
In many environments, the security staff may not _allow_ you to expose it.

You have two options:

* Enable `/ping` on a regular entrypoint
* Enable `/ping` on a dedicated port

### Enable ping health check on a regular entrypoint

To proxy `/ping` from a regular entrypoint to the admin one without exposing the panel, do the following:

```toml
[backends]
  [backends.traefik]
    [backends.traefik.servers.server1]
    url = "http://localhost:8080"
    weight = 10

[frontends]
  [frontends.traefikadmin]
  backend = "traefik"
    [frontends.traefikadmin.routes.ping]
    rule = "Path:/ping"
```

The above creates a new backend called `traefik`, listening on `http://localhost:8080`, i.e. the local admin port.
We only expose the admin panel via the `frontend` named `traefikadmin`, and only expose the `/ping` Path.
Be careful with the `traefikadmin` frontend. If you do _not_ specify a `Path:` rule, you would expose the entire dashboard.

### Enable ping health check on dedicated port

If you do not want to or cannot expose the health-check on a regular entrypoint - e.g. your security rules do not allow it, or you have a conflicting path - then you can enable health-check on its own entrypoint.
Use the following config:

```toml
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.ping]
  address = ":8082"

[backends]
  [backends.traefik]
    [backends.traefik.servers.server1]
    url = "http://localhost:8080"
    weight = 10

[frontends]
  [frontends.traefikadmin]
  backend = "traefik"
  entrypoints = ["ping"]
    [frontends.traefikadmin.routes.ping]
    rule = "Path:/ping"
```

The above is similar to the previous example, but instead of enabling `/ping` on the _default_ entrypoint, we enable it on a _dedicated_ entrypoint.

In the above example, you would access a regular path, admin panel and health-check as follows:

* Regular path: `http://hostname:80/foo`
* Admin panel: `http://hostname:8080/`
* Ping URL: `http://hostname:8082/ping`

Note the dedicated port `:8082` for `/ping`.

In the above example, it is _very_ important to create a named dedicated entrypoint, and do **not** include it in `defaultEntryPoints`.
Otherwise, you are likely to expose _all_ services via that entrypoint.

In the above example, we have two entrypoints, `http` and `ping`, but we only included `http` in `defaultEntryPoints`, while explicitly tying `frontend.traefikadmin` to the `ping` entrypoint.
This ensures that all the "normal" frontends will be exposed via entrypoint `http` and _not_ via entrypoint `ping`.
