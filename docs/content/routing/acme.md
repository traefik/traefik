# ACME

Automatic HTTPS
{: .subtitle }

Traefik can automatically generate certificates for your domains using an ACME provider (like Let's Encrypt).

!!! warning "Let's Encrypt and Rate Limiting"
    Note that Let's Encrypt has [rate limiting](https://letsencrypt.org/docs/rate-limits).

## Configuration Examples

??? example "Configuring ACME on the Https EntryPoint"

    ```toml
    [entryPoints]
      [entryPoints.web]
         address = ":80"

      [entryPoints.http-tls]
         address = ":443"
         [entryPoints.http-tls.tls] # enabling TLS
    
    [acme]
       email = "your-email@your-domain.org"
       storage = "acme.json"
       entryPoint = "http-tls" # acme is enabled on http-tls
       onHostRule = true # dynamic generation based on the Host() matcher
       [acme.httpChallenge]
          entryPoint = "web" # used during the challenge 
    ```
    
??? example "Configuring Wildcard Certificates"

    ```toml
    [entryPoints]
      [entryPoints.web]
         address = ":80"

      [entryPoints.http-tls]
         address = ":443"
         [entryPoints.https.tls] # enabling TLS
    
    [acme]
        email = "your-email@your-domain.org"
        storage = "acme.json"
        entryPoint = "http-tls" # acme is enabled on http-tls
        [acme.dnsChallenge]
            provider = "xxx"
          
        [[acme.domains]]
          main = "*.mydomain.com"
          sans = ["mydomain.com"]
    ```  
    
!!! note "Configuration Reference"

    There are many available options for ACME. For a quick glance at what's possible, browse the [configuration reference](../reference/acme.md).

## Configuration Options

### The Different ACME Challenges

#### tlsChallenge

Use the `TLS-ALPN-01` challenge to generate and renew ACME certificates by provisioning a TLS certificate.

??? example "Using an EntryPoint Called https for the `tlsChallenge`"

    ```toml
    [acme]
    # ...
    entryPoint = "https"
    [acme.tlsChallenge]
    ```

    !!! note
        As described on the Let's Encrypt [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72), when using the `TLS-ALPN-01` challenge, `acme.entryPoint` must be reachable by Let's Encrypt through port 443.

#### `httpChallenge`

Use the `HTTP-01` challenge to generate and renew ACME certificates by provisioning an HTTP resource under a well-known URI.

??? example "Using an EntryPoint Called http for the `httpChallenge`" 

    ```toml
    [acme]
    # ...
    entryPoint = "https"
    [acme.httpChallenge]
      entryPoint = "http"
    ```
    
    !!! note
        As described on the Let's Encrypt [community forum](https://community.letsencrypt.org/t/support-for-ports-other-than-80-and-443/3419/72), when using the `HTTP-01` challenge, `acme.httpChallenge.entryPoint` must be reachable by Let's Encrypt through port 80.
    
    !!! note    
        Redirection is fully compatible with the `HTTP-01` challenge. 

#### `dnsChallenge`

Use the `DNS-01` challenge to generate and renew ACME certificates by provisioning a DNS record.

??? example "Configuring a `dnsChallenge` with the digitalocean Provider"

    ```toml
    [acme]
    # ...
    [acme.dnsChallenge]
      provider = "digitalocean"
      delayBeforeCheck = 0
    # ...
    ```
    
    !!! important
        A `provider` is mandatory.

??? list "Supported Providers"

    Here is a list of supported `providers`, that can automate the DNS verification, along with the required environment variables and their [wildcard & root domain support](#wildcard-domains). 
    
    | Provider Name                                          | Provider Code  | Environment Variables                                                                                                                     | Wildcard & Root Domain Support |
    |--------------------------------------------------------|----------------|-------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------|
    | [ACME DNS](https://github.com/joohoi/acme-dns)         | `acme-dns`     | `ACME_DNS_API_BASE`, `ACME_DNS_STORAGE_PATH`                                                                                              | Not tested yet                 |
    | [Alibaba Cloud](https://www.vultr.com)                 | `alidns`       | `ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION_ID`                                                                        | Not tested yet                 |
    | [Auroradns](https://www.pcextreme.com/aurora/dns)      | `auroradns`    | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                                         | Not tested yet                 |
    | [Azure](https://azure.microsoft.com/services/dns/)     | `azure`        | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`, `AZURE_RESOURCE_GROUP`, `[AZURE_METADATA_ENDPOINT]` | Not tested yet                 |
    | [Blue Cat](https://www.bluecatnetworks.com/)           | `bluecat`      | `BLUECAT_SERVER_URL`, `BLUECAT_USER_NAME`, `BLUECAT_PASSWORD`, `BLUECAT_CONFIG_NAME`, `BLUECAT_DNS_VIEW`                                  | Not tested yet                 |
    | [Cloudflare](https://www.cloudflare.com)               | `cloudflare`   | `CF_API_EMAIL`, `CF_API_KEY` - The `Global API Key` needs to be used, not the `Origin CA Key`                                             | YES                            |
    | [CloudXNS](https://www.cloudxns.net)                   | `cloudxns`     | `CLOUDXNS_API_KEY`, `CLOUDXNS_SECRET_KEY`                                                                                                 | Not tested yet                 |
    | [ConoHa](https://www.conoha.jp)                        | `conoha`       | `CONOHA_TENANT_ID`, `CONOHA_API_USERNAME`, `CONOHA_API_PASSWORD`                                                                          | YES                            |
    | [DigitalOcean](https://www.digitalocean.com)           | `digitalocean` | `DO_AUTH_TOKEN`                                                                                                                           | YES                            |
    | [DNSimple](https://dnsimple.com)                       | `dnsimple`     | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                                               | Not tested yet                 |
    | [DNS Made Easy](https://dnsmadeeasy.com)               | `dnsmadeeasy`  | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                                    | Not tested yet                 |
    | [DNSPod](https://www.dnspod.com/)                      | `dnspod`       | `DNSPOD_API_KEY`                                                                                                                          | Not tested yet                 |
    | [DreamHost](https://www.dreamhost.com/)                | `dreamhost`    | `DREAMHOST_API_KEY`                                                                                                                       | YES                            |
    | [Duck DNS](https://www.duckdns.org/)                   | `duckdns`      | `DUCKDNS_TOKEN`                                                                                                                           | No                             |
    | [Dyn](https://dyn.com)                                 | `dyn`          | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                                      | Not tested yet                 |
    | External Program                                       | `exec`         | `EXEC_PATH`                                                                                                                               | YES                            |
    | [Exoscale](https://www.exoscale.com)                   | `exoscale`     | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                                            | YES                            |
    | [Fast DNS](https://www.akamai.com/)                    | `fastdns`      | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                    | Not tested yet                 |
    | [Gandi](https://www.gandi.net)                         | `gandi`        | `GANDI_API_KEY`                                                                                                                           | Not tested yet                 |
    | [Gandi v5](http://doc.livedns.gandi.net)               | `gandiv5`      | `GANDIV5_API_KEY`                                                                                                                         | YES                            |
    | [Glesys](https://glesys.com/)                          | `glesys`       | `GLESYS_API_USER`, `GLESYS_API_KEY`, `GLESYS_DOMAIN`                                                                                      | Not tested yet                 |
    | [GoDaddy](https://godaddy.com/domains)                 | `godaddy`      | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                                   | Not tested yet                 |
    | [Google Cloud DNS](https://cloud.google.com/dns/docs/) | `gcloud`       | `GCE_PROJECT`, `GCE_SERVICE_ACCOUNT_FILE`                                                                                                 | YES                            |
    | [hosting.de](https://www.hosting.de)                   | `hostingde`    | `HOSTINGDE_API_KEY`, `HOSTINGDE_ZONE_NAME`                                                                                                | Not tested yet                 |
    | HTTP request                                           | `httpreq`      | `HTTPREQ_ENDPOINT`, `HTTPREQ_MODE`, `HTTPREQ_USERNAME`, `HTTPREQ_PASSWORD` (1)                                                            | YES                            |
    | [IIJ](https://www.iij.ad.jp/)                          | `iij`          | `IIJ_API_ACCESS_KEY`, `IIJ_API_SECRET_KEY`, `IIJ_DO_SERVICE_CODE`                                                                         | Not tested yet                 |
    | [INWX](https://www.inwx.de/en)                         | `inwx`         | `INWX_USERNAME`, `INWX_PASSWORD`                                                                                                          | YES                            |
    | [Lightsail](https://aws.amazon.com/lightsail/)         | `lightsail`    | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `DNS_ZONE`                                                                                  | Not tested yet                 |
    | [Linode](https://www.linode.com)                       | `linode`       | `LINODE_API_KEY`                                                                                                                          | Not tested yet                 |
    | [Linode v4](https://www.linode.com)                    | `linodev4`     | `LINODE_TOKEN`                                                                                                                            | Not tested yet                 |
    | manual                                                 | -              | none, but you need to run Traefik interactively, turn on `acmeLogging` to see instructions and press <kbd>Enter</kbd>.                    | YES                            |
    | [MyDNS.jp](https://www.mydns.jp/)                      | `mydnsjp`      | `MYDNSJP_MASTER_ID`, `MYDNSJP_PASSWORD`                                                                                                   | YES                            |
    | [Namecheap](https://www.namecheap.com)                 | `namecheap`    | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                                                 | YES                            |
    | [name.com](https://www.name.com/)                      | `namedotcom`   | `NAMECOM_USERNAME`, `NAMECOM_API_TOKEN`, `NAMECOM_SERVER`                                                                                 | Not tested yet                 |
    | [Netcup](https://www.netcup.eu/)                       | `netcup`       | `NETCUP_CUSTOMER_NUMBER`, `NETCUP_API_KEY`, `NETCUP_API_PASSWORD`                                                                         | Not tested yet                 |
    | [NIFCloud](https://cloud.nifty.com/service/dns.htm)    | `nifcloud`     | `NIFCLOUD_ACCESS_KEY_ID`, `NIFCLOUD_SECRET_ACCESS_KEY`                                                                                    | Not tested yet                 |
    | [Ns1](https://ns1.com/)                                | `ns1`          | `NS1_API_KEY`                                                                                                                             | Not tested yet                 |
    | [Open Telekom Cloud](https://cloud.telekom.de)         | `otc`          | `OTC_DOMAIN_NAME`, `OTC_USER_NAME`, `OTC_PASSWORD`, `OTC_PROJECT_NAME`, `OTC_IDENTITY_ENDPOINT`                                           | Not tested yet                 |
    | [OVH](https://www.ovh.com)                             | `ovh`          | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`                                                       | YES                            |
    | [PowerDNS](https://www.powerdns.com)                   | `pdns`         | `PDNS_API_KEY`, `PDNS_API_URL`                                                                                                            | Not tested yet                 |
    | [Rackspace](https://www.rackspace.com/cloud/dns)       | `rackspace`    | `RACKSPACE_USER`, `RACKSPACE_API_KEY`                                                                                                     | Not tested yet                 |
    | [RFC2136](https://tools.ietf.org/html/rfc2136)         | `rfc2136`      | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                                 | Not tested yet                 |
    | [Route 53](https://aws.amazon.com/route53/)            | `route53`      | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `[AWS_REGION]`, `[AWS_HOSTED_ZONE_ID]` or a configured user/instance IAM profile.           | YES                            |
    | [Sakura Cloud](https://cloud.sakura.ad.jp/)            | `sakuracloud`  | `SAKURACLOUD_ACCESS_TOKEN`, `SAKURACLOUD_ACCESS_TOKEN_SECRET`                                                                             | Not tested yet                 |
    | [Selectel](https://selectel.ru/en/)                    | `selectel`     | `SELECTEL_API_TOKEN`                                                                                                                      | YES                            |
    | [Stackpath](https://www.stackpath.com/)                | `stackpath`    | `STACKPATH_CLIENT_ID`, `STACKPATH_CLIENT_SECRET`, `STACKPATH_STACK_ID`                                                                    | Not tested yet                 |
    | [TransIP](https://www.transip.nl/)                     | `transip`      | `TRANSIP_ACCOUNT_NAME`, `TRANSIP_PRIVATE_KEY_PATH`                                                                                        | YES                            |
    | [VegaDNS](https://github.com/shupp/VegaDNS-API)        | `vegadns`      | `SECRET_VEGADNS_KEY`, `SECRET_VEGADNS_SECRET`, `VEGADNS_URL`                                                                              | Not tested yet                 |
    | [Vscale](https://vscale.io/)                           | `vscale`       | `VSCALE_API_TOKEN`                                                                                                                        | YES                            |
    | [VULTR](https://www.vultr.com)                         | `vultr`        | `VULTR_API_KEY`                                                                                                                           | Not tested yet                 |
    
    - (1): more information about the HTTP message format can be found [here](https://github.com/xenolf/lego/blob/master/providers/dns/httpreq/readme.md)

!!! note "`delayBeforeCheck`"
    By default, the `provider` verifies the TXT record _before_ letting ACME verify.
    You can delay this operation by specifying a delay (in seconds) with `delayBeforeCheck` (value must be greater than zero).
    This option is useful when internal networks block external DNS queries.

!!! note "`resolvers`"

    Use custom DNS servers to resolve the FQDN authority.
    
    ```toml
    [acme]
    # ...
    [acme.dnsChallenge]
      # ...
      resolvers = ["1.1.1.1:53", "8.8.8.8:53"]
    ```

### Known Domains, SANs, and Wildcards

You can set SANs (alternative domains) for each main domain.
Every domain must have A/AAAA records pointing to Traefik.
Each domain & SAN will lead to a certificate request.

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

!!! important
    The certificates for the domains listed in `acme.domains` are negotiated at Traefik startup only.

!!! note
    Wildcard certificates can only be verified through a `DNS-01` challenge.

#### Wildcard Domains

[ACME V2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](#dnschallenge).

```toml
[acme]
# ...
[[acme.domains]]
  main = "*.local1.com"
  sans = ["local1.com"]
# ...
```

!!! note "Double Wildcard Certificates"
    It is not possible to request a double wildcard certificate for a domain (for example `*.*.local.com`).
    
Due to an ACME limitation it is not possible to define wildcards in SANs (alternative domains).
Thus, the wildcard domain has to be defined as a main domain.
Most likely the root domain should receive a certificate too, so it needs to be specified as SAN and 2 `DNS-01` challenges are executed.
In this case the generated DNS TXT record for both domains is the same.
Even though this behavior is [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) compliant, it can lead to problems as all DNS providers keep DNS records cached for a given time (TTL) and this TTL can be greater than the challenge timeout making the `DNS-01` challenge fail.
The Traefik ACME client library [LEGO](https://github.com/xenolf/lego) supports some but not all DNS providers to work around this issue.
The [Supported `provider` table](#dnschallenge) indicates if they allow generating certificates for a wildcard domain and its root domain.

### caServer

??? example "Using the Let's Encrypt staging server"

    ```toml
    [acme]
    # ...
    caServer = "https://acme-staging-v02.api.letsencrypt.org/directory"
    # ...
    ```
    
### onHostRule

Enable certificate generation on [routers](routers.md) `Host` rules (for routers active on the `acme.entryPoint`).

This will request a certificate from Let's Encrypt for each router with a Host rule.

```toml
[acme]
# ...
onHostRule = true
# ...
```

!!! note "Multiple Hosts in a Rule"
    The rule `Host(test1.traefik.io,test2.traefik.io)` will request a certificate with the main domain `test1.traefik.io` and SAN `test2.traefik.io`.

!!! warning
    `onHostRule` option can not be used to generate wildcard certificates. Refer to [wildcard generation](#wildcard-domains) for further information.

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

#### In a File

ACME certificates can be stored in a JSON file that needs to have a `600` file mode .

In Docker you can mount either the JSON file, or the folder containing it:

```bash
docker run -v "/my/host/acme.json:acme.json" traefik
```

```bash
docker run -v "/my/host/acme:/etc/traefik/acme" traefik
```

!!! warning
    For concurrency reason, this file cannot be shared across multiple instances of Traefik. Use a key value store entry instead.

#### In a a Key Value Store Entry

ACME certificates can be stored in a key-value store entry. 

```toml
storage = "traefik/acme/account"
```

!!! note "Storage Size"

    Because key-value stores have limited entry size, the certificates list is compressed _before_ it is saved.
    For example, it is possible to store up to _approximately_ 100 ACME certificates in Consul.

## Fallbacks

If Let's Encrypt is not reachable, the following certificates will apply:

  1. Previously generated ACME certificates (before downtime)
  1. Expired ACME certificates
  1. Provided certificates

!!! note
    For new (sub)domains which need Let's Encrypt authentification, the default Traefik certificate will be used until Traefik is restarted.
