# Consul Key-Value Backend

Træfik can be configured to use Consul as a backend configuration.

```toml
################################################################
# Consul KV configuration backend
################################################################

# Enable Consul KV configuration backend.
[consul]

# Consul server endpoint.
#
# Required
# Default: "127.0.0.1:8500"
#
endpoint = "127.0.0.1:8500"

# Enable watch Consul changes.
#
# Optional
# Default: true
#
watch = true

# Prefix used for KV store.
#
# Optional
# Default: traefik
#
prefix = "traefik"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "consul.tmpl"

# Use Consul user/pass authentication.
#
# Optional
#
# username = foo
# password = bar

# Enable Consul TLS connection.
#
# Optional
#
#    [consul.tls]
#    ca = "/etc/ssl/ca.crt"
#    cert = "/etc/ssl/consul.crt"
#    key = "/etc/ssl/consul.key"
#    insecureskipverify = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on Traefik KV structure.
