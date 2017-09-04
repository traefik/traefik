## ACME (Let's Encrypt) configuration

```toml
# Sample entrypoint configuration when using ACME
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

# Enable ACME (Let's Encrypt): automatic SSL
#
# Optional
#
[acme]

# Email address used for registration
#
# Required
#
email = "test@traefik.io"

# File or key used for certificates storage.
# WARNING, if you use Traefik in Docker, you have 2 options:
#  - create a file on your host and mount it as a volume
#      storageFile = "acme.json"
#      $ docker run -v "/my/host/acme.json:acme.json" traefik
#  - mount the folder containing the file as a volume
#      storageFile = "/etc/traefik/acme/acme.json"
#      $ docker run -v "/my/host/acme:/etc/traefik/acme" traefik
#
# Required
#
storage = "acme.json" # or "traefik/acme/account" if using KV store

# Entrypoint to proxy acme challenge/apply certificates to.
# WARNING, must point to an entrypoint on port 443
#
# Required
#
entryPoint = "https"

# Use a DNS based acme challenge rather than external HTTPS access, e.g. for a firewalled server
# Select the provider that matches the DNS domain that will host the challenge TXT record,
# and provide environment variables with access keys to enable setting it:
#  - cloudflare: CLOUDFLARE_EMAIL, CLOUDFLARE_API_KEY
#  - digitalocean: DO_AUTH_TOKEN
#  - dnsimple: DNSIMPLE_EMAIL, DNSIMPLE_OAUTH_TOKEN
#  - dnsmadeeasy: DNSMADEEASY_API_KEY, DNSMADEEASY_API_SECRET
#  - exoscale: EXOSCALE_API_KEY, EXOSCALE_API_SECRET
#  - gandi: GANDI_API_KEY
#  - linode: LINODE_API_KEY
#  - manual: none, but run traefik interactively & turn on acmeLogging to see instructions & press Enter
#  - namecheap: NAMECHEAP_API_USER, NAMECHEAP_API_KEY
#  - rfc2136: RFC2136_TSIG_KEY, RFC2136_TSIG_SECRET, RFC2136_TSIG_ALGORITHM, RFC2136_NAMESERVER
#  - route53: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, or configured user/instance IAM profile
#  - dyn: DYN_CUSTOMER_NAME, DYN_USER_NAME, DYN_PASSWORD
#  - vultr: VULTR_API_KEY
#  - ovh: OVH_ENDPOINT, OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET, OVH_CONSUMER_KEY
#  - pdns: PDNS_API_KEY, PDNS_API_URL
#
# Optional
#
# dnsProvider = "digitalocean"

# By default, the dnsProvider will verify the TXT DNS challenge record before letting ACME verify
# If delayDontCheckDNS is greater than zero, avoid this & instead just wait so many seconds.
# Useful if internal networks block external DNS queries
#
# Optional
#
# delayDontCheckDNS = 0

# If true, display debug log messages from the acme client library
#
# Optional
#
# acmeLogging = true

# Enable on demand certificate. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate.
# WARNING, TLS handshakes will be slow when requesting a hostname certificate for the first time, this can leads to DoS attacks.
# WARNING, Take note that Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits
#
# Optional
#
# onDemand = true

# Enable certificate generation on frontends Host rules. This will request a certificate from Let's Encrypt for each frontend with a Host rule.
# For example, a rule Host:test1.traefik.io,test2.traefik.io will request a certificate with main domain test1.traefik.io and SAN test2.traefik.io.
#
# Optional
#
# OnHostRule = true

# CA server to use
# Uncomment the line to run on the staging let's encrypt server
# Leave comment to go to prod
#
# Optional
#
# caServer = "https://acme-staging.api.letsencrypt.org/directory"

# Domains list
# You can provide SANs (alternative domains) to each main domain
# All domains must have A/AAAA records pointing to Traefik
# WARNING, Take note that Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits
# Each domain & SANs will lead to a certificate request.
#
# [[acme.domains]]
#   main = "local1.com"
#   sans = ["test1.local1.com", "test2.local1.com"]
# [[acme.domains]]
#   main = "local2.com"
#   sans = ["test1.local2.com", "test2x.local2.com"]
# [[acme.domains]]
#   main = "local3.com"
# [[acme.domains]]
#   main = "local4.com"
[[acme.domains]]
   main = "local1.com"
   sans = ["test1.local1.com", "test2.local1.com"]
[[acme.domains]]
   main = "local3.com"
[[acme.domains]]
   main = "local4.com"
```
