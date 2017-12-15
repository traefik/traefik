# ACME (Let's Encrypt) configuration

See also [Let's Encrypt examples](/user-guide/examples/#lets-encrypt-support) and [Docker & Let's Encrypt user guide](/user-guide/docker-and-lets-encrypt).

## Configuration

```toml
# Sample entrypoint configuration when using ACME.
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

# Enable ACME (Let's Encrypt): automatic SSL.
[acme]

# Email address used for registration.
#
# Required
#
email = "test@traefik.io"

# File or key used for certificates storage.
#
# Required
#
storage = "acme.json"
# or `storage = "traefik/acme/account"` if using KV store.

# Entrypoint to proxy acme challenge/apply certificates to.
# WARNING, must point to an entrypoint on port 443
#
# Required
#
entryPoint = "https"

# Use a DNS based acme challenge rather than external HTTPS access
#
#
# Optional
#
# dnsProvider = "digitalocean"

# By default, the dnsProvider will verify the TXT DNS challenge record before letting ACME verify.
# If delayDontCheckDNS is greater than zero, avoid this & instead just wait so many seconds.
# Useful if internal networks block external DNS queries.
#
# Optional
#
# delayDontCheckDNS = 0

# If true, display debug log messages from the acme client library.
#
# Optional
#
# acmeLogging = true

# Enable on demand certificate.
#
# Optional
#
# onDemand = true

# Enable certificate generation on frontends Host rules.
#
# Optional
#
# onHostRule = true

# CA server to use.
# - Uncomment the line to run on the staging let's encrypt server.
# - Leave comment to go to prod.
#
# Optional
#
# caServer = "https://acme-staging.api.letsencrypt.org/directory"

# Domains list.
#
# [[acme.domains]]
# main = "local1.com"
# sans = ["test1.local1.com", "test2.local1.com"]
# [[acme.domains]]
# main = "local2.com"
# sans = ["test1.local2.com", "test2.local2.com"]
# [[acme.domains]]
# main = "local3.com"
# [[acme.domains]]
# main = "local4.com"
```

### `storage`

```toml
[acme]
# ...
storage = "acme.json"
# ...
```

File or key used for certificates storage.

**WARNING** If you use Traefik in Docker, you have 2 options:

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

### `dnsProvider`

```toml
[acme]
# ...
dnsProvider = "digitalocean"
# ...
```

Use a DNS based acme challenge rather than external HTTPS access, e.g. for a firewalled server.

Select the provider that matches the DNS domain that will host the challenge TXT record, and provide environment variables to enable setting it:

