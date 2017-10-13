# BoltDB Backend

Træfik can be configured to use BoltDB as a backend configuration.

```toml
################################################################
# BoltDB configuration backend
################################################################

# Enable BoltDB configuration backend.
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
#    insecureskipverify = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).
