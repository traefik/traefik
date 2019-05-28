# ACME (Let's Encrypt) Configuration

See [Let's Encrypt examples](/user-guide/examples/#lets-encrypt-support) and [Docker & Let's Encrypt user guide](/user-guide/docker-and-lets-encrypt) as well.

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

# If true, override certificates in key-value store when using storeconfig.
#
# Optional
# Default: false
#
# overrideCertificates = true

# Deprecated. Enable on demand certificate generation.
#
# Optional
# Default: false
#
# onDemand = true

# Enable certificate generation on frontends host rules.
#
# Optional
# Default: false
#
# onHostRule = true

# CA server to use.
# Uncomment the line to use Let's Encrypt's staging server,
# leave commented to go to prod.
#
# Optional
# Default: "https://acme-v02.api.letsencrypt.org/directory"
#
# caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"

# KeyType to use.
#
# Optional
# Default: "RSA4096"
#
# Available values : "EC256", "EC384", "RSA2048", "RSA4096", "RSA8192"
#
# KeyType = "RSA4096"

# Use a TLS-ALPN-01 ACME challenge.
#
# Optional (but recommended)
#
[acme.tlsChallenge]

# Use a HTTP-01 ACME challenge.
#
# Optional
#
# [acme.httpChallenge]

  # EntryPoint to use for the HTTP-01 challenges.
  #
  # Required
  #
  # entryPoint = "http"

# Use a DNS-01 ACME challenge rather than HTTP-01 challenge.
# Note: mandatory for wildcard certificate generation.
#
# Optional
#
# [acme.dnsChallenge]

  # DNS provider used.
  #
  # Required
  #
  # provider = "digitalocean"

  # By default, the provider will verify the TXT DNS challenge record before letting ACME verify.
  # If delayBeforeCheck is greater than zero, this check is delayed for the configured duration in seconds.
  # Useful if internal networks block external DNS queries.
  #
  # Optional
  # Default: 0
  #
  # delayBeforeCheck = 0

  # Use following DNS servers to resolve the FQDN authority.
  #
  # Optional
  # Default: empty
  #
  # resolvers = ["1.1.1.1:53", "8.8.8.8:53"]

  # Disable the DNS propagation checks before notifying ACME that the DNS challenge is ready.
  #
  # NOT RECOMMENDED:
  # Increase the risk of reaching Let's Encrypt's rate limits.
  #
  # Optional
  # Default: false
  #
  # disablePropagationCheck = true

# Domains list.
# Only domains defined here can generate wildcard certificates.
# The certificates for these domains are negotiated at traefik startup only.
#
# [[acme.domains]]
#   main = "local1.com"
#   sans = ["test1.local1.com", "test2.local1.com"]
# [[acme.domains]]
#   main = "local2.com"
# [[acme.domains]]
#   main = "*.local3.com"
#   sans = ["local3.com", "test1.test1.local3.com"]
```

### `caServer`

The CA server to use.

This example shows the usage of Let's Encrypt's staging server:

```toml
[acme]
# ...
caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
# ...
```

### ACME Challenge

#### `tlsChallenge`

Use the `TLS-ALPN-01` challenge to generate and renew ACME certificates by provisioning a TLS certificate.

```toml
[acme]
# ...
entryPoint = "https"
[acme.tlsChallenge]
```

!!! note
    If the `TLS-ALPN-01` challenge is used, `acme.entryPoint` has to be reachable by Let's Encrypt through port 443.
    This is a Let's Encrypt limitation as described on the [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72).

#### `httpChallenge`

Use the `HTTP-01` challenge to generate and renew ACME certificates by provisioning a HTTP resource under a well-known URI.

Redirection is fully compatible with the `HTTP-01` challenge.

```toml
[acme]
# ...
entryPoint = "https"
[acme.httpChallenge]
  entryPoint = "http"
```

!!! note
    If the `HTTP-01` challenge is used, `acme.httpChallenge.entryPoint` has to be defined and reachable by Let's Encrypt through port 80.
    This is a Let's Encrypt limitation as described on the [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72).

##### `entryPoint`

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
    `acme.httpChallenge.entryPoint` has to be reachable through port 80. It's a Let's Encrypt limitation as described on the [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72).

#### `dnsChallenge`

Use the `DNS-01` challenge to generate and renew ACME certificates by provisioning a DNS record.

```toml
[acme]
# ...
[acme.dnsChallenge]
  provider = "digitalocean"
  delayBeforeCheck = 0
