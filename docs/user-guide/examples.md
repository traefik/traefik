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

!!! note
    Even if `TLS-SNI-01` challenge is [disabled](https://community.letsencrypt.org/t/2018-01-11-update-regarding-acme-tls-sni-and-shared-hosting-infrastructure/50188), for the moment, it stays the _by default_ ACME Challenge in Træfik but all the examples use the `HTTP-01` challenge (except DNS challenge examples).
    If `TLS-SNI-01` challenge is not re-enabled in the future, it we will be removed from Træfik.

### Basic example with HTTP challenge

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"
  [acme.httpChallenge]
  entryPoint = "http"

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

This configuration allows generating Let's Encrypt certificates (thanks to `HTTP-01` challenge) for the four domains `local[1-4].com` with described SANs.

Træfik generates these certificates when it starts and it needs to be restart if new domains are added.

### OnHostRule option (with HTTP challenge)

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
onHostRule = true
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"
  [acme.httpChallenge]
  entryPoint = "http"

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

This configuration allows generating Let's Encrypt certificates (thanks to `HTTP-01` challenge) for the four domains `local[1-4].com`.

Træfik generates these certificates when it starts.

If a backend is added with a `onHost` rule, Træfik will automatically generate the Let's Encrypt certificate for the new domain (for frontends wired on the `acme.entryPoint`).

### OnDemand option (with HTTP challenge)

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
onDemand = true
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"
  [acme.httpChallenge]
  entryPoint = "http"
```

This configuration allows generating a Let's Encrypt certificate (thanks to `HTTP-01` challenge) during the first HTTPS request on a new domain.

!!! note
    This option simplifies the configuration but :

    * TLS handshakes will be slow when requesting a host name certificate for the first time, this can leads to DDoS attacks.
    * Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits

    That's why, it's better to use the `onHostRule` option if possible.

### DNS challenge

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
  [acme.dnsChallenge]
  provider = "digitalocean" # DNS Provider name (cloudflare, OVH, gandi...)
  delayBeforeCheck = 0

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
These variables have to be set on the machine/container which host Træfik.

These variables are described [in this section](/configuration/acme/#provider).

### OnHostRule option and provided certificates (with HTTP challenge)

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
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
  [acme.httpChallenge]
  entryPoint = "http"
```

Træfik will only try to generate a Let's encrypt certificate (thanks to `HTTP-01` challenge) if the domain cannot be checked by the provided certificates.

### Cluster mode

#### Prerequisites

Before you use Let's Encrypt in a Traefik cluster, take a look to [the key-value store explanations](/user-guide/kv-config) and more precisely at [this section](/user-guide/kv-config/#store-configuration-in-key-value-store), which will describe how to migrate from a acme local storage *(acme.json file)* to a key-value store configuration.

#### Configuration

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "traefik/acme/account"
caServer = "http://172.18.0.1:4000/directory"
entryPoint = "https"

[acme.httpChallenge]
    entryPoint = "http"

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

## Enable Basic authentication in an entry point

With two user/pass:

- `test`:`test`
- `test2`:`test2`

Passwords are encoded in MD5: you can use `htpasswd` to generate them.

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

## Ping Health Check

The `/ping` health-check URL is enabled with the command-line `--ping` or config file option `[ping]`.
Thus, if you have a regular path for `/foo` and an entrypoint on `:80`, you would access them as follows:

* Regular path: `http://hostname:80/foo`
* Admin panel: `http://hostname:8080/`
* Ping URL: `http://hostname:8080/ping`

However, for security reasons, you may want to be able to expose the `/ping` health-check URL to outside health-checkers, e.g. an Internet service or cloud load-balancer, _without_ exposing your administration panel's port.
In many environments, the security staff may not _allow_ you to expose it.

You have two options:

* Enable `/ping` on a regular entry point
* Enable `/ping` on a dedicated port

### Enable ping health check on a regular entry point

To proxy `/ping` from a regular entry point to the administration one without exposing the panel, do the following:

```toml
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"

[ping]
entryPoint = "http"

```

The above link `ping` on the `http` entry point and then expose it on port `80`

### Enable ping health check on dedicated port

If you do not want to or cannot expose the health-check on a regular entry point - e.g. your security rules do not allow it, or you have a conflicting path - then you can enable health-check on its own entry point.
Use the following configuration:

```toml
defaultEntryPoints = ["http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.ping]
  address = ":8082"

[ping]
entryPoint = "ping"
```

The above is similar to the previous example, but instead of enabling `/ping` on the _default_ entry point, we enable it on a _dedicated_ entry point.

In the above example, you would access a regular path and health-check as follows:

* Regular path: `http://hostname:80/foo`
* Ping URL: `http://hostname:8082/ping`

Note the dedicated port `:8082` for `/ping`.

In the above example, it is _very_ important to create a named dedicated entry point, and do **not** include it in `defaultEntryPoints`.
Otherwise, you are likely to expose _all_ services via this entry point.
