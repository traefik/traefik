---
title: "Traefik Let's Encrypt Documentation"
description: "Learn how to configure Traefik Proxy to use an ACME provider like Let's Encrypt for automatic certificate generation. Read the technical documentation."
---

# Let's Encrypt

Automatic HTTPS
{: .subtitle }

You can configure Traefik to use an ACME provider (like Let's Encrypt) for automatic certificate generation.

!!! warning "Let's Encrypt and Rate Limiting"
    Note that Let's Encrypt API has [rate limiting](https://letsencrypt.org/docs/rate-limits). These last up to **one week**, and cannot be overridden.
    
    When running Traefik in a container the `acme.json` file should be persisted across restarts. 
    If Traefik requests new certificates each time it starts up, a crash-looping container can quickly reach Let's Encrypt's ratelimits.
    To configure where certificates are stored, please take a look at the [storage](#storage) configuration.

    Use Let's Encrypt staging server with the [`caServer`](#caserver) configuration option
    when experimenting to avoid hitting this limit too fast.

## Certificate Resolvers

Traefik requires you to define "Certificate Resolvers" in the [static configuration](../getting-started/configuration-overview.md#the-static-configuration),
which are responsible for retrieving certificates from an ACME server.

Then, each ["router"](../routing/routers/index.md) is configured to enable TLS,
and is associated to a certificate resolver through the [`tls.certresolver` configuration option](../routing/routers/index.md#certresolver).

Certificates are requested for domain names retrieved from the router's [dynamic configuration](../getting-started/configuration-overview.md#the-dynamic-configuration).

You can read more about this retrieval mechanism in the following section: [ACME Domain Definition](#domain-definition).

!!! warning "Defining an [ACME challenge type](#the-different-acme-challenges) is a requirement for a certificate resolver to be functional."

!!! important "Defining a certificate resolver does not result in all routers automatically using it. Each router that is supposed to use the resolver must [reference](../routing/routers/index.md#certresolver) it."

??? note "Configuration Reference"

    There are many available options for ACME.
    For a quick glance at what's possible, browse the configuration reference:

    ```yaml tab="File (YAML)"
    --8<-- "content/https/ref-acme.yaml"
    ```

    ```toml tab="File (TOML)"
    --8<-- "content/https/ref-acme.toml"
    ```

    ```bash tab="CLI"
    --8<-- "content/https/ref-acme.txt"
    ```

## Domain Definition

Certificate resolvers request certificates for a set of the domain names
inferred from routers, with the following logic:

- If the router has a [`tls.domains`](../routing/routers/index.md#domains) option set,
  then the certificate resolver uses the `main` (and optionally `sans`) option of `tls.domains` to know the domain names for this router.

- If no [`tls.domains`](../routing/routers/index.md#domains) option is set,
  then the certificate resolver uses the [router's rule](../routing/routers/index.md#rule),
  by checking the `Host()` matchers.
  Please note that [multiple `Host()` matchers can be used](../routing/routers/index.md#certresolver)) for specifying multiple domain names for this router.

Please note that:

- When multiple domain names are inferred from a given router,
  only **one** certificate is requested with the first domain name as the main domain,
  and the other domains as ["SANs" (Subject Alternative Name)](https://en.wikipedia.org/wiki/Subject_Alternative_Name).

- As [ACME V2 supports "wildcard domains"](#wildcard-domains),
  any router can provide a [wildcard domain](https://en.wikipedia.org/wiki/Wildcard_certificate) name, as "main" domain or as "SAN" domain.

Please check the [configuration examples below](#configuration-examples) for more details.

## Configuration Examples

??? example "Enabling ACME"

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

!!! important "Defining a certificate resolver does not result in all routers automatically using it. Each router that is supposed to use the resolver must [reference](../routing/routers/index.md#certresolver) it."

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

## Automatic Renewals

Traefik automatically tracks the expiry date of ACME certificates it generates.

By default, Traefik manages 90 days certificates,
and starts to renew certificates 30 days before their expiry.

When using a certificate resolver that issues certificates with custom durations,
one can configure the certificates' duration with the [`certificatesDuration`](#certificatesduration) option.

!!! info ""
    Certificates that are no longer used may still be renewed, as Traefik does not currently check if the certificate is being used before renewing.

## Using LetsEncrypt with Kubernetes

When using LetsEncrypt with kubernetes, there are some known caveats with both the [ingress](../providers/kubernetes-ingress.md) and [crd](../providers/kubernetes-crd.md) providers.

!!! info ""
    If you intend to run multiple instances of Traefik with LetsEncrypt, please ensure you read the sections on those provider pages.

## The Different ACME Challenges

!!! warning "Defining one ACME challenge is a requirement for a certificate resolver to be functional."

!!! important "Defining a certificate resolver does not result in all routers automatically using it. Each router that is supposed to use the resolver must [reference](../routing/routers/index.md#certresolver) it."

### `tlsChallenge`

Use the `TLS-ALPN-01` challenge to generate and renew ACME certificates by provisioning a TLS certificate.

As described on the Let's Encrypt [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72),
when using the `TLS-ALPN-01` challenge, Traefik must be reachable by Let's Encrypt through port 443.

??? example "Configuring the `tlsChallenge`"

    ```yaml tab="File (YAML)"
    certificatesResolvers:
      myresolver:
        acme:
          # ...
          tlsChallenge: {}
    ```

    ```toml tab="File (TOML)"
    [certificatesResolvers.myresolver.acme]
      # ...
      [certificatesResolvers.myresolver.acme.tlsChallenge]
    ```

    ```bash tab="CLI"
    # ...
    --certificatesresolvers.myresolver.acme.tlschallenge=true
    ```

### `httpChallenge`

Use the `HTTP-01` challenge to generate and renew ACME certificates by provisioning an HTTP resource under a well-known URI.

As described on the Let's Encrypt [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72),
when using the `HTTP-01` challenge, `certificatesresolvers.myresolver.acme.httpchallenge.entrypoint` must be reachable by Let's Encrypt through port 80.

??? example "Using an EntryPoint Called web for the `httpChallenge`"

    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"

      websecure:
        address: ":443"

    certificatesResolvers:
      myresolver:
        acme:
          # ...
          httpChallenge:
            entryPoint: web
    ```

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"

      [entryPoints.websecure]
        address = ":443"

    [certificatesResolvers.myresolver.acme]
      # ...
      [certificatesResolvers.myresolver.acme.httpChallenge]
        entryPoint = "web"
    ```

    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    # ...
    --certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web
    ```

!!! info ""
    Redirection is fully compatible with the `HTTP-01` challenge.

#### `Delay`

The delay between the creation of the challenge and the validation.
A value lower than or equal to zero means no delay.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      httpChallenge:
        # ...
        delay: 12
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.httpChallenge]
    # ...
    delay = 12
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.httpchallenge.delay=12
```

### `dnsChallenge`

Use the `DNS-01` challenge to generate and renew ACME certificates by provisioning a DNS record.

??? example "Configuring a `dnsChallenge` with the DigitalOcean Provider"

    ```yaml tab="File (YAML)"
    certificatesResolvers:
      myresolver:
        acme:
          # ...
          dnsChallenge:
            provider: digitalocean
            delayBeforeCheck: 0
        # ...
    ```

    ```toml tab="File (TOML)"
    [certificatesResolvers.myresolver.acme]
      # ...
      [certificatesResolvers.myresolver.acme.dnsChallenge]
        provider = "digitalocean"
        delayBeforeCheck = 0
    # ...
    ```

    ```bash tab="CLI"
    # ...
    --certificatesresolvers.myresolver.acme.dnschallenge.provider=digitalocean
    --certificatesresolvers.myresolver.acme.dnschallenge.delaybeforecheck=0
    # ...
    ```

!!! warning "`CNAME` support"

    `CNAME` are supported (and sometimes even [encouraged](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html#the-advantages-of-a-cname)),
    but there are a few cases where they can be [problematic](../../getting-started/faq/#why-does-lets-encrypt-wildcard-certificate-renewalgeneration-with-dns-challenge-fail).

    If needed, `CNAME` support can be disabled with the following environment variable:

    ```bash
    LEGO_DISABLE_CNAME_SUPPORT=true
    ```

!!! warning "Multiple DNS Challenge provider"
    
    Multiple DNS challenge provider are not supported with Traefik, but you can use `CNAME` to handle that.
    For example, if you have `example.org` (account foo) and `example.com` (account bar) you can create a CNAME on `example.org` called `_acme-challenge.example.org` pointing to `challenge.example.com`.
    This way, you can obtain certificates for `example.org` with the `bar` account.

!!! important
    A `provider` is mandatory.

#### `providers`

Here is a list of supported `providers`, that can automate the DNS verification,
along with the required environment variables and their [wildcard & root domain support](#wildcard-domains).
Do not hesitate to complete it.

Many lego environment variables can be overridden by their respective `_FILE` counterpart, which should have a filepath to a file that contains the secret as its value.
For example, `CF_API_EMAIL_FILE=/run/secrets/traefik_cf-api-email` could be used to provide a Cloudflare API email address as a Docker secret named `traefik_cf-api-email`.

For complete details, refer to your provider's _Additional configuration_ link.

| Provider Name                                                          | Provider Code      | Environment Variables                                                                                                                                                            |                                                                                 |
|------------------------------------------------------------------------|--------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| [ACME DNS](https://github.com/joohoi/acme-dns)                         | `acme-dns`         | `ACME_DNS_API_BASE`, `ACME_DNS_STORAGE_PATH`, `ACME_DNS_STORAGE_BASE_URL`                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/acme-dns)         |
| [Active24](https://www.active24.cz)                                    | `active24`         | `ACTIVE24_API_KEY`, `ACTIVE24_SECRET`                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/active24)         |
| [Alibaba Cloud](https://www.alibabacloud.com)                          | `alidns`           | `ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION_ID`                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/alidns)           |
| [all-inkl](https://all-inkl.com)                                       | `allinkl`          | `ALL_INKL_LOGIN`, `ALL_INKL_PASSWORD`                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/allinkl)          |
| [ArvanCloud](https://www.arvancloud.ir/en)                             | `arvancloud`       | `ARVANCLOUD_API_KEY`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/arvancloud)       |
| [Auroradns](https://www.pcextreme.com/dns-health-checks)               | `auroradns`        | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/auroradns)        |
| [Autodns](https://www.internetx.com/domains/autodns/)                  | `autodns`          | `AUTODNS_API_USER`, `AUTODNS_API_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/autodns)          |
| [Axelname](https://axelname.ru)                                        | `axelname`         | `AXELNAME_NICKNAME`, `AXELNAME_TOKEN`                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/axelname)         |
| [Azion](https://zonomi.com)                                            | `azion`            | `AZION_PERSONAL_TOKEN`                                                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/azion)            |
| [Azure](https://azure.microsoft.com/services/dns/) (DEPRECATED)        | `azure`            | DEPRECATED use `azuredns` instead.                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/azure)            |
| [AzureDNS](https://azure.microsoft.com/services/dns/)                  | `azuredns`         | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`, `AZURE_SUBSCRIPTION_ID`, `AZURE_RESOURCE_GROUP`, `[AZURE_ENVIRONMENT]`, `[AZURE_PRIVATE_ZONE]`, `[AZURE_ZONE_NAME]` | [Additional configuration](https://go-acme.github.io/lego/dns/azuredns)         |
| [Baidu Cloud](https://cloud.baidu.com)                                 | `baiducloud`       | `BAIDUCLOUD_ACCESS_KEY_ID`, `BAIDUCLOUD_SECRET_ACCESS_KEY`                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/baiducloud)       |
| [Bindman](https://github.com/labbsr0x/bindman-dns-webhook)             | `bindman`          | `BINDMAN_MANAGER_ADDRESS`                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/bindman)          |
| [Blue Cat](https://www.bluecatnetworks.com/)                           | `bluecat`          | `BLUECAT_SERVER_URL`, `BLUECAT_USER_NAME`, `BLUECAT_PASSWORD`, `BLUECAT_CONFIG_NAME`, `BLUECAT_DNS_VIEW`                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/bluecat)          |
| [BookMyName](https://www.bookmyname.com)                               | `bookmyname`       | `BOOKMYNAME_USERNAME`, `BOOKMYNAME_PASSWORD`                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/bookmyname)       |
| [Brandit](https://www.brandit.com) (DEPRECATED)                        | `brandit`          | DEPRECATED                                                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/brandit)          |
| [Bunny](https://bunny.net)                                             | `bunny`            | `BUNNY_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/bunny)            |
| [Checkdomain](https://www.checkdomain.de/)                             | `checkdomain`      | `CHECKDOMAIN_TOKEN`,                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/checkdomain/)     |
| [Civo](https://www.civo.com/)                                          | `civo`             | `CIVO_TOKEN`                                                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/civo)             |
| [Cloud.ru](https://cloud.ru)                                           | `cloudru`          | `CLOUDRU_SERVICE_INSTANCE_ID`, `CLOUDRU_KEY_ID`, `CLOUDRU_SECRET`                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/cloudru)          |
| [CloudDNS](https://vshosting.eu/)                                      | `clouddns`         | `CLOUDDNS_CLIENT_ID`, `CLOUDDNS_EMAIL`, `CLOUDDNS_PASSWORD`                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/clouddns)         |
| [Cloudflare](https://www.cloudflare.com)                               | `cloudflare`       | `CF_API_EMAIL`, `CF_API_KEY` [^5] or `CF_DNS_API_TOKEN`, `[CF_ZONE_API_TOKEN]`                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/cloudflare)       |
| [ClouDNS](https://www.cloudns.net/)                                    | `cloudns`          | `CLOUDNS_AUTH_ID`, `CLOUDNS_AUTH_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/cloudns)          |
| [CloudXNS](https://www.cloudxns.net) (DEPRECATED)                      | `cloudxns`         | DEPRECATED                                                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/cloudxns)         |
| [ConoHa v3](https://www.conoha.jp/)                                    | `conohav3`         | `CONOHAV3_TENANT_ID`, `CONOHAV3_API_USER_ID`, `CONOHAV3_API_PASSWORD`                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/conohav3)         |
| [ConoHa](https://www.conoha.jp)                                        | `conoha`           | `CONOHA_TENANT_ID`, `CONOHA_API_USERNAME`, `CONOHA_API_PASSWORD`                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/conoha)           |
| [Constellix](https://constellix.com)                                   | `constellix`       | `CONSTELLIX_API_KEY`, `CONSTELLIX_SECRET_KEY`                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/constellix)       |
| [Core-Networks](https://www.core-networks.de)                          | `corenetworks`     | `CORENETWORKS_LOGIN`, `CORENETWORKS_PASSWORD`                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/corenetworks)     |
| [CPanel and WHM](https://cpanel.net/)                                  | `cpanel`           | `CPANEL_MODE`, `CPANEL_USERNAME`, `CPANEL_TOKEN`, `CPANEL_BASE_URL`                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/cpanel)           |
| [Derak Cloud](https://derak.cloud/)                                    | `derak`            | `DERAK_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/derak)            |
| [deSEC](https://desec.io)                                              | `desec`            | `DESEC_TOKEN`                                                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/desec)            |
| [DigitalOcean](https://www.digitalocean.com)                           | `digitalocean`     | `DO_AUTH_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/digitalocean)     |
| [DirectAdmin](https://www.directadmin.com)                             | `directadmin`      | `DIRECTADMIN_API_URL` , `DIRECTADMIN_USERNAME`, `DIRECTADMIN_PASSWORD`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/directadmin)      |
| [DNS Made Easy](https://dnsmadeeasy.com)                               | `dnsmadeeasy`      | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/dnsmadeeasy)      |
| [dnsHome.de](https://www.dnshome.de)                                   | `dnsHomede`        | `DNSHOMEDE_CREDENTIALS`                                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/dnshomede)        |
| [DNSimple](https://dnsimple.com)                                       | `dnsimple`         | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/dnsimple)         |
| [DNSPod](https://www.dnspod.com/) (DEPRECATED)                         | `dnspod`           | DEPRECATED  use `tencentcloud` instead.                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/dnspod)           |
| [Domain Offensive (do.de)](https://www.do.de/)                         | `dode`             | `DODE_TOKEN`                                                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/dode)             |
| [Domeneshop](https://domene.shop)                                      | `domeneshop`       | `DOMENESHOP_API_TOKEN`, `DOMENESHOP_API_SECRET`                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/domeneshop)       |
| [DreamHost](https://www.dreamhost.com/)                                | `dreamhost`        | `DREAMHOST_API_KEY`                                                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/dreamhost)        |
| [Duck DNS](https://www.duckdns.org/)                                   | `duckdns`          | `DUCKDNS_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/duckdns)          |
| [Dyn](https://dyn.com)                                                 | `dyn`              | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/dyn)              |
| [DynDnsFree.de](https://www.dyndnsfree.de)                             | `dyndnsfree`       | `DYNDNSFREE_USERNAME`, `DYNDNSFREE_PASSWORD`                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/dyndnsfree)       |
| [Dynu](https://www.dynu.com)                                           | `dynu`             | `DYNU_API_KEY`                                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/dynu)             |
| [EasyDNS](https://easydns.com/)                                        | `easydns`          | `EASYDNS_TOKEN`, `EASYDNS_KEY`                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/easydns)          |
| [EdgeDNS](https://www.akamai.com/)                                     | `edgedns`          | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/edgedns)          |
| [Efficient IP](https://efficientip.com)                                | `efficientip`      | `EFFICIENTIP_USERNAME`, `EFFICIENTIP_PASSWORD`, `EFFICIENTIP_HOSTNAME`, `EFFICIENTIP_DNS_NAME`                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/efficientip)      |
| [Epik](https://www.epik.com)                                           | `epik`             | `EPIK_SIGNATURE`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/epik)             |
| [Exoscale](https://www.exoscale.com)                                   | `exoscale`         | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/exoscale)         |
| [F5 XC](https://www.f5.com/products/distributed-cloud-services)        | `f5xc`             | `F5XC_API_TOKEN`, `F5XC_TENANT_NAME`, `F5XC_GROUP_NAME`                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/f5xc)             |
| [Fast DNS](https://www.akamai.com/)                                    | `fastdns`          | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/edgedns)          |
| [Freemyip.com](https://freemyip.com)                                   | `freemyip`         | `FREEMYIP_TOKEN`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/freemyip)         |
| [G-Core](https://gcore.com/dns/)                                       | `gcore`            | `GCORE_PERMANENT_API_TOKEN`                                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/gcore)            |
| [Gandi v5](https://doc.livedns.gandi.net)                              | `gandiv5`          | `GANDIV5_PERSONAL_ACCESS_TOKEN`                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/gandiv5)          |
| [Gandi](https://www.gandi.net)                                         | `gandi`            | `GANDI_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/gandi)            |
| [Glesys](https://glesys.com/)                                          | `glesys`           | `GLESYS_API_USER`, `GLESYS_API_KEY`, `GLESYS_DOMAIN`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/glesys)           |
| [GoDaddy](https://www.godaddy.com)                                     | `godaddy`          | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/godaddy)          |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/)                 | `gcloud`           | `GCE_PROJECT`, Application Default Credentials [^2] [^3], [`GCE_SERVICE_ACCOUNT_FILE`]                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/gcloud)           |
| [Google Domains](https://domains.google) (DEPRECATED)                  | `googledomains`    | DEPRECATED                                                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/googledomains)    |
| [Hetzner](https://hetzner.com)                                         | `hetzner`          | `HETZNER_API_KEY`                                                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/hetzner)          |
| [hosting.de](https://www.hosting.de)                                   | `hostingde`        | `HOSTINGDE_API_KEY`, `HOSTINGDE_ZONE_NAME`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/hostingde)        |
| [Hosttech](https://www.hosttech.eu)                                    | `hosttech`         | `HOSTTECH_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/hosttech)         |
| [http.net](https://www.http.net/)                                      | `httpnet`          | `HTTPNET_API_KEY`                                                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/httpnet)          |
| [Huawei Cloud](https://huaweicloud.com)                                | `huaweicloud`      | `HUAWEICLOUD_ACCESS_KEY_ID`, `HUAWEICLOUD_SECRET_ACCESS_KEY`, `HUAWEICLOUD_REGION`                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/huaweicloud)      |
| [Hurricane Electric](https://dns.he.net)                               | `hurricane`        | `HURRICANE_TOKENS` [^6]                                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/hurricane)        |
| [HyperOne](https://www.hyperone.com)                                   | `hyperone`         | `HYPERONE_PASSPORT_LOCATION`, `HYPERONE_LOCATION_ID`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/hyperone)         |
| [IBM Cloud (SoftLayer)](https://www.ibm.com/cloud/)                    | `ibmcloud`         | `SOFTLAYER_USERNAME`, `SOFTLAYER_API_KEY`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/ibmcloud)         |
| [IIJ DNS Platform Service](https://www.iij.ad.jp)                      | `iijdpf`           | `IIJ_DPF_API_TOKEN` , `IIJ_DPF_DPM_SERVICE_CODE`                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/iijdpf)           |
| [IIJ](https://www.iij.ad.jp/)                                          | `iij`              | `IIJ_API_ACCESS_KEY`, `IIJ_API_SECRET_KEY`, `IIJ_DO_SERVICE_CODE`                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/iij)              |
| [Infoblox](https://www.infoblox.com/)                                  | `infoblox`         | `INFOBLOX_USERNAME`, `INFOBLOX_PASSWORD`, `INFOBLOX_HOST`                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/infoblox)         |
| [Infomaniak](https://www.infomaniak.com)                               | `infomaniak`       | `INFOMANIAK_ACCESS_TOKEN`                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/infomaniak)       |
| [Internet.bs](https://internetbs.net)                                  | `internetbs`       | `INTERNET_BS_API_KEY`, `INTERNET_BS_PASSWORD`                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/internetbs)       |
| [INWX](https://www.inwx.de/en)                                         | `inwx`             | `INWX_USERNAME`, `INWX_PASSWORD`                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/inwx)             |
| [ionos](https://ionos.com/)                                            | `ionos`            | `IONOS_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/ionos)            |
| [IPv64](https://ipv64.net)                                             | `ipv64`            | `IPV64_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/ipv64)            |
| [iwantmyname](https://iwantmyname.com)                                 | `iwantmyname`      | `IWANTMYNAME_USERNAME` , `IWANTMYNAME_PASSWORD`                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/iwantmyname)      |
| [Joker.com](https://joker.com)                                         | `joker`            | `JOKER_API_MODE` with `JOKER_API_KEY` or `JOKER_USERNAME`, `JOKER_PASSWORD`                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/joker)            |
| [Liara](https://liara.ir)                                              | `liara`            | `LIARA_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/liara)            |
| [Lightsail](https://aws.amazon.com/lightsail/)                         | `lightsail`        | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `DNS_ZONE`                                                                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/lightsail)        |
| [Lima-City](https://www.lima-city.de)                                  | `limacity`         | `LIMACITY_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/limacity)         |
| [Linode v4](https://www.linode.com)                                    | `linode`           | `LINODE_TOKEN`                                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/linode)           |
| [Liquid Web](https://www.liquidweb.com/)                               | `liquidweb`        | `LIQUID_WEB_PASSWORD`, `LIQUID_WEB_USERNAME`, `LIQUID_WEB_ZONE`                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/liquidweb)        |
| [Loopia](https://loopia.com/)                                          | `loopia`           | `LOOPIA_API_PASSWORD`, `LOOPIA_API_USER`                                                                                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/loopia)           |
| [LuaDNS](https://luadns.com)                                           | `luadns`           | `LUADNS_API_USERNAME`, `LUADNS_API_TOKEN`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/luadns)           |
| [Mail-in-a-Box](https://mailinabox.email)                              | `mailinabox`       | `MAILINABOX_EMAIL`, `MAILINABOX_PASSWORD`, `MAILINABOX_BASE_URL`                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/mailinabox)       |
| [ManageEngine CloudDNS](https://clouddns.manageengine.com)             | `manageengine`     | `MANAGEENGINE_CLIENT_ID`, `MANAGEENGINE_CLIENT_SECRET`                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/manageengine)     |
| [Metaname](https://metaname.net)                                       | `metaname`         | `METANAME_ACCOUNT_REFERENCE`, `METANAME_API_KEY`                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/metaname)         |
| [Metaregistrar](https://metaregistrar.com)                             | `metaregistrar`    | `METAREGISTRAR_API_TOKEN`                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/metaregistrar)    |
| [mijn.host](https://mijn.host/)                                        | `mijnhost`         | `MIJNHOST_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/mijnhost)         |
| [Mittwald](https://www.mittwald.de)                                    | `mittwald`         | `MITTWALD_TOKEN`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/mittwald)         |
| [myaddr.{tools,dev,io}](https://myaddr.tools/)                         | `myaddr`           | `MYADDR_PRIVATE_KEYS_MAPPING`                                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/myaddr)           |
| [MyDNS.jp](https://www.mydns.jp/)                                      | `mydnsjp`          | `MYDNSJP_MASTER_ID`, `MYDNSJP_PASSWORD`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/mydnsjp)          |
| [Mythic Beasts](https://www.mythic-beasts.com)                         | `mythicbeasts`     | `MYTHICBEASTS_USER_NAME`, `MYTHICBEASTS_PASSWORD`                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/mythicbeasts)     |
| [name.com](https://www.name.com/)                                      | `namedotcom`       | `NAMECOM_USERNAME`, `NAMECOM_API_TOKEN`, `NAMECOM_SERVER`                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/namedotcom)       |
| [Namecheap](https://www.namecheap.com)                                 | `namecheap`        | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/namecheap)        |
| [Namesilo](https://www.namesilo.com/)                                  | `namesilo`         | `NAMESILO_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/namesilo)         |
| [NearlyFreeSpeech.NET](https://www.nearlyfreespeech.net/)              | `nearlyfreespeech` | `NEARLYFREESPEECH_API_KEY`, `NEARLYFREESPEECH_LOGIN`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/nearlyfreespeech) |
| [Netcup](https://www.netcup.eu/)                                       | `netcup`           | `NETCUP_CUSTOMER_NUMBER`, `NETCUP_API_KEY`, `NETCUP_API_PASSWORD`                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/netcup)           |
| [Netlify](https://www.netlify.com)                                     | `netlify`          | `NETLIFY_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/netlify)          |
| [Nicmanager](https://www.nicmanager.com)                               | `nicmanager`       | `NICMANAGER_API_EMAIL`, `NICMANAGER_API_PASSWORD`                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/nicmanager)       |
| [NIFCloud](https://cloud.nifty.com/service/dns.htm)                    | `nifcloud`         | `NIFCLOUD_ACCESS_KEY_ID`, `NIFCLOUD_SECRET_ACCESS_KEY`                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/nifcloud)         |
| [Njalla](https://njal.la)                                              | `njalla`           | `NJALLA_TOKEN`                                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/njalla)           |
| [Nodion](https://www.nodion.com)                                       | `nodion`           | `NODION_API_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/nodion)           |
| [NS1](https://ns1.com/)                                                | `ns1`              | `NS1_API_KEY`                                                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/ns1)              |
| [Open Telekom Cloud](https://cloud.telekom.de)                         | `otc`              | `OTC_DOMAIN_NAME`, `OTC_USER_NAME`, `OTC_PASSWORD`, `OTC_PROJECT_NAME`, `OTC_IDENTITY_ENDPOINT`                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/otc)              |
| [Openstack Designate](https://docs.openstack.org/designate)            | `designate`        | `OS_AUTH_URL`, `OS_USERNAME`, `OS_PASSWORD`, `OS_TENANT_NAME`, `OS_REGION_NAME`                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/designate)        |
| [Oracle Cloud](https://cloud.oracle.com/home)                          | `oraclecloud`      | `OCI_COMPARTMENT_OCID`, `OCI_PRIVKEY_FILE`, `OCI_PRIVKEY_PASS`, `OCI_PUBKEY_FINGERPRINT`, `OCI_REGION`, `OCI_TENANCY_OCID`, `OCI_USER_OCID`                                      | [Additional configuration](https://go-acme.github.io/lego/dns/oraclecloud)      |
| [OVH](https://www.ovh.com)                                             | `ovh`              | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`, `OVH_CLIENT_ID`, `OVH_CLIENT_SECRET`                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/ovh)              |
| [Plesk](https://www.plesk.com)                                         | `plesk`            | `PLESK_SERVER_BASE_URL`, `PLESK_USERNAME`, `PLESK_PASSWORD`                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/plesk)            |
| [Porkbun](https://porkbun.com/)                                        | `porkbun`          | `PORKBUN_SECRET_API_KEY`, `PORKBUN_API_KEY`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/porkbun)          |
| [PowerDNS](https://www.powerdns.com)                                   | `pdns`             | `PDNS_API_KEY`, `PDNS_API_URL`                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/pdns)             |
| [Rackspace](https://www.rackspace.com/cloud/dns)                       | `rackspace`        | `RACKSPACE_USER`, `RACKSPACE_API_KEY`                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/rackspace)        |
| [Rainyun/雨云](https://www.rainyun.com)                                  | `rainyun`          | `RAINYUN_API_KEY`                                                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/rainyun)          |
| [RcodeZero](https://www.rcodezero.at)                                  | `rcodezero`        | `RCODEZERO_API_TOKEN`                                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/rcodezero)        |
| [reg.ru](https://www.reg.ru)                                           | `regru`            | `REGRU_USERNAME`, `REGRU_PASSWORD`                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/regru)            |
| [Regfish](https://regfish.de)                                          | `regfish`          | `regfish`                                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/regfish)          |
| [RFC2136](https://tools.ietf.org/html/rfc2136)                         | `rfc2136`          | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/rfc2136)          |
| [RimuHosting](https://rimuhosting.com)                                 | `rimuhosting`      | `RIMUHOSTING_API_KEY`                                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/rimuhosting)      |
| [Route 53](https://aws.amazon.com/route53/)                            | `route53`          | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `[AWS_REGION]`, `[AWS_HOSTED_ZONE_ID]` or a configured user/instance IAM profile.                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/route53)          |
| [RU Center](https://nic.ru/)                                           | `nicru`            | `NICRU_USER`, `NICRU_PASSWORD`, `NICRU_SERVICE_ID`, `NICRU_SECRET`, `NICRU_SERVICE_NAME`                                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/nicru)            |
| [Sakura Cloud](https://cloud.sakura.ad.jp/)                            | `sakuracloud`      | `SAKURACLOUD_ACCESS_TOKEN`, `SAKURACLOUD_ACCESS_TOKEN_SECRET`                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/sakuracloud)      |
| [Scaleway](https://www.scaleway.com)                                   | `scaleway`         | `SCW_API_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/scaleway)         |
| [Selectel v2](https://selectel.ru/en/)                                 | `selectelv2`       | `SELECTELV2_ACCOUNT_ID`, `SELECTELV2_PASSWORD`, `SELECTELV2_PROJECT_ID`, `SELECTELV2_USERNAME`                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/selectelv2)       |
| [Selectel](https://selectel.ru/en/)                                    | `selectel`         | `SELECTEL_API_TOKEN`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/selectel)         |
| [SelfHost.(de/eu)](https://www.selfhost.de)                            | `selfhostde`       | `SELFHOSTDE_USERNAME`, `SELFHOSTDE_PASSWORD`, `SELFHOSTDE_RECORDS_MAPPING`                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/selfhostde)       |
| [Servercow](https://servercow.de)                                      | `servercow`        | `SERVERCOW_USERNAME`, `SERVERCOW_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/servercow)        |
| [Shellrent](https://www.shellrent.com)                                 | `shellrent`        | `SHELLRENT_USERNAME`, `SHELLRENT_TOKEN`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/shellrent)        |
| [Simply.com](https://www.simply.com/en/domains/)                       | `simply`           | `SIMPLY_ACCOUNT_NAME`, `SIMPLY_API_KEY`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/simply)           |
| [Sonic](https://www.sonic.com/)                                        | `sonic`            | `SONIC_USER_ID`, `SONIC_API_KEY`                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/sonic)            |
| [Spaceship](https://spaceship.com)                                     | `spaceship`        | `SPACESHIP_API_KEY`, `SPACESHIP_API_SECRET`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/spaceship)        |
| [Stackpath](https://www.stackpath.com/)                                | `stackpath`        | `STACKPATH_CLIENT_ID`, `STACKPATH_CLIENT_SECRET`, `STACKPATH_STACK_ID`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/stackpath)        |
| [Technitium](https://technitium.com)                                   | `technitium`       | `TECHNITIUM_SERVER_BASE_URL`, `TECHNITIUM_API_TOKEN`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/technitium)       |
| [Tencent Cloud DNS](https://cloud.tencent.com/product/cns)             | `tencentcloud`     | `TENCENTCLOUD_SECRET_ID`, `TENCENTCLOUD_SECRET_KEY`                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/tencentcloud)     |
| [Timeweb Cloud](https://timeweb.cloud)                                 | `timewebcloud`     | `TIMEWEBCLOUD_AUTH_TOKEN`                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/timewebcloud)     |
| [TransIP](https://www.transip.nl/)                                     | `transip`          | `TRANSIP_ACCOUNT_NAME`, `TRANSIP_PRIVATE_KEY_PATH`                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/transip)          |
| [UKFast SafeDNS](https://docs.ukfast.co.uk/domains/safedns/index.html) | `safedns`          | `SAFEDNS_AUTH_TOKEN`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/safedns)          |
| [Ultradns](https://neustarsecurityservices.com/dns-services)           | `ultradns`         | `ULTRADNS_USERNAME`, `ULTRADNS_PASSWORD`                                                                                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/ultradns)         |
| [Variomedia](https://www.variomedia.de/)                               | `variomedia`       | `VARIOMEDIA_API_TOKEN`                                                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/variomedia)       |
| [VegaDNS](https://github.com/shupp/VegaDNS-API)                        | `vegadns`          | `SECRET_VEGADNS_KEY`, `SECRET_VEGADNS_SECRET`, `VEGADNS_URL`                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/vegadns)          |
| [Vercel](https://vercel.com)                                           | `vercel`           | `VERCEL_API_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/vercel)           |
| [Versio](https://www.versio.nl/domeinnamen)                            | `versio`           | `VERSIO_USERNAME`, `VERSIO_PASSWORD`                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/versio)           |
| [VinylDNS](https://www.vinyldns.io)                                    | `vinyldns`         | `VINYLDNS_ACCESS_KEY`, `VINYLDNS_SECRET_KEY`, `VINYLDNS_HOST`                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/vinyldns)         |
| [VK Cloud](https://mcs.mail.ru/)                                       | `vkcloud`          | `VK_CLOUD_PASSWORD`, `VK_CLOUD_PROJECT_ID`, `VK_CLOUD_USERNAME`                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/vkcloud)          |
| [Volcano Engine](https://www.volcengine.com)                           | `volcengine`       | `VOLC_ACCESSKEY`, `VOLC_SECRETKEY`                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/volcengine)       |
| [Vscale](https://vscale.io/)                                           | `vscale`           | `VSCALE_API_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/vscale)           |
| [VULTR](https://www.vultr.com)                                         | `vultr`            | `VULTR_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/vultr)            |
| [Webnames](https://www.webnames.ru/)                                   | `webnames`         | `WEBNAMES_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/webnames)         |
| [Websupport](https://websupport.sk)                                    | `websupport`       | `WEBSUPPORT_API_KEY`, `WEBSUPPORT_SECRET`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/websupport)       |
| [WEDOS](https://www.wedos.com)                                         | `wedos`            | `WEDOS_USERNAME`, `WEDOS_WAPI_PASSWORD`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/wedos)            |
| [West.cn/西部数码](https://www.west.cn)                                    | `westcn`           | `WESTCN_USERNAME`, `WESTCN_PASSWORD`                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/westcn)           |
| [Yandex 360](https://360.yandex.ru)                                    | `yandex360`        | `YANDEX360_OAUTH_TOKEN`, `YANDEX360_ORG_ID`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/yandex360)        |
| [Yandex Cloud](https://cloud.yandex.com/en/)                           | `yandexcloud`      | `YANDEX_CLOUD_FOLDER_ID`, `YANDEX_CLOUD_IAM_TOKEN`                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/yandexcloud)      |
| [Yandex](https://yandex.com)                                           | `yandex`           | `YANDEX_PDD_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/yandex)           |
| [Zone.ee](https://www.zone.ee)                                         | `zoneee`           | `ZONEEE_API_USER`, `ZONEEE_API_KEY`                                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/zoneee)           |
| [ZoneEdit](https://www.zoneedit.com)                                   | `zoneedit`         | `ZONEEDIT_USER`, `ZONEEDIT_AUTH_TOKEN`                                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/zoneedit)         |
| [Zonomi](https://zonomi.com)                                           | `zonomi`           | `ZONOMI_API_KEY`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/zonomi)           |
| External Program                                                       | `exec`             | `EXEC_PATH`                                                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/exec)             |
| HTTP request                                                           | `httpreq`          | `HTTPREQ_ENDPOINT`, `HTTPREQ_MODE`, `HTTPREQ_USERNAME`, `HTTPREQ_PASSWORD` [^1]                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/httpreq)          |
| manual                                                                 | `manual`           | none, but you need to run Traefik interactively [^4], turn on debug log to see instructions and press <kbd>Enter</kbd>.                                                          |                                                                                 |

[^1]: More information about the HTTP message format can be found [here](https://go-acme.github.io/lego/dns/httpreq/).
[^2]: [Providing credentials to your application](https://cloud.google.com/docs/authentication/production).
[^3]: [google/default.go](https://github.com/golang/oauth2/blob/36a7019397c4c86cf59eeab3bc0d188bac444277/google/default.go#L61-L76)
[^4]: `docker stack` remark: there is no way to support terminal attached to container when deploying with `docker stack`, so you might need to run container with `docker run -it` to generate certificates using `manual` provider.
[^5]: The `Global API Key` needs to be used, not the `Origin CA Key`.
[^6]: As explained in the [LEGO hurricane configuration](https://go-acme.github.io/lego/dns/hurricane/#credentials), each domain or wildcard (record name) needs a token. So each update of record name must be followed by an update of the `HURRICANE_TOKENS` variable, and a restart of Traefik.

#### `resolvers`

Use custom DNS servers to resolve the FQDN authority.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      dnsChallenge:
        # ...
        resolvers:
          - "1.1.1.1:53"
          - "8.8.8.8:53"
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.dnsChallenge]
    # ...
    resolvers = ["1.1.1.1:53", "8.8.8.8:53"]
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.dnschallenge.resolvers=1.1.1.1:53,8.8.8.8:53
```

#### `propagation.delayBeforeChecks`

By default, the `provider` verifies the TXT record _before_ letting ACME verify.

You can delay this operation by specifying a delay (in seconds) with `delayBeforeChecks` (value must be greater than zero).

This option is useful when internal networks block external DNS queries.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      dnsChallenge:
        # ...
        propagation:
          # ...
          delayBeforeChecks: 2s
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.dnsChallenge]
    # ...
    [certificatesResolvers.myresolver.acme.dnsChallenge.propagation]
      # ...
      delayBeforeChecks = "2s"
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.dnschallenge.propagation.delayBeforeChecks=2s
```

#### `propagation.disableChecks`

Disables the challenge TXT record propagation checks, before notifying ACME that the DNS challenge is ready. 

Please note that disabling checks can prevent the challenge from succeeding.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      dnsChallenge:
        # ...
        propagation:
          # ...
          disableChecks: true
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.dnsChallenge]
    # ...
    [certificatesResolvers.myresolver.acme.dnsChallenge.propagation]
      # ...
      disableChecks = true
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.dnschallenge.propagation.disableChecks=true
```

#### `propagation.requireAllRNS`

Requires the challenge TXT record to be propagated to all recursive nameservers.

!!! note

    If you have disabled authoritative nameservers checks (with `propagation.disableANSChecks`),
    it is recommended to check all recursive nameservers instead.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      dnsChallenge:
        # ...
        propagation:
          # ...
          requireAllRNS: true
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.dnsChallenge]
    # ...
    [certificatesResolvers.myresolver.acme.dnsChallenge.propagation]
      # ...
      requireAllRNS = true
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.dnschallenge.propagation.requireAllRNS=true
```

#### `propagation.disableANSChecks`

Disables the challenge TXT record propagation checks against authoritative nameservers.

This option will skip the propagation check against the nameservers of the authority (SOA).

It should be used only if the nameservers of the authority are not reachable.

!!! note

    If you have disabled authoritative nameservers checks,
    it is recommended to check all recursive nameservers instead.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      dnsChallenge:
        # ...
        propagation:
          # ...
          disableANSChecks: true
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  [certificatesResolvers.myresolver.acme.dnsChallenge]
    # ...
    [certificatesResolvers.myresolver.acme.dnsChallenge.propagation]
      # ...
      disableANSChecks = true
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.dnschallenge.propagation.disableANSChecks=true
```

#### Wildcard Domains

[ACME V2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](#dnschallenge).

## External Account Binding

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

## More Configuration

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

`certificatesDuration` specifies the duration (in hours) of the certificates issued by the CA server. It is used to determine when to renew the certificate, but it **doesn't** define the duration of the certificates, that is up to the CA server. 

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

### `clientTimeout`

_Optional, Default=2m_

`clientTimeout` is the total timeout for a complete HTTP transaction (including TCP connection, sending request and receiving response) with the ACME server.
It defaults to 2 minutes.

!!! warning "This timeout encompasses the entire request-response cycle, including the response headers timeout. It must be at least `clientResponseHeaderTimeout`, otherwise the certificate resolver will fail to start."

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      clientTimeout: 1m
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  clientTimeout=1m
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.clientTimeout=1m
# ...
```

!!! warning
    This should not be confused with any timeouts used for validating challenges.

### `clientResponseHeaderTimeout`

_Optional, Default=30s_

`clientResponseHeaderTimeout` defines how long the HTTP client waits for response headers when communicating with the `caServer`.
It defaults to 30 seconds. 

!!! warning "It must be lower than `clientTimeout`, otherwise the certificate resolver will fail to start."

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      clientResponseHeaderTimeout: 1m
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  clientResponseHeaderTimeout=1m
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.clientResponseHeaderTimeout=1m
# ...
```

### `preferredChain`

_Optional, Default=""_

Preferred chain to use.

If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name.
If no match, the default offered chain will be used.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      preferredChain: 'ISRG Root X1'
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  preferredChain = "ISRG Root X1"
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.preferredChain=ISRG Root X1
# ...
```

### `profile`

_Optional, Default=""_

Certificate profile to use.

For more information, please check out the [Let's Encrypt blog post](https://letsencrypt.org/2025/01/09/acme-profiles/) about certificate profile selection.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      profile: tlsserver
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  profile = "tlsserver"
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.profile=tlsserver
# ...
```

### `emailAddresses`

_Optional, Default=""_

CSR email addresses to use.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      emailAddresses:
        - foo@example.com
        - bar@example.org
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  emailAddresses = ["foo@example.com", "bar@example.org"]
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.emailaddresses=foo@example.com,bar@example.org
# ...
```

### `keyType`

_Optional, Default="RSA4096"_

KeyType used for generating certificate private key. Allow value 'EC256', 'EC384', 'RSA2048', 'RSA4096', 'RSA8192'.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      keyType: 'RSA4096'
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  keyType = "RSA4096"
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.keyType=RSA4096
# ...
```

### `caCertificates`

_Optional, Default=[]_

The `caCertificates` option specifies the paths to PEM encoded CA Certificates that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      caCertificates:
        - path/certificates1.pem
        - path/certificates2.pem
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  caCertificates = [ "path/certificates1.pem", "path/certificates2.pem" ]
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.caCertificates="path/certificates1.pem,path/certificates2.pem"
# ...
```

??? note "LEGO Environment Variable"

    It can be defined globally by using the environment variable `LEGO_CA_CERTIFICATES`.
    This environment variable is neither a fallback nor an override of the configuration option.

### `caSystemCertPool`

_Optional, Default=false_

The `caSystemCertPool` option defines if the certificates pool must use a copy of the system cert pool.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      caSystemCertPool: true
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  caSystemCertPool = true
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.caSystemCertPool=true
# ...
```

??? note "LEGO Environment Variable"

    It can be defined globally by using the environment variable `LEGO_CA_SYSTEM_CERT_POOL`.
    `LEGO_CA_SYSTEM_CERT_POOL` is ignored if `LEGO_CA_CERTIFICATES` is not set or empty.
    This environment variable is neither a fallback nor an override of the configuration option.

### `caServerName`

_Optional, Default=""_

The `caServerName` option specifies the CA server name that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

```yaml tab="File (YAML)"
certificatesResolvers:
  myresolver:
    acme:
      # ...
      caServerName: "my-server"
      # ...
```

```toml tab="File (TOML)"
[certificatesResolvers.myresolver.acme]
  # ...
  caServerName = "my-server"
  # ...
```

```bash tab="CLI"
# ...
--certificatesresolvers.myresolver.acme.caServerName="my-server"
# ...
```

??? note "LEGO Environment Variable"

    It can be defined globally by using the environment variable `LEGO_CA_SERVER_NAME`.
    `LEGO_CA_SERVER_NAME` is ignored if `LEGO_CA_CERTIFICATES` is not set or empty.
    This environment variable is neither a fallback nor an override of the configuration option.

## Fallback

If Let's Encrypt is not reachable, the following certificates will apply:

  1. Previously generated ACME certificates (before downtime)
  2. Expired ACME certificates
  3. Provided certificates

!!! important
    For new (sub)domains which need Let's Encrypt authentication, the default Traefik certificate will be used until Traefik is restarted.

{!traefik-for-business-applications.md!}
