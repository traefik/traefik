# Eureka Provider

Traefik can be configured to use Eureka as a provider.

```toml
################################################################
# Eureka Provider
################################################################

# Enable Eureka Provider.
[eureka]

# Eureka server endpoint.
#
# Required
#
endpoint = "http://my.eureka.server/eureka"

# Override default configuration time between refresh.
#
# Optional
# Default: 30s
#
refreshSeconds = "1m"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "eureka.tmpl"
```
