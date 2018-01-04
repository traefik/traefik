# Consul Catalog backend

Træfik can be configured to use service discovery catalog of Consul as a backend configuration.

```toml
################################################################
# Consul Catalog configuration backend
################################################################

# Enable Consul Catalog configuration backend.
[consulCatalog]

# Consul server endpoint.
#
# Required
# Default: "127.0.0.1:8500"
#
endpoint = "127.0.0.1:8500"

# Expose Consul catalog services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Default domain used.
#
# Optional
#
domain = "consul.localhost"

# Prefix for Consul catalog tags.
#
# Optional
# Default: "traefik"
#
prefix = "traefik"

# Default frontEnd Rule for Consul services.
#
# The format is a Go Template with:
# - ".ServiceName", ".Domain" and ".Attributes" available
# - "getTag(name, tags, defaultValue)", "hasTag(name, tags)" and "getAttribute(name, tags, defaultValue)" functions are available
# - "getAttribute(...)" function uses prefixed tag names based on "prefix" value
#
# Optional
# Default: "Host:{{.ServiceName}}.{{.Domain}}"
#
#frontEndRule = "Host:{{.ServiceName}}.{{.Domain}}"
```

This backend will create routes matching on hostname based on the service name used in Consul.

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

### Tags

Additional settings can be defined using Consul Catalog tags.

| Tag                                                       | Description                                                                                                                                                                        |
|-----------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.enable=false`                                    | Disable this container in Træfik                                                                                                                                                   |
| `traefik.protocol=https`                                  | Override the default `http` protocol                                                                                                                                               |
| `traefik.backend.weight=10`                               | Assign this weight to the container                                                                                                                                                |
| `traefik.backend.circuitbreaker=EXPR`                     | Create a [circuit breaker](/basics/#backends) to be used against the backend, ex: `NetworkErrorRatio() > 0.`                                                                       |
| `traefik.backend.maxconn.amount=10`                       | Set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.                                                               |
| `traefik.backend.maxconn.extractorfunc=client.ip`         | Set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect. |
| `traefik.frontend.rule=Host:test.traefik.io`              | Override the default frontend rule (Default: `Host:{{.ServiceName}}.{{.Domain}}`).                                                                                                 |
| `traefik.frontend.passHostHeader=true`                    | Forward client `Host` header to the backend.                                                                                                                                       |
| `traefik.frontend.priority=10`                            | Override default frontend priority                                                                                                                                                 |
| `traefik.frontend.entryPoints=http,https`                 | Assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`.                                                                                           |
| `traefik.frontend.auth.basic=EXPR`                        | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`                                                                                                   |
| `traefik.backend.loadbalancer=drr`                        | override the default `wrr` load balancer algorithm                                                                                                                                 |
| `traefik.backend.loadbalancer.stickiness=true`            | enable backend sticky sessions                                                                                                                                                     |
| `traefik.backend.loadbalancer.stickiness.cookieName=NAME` | Manually set the cookie name for sticky sessions                                                                                                                                   |
| `traefik.backend.loadbalancer.sticky=true`                | enable backend sticky sessions (DEPRECATED)                                                                                                                                        |

### Examples

If you want that Træfik uses Consul tags correctly you need to defined them like that:
```json
traefik.enable=true
traefik.tags=api
traefik.tags=external
```

If the prefix defined in Træfik configuration is `bla`, tags need to be defined like that:
```json
bla.enable=true
bla.tags=api
bla.tags=external
```