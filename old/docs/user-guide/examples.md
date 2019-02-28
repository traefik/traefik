# Examples

You will find here some configuration examples of Traefik.

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
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
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

Traefik generates these certificates when it starts and it needs to be restart if new domains are added.

### onHostRule option (with HTTP challenge)

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
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
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

Traefik generates these certificates when it starts.

If a backend is added with a `onHost` rule, Traefik will automatically generate the Let's Encrypt certificate for the new domain (for frontends wired on the `acme.entryPoint`).

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
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
entryPoint = "https"
  [acme.httpChallenge]
  entryPoint = "http"
```

This configuration allows generating a Let's Encrypt certificate (thanks to `HTTP-01` challenge) during the first HTTPS request on a new domain.

!!! note
    This option simplifies the configuration but :

    * TLS handshakes will be slow when requesting a hostname certificate for the first time, which can lead to DDoS attacks.
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
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
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
These variables have to be set on the machine/container that host Traefik.

These variables are described [in this section](/configuration/acme/#provider).

### DNS challenge with wildcard domains

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[acme]
email = "test@traefik.io"
storage = "acme.json"
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
entryPoint = "https"
  [acme.dnsChallenge]
  provider = "digitalocean" # DNS Provider name (cloudflare, OVH, gandi...)
  delayBeforeCheck = 0

[[acme.domains]]
  main = "*.local1.com"
[[acme.domains]]
  main = "local2.com"
  sans = ["test1.local2.com", "test2x.local2.com"]
[[acme.domains]]
  main = "*.local3.com"
[[acme.domains]]
  main = "*.local4.com"
```

DNS challenge needs environment variables to be executed.
These variables have to be set on the machine/container that host Traefik.

These variables are described [in this section](/configuration/acme/#provider).

More information about wildcard certificates are available [in this section](/configuration/acme/#wildcard-domains).

### onHostRule option and provided certificates (with HTTP challenge)

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

Traefik will only try to generate a Let's encrypt certificate (thanks to `HTTP-01` challenge) if the domain cannot be checked by the provided certificates.

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
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"

  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
  rule = "Path:/test"
```

## Override the Traefik HTTP server idleTimeout and/or throttle configurations from re-loading too quickly

```toml
providersThrottleDuration = "5s"

[respondingTimeouts]
idleTimeout = "360s"
```

## Using labels in docker-compose.yml

Pay attention to the **labels** section:

```
home:
image: abiosoft/caddy:0.10.14
networks:
  - ntw_front
volumes:
  - ./www/home/srv/:/srv/
deploy:
  mode: replicated
  replicas: 2
  #placement:
  #  constraints: [node.role==manager]
  restart_policy:
    condition: on-failure
    max_attempts: 5
  resources:
    limits:
      cpus: '0.20'
      memory: 9M
    reservations:
      cpus: '0.05'
      memory: 9M
  labels:
  - "traefik.frontend.rule=PathPrefixStrip:/"
  - "traefik.backend=home"
  - "traefik.port=2015"
  - "traefik.weight=10"
  - "traefik.enable=true"
  - "traefik.passHostHeader=true"
  - "traefik.docker.network=ntw_front"
  - "traefik.frontend.entryPoints=http"
  - "traefik.backend.loadbalancer.swarm=true"
  - "traefik.backend.loadbalancer.method=drr"
```

Something more tricky using `regex`.

In this case a slash is added to `siteexample.io/portainer` and redirect to `siteexample.io/portainer/`. For more details: https://github.com/containous/traefik/issues/563

The double sign `$$` are variables managed by the docker compose file ([documentation](https://docs.docker.com/compose/compose-file/#variable-substitution)). 

```
portainer:
image: portainer/portainer:1.16.5
networks:
  - ntw_front
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
deploy:
  mode: replicated
  replicas: 1
  placement:
    constraints: [node.role==manager]
  restart_policy:
    condition: on-failure
    max_attempts: 5
  resources:
    limits:
      cpus: '0.33'
      memory: 20M
    reservations:
      cpus: '0.05'
      memory: 10M
  labels:
    - "traefik.frontend.rule=PathPrefixStrip:/portainer"
    - "traefik.backend=portainer"
    - "traefik.port=9000"
    - "traefik.weight=10"
    - "traefik.enable=true"
    - "traefik.passHostHeader=true"
    - "traefik.docker.network=ntw_front"
    - "traefik.frontend.entryPoints=http"
    - "traefik.backend.loadbalancer.swarm=true"
    - "traefik.backend.loadbalancer.method=drr"
    # https://github.com/containous/traefik/issues/563#issuecomment-421360934
    - "traefik.frontend.redirect.regex=^(.*)/portainer$$"
    - "traefik.frontend.redirect.replacement=$$1/portainer/"
    - "traefik.frontend.rule=PathPrefix:/portainer;ReplacePathRegex: ^/portainer/(.*) /$$1"
```