# ...
```

##### `delayBeforeCheck`

By default, the `provider` will verify the TXT DNS challenge record before letting ACME verify.
If `delayBeforeCheck` is greater than zero, this check is delayed for the configured duration in seconds.

Useful if internal networks block external DNS queries.

!!! note
    A `provider` is mandatory.

##### `provider`

Here is a list of supported `provider`s, that can automate the DNS verification, along with the required environment variables and their [wildcard & root domain support](/configuration/acme/#wildcard-domains) for each.
Do not hesitate to complete it.
Every lego environment variable can be overridden by their respective `_FILE` counterpart, which should have a filepath to a file that contains the secret as its value.
For example, `CF_API_EMAIL_FILE=/run/secrets/traefik_cf-api-email` could be used to provide a Cloudflare API email address as a Docker secret named `traefik_cf-api-email`.

| Provider Name                                               | Provider Code  | Environment Variables                                                                                                                       | Wildcard & Root Domain Support |
|-------------------------------------------------------------|----------------|---------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------|
| [ACME DNS](https://github.com/joohoi/acme-dns)              | `acme-dns`     | `ACME_DNS_API_BASE`, `ACME_DNS_STORAGE_PATH`                                                                                                | Not tested yet                 |
| [Alibaba Cloud](https://www.vultr.com)                      | `alidns`       | `ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION_ID`                                                                          | Not tested yet                 |
| [Auroradns](https://www.pcextreme.com/aurora/dns)           | `auroradns`    | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                                           | Not tested yet                 |
| [Azure](https://azure.microsoft.com/services/dns/)          | `azure`        | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`, `AZURE_RESOURCE_GROUP`, `[AZURE_METADATA_ENDPOINT]`   | Not tested yet                 |
| [Bindman](https://github.com/labbsr0x/bindman-dns-webhook)  | `bindman`      | `BINDMAN_MANAGER_ADDRESS`                                                                                                                   | YES                            |
| [Blue Cat](https://www.bluecatnetworks.com/)                | `bluecat`      | `BLUECAT_SERVER_URL`, `BLUECAT_USER_NAME`, `BLUECAT_PASSWORD`, `BLUECAT_CONFIG_NAME`, `BLUECAT_DNS_VIEW`                                    | Not tested yet                 |
| [ClouDNS](https://www.cloudns.net/)                         | `cloudns`      | `CLOUDNS_AUTH_ID`, `CLOUDNS_AUTH_PASSWORD`                                                                                                  | YES                            |
| [Cloudflare](https://www.cloudflare.com)                    | `cloudflare`   | `CF_API_EMAIL`, `CF_API_KEY` - The `Global API Key` needs to be used, not the `Origin CA Key`                                               | YES                            |
| [CloudXNS](https://www.cloudxns.net)                        | `cloudxns`     | `CLOUDXNS_API_KEY`, `CLOUDXNS_SECRET_KEY`                                                                                                   | Not tested yet                 |
| [ConoHa](https://www.conoha.jp)                             | `conoha`       | `CONOHA_TENANT_ID`, `CONOHA_API_USERNAME`, `CONOHA_API_PASSWORD`                                                                            | YES                            |
| [DigitalOcean](https://www.digitalocean.com)                | `digitalocean` | `DO_AUTH_TOKEN`                                                                                                                             | YES                            |
| [DNSimple](https://dnsimple.com)                            | `dnsimple`     | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                                                 | YES                            |
| [DNS Made Easy](https://dnsmadeeasy.com)                    | `dnsmadeeasy`  | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                                      | Not tested yet                 |
| [DNSPod](https://www.dnspod.com/)                           | `dnspod`       | `DNSPOD_API_KEY`                                                                                                                            | Not tested yet                 |
| [Domain Offensive (do.de)](https://www.do.de/)              | `dode`         | `DODE_TOKEN`                                                                                                                                | YES                            |
| [DreamHost](https://www.dreamhost.com/)                     | `dreamhost`    | `DREAMHOST_API_KEY`                                                                                                                         | YES                            |
| [Duck DNS](https://www.duckdns.org/)                        | `duckdns`      | `DUCKDNS_TOKEN`                                                                                                                             | YES                            |
| [Dyn](https://dyn.com)                                      | `dyn`          | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                                        | Not tested yet                 |
| [EasyDNS](https://easydns.com/)                             | `easydns`      | `EASYDNS_TOKEN`, `EASYDNS_KEY`                                                                                                              | YES                            |
| External Program                                            | `exec`         | `EXEC_PATH`                                                                                                                                 | YES                            |
| [Exoscale](https://www.exoscale.com)                        | `exoscale`     | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                                              | YES                            |
| [Fast DNS](https://www.akamai.com/)                         | `fastdns`      | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                      | YES                            |
| [Gandi](https://www.gandi.net)                              | `gandi`        | `GANDI_API_KEY`                                                                                                                             | Not tested yet                 |
| [Gandi v5](http://doc.livedns.gandi.net)                    | `gandiv5`      | `GANDIV5_API_KEY`                                                                                                                           | YES                            |
| [Glesys](https://glesys.com/)                               | `glesys`       | `GLESYS_API_USER`, `GLESYS_API_KEY`, `GLESYS_DOMAIN`                                                                                        | Not tested yet                 |
| [GoDaddy](https://godaddy.com/domains)                      | `godaddy`      | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                                     | Not tested yet                 |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/)      | `gcloud`       | `GCE_PROJECT`, Application Default Credentials (2) (3), [`GCE_SERVICE_ACCOUNT_FILE`]                                                        | YES                            |
| [hosting.de](https://www.hosting.de)                        | `hostingde`    | `HOSTINGDE_API_KEY`, `HOSTINGDE_ZONE_NAME`                                                                                                  | YES                            |
| HTTP request                                                | `httpreq`      | `HTTPREQ_ENDPOINT`, `HTTPREQ_MODE`, `HTTPREQ_USERNAME`, `HTTPREQ_PASSWORD` (1)                                                              | YES                            |
| [IIJ](https://www.iij.ad.jp/)                               | `iij`          | `IIJ_API_ACCESS_KEY`, `IIJ_API_SECRET_KEY`, `IIJ_DO_SERVICE_CODE`                                                                           | Not tested yet                 |
| [INWX](https://www.inwx.de/en)                              | `inwx`         | `INWX_USERNAME`, `INWX_PASSWORD`                                                                                                            | YES                            |
| [Joker.com](https://joker.com)                              | `joker`        | `JOKER_API_KEY`                                                                                                                             | YES                            |
| [Lightsail](https://aws.amazon.com/lightsail/)              | `lightsail`    | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `DNS_ZONE`                                                                                    | Not tested yet                 |
| [Linode](https://www.linode.com)                            | `linode`       | `LINODE_API_KEY`                                                                                                                            | Not tested yet                 |
| [Linode v4](https://www.linode.com)                         | `linodev4`     | `LINODE_TOKEN`                                                                                                                              | Not tested yet                 |
| manual                                                      | -              | none, but you need to run Traefik interactively, turn on `acmeLogging` to see instructions and press <kbd>Enter</kbd>.                      | YES                            |
| [MyDNS.jp](https://www.mydns.jp/)                           | `mydnsjp`      | `MYDNSJP_MASTER_ID`, `MYDNSJP_PASSWORD`                                                                                                     | YES                            |
| [Namecheap](https://www.namecheap.com)                      | `namecheap`    | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                                                   | YES                            |
| [name.com](https://www.name.com/)                           | `namedotcom`   | `NAMECOM_USERNAME`, `NAMECOM_API_TOKEN`, `NAMECOM_SERVER`                                                                                   | Not tested yet                 |
| [Netcup](https://www.netcup.eu/)                            | `netcup`       | `NETCUP_CUSTOMER_NUMBER`, `NETCUP_API_KEY`, `NETCUP_API_PASSWORD`                                                                           | Not tested yet                 |
| [NIFCloud](https://cloud.nifty.com/service/dns.htm)         | `nifcloud`     | `NIFCLOUD_ACCESS_KEY_ID`, `NIFCLOUD_SECRET_ACCESS_KEY`                                                                                      | Not tested yet                 |
| [Ns1](https://ns1.com/)                                     | `ns1`          | `NS1_API_KEY`                                                                                                                               | Not tested yet                 |
| [Open Telekom Cloud](https://cloud.telekom.de)              | `otc`          | `OTC_DOMAIN_NAME`, `OTC_USER_NAME`, `OTC_PASSWORD`, `OTC_PROJECT_NAME`, `OTC_IDENTITY_ENDPOINT`                                             | Not tested yet                 |
| [OVH](https://www.ovh.com)                                  | `ovh`          | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`                                                         | YES                            |
| [Openstack Designate](https://docs.openstack.org/designate) | `designate`    | `OS_AUTH_URL`, `OS_USERNAME`, `OS_PASSWORD`, `OS_TENANT_NAME`, `OS_REGION_NAME`                                                             | YES                            |
| [Oracle Cloud](https://cloud.oracle.com/home)               | `oraclecloud`  | `OCI_COMPARTMENT_OCID`, `OCI_PRIVKEY_FILE`, `OCI_PRIVKEY_PASS`, `OCI_PUBKEY_FINGERPRINT`, `OCI_REGION`, `OCI_TENANCY_OCID`, `OCI_USER_OCID` | YES                            |
| [PowerDNS](https://www.powerdns.com)                        | `pdns`         | `PDNS_API_KEY`, `PDNS_API_URL`                                                                                                              | Not tested yet                 |
| [Rackspace](https://www.rackspace.com/cloud/dns)            | `rackspace`    | `RACKSPACE_USER`, `RACKSPACE_API_KEY`                                                                                                       | Not tested yet                 |
| [RFC2136](https://tools.ietf.org/html/rfc2136)              | `rfc2136`      | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                                   | Not tested yet                 |
| [Route 53](https://aws.amazon.com/route53/)                 | `route53`      | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `[AWS_REGION]`, `[AWS_HOSTED_ZONE_ID]` or a configured user/instance IAM profile.             | YES                            |
| [Sakura Cloud](https://cloud.sakura.ad.jp/)                 | `sakuracloud`  | `SAKURACLOUD_ACCESS_TOKEN`, `SAKURACLOUD_ACCESS_TOKEN_SECRET`                                                                               | Not tested yet                 |
| [Selectel](https://selectel.ru/en/)                         | `selectel`     | `SELECTEL_API_TOKEN`                                                                                                                        | YES                            |
| [Stackpath](https://www.stackpath.com/)                     | `stackpath`    | `STACKPATH_CLIENT_ID`, `STACKPATH_CLIENT_SECRET`, `STACKPATH_STACK_ID`                                                                      | Not tested yet                 |
| [TransIP](https://www.transip.nl/)                          | `transip`      | `TRANSIP_ACCOUNT_NAME`, `TRANSIP_PRIVATE_KEY_PATH`                                                                                          | YES                            |
| [VegaDNS](https://github.com/shupp/VegaDNS-API)             | `vegadns`      | `SECRET_VEGADNS_KEY`, `SECRET_VEGADNS_SECRET`, `VEGADNS_URL`                                                                                | Not tested yet                 |
| [Vscale](https://vscale.io/)                                | `vscale`       | `VSCALE_API_TOKEN`                                                                                                                          | YES                            |
| [VULTR](https://www.vultr.com)                              | `vultr`        | `VULTR_API_KEY`                                                                                                                             | Not tested yet                 |
| [Zone.ee](https://www.zone.ee)                              | `zoneee`       | `ZONEEE_API_USER`, `ZONEEE_API_KEY`                                                                                                         | YES                            |

- (1): more information about the HTTP message format can be found [here](https://go-acme.github.io/lego/dns/httpreq/)
- (2): https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application
- (3): https://github.com/golang/oauth2/blob/36a7019397c4c86cf59eeab3bc0d188bac444277/google/default.go#L61-L76

#### `resolvers`

Use custom DNS servers to resolve the FQDN authority.

```toml
[acme]
# ...
[acme.dnsChallenge]
  # ...
  resolvers = ["1.1.1.1:53", "8.8.8.8:53"]
```

### `domains`

You can provide SANs (alternative domains) to each main domain.
All domains must have A/AAAA records pointing to Traefik.
Each domain & SAN will lead to a certificate request.

!!! note
    The certificates for the domains listed in `acme.domains` are negotiated at traefik startup only.

```toml
[acme]
# ...
[[acme.domains]]
  main = "local1.com"
  sans = ["test1.local1.com", "test2.local1.com"]
[[acme.domains]]
  main = "local2.com"
[[acme.domains]]
  main = "*.local3.com"
  sans = ["local3.com", "test1.test1.local3.com"]
# ...
```

!!! warning
    Take note that Let's Encrypt applies [rate limiting](https://letsencrypt.org/docs/rate-limits).

!!! note
    Wildcard certificates can only be verified through a `DNS-01` challenge.

#### Wildcard Domains

[ACME V2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) allows wildcard certificate support.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](/configuration/acme/#dnschallenge).

```toml
[acme]
# ...
[[acme.domains]]
  main = "*.local1.com"
  sans = ["local1.com"]
# ...
```

It is not possible to request a double wildcard certificate for a domain (for example `*.*.local.com`).
Most likely the root domain should receive a certificate too, so it needs to be specified as SAN and 2 `DNS-01` challenges are executed.
In this case the generated DNS TXT record for both domains is the same.
Even though this behaviour is [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) compliant, it can lead to problems as all DNS providers keep DNS records cached for a certain time (TTL) and this TTL can be superior to the challenge timeout making the `DNS-01` challenge fail.
The Traefik ACME client library [LEGO](https://github.com/go-acme/lego) supports some but not all DNS providers to work around this issue.
The [`provider` table](/configuration/acme/#provider) indicates if they allow generating certificates for a wildcard domain and its root domain.

### `onDemand` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated.

```toml
[acme]
# ...
onDemand = true
# ...
```

Enable on demand certificate generation.

This will request certificates from Let's Encrypt during the first TLS handshake for host names that do not yet have certificates.

!!! warning
    TLS handshakes are slow when requesting a host name certificate for the first time. This can lead to DoS attacks!

!!! warning
    Take note that Let's Encrypt applies [rate limiting](https://letsencrypt.org/docs/rate-limits).

### `onHostRule`

```toml
[acme]
# ...
onHostRule = true
# ...
```

Enable certificate generation on frontend `Host` rules (for frontends wired to the `acme.entryPoint`).

This will request a certificate from Let's Encrypt for each frontend with a Host rule.

For example, the rule `Host:test1.traefik.io,test2.traefik.io` will request a certificate with main domain `test1.traefik.io` and SAN `test2.traefik.io`.

!!! warning
    `onHostRule` option can not be used to generate wildcard certificates.
    Refer to [wildcard generation](/configuration/acme/#wildcard-domains) for further information.

### `storage`

The `storage` option sets the location where your ACME certificates are saved to.

```toml
[acme]
# ...
storage = "acme.json"
# ...
```

The value can refer to two kinds of storage:

- a JSON file
- a KV store entry

!!! danger "DEPRECATED"
    `storage` replaces `storageFile` which is deprecated.

!!! note
    During migration to a KV store use both `storageFile` and `storage` to migrate ACME certificates too. See [`storeconfig` subcommand](/user-guide/kv-config/#store-configuration-in-key-value-store) for further information.

#### As a File

ACME certificates can be stored in a JSON file that needs to have file mode `600`.

In Docker you can either mount the JSON file or the folder containing it:

```bash
docker run -v "/my/host/acme.json:acme.json" traefik
```
```bash
docker run -v "/my/host/acme:/etc/traefik/acme" traefik
```

!!! warning
    This file cannot be shared across multiple instances of Traefik at the same time. Please use a [KV Store entry](/configuration/acme/#as-a-key-value-store-entry) instead.

#### As a Key Value Store Entry

ACME certificates can be stored in a KV Store entry. This kind of storage is **mandatory in cluster mode**.

```toml
storage = "traefik/acme/account"
```

Because KV stores (like Consul) have limited entry size the certificates list is compressed before it is saved as KV store entry.

!!! note
    It is possible to store up to approximately 100 ACME certificates in Consul.

#### ACME v2 Migration

During migration from ACME v1 to ACME v2, using a storage file, a backup of the original file is created in the same place as the latter (with a `.bak` extension).

For example: if `acme.storage`'s value is `/etc/traefik/acme/acme.json`, the backup file will be `/etc/traefik/acme/acme.json.bak`.

!!! note
    When Traefik is launched in a container, the storage file's parent directory needs to be mounted to be able to access the backup file on the host.
    Otherwise the backup file will be deleted when the container is stopped. Traefik will only generate it once!

### `dnsProvider` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated. Please use [dnsChallenge.provider](/configuration/acme/#provider) instead.

### `delayDontCheckDNS` (Deprecated)

!!! danger "DEPRECATED"
    This option is deprecated. Please use [dnsChallenge.delayBeforeCheck](/configuration/acme/#dnschallenge) instead.

## Fallbacks

If Let's Encrypt is not reachable, these certificates will be used:

  1. ACME certificates already generated before downtime
  1. Expired ACME certificates
  1. Provided certificates

!!! note
    For new (sub)domains which need Let's Encrypt authentification, the default Traefik certificate will be used until Traefik is restarted.
