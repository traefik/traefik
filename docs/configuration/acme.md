# ACME (Let's Encrypt) configuration

See also [Let's Encrypt examples](/user-guide/examples/#lets-encrypt-support) and [Docker & Let's Encrypt user guide](/user-guide/docker-and-lets-encrypt).

## Configuration

```toml
# Sample entrypoint configuration when using ACME.
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
```

```toml
# Enable ACME (Let's Encrypt): automatic SSL.
[acme]

# Email address used for registration.
#
# Required
#
email = "test@traefik.io"

# File used for certificates storage.
#
# Optional (Deprecated)
#
#storageFile = "acme.json"

# File or key used for certificates storage.
#
# Required
#
storage = "acme.json"
# or `storage = "traefik/acme/account"` if using KV store.

# Entrypoint to proxy acme apply certificates to.
#
# Required
#
entryPoint = "https"

# Deprecated, replaced by [acme.dnsChallenge].
#
# Optional.
#
# dnsProvider = "digitalocean"

# Deprecated, replaced by [acme.dnsChallenge.delayBeforeCheck].
#
# Optional
# Default: 0
#
# delayDontCheckDNS = 0

# If true, display debug log messages from the acme client library.
#
# Optional
# Default: false
#
# acmeLogging = true

# Enable on demand certificate generation.
#
# Optional (Deprecated)
# Default: false
#
# onDemand = true

# Enable certificate generation on frontends Host rules.
#
# Optional
# Default: false
#
# onHostRule = true

# CA server to use.
# - Uncomment the line to run on the staging let's encrypt server.
# - Leave comment to go to prod.
#
# Optional
# Default: "https://acme-v02.api.letsencrypt.org/directory"
#
# caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"

# Domains list.
# Only domains defined here can generate wildcard certificates.
#
# [[acme.domains]]
#   main = "local1.com"
#   sans = ["test1.local1.com", "test2.local1.com"]
# [[acme.domains]]
#   main = "local2.com"
#   sans = ["test1.local2.com", "test2.local2.com"]
# [[acme.domains]]
#   main = "local3.com"
# [[acme.domains]]
#   main = "local4.com"

# Use a HTTP-01 acme challenge.
#
# Optional but recommend
#
[acme.httpChallenge]

  # EntryPoint to use for the HTTP-01 challenges.
  #
  # Required
  #
  entryPoint = "http"

# Use a DNS-01/DNS-01 acme challenge rather than HTTP-01 challenge.
# Note : Mandatory for wildcard certificates generation.
#
# Optional
#
# [acme.dnsChallenge]

  # Provider used.
  #
  # Required
  #
  # provider = "digitalocean"

  # By default, the provider will verify the TXT DNS challenge record before letting ACME verify.
  # If delayBeforeCheck is greater than zero, avoid this & instead just wait so many seconds.
  # Useful if internal networks block external DNS queries.
  #
  # Optional
  # Default: 0
  #
  # delayBeforeCheck = 0
```

!!! note
    If `HTTP-01` challenge is used, `acme.httpChallenge.entryPoint` has to be defined and reachable by Let's Encrypt through the port 80.
    These are Let's Encrypt limitations as described on the [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72).

!!! note
    Wildcard certificates can be generated only if `acme.dnsChallenge` option is enable.

### Let's Encrypt downtime

Let's Encrypt functionality will be limited until Træfik is restarted.

If Let's Encrypt is not reachable, these certificates will be used :

  - ACME certificates already generated before downtime
  - Expired ACME certificates
  - Provided certificates

