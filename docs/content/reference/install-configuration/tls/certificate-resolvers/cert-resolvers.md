---
title: "Traefik & ACME Certificates Resolver"
description: "Automatic Certificate Management Environment using Let's Encrypt."
---


In Traefik, TLS Certificates can be generated using Certificates Resolvers.

In Traefik, two certificate resolvers exist:

- `acme`: It allows generating ACME certificates stored in a file (not distributed).
- `tailscale`: It allows provisioning TLS certificates for internal Tailscale services.

The Certificates resolvers are defined in the static configuration.

!!! note Referencing a certificate resolver
    Defining a certificate resolver does not imply that routers are going to use it automatically.
    Each router or entrypoint that is meant to use the resolver must explicitly reference it.

## Configuration Example

```yaml tab="Static configuration"
certificatesResolvers:
  my-resolver:
    acme:
      email: "your-email@example.com"
      storage: "/path/to/acme.json"
      httpChallenge:
        entryPoint: "web"
```

## Configuration Options

ACME certificate resolvers have the following configuration options:

| Field   | Description  | Default    | Required |
|:------------------|:--------------------|:-----------------------------------------------|:---------|
| `acme.email` | Email address used for registration. |   | Yes      |
| `acme.caServer` | CA server to use. | https://acme-v02.api.letsencrypt.org/directory | No       |
| `acme.preferredChain`  | Preferred chain to use. If the CA offers multiple certificate chains, prefer the chain with an issuer matching this Subject Common Name. If no match, the default offered chain will be used. |   | No  |
| `acme.keyType` | KeyType to use. | RSA4096  | No       |
| `acme.eab` | Enable external account binding.|   | No  |
| `acme.eab.kid` | Key identifier from External CA. | No   | |
| `acme.eab.hmacEncoded`  | HMAC key from External CA, should be in Base64 URL Encoding without padding format.  |        | No    |
| `acme.certificatesDuration`  | The certificates' duration in hours, exclusively used to determine renewal dates. | 2160      | No       |
| `acme.dnsChallenge`  | Enable DNS-01 challenge. More information [here](#dnschallenge)   |       | No       |
| `acme.dnsChallenge.provider`    | DNS provider to use.  |      | No   |
| `acme.dnsChallenge.delayBeforeCheck`  | By default, the provider will verify the TXT DNS challenge record before letting ACME verify. If `delayBeforeCheck` is greater than zero, this check is delayed for the configured duration in seconds. Useful if internal networks block external DNS queries. |       | No  |
| `acme.dnsChallenge.resolvers` | DNS servers to resolve the FQDN authority.   |    | No       |
| `acme.dnsChallenge.disablePropagationCheck`  | Disable the DNS propagation checks before notifying ACME that the DNS challenge is ready.   |     | No       |
| `acme.httpChallenge`    | Enable HTTP-01 challenge. More information [here](#httpchallenge)   |    | No   |
| `acme.httpChallenge.entryPoint`  | EntryPoint to use for the HTTP-01 challenges. Must be reachable by Let's Encrypt through port 80 |    | Yes      |
| `acme.tlsChallenge` | Enable TLS-ALPN-01 challenge. Traefik must be reachable by Let's Encrypt through port 443. More information [here](#tlschallenge) |        | No       |
| `acme.storage`  | File path used for certificates storage.  |     | Yes      |

### dnsChallenge

The DNS-01 challenge to generate and renew ACME certificates by provisioning a DNS record.

Traefik relies internally on [Lego](https://go-acme.github.io/lego/ "Link to Lego website") for ACME.
You can find the list of all the supported DNS providers in their [documentation](https://go-acme.github.io/lego/dns/ "Link to Lego DNS challenge documentation page")
with instructions about which environment variables need to be setup.

!!! note

      CNAME are supported and even [encouraged](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html#the-advantages-of-a-cname "Link to The Advantages of a CNAME article").

      If needed, CNAME support can be turned off with the following environment variable:

      ```env
      LEGO_DISABLE_CNAME_SUPPORT=true
      ```

!!! warning "Multiple DNS challenge"

      Multiple DNS challenge provider are not supported with Traefik, but you can use CNAME to handle that.
      For example, if you have `example.org` (account foo) and `example.com` (account bar) you can create a CNAME on `example.org` called `_acme-challenge.example.org` pointing to `challenge.example.com`.
      This way, you can obtain certificates for `example.com` with the foo account.
      
<!-- 
#### `providers`

Here is a list of supported `providers`, that can automate the DNS verification,
along with the required environment variables and their [wildcard & root domain support](#wildcard-domains).
Do not hesitate to complete it.

Many lego environment variables can be overridden by their respective `_FILE` counterpart, which should have a filepath to a file that contains the secret as its value.
For example, `CF_API_EMAIL_FILE=/run/secrets/traefik_cf-api-email` could be used to provide a Cloudflare API email address as a Docker secret named `traefik_cf-api-email`.

You can find the list of all the supported DNS providers in their [documentation](https://go-acme.github.io/lego/dns/) with instructions about which environment variables need to be setup. -->

<!-- | Provider Name   | Provider Code      | Environment Variables   |  |
|---------------------------|--------------------|--------------------------------------|--------|
| [ACME DNS](https://github.com/joohoi/acme-dns)   | `acme-dns`  | `ACME_DNS_API_BASE`, `ACME_DNS_STORAGE_PATH`  | [Additional configuration](https://go-acme.github.io/lego/dns/acme-dns)         |
| [Alibaba Cloud](https://www.alibabacloud.com)                          | `alidns`           | `ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION_ID`                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/alidns)           |
| [all-inkl](https://all-inkl.com)                                       | `allinkl`          | `ALL_INKL_LOGIN`, `ALL_INKL_PASSWORD`                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/allinkl)          |
| [ArvanCloud](https://www.arvancloud.ir/en)                             | `arvancloud`       | `ARVANCLOUD_API_KEY`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/arvancloud)       |
| [Auroradns](https://www.pcextreme.com/dns-health-checks)               | `auroradns`        | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/auroradns)        |
| [Autodns](https://www.internetx.com/domains/autodns/)                  | `autodns`          | `AUTODNS_API_USER`, `AUTODNS_API_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/autodns)          |
| [AzureDNS](https://azure.microsoft.com/services/dns/)                  | `azuredns`         | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`, `AZURE_SUBSCRIPTION_ID`, `AZURE_RESOURCE_GROUP`, `[AZURE_ENVIRONMENT]`, `[AZURE_PRIVATE_ZONE]`, `[AZURE_ZONE_NAME]` | [Additional configuration](https://go-acme.github.io/lego/dns/azuredns)         |
| [Bindman](https://github.com/labbsr0x/bindman-dns-webhook)             | `bindman`          | `BINDMAN_MANAGER_ADDRESS`                                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/bindman)          |
| [Blue Cat](https://www.bluecatnetworks.com/)                           | `bluecat`          | `BLUECAT_SERVER_URL`, `BLUECAT_USER_NAME`, `BLUECAT_PASSWORD`, `BLUECAT_CONFIG_NAME`, `BLUECAT_DNS_VIEW`                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/bluecat)          |
| [Brandit](https://www.brandit.com)                                     | `brandit`          | `BRANDIT_API_USERNAME`, `BRANDIT_API_KEY`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/brandit)          |
| [Bunny](https://bunny.net)                                             | `bunny`            | `BUNNY_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/bunny)            |
| [Checkdomain](https://www.checkdomain.de/)                             | `checkdomain`      | `CHECKDOMAIN_TOKEN`,                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/checkdomain/)     |
| [Civo](https://www.civo.com/)                                          | `civo`             | `CIVO_TOKEN`                                                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/civo)             |
| [Cloud.ru](https://cloud.ru)                                           | `cloudru`          | `CLOUDRU_SERVICE_INSTANCE_ID`, `CLOUDRU_KEY_ID`, `CLOUDRU_SECRET`                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/cloudru)          |
| [CloudDNS](https://vshosting.eu/)                                      | `clouddns`         | `CLOUDDNS_CLIENT_ID`, `CLOUDDNS_EMAIL`, `CLOUDDNS_PASSWORD`                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/clouddns)         |
| [Cloudflare](https://www.cloudflare.com)                               | `cloudflare`       | `CF_API_EMAIL`, `CF_API_KEY` [^5] or `CF_DNS_API_TOKEN`, `[CF_ZONE_API_TOKEN]`                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/cloudflare)       |
| [ClouDNS](https://www.cloudns.net/)                                    | `cloudns`          | `CLOUDNS_AUTH_ID`, `CLOUDNS_AUTH_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/cloudns)          |
| [CloudXNS](https://www.cloudxns.net)                                   | `cloudxns`         | `CLOUDXNS_API_KEY`, `CLOUDXNS_SECRET_KEY`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/cloudxns)         |
| [ConoHa](https://www.conoha.jp)                                        | `conoha`           | `CONOHA_TENANT_ID`, `CONOHA_API_USERNAME`, `CONOHA_API_PASSWORD`                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/conoha)           |
| [Constellix](https://constellix.com)                                   | `constellix`       | `CONSTELLIX_API_KEY`, `CONSTELLIX_SECRET_KEY`                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/constellix)       |
| [CPanel and WHM](https://cpanel.net/)                                  | `cpanel`           | `CPANEL_MODE`, `CPANEL_USERNAME`, `CPANEL_TOKEN`, `CPANEL_BASE_URL`                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/cpanel)           |
| [Derak Cloud](https://derak.cloud/)                                    | `derak`            | `DERAK_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/derak)            |
| [deSEC](https://desec.io)                                              | `desec`            | `DESEC_TOKEN`                                                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/desec)            |
| [DigitalOcean](https://www.digitalocean.com)                           | `digitalocean`     | `DO_AUTH_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/digitalocean)     |
| [DirectAdmin](https://www.directadmin.com)                             | `directadmin`      | `DIRECTADMIN_API_URL` , `DIRECTADMIN_USERNAME`, `DIRECTADMIN_PASSWORD`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/directadmin)      |
| [DNS Made Easy](https://dnsmadeeasy.com)                               | `dnsmadeeasy`      | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/dnsmadeeasy)      |
| [dnsHome.de](https://www.dnshome.de)                                   | `dnsHomede`        | `DNSHOMEDE_CREDENTIALS`                                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/dnshomede)        |
| [DNSimple](https://dnsimple.com)                                       | `dnsimple`         | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/dnsimple)         |
| [DNSPod](https://www.dnspod.com/)                                      | `dnspod`           | `DNSPOD_API_KEY`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/dnspod)           |
| [Domain Offensive (do.de)](https://www.do.de/)                         | `dode`             | `DODE_TOKEN`                                                                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/dode)             |
| [Domeneshop](https://domene.shop)                                      | `domeneshop`       | `DOMENESHOP_API_TOKEN`, `DOMENESHOP_API_SECRET`                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/domeneshop)       |
| [DreamHost](https://www.dreamhost.com/)                                | `dreamhost`        | `DREAMHOST_API_KEY`                                                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/dreamhost)        |
| [Duck DNS](https://www.duckdns.org/)                                   | `duckdns`          | `DUCKDNS_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/duckdns)          |
| [Dyn](https://dyn.com)                                                 | `dyn`              | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/dyn)              |
| [Dynu](https://www.dynu.com)                                           | `dynu`             | `DYNU_API_KEY`                                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/dynu)             |
| [EasyDNS](https://easydns.com/)                                        | `easydns`          | `EASYDNS_TOKEN`, `EASYDNS_KEY`                                                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/easydns)          |
| [EdgeDNS](https://www.akamai.com/)                                     | `edgedns`          | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/edgedns)          |
| [Efficient IP](https://efficientip.com)                                | `efficientip`      | `EFFICIENTIP_USERNAME`, `EFFICIENTIP_PASSWORD`, `EFFICIENTIP_HOSTNAME`, `EFFICIENTIP_DNS_NAME`                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/efficientip)      |
| [Epik](https://www.epik.com)                                           | `epik`             | `EPIK_SIGNATURE`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/epik)             |
| [Exoscale](https://www.exoscale.com)                                   | `exoscale`         | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/exoscale)         |
| [Fast DNS](https://www.akamai.com/)                                    | `fastdns`          | `AKAMAI_CLIENT_TOKEN`,  `AKAMAI_CLIENT_SECRET`,  `AKAMAI_ACCESS_TOKEN`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/edgedns)          |
| [Freemyip.com](https://freemyip.com)                                   | `freemyip`         | `FREEMYIP_TOKEN`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/freemyip)         |
| [G-Core](https://gcore.com/dns/)                                       | `gcore`            | `GCORE_PERMANENT_API_TOKEN`                                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/gcore)            |
| [Gandi v5](https://doc.livedns.gandi.net)                              | `gandiv5`          | `GANDIV5_PERSONAL_ACCESS_TOKEN`                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/gandiv5)          |
| [Gandi](https://www.gandi.net)                                         | `gandi`            | `GANDI_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/gandi)            |
| [Glesys](https://glesys.com/)                                          | `glesys`           | `GLESYS_API_USER`, `GLESYS_API_KEY`, `GLESYS_DOMAIN`                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/glesys)           |
| [GoDaddy](https://www.godaddy.com)                                     | `godaddy`          | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/godaddy)          |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/)                 | `gcloud`           | `GCE_PROJECT`, Application Default Credentials [^2] [^3], [`GCE_SERVICE_ACCOUNT_FILE`]                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/gcloud)           |
| [Google Domains](https://domains.google)                               | `googledomains`    | `GOOGLE_DOMAINS_ACCESS_TOKEN`                                                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/googledomains)    |
| [Hetzner](https://hetzner.com)                                         | `hetzner`          | `HETZNER_API_KEY`                                                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/hetzner)          |
| [hosting.de](https://www.hosting.de)                                   | `hostingde`        | `HOSTINGDE_API_KEY`, `HOSTINGDE_ZONE_NAME`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/hostingde)        |
| [Hosttech](https://www.hosttech.eu)                                    | `hosttech`         | `HOSTTECH_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/hosttech)         |
| [http.net](https://www.http.net/)                                      | `httpnet`          | `HTTPNET_API_KEY`                                                                                                                                                                | [Additional configuration](https://go-acme.github.io/lego/dns/httpnet)          |
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
| [Metaname](https://metaname.net)                                       | `metaname`         | `METANAME_ACCOUNT_REFERENCE`, `METANAME_API_KEY`                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/metaname)         |
| [mijn.host](https://mijn.host/)                                        | `mijnhost`         | `MIJNHOST_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/mijnhost)         |
| [Mittwald](https://www.mittwald.de)                                    | `mittwald`         | `MITTWALD_TOKEN`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/mittwald)         |
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
| [RcodeZero](https://www.rcodezero.at)                                  | `rcodezero`        | `RCODEZERO_API_TOKEN`                                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/rcodezero)        |
| [reg.ru](https://www.reg.ru)                                           | `regru`            | `REGRU_USERNAME`, `REGRU_PASSWORD`                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/regru)            |
| [RFC2136](https://tools.ietf.org/html/rfc2136)                         | `rfc2136`          | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/rfc2136)          |
| [RimuHosting](https://rimuhosting.com)                                 | `rimuhosting`      | `RIMUHOSTING_API_KEY`                                                                                                                                                            | [Additional configuration](https://go-acme.github.io/lego/dns/rimuhosting)      |
| [Route 53](https://aws.amazon.com/route53/)                            | `route53`          | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `[AWS_REGION]`, `[AWS_HOSTED_ZONE_ID]` or a configured user/instance IAM profile.                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/route53)          |
| [Sakura Cloud](https://cloud.sakura.ad.jp/)                            | `sakuracloud`      | `SAKURACLOUD_ACCESS_TOKEN`, `SAKURACLOUD_ACCESS_TOKEN_SECRET`                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/sakuracloud)      |
| [Scaleway](https://www.scaleway.com)                                   | `scaleway`         | `SCW_API_TOKEN`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/scaleway)         |
| [Selectel v2](https://selectel.ru/en/)                                 | `selectelv2`       | `SELECTELV2_ACCOUNT_ID`, `SELECTELV2_PASSWORD`, `SELECTELV2_PROJECT_ID`, `SELECTELV2_USERNAME`                                                                                   | [Additional configuration](https://go-acme.github.io/lego/dns/selectelv2)       |
| [Selectel](https://selectel.ru/en/)                                    | `selectel`         | `SELECTEL_API_TOKEN`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/selectel)         |
| [Servercow](https://servercow.de)                                      | `servercow`        | `SERVERCOW_USERNAME`, `SERVERCOW_PASSWORD`                                                                                                                                       | [Additional configuration](https://go-acme.github.io/lego/dns/servercow)        |
| [Shellrent](https://www.shellrent.com)                                 | `shellrent`        | `SHELLRENT_USERNAME`, `SHELLRENT_TOKEN`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/shellrent)        |
| [Simply.com](https://www.simply.com/en/domains/)                       | `simply`           | `SIMPLY_ACCOUNT_NAME`, `SIMPLY_API_KEY`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/simply)           |
| [Sonic](https://www.sonic.com/)                                        | `sonic`            | `SONIC_USER_ID`, `SONIC_API_KEY`                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/sonic)            |
| [Stackpath](https://www.stackpath.com/)                                | `stackpath`        | `STACKPATH_CLIENT_ID`, `STACKPATH_CLIENT_SECRET`, `STACKPATH_STACK_ID`                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/stackpath)        |
| [Tencent Cloud DNS](https://cloud.tencent.com/product/cns)             | `tencentcloud`     | `TENCENTCLOUD_SECRET_ID`, `TENCENTCLOUD_SECRET_KEY`                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/tencentcloud)     |
| [TransIP](https://www.transip.nl/)                                     | `transip`          | `TRANSIP_ACCOUNT_NAME`, `TRANSIP_PRIVATE_KEY_PATH`                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/transip)          |
| [UKFast SafeDNS](https://docs.ukfast.co.uk/domains/safedns/index.html) | `safedns`          | `SAFEDNS_AUTH_TOKEN`                                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/safedns)          |
| [Ultradns](https://neustarsecurityservices.com/dns-services)           | `ultradns`         | `ULTRADNS_USERNAME`, `ULTRADNS_PASSWORD`                                                                                                                                         | [Additional configuration](https://go-acme.github.io/lego/dns/ultradns)         |
| [Variomedia](https://www.variomedia.de/)                               | `variomedia`       | `VARIOMEDIA_API_TOKEN`                                                                                                                                                           | [Additional configuration](https://go-acme.github.io/lego/dns/variomedia)       |
| [VegaDNS](https://github.com/shupp/VegaDNS-API)                        | `vegadns`          | `SECRET_VEGADNS_KEY`, `SECRET_VEGADNS_SECRET`, `VEGADNS_URL`                                                                                                                     | [Additional configuration](https://go-acme.github.io/lego/dns/vegadns)          |
| [Vercel](https://vercel.com)                                           | `vercel`           | `VERCEL_API_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/vercel)           |
| [Versio](https://www.versio.nl/domeinnamen)                            | `versio`           | `VERSIO_USERNAME`, `VERSIO_PASSWORD`                                                                                                                                             | [Additional configuration](https://go-acme.github.io/lego/dns/versio)           |
| [VinylDNS](https://www.vinyldns.io)                                    | `vinyldns`         | `VINYLDNS_ACCESS_KEY`, `VINYLDNS_SECRET_KEY`, `VINYLDNS_HOST`                                                                                                                    | [Additional configuration](https://go-acme.github.io/lego/dns/vinyldns)         |
| [VK Cloud](https://mcs.mail.ru/)                                       | `vkcloud`          | `VK_CLOUD_PASSWORD`, `VK_CLOUD_PROJECT_ID`, `VK_CLOUD_USERNAME`                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/vkcloud)          |
| [Vscale](https://vscale.io/)                                           | `vscale`           | `VSCALE_API_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/vscale)           |
| [VULTR](https://www.vultr.com)                                         | `vultr`            | `VULTR_API_KEY`                                                                                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/vultr)            |
| [Webnames](https://www.webnames.ru/)                                   | `webnames`         | `WEBNAMES_API_KEY`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/webnames)         |
| [Websupport](https://websupport.sk)                                    | `websupport`       | `WEBSUPPORT_API_KEY`, `WEBSUPPORT_SECRET`                                                                                                                                        | [Additional configuration](https://go-acme.github.io/lego/dns/websupport)       |
| [WEDOS](https://www.wedos.com)                                         | `wedos`            | `WEDOS_USERNAME`, `WEDOS_WAPI_PASSWORD`                                                                                                                                          | [Additional configuration](https://go-acme.github.io/lego/dns/wedos)            |
| [Yandex 360](https://360.yandex.ru)                                    | `yandex360`        | `YANDEX360_OAUTH_TOKEN`, `YANDEX360_ORG_ID`                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/yandex360)        |
| [Yandex Cloud](https://cloud.yandex.com/en/)                           | `yandexcloud`      | `YANDEX_CLOUD_FOLDER_ID`, `YANDEX_CLOUD_IAM_TOKEN`                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/yandexcloud)      |
| [Yandex](https://yandex.com)                                           | `yandex`           | `YANDEX_PDD_TOKEN`                                                                                                                                                               | [Additional configuration](https://go-acme.github.io/lego/dns/yandex)           |
| [Zone.ee](https://www.zone.ee)                                         | `zoneee`           | `ZONEEE_API_USER`, `ZONEEE_API_KEY`                                                                                                                                              | [Additional configuration](https://go-acme.github.io/lego/dns/zoneee)           |
| [Zonomi](https://zonomi.com)                                           | `zonomi`           | `ZONOMI_API_KEY`                                                                                                                                                                 | [Additional configuration](https://go-acme.github.io/lego/dns/zonomi)           |
| External Program                                                       | `exec`             | `EXEC_PATH`                                                                                                                                                                      | [Additional configuration](https://go-acme.github.io/lego/dns/exec)             |
| HTTP request                                                           | `httpreq`          | `HTTPREQ_ENDPOINT`, `HTTPREQ_MODE`, `HTTPREQ_USERNAME`, `HTTPREQ_PASSWORD` [^1]                                                                                                  | [Additional configuration](https://go-acme.github.io/lego/dns/httpreq)          |
| manual                                                                 | `manual`           | none, but you need to run Traefik interactively [^4], turn on debug log to see instructions and press <kbd>Enter</kbd>.                                                          |                                                                                 |

[^1]: More information about the HTTP message format can be found [here](https://go-acme.github.io/lego/dns/httpreq/).
[^2]: [Providing credentials to your application](https://cloud.google.com/docs/authentication/production).
[^3]: [google/default.go](https://github.com/golang/oauth2/blob/36a7019397c4c86cf59eeab3bc0d188bac444277/google/default.go#L61-L76)
[^4]: `docker stack` remark: there is no way to support terminal attached to container when deploying with `docker stack`, so you might need to run container with `docker run -it` to generate certificates using `manual` provider.
[^5]: The `Global API Key` needs to be used, not the `Origin CA Key`.
[^6]: As explained in the [LEGO hurricane configuration](https://go-acme.github.io/lego/dns/hurricane/#credentials), each domain or wildcard (record name) needs a token. So each update of record name must be followed by an update of the `HURRICANE_TOKENS` variable, and a restart of Traefik. -->

!!! info ""
    CNAME are supported and even encouraged.
    If needed, CNAME support can be turned off with the following environment variable:
    ```sh
    LEGO_DISABLE_CNAME_SUPPORT=true
    ```

!!! info "`delayBeforeCheck`"
    By default, the `provider` verifies the TXT record _before_ letting ACME verify.
    You can delay this operation by specifying a delay (in seconds) with `delayBeforeCheck` (value must be greater than zero).
    This option is useful when internal networks block external DNS queries.

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

## Domain Definition

A certificate resolver requests certificates for a set of domain names inferred from routers, according to the following:

- If the IngressRoute has a `tls.domains` option set,
  then the certificate resolver derives this router domain name from the `main` option of `tls.domains`.

- Otherwise, the certificate resolver derives the domain name from any `Host()` or `HostSNI()` matchers
  in the IngressRoute's rule.

You can set SANs (alternative domains) for each main domain.
Every domain must have A/AAAA records pointing to Traefik.
Each domain & SAN will lead to a certificate request.

[ACME v2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a `DNS-01` challenge.
It is not possible to request a double wildcard certificate for a domain (for example `*.*.local.com`).

Most likely the root domain should receive a certificate too, so it needs to be specified as SAN and 2 `DNS-01` challenges are invoked.
In such a case the generated DNS TXT record for both domains is the same.
Even though this behavior is [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) compliant,
it can lead to problems as all DNS providers keep DNS records cached for a given time (TTL) and this TTL can be greater than the challenge timeout making the `DNS-01` challenge fail.

The Traefik ACME client library [lego](https://github.com/go-acme/lego) supports some but not all DNS providers to work around this issue.
The supported `provider` table indicates if they allow generating certificates for a wildcard domain and its root domain.

### Wildcard Domains

[ACME V2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](#dnschallenge).

## Certificate Renewal

Traefik automatically tracks the expiry date of certificates it generates.

!!! note Not used Certificates

      Certificates that are no longer used may still be renewed,
      as Traefik does not currently check if the certificate is being used before renewing.

### ACME Certificates

By default, Traefik manages 90-day certificates and starts renewing them 30 days before their expiry.

When using a certificate resolver that issues certificates with custom durations, the `certificatesDuration` option can be used to configure the certificates' duration.

### Tailscale Certificates

Traefik automatically tracks the expiry date of each Tailscale certificate it fetches and starts to renew a certificate 14 days before its expiry to match the Tailscale daemon renewal policy.
