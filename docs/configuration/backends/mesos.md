# Mesos Generic Backend

Træfik can be configured to use Mesos as a backend configuration:

```toml
################################################################
# Mesos configuration backend
################################################################

# Enable Mesos configuration backend
[mesos]

# Mesos server endpoint.
# You can also specify multiple endpoint for Mesos:
# endpoint = "192.168.35.40:5050,192.168.35.41:5050,192.168.35.42:5050"
# endpoint = "zk://192.168.35.20:2181,192.168.35.21:2181,192.168.35.22:2181/mesos"
#
# Required
#
endpoint = "http://127.0.0.1:8080"

# Enable watch Mesos changes
#
# Optional
#
watch = true

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on an application.
#
# Required
#
domain = "mesos.localhost"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "mesos.tmpl"

# Expose Mesos apps by default in traefik
#
# Optional
# Default: false
#
# ExposedByDefault = true

# TLS client configuration. https://golang.org/pkg/crypto/tls/#Config
#
# Optional
#
# [mesos.TLS]
# InsecureSkipVerify = true

# Zookeeper timeout (in seconds)
#
# Optional
# Default: 30
#
# ZkDetectionTimeout = 30

# Polling interval (in seconds)
#
# Optional
# Default: 30
#
# RefreshSeconds = 30

# IP sources (e.g. host, docker, mesos, rkt)
#
# Optional
#
# IPSources = "host"

# HTTP Timeout (in seconds)
#
# Optional
# Default: 30
#
# StateTimeoutSecond = "30"
```
