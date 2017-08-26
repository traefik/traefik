# Marathon Backend

Træfik can be configured to use Marathon as a backend configuration:

```toml
################################################################
# Mesos/Marathon configuration backend
################################################################

# Enable Marathon configuration backend
[marathon]

# Marathon server endpoint.
# You can also specify multiple endpoint for Marathon:
# endpoint := "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
#
# Required
#
endpoint = "http://127.0.0.1:8080"

# Enable watch Marathon changes
#
# Optional
#
watch = true

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on an application.
#
# Required
#
domain = "marathon.localhost"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "marathon.tmpl"

# Expose Marathon apps by default in traefik
#
# Optional
# Default: true
#
# exposedByDefault = true

# Convert Marathon groups to subdomains
# Default behavior: /foo/bar/myapp => foo-bar-myapp.{defaultDomain}
# with groupsAsSubDomains enabled: /foo/bar/myapp => myapp.bar.foo.{defaultDomain}
#
# Optional
# Default: false
#
# groupsAsSubDomains = true

# Enable compatibility with marathon-lb labels
#
# Optional
# Default: false
#
# marathonLBCompatibility = true

# Enable Marathon basic authentication
#
# Optional
#
#  [marathon.basic]
#  httpBasicAuthUser = "foo"
#  httpBasicPassword = "bar"

# TLS client configuration. https://golang.org/pkg/crypto/tls/#Config
#
# Optional
#
# [marathon.TLS]
# CA = "/etc/ssl/ca.crt"
# Cert = "/etc/ssl/marathon.cert"
# Key = "/etc/ssl/marathon.key"
# InsecureSkipVerify = true

# DCOSToken for DCOS environment, This will override the Authorization header
#
# Optional
#
# dcosToken = "xxxxxx"

# Override DialerTimeout
# Amount of time to allow the Marathon provider to wait to open a TCP connection
# to a Marathon master.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming
# seconds.
#
# Optional
# Default: "60s"
# dialerTimeout = "60s"

# Set the TCP Keep Alive interval for the Marathon HTTP Client.
# Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw
# values (digits). If no units are provided, the value is parsed assuming
# seconds.
#
# Optional
# Default: "10s"
#
# keepAlive = "10s"

# By default, a task's IP address (as returned by the Marathon API) is used as
# backend server if an IP-per-task configuration can be found; otherwise, the
# name of the host running the task is used.
# The latter behavior can be enforced by enabling this switch.
#
# Optional
# Default: false
#
# forceTaskHostname = false

# Applications may define readiness checks which are probed by Marathon during
# deployments periodically and the results exposed via the API. Enabling the
# following parameter causes Traefik to filter out tasks whose readiness checks
# have not succeeded.
# Note that the checks are only valid at deployment times. See the Marathon
# guide for details.
#
# Optional
# Default: false
#
# respectReadinessChecks = false
```

Labels can be used on containers to override default behaviour:

| Label                                                                                                                | Description                                                                                                                                                                        |
|----------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `traefik.backend=foo`                                                                                                | assign the application to `foo` backend                                                                                                                                            |
| `traefik.backend.maxconn.amount=10`                                                                                  | set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.                                                               |
| `traefik.backend.maxconn.extractorfunc=client.ip`                                                                    | set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect. |
| `traefik.backend.loadbalancer.method=drr`                                                                            | override the default `wrr` load balancer algorithm                                                                                                                                 |
| `traefik.backend.loadbalancer.sticky=true`                                                                           | enable backend sticky sessions                                                                                                                                                     |
| `traefik.backend.circuitbreaker.expression=NetworkErrorRatio() > 0.5`                                                | create a [circuit breaker](/basics/#backends) to be used against the backend                                                                                                       |
| `traefik.backend.healthcheck.path=/health`                                                                           | set the Traefik health check path [default: no health checks]                                                                                                                      |
| `traefik.backend.healthcheck.interval=5s`                                                                            | sets a custom health check interval in Go-parseable (`time.ParseDuration`) format [default: 30s]                                                                                   |
| `traefik.portIndex=1`                                                                                                | register port by index in the application's ports array. Useful when the application exposes multiple ports.                                                                       |
| `traefik.port=80`                                                                                                    | register the explicit application port value. Cannot be used alongside `traefik.portIndex`.                                                                                        |
| `traefik.protocol=https`                                                                                             | override the default `http` protocol                                                                                                                                               |
| `traefik.weight=10`                                                                                                  | assign this weight to the application                                                                                                                                              |
| `traefik.enable=false`                                                                                               | disable this application in Træfik                                                                                                                                                 |
| `traefik.frontend.rule=Host:test.traefik.io`                                                                         | override the default frontend rule (Default: `Host:{containerName}.{domain}`).                                                                                                     |
| `traefik.frontend.passHostHeader=true`                                                                               | forward client `Host` header to the backend.                                                                                                                                       |
| `traefik.frontend.priority=10`                                                                                       | override default frontend priority                                                                                                                                                 |
| `traefik.frontend.entryPoints=http,https`                                                                            | assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`.                                                                                           |
| `traefik.frontend.auth.basic=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0` | Sets basic authentication for that frontend with the usernames and passwords test:test and test2:test2, respectively                                                               |

If several ports need to be exposed from a container, the services labels can be used:

| Label                                                                                                                               | Description                                                                                          |
|-------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| `traefik.<service-name>.port=443`                                                                                                   | create a service binding with frontend/backend using this port. Overrides `traefik.port`.            |
| `traefik.<service-name>.portIndex=1`                                                                                                | create a service binding with frontend/backend using this port index. Overrides `traefik.portIndex`. |
| `traefik.<service-name>.protocol=https`                                                                                             | assign `https` protocol. Overrides `traefik.protocol`.                                               |
| `traefik.<service-name>.weight=10`                                                                                                  | assign this service weight. Overrides `traefik.weight`.                                              |
| `traefik.<service-name>.frontend.backend=fooBackend`                                                                                | assign this service frontend to `foobackend`. Default is to assign to the service backend.           |
| `traefik.<service-name>.frontend.entryPoints=http`                                                                                  | assign this service entrypoints. Overrides `traefik.frontend.entrypoints`.                           |
| `traefik.<service-name>.frontend.auth.basic=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0` | Sets a Basic Auth for that frontend with the users test:test and test2:test2.                        |
| `traefik.<service-name>.frontend.passHostHeader=true`                                                                               | Forward client `Host` header to the backend. Overrides `traefik.frontend.passHostHeader`.            |
| `traefik.<service-name>.frontend.priority=10`                                                                                       | assign the service frontend priority. Overrides `traefik.frontend.priority`.                         |
| `traefik.<service-name>.frontend.rule=Path:/foo`                                                                                    | assign the service frontend rule. Overrides `traefik.frontend.rule`.                                 |
