# Etcd Provider

Traefik can be configured to use Etcd as a provider.

```toml
################################################################
# Etcd Provider
################################################################

# Enable Etcd Provider.
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

# Force to use API V3 (otherwise still use API V2)
#
# Deprecated
#
# Optional
# Default: false
#
useAPIV3 = true


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
#    insecureSkipVerify = true
```

To enable constraints see [provider-specific constraints section](/configuration/commons/#provider-specific).

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on Traefik KV structure.

!!! note
    The option `useAPIV3` allows using Etcd API V3 only if it's set to true.
    This option is **deprecated** and API V2 won't be supported in the future.