!!! note
    Default Træfik certificate will be used instead of ACME certificates for new (sub)domains (which need Let's Encrypt challenge).

### `storage`

```toml
[acme]
# ...
storage = "acme.json"
# ...
```

The `storage` option sets where are stored your ACME certificates.

There are two kind of `storage` :

- a JSON file,
- a KV store entry.

!!! danger "DEPRECATED"
    `storage` replaces `storageFile` which is deprecated.

!!! note
    During Træfik configuration migration from a configuration file to a KV store (thanks to `storeconfig` subcommand as described [here](/user-guide/kv-config/#store-configuration-in-key-value-store)), if ACME certificates have to be migrated too, use both `storageFile` and `storage`.

    - `storageFile` will contain the path to the `acme.json` file to migrate.
    - `storage` will contain the key where the certificates will be stored.

#### Store data in a file

ACME certificates can be stored in a JSON file which with the `600` right mode.

There are two ways to store ACME certificates in a file from Docker:

- create a file on your host and mount it as a volume:
```toml
storage = "acme.json"
```
```bash
docker run -v "/my/host/acme.json:acme.json" traefik
```
- mount the folder containing the file as a volume
```toml
storage = "/etc/traefik/acme/acme.json"
```
```bash
docker run -v "/my/host/acme:/etc/traefik/acme" traefik
```

!!! warning
    This file cannot be shared per many instances of Træfik at the same time.
    If you have to use Træfik cluster mode, please use [a KV Store entry](/configuration/acme/#storage-kv-entry).

#### Store data in a KV store entry

ACME certificates can be stored in a KV Store entry.

```toml
storage = "traefik/acme/account"
```

**This kind of storage is mandatory in cluster mode.**

Because KV stores (like Consul) have limited entries size, the certificates list is compressed before to be set in a KV store entry.

!!! note
    It's possible to store up to approximately 100 ACME certificates in Consul.

### `httpChallenge`

Use `HTTP-01` challenge to generate/renew ACME certificates.

The redirection is fully compatible with the HTTP-01 challenge.
You can use redirection with HTTP-01 challenge without problem.

```toml
[acme]
# ...
entryPoint = "https"
[acme.httpChallenge]
  entryPoint = "http"
```

#### `entryPoint`

Specify the entryPoint to use during the challenges.

```toml
defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
# ...

[acme]
  # ...
  entryPoint = "https"
  [acme.httpChallenge]
    entryPoint = "http"
```

!!! note
    `acme.httpChallenge.entryPoint` has to be reachable by Let's Encrypt through the port 80.
    It's a Let's Encrypt limitation as described on the [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72).

### `dnsChallenge`

Use `DNS-01/DNS-01` challenge to generate/renew ACME certificates.

```toml
[acme]
# ...
[acme.dnsChallenge]
  provider = "digitalocean"
  delayBeforeCheck = 0
# ...
```

!!! note
    ACME wildcard certificates can only be generated thanks to a `DNS-01` challenge.

#### `provider`

Select the provider that matches the DNS domain that will host the challenge TXT record, and provide environment variables to enable setting it:

| Provider Name                                          | Provider code  | Configuration                                                                                                             |
|--------------------------------------------------------|----------------|---------------------------------------------------------------------------------------------------------------------------|
| [Auroradns](https://www.pcextreme.com/aurora/dns)      | `auroradns`    | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                         |
| [Azure](https://azure.microsoft.com/services/dns/)     | `azure`        | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`, `AZURE_RESOURCE_GROUP`              |
| [Blue Cat](https://www.bluecatnetworks.com/)           | `bluecat`      | `BLUECAT_SERVER_URL`, `BLUECAT_USER_NAME`, `BLUECAT_PASSWORD`, `BLUECAT_CONFIG_NAME`, `BLUECAT_DNS_VIEW`                  |
| [Cloudflare](https://www.cloudflare.com)               | `cloudflare`   | `CLOUDFLARE_EMAIL`, `CLOUDFLARE_API_KEY` - The Cloudflare `Global API Key` needs to be used and not the `Origin CA Key`   |
| [CloudXNS](https://www.cloudxns.net)                   | `cloudxns`     | `CLOUDXNS_API_KEY`, `CLOUDXNS_SECRET_KEY`                                                                                 |
| [DigitalOcean](https://www.digitalocean.com)           | `digitalocean` | `DO_AUTH_TOKEN`                                                                                                           |
| [DNSimple](https://dnsimple.com)                       | `dnsimple`     | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                               |
| [DNS Made Easy](https://dnsmadeeasy.com)               | `dnsmadeeasy`  | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                    |
| [DNSPod](http://www.dnspod.net/)                       | `dnspod`       | `DNSPOD_API_KEY`                                                                                                          |
| [Duck DNS](https://www.duckdns.org/)                   | `duckdns`      | `DUCKDNS_TOKEN`                                                                                                           |
| [Dyn](https://dyn.com)                                 | `dyn`          | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                      |
| External Program                                       | `exec`         | `EXEC_PATH`                                                                                                               |
| [Exoscale](https://www.exoscale.ch)                    | `exoscale`     | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                            |
| [Fast DNS](https://www.akamai.com/)                    | `fastdns`      | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                    |
| [Gandi](https://www.gandi.net)                         | `gandi`        | `GANDI_API_KEY`                                                                                                           |
| [Gandi V5](http://doc.livedns.gandi.net)               | `gandiv5`      | `GANDIV5_API_KEY`                                                                                                         |
| [Glesys](https://glesys.com/)                          | `glesys`       | `GLESYS_API_USER`, `GLESYS_API_KEY`, `GLESYS_DOMAIN`                                                                      |
| [GoDaddy](https://godaddy.com/domains)                 | `godaddy`      | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                   |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/) | `gcloud`       | `GCE_PROJECT`, `GCE_SERVICE_ACCOUNT_FILE`                                                                                 |
| [Lightsail](https://aws.amazon.com/lightsail/)         | `lightsail`    | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `DNS_ZONE`                                                                  |
| [Linode](https://www.linode.com)                       | `linode`       | `LINODE_API_KEY`                                                                                                          |
| manual                                                 | -              | none, but run Træfik interactively & turn on `acmeLogging` to see instructions & press <kbd>Enter</kbd>.                  |
| [Namecheap](https://www.namecheap.com)                 | `namecheap`    | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                                 |
| [name.com](https://www.name.com/)                      | `namedotcom`   | `NAMECOM_USERNAME`, `NAMECOM_API_TOKEN`, `NAMECOM_SERVER`                                                                 |
| [Ns1](https://ns1.com/)                                | `ns1`          | `NS1_API_KEY`                                                                                                             |
| [Open Telekom Cloud](https://cloud.telekom.de/en/)     | `otc`          | `OTC_DOMAIN_NAME`, `OTC_USER_NAME`, `OTC_PASSWORD`, `OTC_PROJECT_NAME`, `OTC_IDENTITY_ENDPOINT`                           |
| [OVH](https://www.ovh.com)                             | `ovh`          | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`                                       |
| [PowerDNS](https://www.powerdns.com)                   | `pdns`         | `PDNS_API_KEY`, `PDNS_API_URL`                                                                                            |
| [Rackspace](https://www.rackspace.com/cloud/dns)       | `rackspace`    | `RACKSPACE_USER`, `RACKSPACE_API_KEY`                                                                                     |
| [RFC2136](https://tools.ietf.org/html/rfc2136)         | `rfc2136`      | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                 |
| [Route 53](https://aws.amazon.com/route53/)            | `route53`      | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `AWS_HOSTED_ZONE_ID` or configured user/instance IAM profile. |
| [VULTR](https://www.vultr.com)                         | `vultr`        | `VULTR_API_KEY`                                                                                                           |

#### `delayBeforeCheck`

By default, the `provider` will verify the TXT DNS challenge record before letting ACME verify.  
If `delayBeforeCheck` is greater than zero, avoid this & instead just wait so many seconds.

Useful if internal networks block external DNS queries.

!!! note
    This field has no sense if a `provider` is not defined.

### `onDemand` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated.

```toml
[acme]
# ...
onDemand = true
# ...
```

Enable on demand certificate.

This will request a certificate from Let's Encrypt during the first TLS handshake for a host name that does not yet have a certificate.

!!! warning
    TLS handshakes will be slow when requesting a host name certificate for the first time, this can lead to DoS attacks.

!!! warning
    Take note that Let's Encrypt have [rate limiting](https://letsencrypt.org/docs/rate-limits).

### `onHostRule`

```toml
[acme]
# ...
onHostRule = true
# ...
```

Enable certificate generation on frontends `Host` rules (for frontends wired on the `acme.entryPoint`).

This will request a certificate from Let's Encrypt for each frontend with a Host rule.

For example, a rule `Host:test1.traefik.io,test2.traefik.io` will request a certificate with main domain `test1.traefik.io` and SAN `test2.traefik.io`.

!!! warning
    `onHostRule` option can not be used to generate wildcard certificates.
    Refer to [the wildcard generation section](/configuration/acme/#wildcard-domain) for more information.

### `caServer`

```toml
[acme]
# ...
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
# ...
```

CA server to use.

- Uncomment the line to run on the staging Let's Encrypt server.
- Leave comment to go to prod.

### `domains`

```toml
[acme]
# ...
[[acme.domains]]
  main = "local1.com"
  sans = ["test1.local1.com", "test2.local1.com"]
[[acme.domains]]
  main = "local2.com"
  sans = ["test1.local2.com", "test2.local2.com"]
[[acme.domains]]
  main = "local3.com"
[[acme.domains]]
  main = "*.local4.com"
  sans = ["local4.com", "test1.test1.local4.com"]
# ...
```

#### Wildcard domains

Wildcard domain has to be defined as a main domain.
All domains must have A/AAAA records pointing to Træfik.

Due to ACME limitation, it's not possible to define a wildcard as a SAN (alternative domains).
It's neither possible to define a wildcard on a wildcard domain (for example `*.*.local.com`).

!!! warning
    Note that Let's Encrypt has [rate limiting](https://letsencrypt.org/docs/rate-limits).

Each domain & SANs will lead to a certificate request.

#### Others domains

You can provide SANs (alternative domains) to each main domain.
All domains must have A/AAAA records pointing to Træfik.

!!! warning
    Take note that Let's Encrypt have [rate limiting](https://letsencrypt.org/docs/rate-limits).

Each domain & SANs will lead to a certificate request.

### `dnsProvider` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated, use [dnsChallenge.provider](/configuration/acme/#dnschallenge) instead.

### `delayDontCheckDNS` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated, use [dnsChallenge.delayBeforeCheck](/configuration/acme/#dnschallenge) instead.

## Wildcard certificates

[ACME V2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) allows wildcard certificate support.
However, this feature needs a specific configuration.

### DNS-01 Challenge

As described in [Let's Encrypt post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605), wildcard certificates can only be generated through a `DNS-01` Challenge.
This challenge is linked to the Træfik option `acme.dnsChallenge`.

```toml
[acme]
# ...
[acme.dnsChallenge]
  provider = "digitalocean"
  delayBeforeCheck = 0
# ...
```

For more information about this option, please refer to the [dnsChallenge section](/configuration/acme/#dnschallenge).

### Wildcard domain

Wildcard domains can currently be provided only by to the `acme.domains` option.

```toml
[acme]
# ...
[[acme.domains]]
  main = "*.local1.com"
  sans = ["local1.com"]
[[acme.domains]]
  main = "*.local2.com"
# ...
```

For more information about this option, please refer to the [domains section](/configuration/acme/#domains).

### Limitations

Let's Encrypt wildcard support have some limitations to take into account :

- Wildcard domain can not be a SAN (alternative domain),
- Wildcard domain on a wildcard domain is forbidden (for example `*.*.local.com`),
- A DNS-01 Challenge is executed for each domain (CN and SANs), DNS provider can not manage correctly this behavior as explained in the [DNS provider support section](/configuration/acme/#dns-provider-support)


### DNS provider support

All DNS providers allow creating ACME wildcard certificates.
However, many troubles can appear for wildcard domains with SANs.

If a wildcard domain is defined with it root domain as SAN, as described below, 2 DNS-01 Challenges will be executed.

```toml
[acme]
# ...
[[acme.domains]]
  main = "*.local1.com"
  sans = ["local1.com"]
# ...
```

When a DNS-01 Challenge is done, Let's Encrypt checks if a TXT record is created with a given name and a given value.
When a certificate is generated for a wildcard domain is defined with it root domain as SAN, the requested TXT record name for both the wildcard domain and the root domain is the same.

The [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) allows this behavior.
But all DNS providers keep TXT records values in a cache with a TTL.
In function of the parameters given by the Træfik ACME client library ([LEGO](https://github.com/xenolf/lego)), the TXT record TTL can be superior to challenge Timeout.
In that event, the DNS-01 Challenge will not work correctly.
 
[LEGO](https://github.com/xenolf/lego) will involve in the way to be adapted to all of DNS providers.
Meanwhile, the table described below contains all the DNS providers supported by Træfik and indicates if they allow generating certificates for a wildcard domain and its root domain.
Do not hesitate to complete it.

| Provider Name                                          | Provider code  | Wildcard and Root Domain Support |
|--------------------------------------------------------|----------------|----------------------------------|
| [Auroradns](https://www.pcextreme.com/aurora/dns)      | `auroradns`    | Not tested yet                   |
| [Azure](https://azure.microsoft.com/services/dns/)     | `azure`        | Not tested yet                   |
| [Blue Cat](https://www.bluecatnetworks.com/)           | `bluecat`      | Not tested yet                   |
| [Cloudflare](https://www.cloudflare.com)               | `cloudflare`   | YES                              |
| [CloudXNS](https://www.cloudxns.net)                   | `cloudxns`     | Not tested yet                   |
| [DigitalOcean](https://www.digitalocean.com)           | `digitalocean` | YES                              |
| [DNSimple](https://dnsimple.com)                       | `dnsimple`     | Not tested yet                   |
| [DNS Made Easy](https://dnsmadeeasy.com)               | `dnsmadeeasy`  | Not tested yet                   |
| [DNSPod](http://www.dnspod.net/)                       | `dnspod`       | Not tested yet                   |
| [Duck DNS](https://www.duckdns.org/)                   | `duckdns`      | Not tested yet                   |
| [Dyn](https://dyn.com)                                 | `dyn`          | Not tested yet                   |
| External Program                                       | `exec`         | Not tested yet                   |
| [Exoscale](https://www.exoscale.ch)                    | `exoscale`     | Not tested yet                   |
| [Fast DNS](https://www.akamai.com/)                    | `fastdns`      | Not tested yet                   |
| [Gandi](https://www.gandi.net)                         | `gandi`        | Not tested yet                   |
| [Gandi V5](http://doc.livedns.gandi.net)               | `gandiv5`      | Not tested yet                   |
| [Glesys](https://glesys.com/)                          | `glesys`       | Not tested yet                   |
| [GoDaddy](https://godaddy.com/domains)                 | `godaddy`      | Not tested yet                   |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/) | `gcloud`       | YES                              |
| [Lightsail](https://aws.amazon.com/lightsail/)         | `lightsail`    | Not tested yet                   |
| [Linode](https://www.linode.com)                       | `linode`       | Not tested yet                   |
| manual                                                 | -              | YES                              |
| [Namecheap](https://www.namecheap.com)                 | `namecheap`    | Not tested yet                   |
| [name.com](https://www.name.com/)                      | `namedotcom`   | Not tested yet                   |
| [Ns1](https://ns1.com/)                                | `ns1`          | Not tested yet                   |
| [Open Telekom Cloud](https://cloud.telekom.de/en/)     | `otc`          | Not tested yet                   |
| [OVH](https://www.ovh.com)                             | `ovh`          | YES                              |
| [PowerDNS](https://www.powerdns.com)                   | `pdns`         | Not tested yet                   |
| [Rackspace](https://www.rackspace.com/cloud/dns)       | `rackspace`    | Not tested yet                   |
| [RFC2136](https://tools.ietf.org/html/rfc2136)         | `rfc2136`      | Not tested yet                   |
| [Route 53](https://aws.amazon.com/route53/)            | `route53`      | YES                              |
| [VULTR](https://www.vultr.com)                         | `vultr`        | Not tested yet                   |

## ACME V2 migration

During migration from ACME V1 to ACME V2 with a storage file, a backup is created with the content of the ACME V1 file.
To obtain the name of the backup file, Træfik concatenates the option `acme.storage` and the suffix `.bak`.

For example : if `acme.storage` value is `/etc/traefik/acme/acme.json`, the backup file will be named `/etc/traefik/acme/acme.json.bak`.

!!! note
    When Træfik is launched in a container, do not forget to create a volume of the parent folder to get the backup file on the host.
    Otherwise, the backup file will be deleted when the container will be stopped and Træfik will not generate it again.
