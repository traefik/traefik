# Docker Backend

Træfik can be configured to use Docker as a backend configuration.

## Docker

```toml
################################################################
# Docker configuration backend
################################################################

# Enable Docker configuration backend
[docker]

# Docker server endpoint. Can be a tcp or a unix socket endpoint.
#
# Required
#
endpoint = "unix:///var/run/docker.sock"

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on a container.
#
# Required
#
domain = "docker.localhost"

# Enable watch docker changes
#
# Optional
#
watch = true

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "docker.tmpl"

# Expose containers by default in traefik
# If set to false, containers that don't have `traefik.enable=true` will be ignored
#
# Optional
# Default: true
#
exposedbydefault = true

# Use the IP address from the binded port instead of the inner network one. For specific use-case :)
#
# Optional
# Default: false
#
usebindportip = true

# Use Swarm Mode services as data provider
#
# Optional
# Default: false
#
swarmmode = false

# Enable docker TLS connection
#
# Optional
#
#  [docker.tls]
#  ca = "/etc/ssl/ca.crt"
#  cert = "/etc/ssl/docker.crt"
#  key = "/etc/ssl/docker.key"
#  insecureskipverify = true
```

## Docker Swarm Mode

```toml
################################################################
# Docker Swarmmode configuration backend
################################################################

# Enable Docker configuration backend
[docker]

# Docker server endpoint. Can be a tcp or a unix socket endpoint.
#
# Required
# Default: "unix:///var/run/docker.sock"
#
endpoint = "tcp://127.0.0.1:2375"

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on a services.
#
# Optional
# Default: ""
#
domain = "docker.localhost"

# Enable watch docker changes
#
# Optional
#
watch = true

# Use Docker Swarm Mode as data provider
swarmmode = true

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "docker.tmpl"

# Expose services by default in traefik
#
# Optional
# Default: true
#
exposedbydefault = false

# Enable docker TLS connection
#
# Optional
#
#  [swarm.tls]
#  ca = "/etc/ssl/ca.crt"
#  cert = "/etc/ssl/docker.crt"
#  key = "/etc/ssl/docker.key"
#  insecureskipverify = true
```

## Labels can be used on containers to override default behaviour

| Label                                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                   |
|---------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.backend=foo`                             | Give the name `foo` to the generated backend for this container.                                                                                                                                                                                                                                                                                                                                                              |
| `traefik.backend.maxconn.amount=10`               | Set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.                                                                                                                                                                                                                                                                                                          |
| `traefik.backend.maxconn.extractorfunc=client.ip` | Set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect.                                                                                                                                                                                                                                            |
| `traefik.backend.loadbalancer.method=drr`         | Override the default `wrr` load balancer algorithm                                                                                                                                                                                                                                                                                                                                                                            |
| `traefik.backend.loadbalancer.sticky=true`        | Enable backend sticky sessions                                                                                                                                                                                                                                                                                                                                                                                                |
| `traefik.backend.loadbalancer.swarm=true`         | Use Swarm's inbuilt load balancer (only relevant under Swarm Mode).                                                                                                                                                                                                                                                                                                                                                           |
| `traefik.backend.circuitbreaker.expression=EXPR`  | Create a [circuit breaker](/basics/#backends) to be used against the backend                                                                                                                                                                                                                                                                                                                                                  |
| `traefik.port=80`                                 | Register this port. Useful when the container exposes multiples ports.                                                                                                                                                                                                                                                                                                                                                        |
| `traefik.protocol=https`                          | Override the default `http` protocol                                                                                                                                                                                                                                                                                                                                                                                          |
| `traefik.weight=10`                               | Assign this weight to the container                                                                                                                                                                                                                                                                                                                                                                                           |
| `traefik.enable=false`                            | Disable this container in Træfik                                                                                                                                                                                                                                                                                                                                                                                              |
| `traefik.frontend.rule=EXPR`                      | Override the default frontend rule. Default: `Host:{containerName}.{domain}` or `Host:{service}.{project_name}.{domain}` if you are using `docker-compose`.                                                                                                                                                                                                                                                                   |
| `traefik.frontend.passHostHeader=true`            | Forward client `Host` header to the backend.                                                                                                                                                                                                                                                                                                                                                                                  |
| `traefik.frontend.priority=10`                    | Override default frontend priority                                                                                                                                                                                                                                                                                                                                                                                            |
| `traefik.frontend.entryPoints=http,https`         | Assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`                                                                                                                                                                                                                                                                                                                                       |
| `traefik.frontend.auth.basic=EXPR`                | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`                                                                                                                                                                                                                                                                                                                                              |
| `traefik.frontend.whitelistSourceRange:RANGE`     | List of IP-Ranges which are allowed to access. An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access.                                                                                                                                                                                                           |
| `traefik.docker.network`                          | Set the docker network to use for connections to this container. If a container is linked to several networks, be sure to set the proper network name (you can check with docker inspect <container_id>) otherwise it will randomly pick one (depending on how docker is returning them). For instance when deploying docker `stack` from compose files, the compose defined networks will be prefixed with the `stack` name. |

### Services labels can be used for overriding default behaviour

| Label                                             | Description                                                                                      |
|---------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `traefik.<service-name>.port=PORT`                | Overrides `traefik.port`. If several ports need to be exposed, the service labels could be used. |
| `traefik.<service-name>.protocol`                 | Overrides `traefik.protocol`.                                                                    |
| `traefik.<service-name>.weight`                   | Assign this service weight. Overrides `traefik.weight`.                                          |
| `traefik.<service-name>.frontend.backend=BACKEND` | Assign this service frontend to `BACKEND`. Default is to assign to the service backend.          |
| `traefik.<service-name>.frontend.entryPoints`     | Overrides `traefik.frontend.entrypoints`                                                         |
| `traefik.<service-name>.frontend.auth.basic`      | Sets a Basic Auth for that frontend                                                              |
| `traefik.<service-name>.frontend.passHostHeader`  | Overrides `traefik.frontend.passHostHeader`.                                                     |
| `traefik.<service-name>.frontend.priority`        | Overrides `traefik.frontend.priority`.                                                           |
| `traefik.<service-name>.frontend.rule`            | Overrides `traefik.frontend.rule`.                                                               |

!!! warning
    when running inside a container, Træfik will need network access through:

    `docker network connect <network> <traefik-container>`