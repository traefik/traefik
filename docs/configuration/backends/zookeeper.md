# Zookeeper Backend

Træfik can be configured to use Zookeeper as a backend configuration.

```toml
################################################################
# Zookeeper configuration backend
################################################################

# Enable Zookeeperconfiguration backend.
[zookeeper]

# Zookeeper server endpoint.
#
# Required
# Default: "127.0.0.1:2181"
#
endpoint = "127.0.0.1:2181"

# Enable watch Zookeeper changes.
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
# filename = "zookeeper.tmpl"

# Use Zookeeper user/pass authentication.
#
# Optional
#
# username = foo
# password = bar

# Enable Zookeeper TLS connection.
#
# Optional
#
#    [zookeeper.tls]
#    ca = "/etc/ssl/ca.crt"
#    cert = "/etc/ssl/zookeeper.crt"
#    key = "/etc/ssl/zookeeper.key"
#    insecureskipverify = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on Traefik KV structure.