| Provider Name                                          | Provider code  | Configuration                                                                                                             |
|--------------------------------------------------------|----------------|---------------------------------------------------------------------------------------------------------------------------|
| [Auroradns](https://www.pcextreme.com/aurora/dns)      | `auroradns`    | `AURORA_USER_ID`, `AURORA_KEY`, `AURORA_ENDPOINT`                                                                         |
| [Azure](https://azure.microsoft.com/services/dns/)     | `azure`        | `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`, `AZURE_RESOURCE_GROUP`              |
| [Cloudflare](https://www.cloudflare.com)               | `cloudflare`   | `CLOUDFLARE_EMAIL`, `CLOUDFLARE_API_KEY` - The Cloudflare `Global API Key` needs to be used and not the `Origin CA Key`   |
| [DigitalOcean](https://www.digitalocean.com)           | `digitalocean` | `DO_AUTH_TOKEN`                                                                                                           |
| [DNSimple](https://dnsimple.com)                       | `dnsimple`     | `DNSIMPLE_OAUTH_TOKEN`, `DNSIMPLE_BASE_URL`                                                                               |
| [DNS Made Easy](https://dnsmadeeasy.com)               | `dnsmadeeasy`  | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`, `DNSMADEEASY_SANDBOX`                                                    |
| [DNSPod](http://www.dnspod.net/)                       | `dnspod`       | `DNSPOD_API_KEY`                                                                                                          |
| [Dyn](https://dyn.com)                                 | `dyn`          | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                                      |
| [Exoscale](https://www.exoscale.ch)                    | `exoscale`     | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`, `EXOSCALE_ENDPOINT`                                                            |
| [Gandi](https://www.gandi.net)                         | `gandi`        | `GANDI_API_KEY`                                                                                                           |
| [GoDaddy](https://godaddy.com/domains)                 | `godaddy`      | `GODADDY_API_KEY`, `GODADDY_API_SECRET`                                                                                   |
| [Google Cloud DNS](https://cloud.google.com/dns/docs/) | `gcloud`       | `GCE_PROJECT`, `GCE_SERVICE_ACCOUNT_FILE`                                                                                 |
| [Linode](https://www.linode.com)                       | `linode`       | `LINODE_API_KEY`                                                                                                          |
| manual                                                 | -              | none, but run Traefik interactively & turn on `acmeLogging` to see instructions & press <kbd>Enter</kbd>.                 |
| [Namecheap](https://www.namecheap.com)                 | `namecheap`    | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                                 |
| [Ns1](https://ns1.com/)                                | `ns1`          | `NS1_API_KEY`                                                                                                             |
| [Open Telekom Cloud](https://cloud.telekom.de/en/)     | `otc`          | `OTC_DOMAIN_NAME`, `OTC_USER_NAME`, `OTC_PASSWORD`, `OTC_PROJECT_NAME`, `OTC_IDENTITY_ENDPOINT`                           |
| [OVH](https://www.ovh.com)                             | `ovh`          | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`                                       |
| [PowerDNS](https://www.powerdns.com)                   | `pdns`         | `PDNS_API_KEY`, `PDNS_API_URL`                                                                                            |
| [Rackspace](https://www.rackspace.com/cloud/dns)       | `rackspace`    | `RACKSPACE_USER`, `RACKSPACE_API_KEY`                                                                                     |
| [RFC2136](https://tools.ietf.org/html/rfc2136)         | `rfc2136`      | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                                 |
| [Route 53](https://aws.amazon.com/route53/)            | `route53`      | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `AWS_HOSTED_ZONE_ID` or configured user/instance IAM profile. |
| [VULTR](https://www.vultr.com)                         | `vultr`        | `VULTR_API_KEY`                                                                                                           |

### `delayDontCheckDNS`

```toml
[acme]
# ...
delayDontCheckDNS = 0
# ...
```

By default, the dnsProvider will verify the TXT DNS challenge record before letting ACME verify.  
If `delayDontCheckDNS` is greater than zero, avoid this & instead just wait so many seconds.

Useful if internal networks block external DNS queries.

### `onDemand`

```toml
[acme]
# ...
onDemand = true
# ...
```

Enable on demand certificate.

This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate.

!!! warning
    TLS handshakes will be slow when requesting a hostname certificate for the first time, this can lead to DoS attacks.
    
!!! warning
    Take note that Let's Encrypt have [rate limiting](https://letsencrypt.org/docs/rate-limits)

### `onHostRule`

```toml
[acme]
# ...
onHostRule = true
# ...
```

Enable certificate generation on frontends Host rules.

This will request a certificate from Let's Encrypt for each frontend with a Host rule.

For example, a rule `Host:test1.traefik.io,test2.traefik.io` will request a certificate with main domain `test1.traefik.io` and SAN `test2.traefik.io`.

### `caServer`

```toml
[acme]
# ...
caServer = "https://acme-staging.api.letsencrypt.org/directory"
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
main = "local4.com"
# ...
```

You can provide SANs (alternative domains) to each main domain.
All domains must have A/AAAA records pointing to Traefik.

!!! warning
    Take note that Let's Encrypt have [rate limiting](https://letsencrypt.org/docs/rate-limits).

Each domain & SANs will lead to a certificate request.
