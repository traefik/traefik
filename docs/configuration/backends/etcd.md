# Etcd Backend

Træfik can be configured to use Etcd as a backend configuration.

```toml
################################################################
# Etcd configuration backend
################################################################

# Enable Etcd configuration backend.
[etcd]

# Etcd server endpoint.
#
# Required
# Default: "127.0.0.1:2379"
#
endpoint = "127.0.0.1:2379"

# Enable watch Etcd changes.
#
# Optional
# Default: true
#
watch = true

# Prefix used for KV store.
#
# Optional
# Default: "/traefik"
#
prefix = "/traefik"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "etcd.tmpl"

# Use etcd user/pass authentication.
#
# Optional
#
# username = foo
# password = bar

# Enable etcd TLS connection.
#
# Optional
#
#    [etcd.tls]
#    ca = "/etc/ssl/ca.crt"
#    cert = "/etc/ssl/etcd.crt"
#    key = "/etc/ssl/etcd.key"
#    insecureskipverify = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on Traefik KV structure.
