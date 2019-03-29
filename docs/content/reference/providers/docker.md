# Docker -- Reference

## Docker

```toml
################################################################
# Docker Provider
################################################################

# Enable Docker Provider.
[docker]

# Docker server endpoint. Can be a tcp or a unix socket endpoint.
#
# Required
#
endpoint = "unix:///var/run/docker.sock"

# Default base domain used for the frontend rules.
# Can be overridden by setting the "traefik.domain" label on a container.
#
# Optional
#
domain = "docker.localhost"

# Enable watch docker changes.
#
# Optional
#
watch = true

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "docker.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = 2

# Expose containers by default in Traefik.
# If set to false, containers that don't have `traefik.enable=true` will be ignored.
#
# Optional
# Default: true
#
exposedByDefault = true

# Use the IP address from the bound port instead of the inner network one.
#
# In case no IP address is attached to the bound port (or in case
# there is no bind), the inner network one will be used as a fallback.
#
# Optional
# Default: false
#
usebindportip = true

# Use Swarm Mode services as data provider.
#
# Optional
# Default: false
#
swarmMode = false

# Polling interval (in seconds) for Swarm Mode.
#
# Optional
# Default: 15
#
swarmModeRefreshSeconds = 15

# Define a default docker network to use for connections to all containers.
# Can be overridden by the traefik.docker.network label.
#
# Optional
#
network = "web"

# Enable docker TLS connection.
#
# Optional
#
#  [docker.tls]
#  ca = "/etc/ssl/ca.crt"
#  cert = "/etc/ssl/docker.crt"
#  key = "/etc/ssl/docker.key"
#  insecureSkipVerify = true
```

## Docker Swarm Mode

```toml
################################################################
# Docker Swarm Mode Provider
################################################################

# Enable Docker Provider.
[docker]

# Docker server endpoint.
# Can be a tcp or a unix socket endpoint.
#
# Required
# Default: "unix:///var/run/docker.sock"
#
# swarm classic (1.12-)
# endpoint = "tcp://127.0.0.1:2375"
# docker swarm mode (1.12+)
endpoint = "tcp://127.0.0.1:2377"

# Default base domain used for the frontend rules.
# Can be overridden by setting the "traefik.domain" label on a services.
#
# Optional
# Default: ""
#
domain = "docker.localhost"

# Enable watch docker changes.
#
# Optional
# Default: true
#
watch = true

# Use Docker Swarm Mode as data provider.
#
# Optional
# Default: false
#
swarmMode = true

# Define a default docker network to use for connections to all containers.
# Can be overridden by the traefik.docker.network label.
#
# Optional
#
network = "web"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "docker.tmpl"

# Override template version
# For advanced users :)
#
# Optional
# - "1": previous template version (must be used only with older custom templates, see "filename")
# - "2": current template version (must be used to force template version when "filename" is used)
#
# templateVersion = 2

# Expose services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Enable docker TLS connection.
#
# Optional
#
#  [docker.tls]
#  ca = "/etc/ssl/ca.crt"
#  cert = "/etc/ssl/docker.crt"
#  key = "/etc/ssl/docker.key"
#  insecureSkipVerify = true
```