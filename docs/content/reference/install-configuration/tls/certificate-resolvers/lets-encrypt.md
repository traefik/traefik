---
title: "Traefik Let's Encrypt Documentation"
description: "Learn how to configure Traefik Proxy to use an ACME provider like Let's Encrypt for automatic certificate generation. Read the technical documentation."
---

# Let's Encrypt

You can configure Traefik to use an ACME provider (like Let's Encrypt) for automatic certificate generation.

!!! warning "Let's Encrypt and Rate Limiting"
    Note that Let's Encrypt API has [rate limiting](https://letsencrypt.org/docs/rate-limits). These last up to **one week**, and cannot be overridden.
    
    When running Traefik in a container this file should be persisted across restarts. 
    If Traefik requests new certificates each time it starts up, a crash-looping container can quickly reach Let's Encrypt's ratelimits.
    To configure where certificates are stored, please take a look at the [storage](#storage) configuration.

    Use Let's Encrypt staging server with the [`caServer`](#caserver) configuration option
    when experimenting to avoid hitting this limit too fast.

## Configuration Examples

Enabling ACME

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: ":80"

  websecure:
    address: ":443"

certificatesResolvers:
  myresolver:
    acme:
      email: your-email@example.com
      storage: acme.json
      httpChallenge:
        # used during the challenge
        entryPoint: web
```

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.web]
    address = ":80"

  [entryPoints.websecure]
    address = ":443"

[certificatesResolvers.myresolver.acme]
  email = "your-email@example.com"
  storage = "acme.json"
  [certificatesResolvers.myresolver.acme.httpChallenge]
    # used during the challenge
    entryPoint = "web"
```

```bash tab="CLI"
--entryPoints.web.address=:80
--entryPoints.websecure.address=:443
# ...
--certificatesresolvers.myresolver.acme.email=your-email@example.com
--certificatesresolvers.myresolver.acme.storage=acme.json
# used during the challenge
--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web
```

```yaml tab="Helm Chart Values"
# Traefik entryPoints configuration for HTTP and HTTPS
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

# Let's Encrypt configuration
certificatesResolvers:
  myresolver:
    acme:
      email: "your-email@example.com"
      storage: "/data/acme.json"       # Path to store the certificate information
      httpChallenge:
        # Entry point to use during the ACME HTTP-01 challenge
        entryPoint: "web"
```

!!! important "Defining a certificate resolver does not result in all routers automatically using it. Each router that is supposed to use the resolver must [reference](../../../../routing/routers/index.md#certresolver) it."

??? example "Single Domain from Router's Rule Example"

    * A certificate for the domain `example.com` is requested:

    --8<-- "content/https/include-acme-single-domain-example.md"

??? example "Multiple Domains from Router's Rule Example"

    * A certificate for the domains `example.com` (main) and `blog.example.org`
      is requested:

    --8<-- "content/https/include-acme-multiple-domains-from-rule-example.md"

??? example "Multiple Domains from Router's `tls.domain` Example"

    * A certificate for the domains `example.com` (main) and `*.example.org` (SAN)
      is requested:

    --8<-- "content/https/include-acme-multiple-domains-example.md"

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `caServer` | Defines the CA server to use. More information [here](#caserver). | "https://acme-v02.api.letsencrypt.org/directory" | Yes |
| `storage` | Defines the location where the ACME certificates are saved to. More information [here](#storage) | "acme.json" | Yes |
| `certificatesDuration` | Defines the renewal period and interval for a certificate. More information [here](#certificatesduration) | 2160 | No |
| `preferredChain` | Defines the preferred chain to use. | 2160 | No |
| `keyType` | Defines the key type used for generating certificate private key. It supports 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'. | RSA4096 | No |
| `caCertificates` | Defines the the paths to PEM encoded CA Certificates that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list. | [] | No |
| `caSystemCertPool` | Defines if the certificates pool must use a copy of the system cert pool. | false | No |
| `caServerName` | Defines the CA server name that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list. | "" | No |
| `eab` | Defines the external CA. More information [here](#external-account-binding) | "" | No |

### `caServer`

_Required, Default="https://acme-v02.api.letsencrypt.org/directory"_

The CA server to use:

- Let's Encrypt production server: https://acme-v02.api.letsencrypt.org/directory
- Let's Encrypt staging server: https://acme-staging-v02.api.letsencrypt.org/directory

??? example "Using the Let's Encrypt staging server"

    ```yaml tab="File (YAML)"
    certificatesResolvers:
      myresolver:
        acme:
          # ...
          caServer: https://acme-staging-v02.api.letsencrypt.org/directory
          # ...
    ```

    ```toml tab="File (TOML)"
    [certificatesResolvers.myresolver.acme]
      # ...
      caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
      # ...
    ```

    ```bash tab="CLI"
    # ...
    --certificatesresolvers.myresolver.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory
    # ...
    ```

### `storage`

_Required, Default="acme.json"_

The `storage` option sets the location where your ACME certificates are saved to.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      storage: acme.json
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  storage = "acme.json"
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.storage=acme.json
# ...
```

ACME certificates are stored in a JSON file that needs to have a `600` file mode.

In Docker you can mount either the JSON file, or the folder containing it:

```bash
docker run -v "/my/host/acme.json:/acme.json" traefik
```

```bash
docker run -v "/my/host/acme:/etc/traefik/acme" traefik
```

!!! warning
    For concurrency reasons, this file cannot be shared across multiple instances of Traefik.

### `certificatesDuration`

_Optional, Default=2160_

`certificatesDuration` is used to calculate two durations:

- `Renew Period`: the period before the end of the certificate duration, during which the certificate should be renewed.
- `Renew Interval`: the interval between renew attempts.

It defaults to `2160` (90 days) to follow Let's Encrypt certificates' duration.

| Certificate Duration | Renew Period      | Renew Interval          |
|----------------------|-------------------|-------------------------|
| >= 1 year            | 4 months          | 1 week                  |
| >= 90 days           | 30 days           | 1 day                   |
| >= 30 days           | 10 days           | 12 hours                |
| >= 7 days            | 1 day             | 1 hour                  |
| >= 24 hours          | 6 hours           | 10 min                  |
| < 24 hours           | 20 min            | 1 min                   |

!!! warning "Traefik cannot manage certificates with a duration lower than 1 hour."

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      certificatesDuration: 72
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  certificatesDuration=72
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.certificatesduration=72
# ...
```

### External Account Binding

- `kid`: Key identifier from External CA
- `hmacEncoded`: HMAC key from External CA, should be in Base64 URL Encoding without padding format

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      eab:
        kid: abc-keyID-xyz
        hmacEncoded: abc-hmac-xyz
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.eab]
    kid = "abc-keyID-xyz"
    hmacEncoded = "abc-hmac-xyz"
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.eab.kid=abc-keyID-xyz
--certificatesresolvers.myresolver.acme.eab.hmacencoded=abc-hmac-xyz
```

## LetsEncrypt Support with the Ingress Provider

By design, Traefik is a stateless application,
meaning that it only derives its configuration from the environment it runs in,
without additional configuration.
For this reason, users can run multiple instances of Traefik at the same time to
achieve HA, as is a common pattern in the kubernetes ecosystem.

When using a single instance of Traefik Proxy with Let's Encrypt, 
you should encounter no issues. However, this could be a single point of failure.
Unfortunately, it is not possible to run multiple instances of Traefik 2.0 
with Let's Encrypt enabled, because there is no way to ensure that the correct 
instance of Traefik receives the challenge request, and subsequent responses.
Early versions (v1.x) of Traefik used a 
[KV store](https://doc.traefik.io/traefik/v1.7/configuration/acme/#storage) 
to attempt to achieve this, but due to sub-optimal performance that feature 
was dropped in 2.0.

If you need Let's Encrypt with high availability in a Kubernetes environment,
we recommend using [Traefik Enterprise](https://traefik.io/traefik-enterprise/) 
which includes distributed Let's Encrypt as a supported feature.

If you want to keep using Traefik Proxy,
LetsEncrypt HA can be achieved by using a Certificate Controller such as [Cert-Manager](https://cert-manager.io/docs/).
When using Cert-Manager to manage certificates,
it creates secrets in your namespaces that can be referenced as TLS secrets in 
your [ingress objects](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls)
.

## Fallback

If Let's Encrypt is not reachable, the following certificates will apply:

  1. Previously generated ACME certificates (before downtime)
  2. Expired ACME certificates
  3. Provided certificates

!!! important
    For new (sub)domains which need Let's Encrypt authentication, the default Traefik certificate will be used until Traefik is restarted.

{!traefik-for-business-applications.md!}
