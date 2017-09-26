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

Select the provider that matches the DNS domain that will host the challenge TXT record, and provide environment variables with access keys to enable setting it:

| Provider                                     | Configuration                                                                                             |
|----------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| [Cloudflare](https://www.cloudflare.com)     | `CLOUDFLARE_EMAIL`, `CLOUDFLARE_API_KEY`                                                                  |
| [DigitalOcean](https://www.digitalocean.com) | `DO_AUTH_TOKEN`                                                                                           |
| [DNSimple](https://dnsimple.com)             | `DNSIMPLE_EMAIL`, `DNSIMPLE_OAUTH_TOKEN`                                                                  |
| [DNS Made Easy](https://dnsmadeeasy.com)     | `DNSMADEEASY_API_KEY`, `DNSMADEEASY_API_SECRET`                                                           |
| [Exoscale](https://www.exoscale.ch)          | `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`                                                                 |
| [Gandi](https://www.gandi.net)               | `GANDI_API_KEY`                                                                                           |
| [Linode](https://www.linode.com)             | `LINODE_API_KEY`                                                                                          |
| manual                                       | none, but run Traefik interactively & turn on `acmeLogging` to see instructions & press <kbd>Enter</kbd>. |
| [Namecheap](https://www.namecheap.com)       | `NAMECHEAP_API_USER`, `NAMECHEAP_API_KEY`                                                                 |
| RFC2136                                      | `RFC2136_TSIG_KEY`, `RFC2136_TSIG_SECRET`, `RFC2136_TSIG_ALGORITHM`, `RFC2136_NAMESERVER`                 |
| [Route 53](https://aws.amazon.com/route53/)  | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, or configured user/instance IAM profile.      |
| [dyn](https://dyn.com)                       | `DYN_CUSTOMER_NAME`, `DYN_USER_NAME`, `DYN_PASSWORD`                                                      |
| [VULTR](https://www.vultr.com)               | `VULTR_API_KEY`                                                                                           |
| [OVH](https://www.ovh.com)                   | `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`                       |
| [pdns](https://www.powerdns.com)             | `PDNS_API_KEY`, `PDNS_API_URL`                                                                            |

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
    TLS handshakes will be slow when requesting a hostname certificate for the first time, this can leads to DoS attacks.
    
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
