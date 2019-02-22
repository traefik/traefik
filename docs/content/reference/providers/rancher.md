# Rancher -- Reference

## Rancher

```toml
################################################################
# Rancher Provider
################################################################

# Enable Docker Provider.
[rancher]

# The default host rule for all services.
#
# Optionnal
#
DefaultRule = "unix:///var/run/docker.sock"

# Expose Rancher services by default in Traefik.
#
# Optional
#
ExposedByDefault = "docker.localhost"

# Enable watch docker changes.
#
# Optional
#
watch = true

# Filter services with unhealthy states and inactive states.
#
# Optional
#
EnableServiceHealthFilter = true

# Defines the polling interval (in seconds).
#
# Optional
#
RefreshSeconds = true

# Poll the Rancher metadata service for changes every `rancher.refreshSeconds`, which is less accurate
#
# Optional
#
IntervalPoll = false

# Prefix used for accessing the Rancher metadata service
#
# Optional
#
Prefix = 15
```

