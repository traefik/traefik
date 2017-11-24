# Remote Backend

Træfik can be configured to use remote HTTP endpoint as a source of backend configuration.

```toml
################################################################
# Remote configuration backend
################################################################

# Enable Remote HTTP configuration backend.
[remote]

# Remote server endpoint.
#
# Required
#
url = "http://my.config.server/toml"

# Configure time to wait for configuration before trying again.
# The configured value will be appended to 'url' query parameter as 'wait=%ds' in seconds
#
#  
#
# Optional
#
#longPollDuration = "1m"

# Configure refresh interval if long-poll is not used/supported.
#
# Optional
#
# repeatInterval = "30s"

# TLS client configuration. https://golang.org/pkg/crypto/tls/#Config
#
# Optional
#
#    [remote.TLS]
#    CA = "/etc/ssl/ca.crt"
#    Cert = "/etc/ssl/https.cert"
#    Key = "/etc/ssl/https.key"
#    InsecureSkipVerify = true

#If you want Træfik to watch file changes automatically, just add:
watch = true
```

!!! warning
    If you are using `longPollDuration` make sure the server indeed supports Long Polling.  
    Otherwise requests will be fired immediately as soon as the previous ones succeed, possibly resulting in DDoS!
    If you want to query remote server for config at a pre-determinted interval, make sure to use `repeatInterval` option instead!  