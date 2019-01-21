# Zookeeper Provider

Traefik can be configured to use Zookeeper as a provider.

```toml
################################################################
# Zookeeper Provider
################################################################

# Enable Zookeeper Provider.
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
# Default: "traefik"
#
prefix = "traefik"

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
#    insecureSkipVerify = true
```

To enable constraints see [provider-specific constraints section](/configuration/commons/#provider-specific).

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on Traefik KV structure.
