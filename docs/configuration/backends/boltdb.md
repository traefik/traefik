# BoltDB Provider

Traefik can be configured to use BoltDB as a provider.

```toml
################################################################
# BoltDB Provider
################################################################

# Enable BoltDB Provider.
[boltdb]

# BoltDB file.
#
# Required
# Default: "127.0.0.1:4001"
#
endpoint = "/my.db"

# Enable watch BoltDB changes.
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
filename = "boltdb.tmpl"

# Use BoltDB user/pass authentication.
#
# Optional
#
# username = foo
# password = bar

# Enable BoltDB TLS connection.
#
# Optional
#
#    [boltdb.tls]
#    ca = "/etc/ssl/ca.crt"
#    cert = "/etc/ssl/boltdb.crt"
#    key = "/etc/ssl/boltdb.key"
#    insecureSkipVerify = true
```

To enable constraints see [provider-specific constraints section](/configuration/commons/#provider-specific).
