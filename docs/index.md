![Træfɪk](http://traefik.github.io/traefik.logo.svg "Træfɪk")
___


# <a id="top"></a> Documentation

- [Basics](#basics)
- [Launch configuration](#launch)
- [Global configuration](#global)
- [File backend](#file)
- [API backend](#api)
- [Docker backend](#docker)
- [Mesos/Marathon backend](#marathon)
- [Consul backend](#consul)
- [Etcd backend](#etcd)
- [Zookeeper backend](#zk)
- [Boltdb backend](#boltdb)
- [Benchmarks](#benchmarks)


## <a id="basics"></a> Basics


Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends ([Docker :whale:](https://www.docker.com/), [Mesos/Marathon](https://mesosphere.github.io/marathon/), [Consul](https://consul.io/), [Etcd](https://coreos.com/etcd/), Rest API, file...) to manage its configuration automatically and dynamically.

Basically, Træfɪk is a http router, which sends traffic from frontends to http backends, following rules you have configured.

### <a id="frontends"></a> Frontends

Frontends can be defined using the following rules:

- `Headers`: Headers adds a matcher for request header values. It accepts a sequence of key/value pairs to be matched. For example: `application/json`
- `HeadersRegexp`: Regular expressions can be used with headers as well. It accepts a sequence of key/value pairs, where the value has regex support. For example: `application/(text|json)`
- `Host`: Host adds a matcher for the URL host. It accepts a template with zero or more URL variables enclosed by `{}`. Variables can define an optional regexp pattern to be matched: `www.traefik.io`, `{subdomain:[a-z]+}.traefik.io`
- `Methods`: Methods adds a matcher for HTTP methods. It accepts a sequence of one or more methods to be matched, e.g.: `GET`, `POST`, `PUT`
- `Path`: Path adds a matcher for the URL path. It accepts a template with zero or more URL variables enclosed by `{}`. The template must start with a `/`. For exemple `/products/` `/articles/{category}/{id:[0-9]+}`
- `PathPrefix`: PathPrefix adds a matcher for the URL path prefix. This matches if the given template is a prefix of the full URL path.


 A frontend is a set of rules that forwards the incoming http traffic to a backend.
 You can optionally enable `passHostHeader` to
- []forward client `Host` header to the backend.

### HTTP Backends

A backend is responsible to load-balance the traffic coming from one or more frontends to a set of http servers.
Various methods of load-balancing is supported:

- `wrr`: Weighted Round Robin
- `drr`: Dynamic Round Robin: increases weights on servers that perform better than others. It also rolls back to original weights if the servers have changed.

A circuit breaker can also be applied to a backend, preventing high loads on failing servers.
It can be configured using:

- Methods: `LatencyAtQuantileMS`, `NetworkErrorRatio`, `ResponseCodeRatio`
- Operators:  `AND`, `OR`, `EQ`, `NEQ`, `LT`, `LE`, `GT`, `GE`

For example:
- `NetworkErrorRatio() > 0.5`
- `LatencyAtQuantileMS(50.0) > 50`
- `ResponseCodeRatio(500, 600, 0, 600) > 0.5`


## <a id="launch"></a> Launch configuration

Træfɪk can be configured using a TOML file configuration, arguments, or both.
By default, Træfɪk will try to find a `traefik.toml` in the following places:
- `/etc/traefik/`
- `$HOME/.traefik/`
- `.` the working directory

You can override this by setting a `configFile` argument:

```bash
$ traefik --configFile=foo/bar/myconfigfile.toml
```

Træfɪk uses the following precedence order. Each item takes precedence over the item below it:

- arguments
- configuration file
- default

It means that arguments overrides configuration file.
Each argument is described in the help section:
```bash
$ traefik --help
traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at http://traefik.io

Usage:
  traefik [flags]
  traefik [command]

Available Commands:
  version     Print version
  help        Help about any command

Flags:
      --accessLogsFile="log/access.log": Access logs file
      --boltdb=false: Enable Boltdb backend
      --boltdb.endpoint="127.0.0.1:4001": Boltdb server endpoint
      --boltdb.filename="": Override default configuration template. For advanced users :)
      --boltdb.prefix="/traefik": Prefix used for KV store
      --boltdb.watch=true: Watch provider
      --certificates=: SSL certificates and keys. You may add several certificate/key pairs to terminate HTTPS for multiple domain names using TLS SNI
  -c, --configFile="": Configuration file to use (TOML, JSON, YAML, HCL).
      --consul=false: Enable Consul backend
      --consul.endpoint="127.0.0.1:8500": Consul server endpoint
      --consul.filename="": Override default configuration template. For advanced users :)
      --consul.prefix="/traefik": Prefix used for KV store
      --consul.watch=true: Watch provider
      --docker=false: Enable Docker backend
      --docker.domain="": Default domain used
      --docker.endpoint="unix:///var/run/docker.sock": Docker server endpoint. Can be a tcp or a unix socket endpoint
      --docker.filename="": Override default configuration template. For advanced users :)
      --docker.tls=false: Enable Docker TLS support
      --docker.tls.ca="": TLS CA
      --docker.tls.cert="": TLS cert
      --docker.tls.insecureSkipVerify=false: TLS insecure skip verify
      --docker.tls.key="": TLS key
      --docker.watch=true: Watch provider
      --etcd=false: Enable Etcd backend
      --etcd.endpoint="127.0.0.1:4001": Etcd server endpoint
      --etcd.filename="": Override default configuration template. For advanced users :)
      --etcd.prefix="/traefik": Prefix used for KV store
      --etcd.watch=true: Watch provider
      --file=false: Enable File backend
      --file.filename="": Override default configuration template. For advanced users :)
      --file.watch=true: Watch provider
  -g, --graceTimeOut="10": Timeout in seconds. Duration to give active requests a chance to finish during hot-reloads
  -h, --help=false: help for traefik
  -l, --logLevel="ERROR": Log level
      --marathon=false: Enable Marathon backend
      --marathon.domain="": Default domain used
      --marathon.endpoint="http://127.0.0.1:8080": Marathon server endpoint. You can also specify multiple endpoint for Marathon
      --marathon.filename="": Override default configuration template. For advanced users :)
      --marathon.networkInterface="eth0": Network interface used to call Marathon web services. Needed in case of multiple network interfaces
      --marathon.watch=true: Watch provider
  -p, --port=":80": Reverse proxy port
      --providersThrottleDuration=2s: Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time.
      --traefikLogsFile="log/traefik.log": Traefik logs file
      --web=false: Enable Web backend
      --web.address=":8080": Web administration port
      --web.cerFile="": SSL certificate
      --web.keyFile="": SSL certificate
      --web.readOnly=false: Enable read only API
      --zookeeper=false: Enable Zookeeper backend
      --zookeeper.endpoint="127.0.0.1:2181": Zookeeper server endpoint
      --zookeeper.filename="": Override default configuration template. For advanced users :)
      --zookeeper.prefix="/traefik": Prefix used for KV store
      --zookeeper.watch=true: Watch provider


Use "traefik help [command]" for more information about a command.
```

## <a id="global"></a> Global configuration

```toml
# traefik.toml
################################################################
# Global configuration
################################################################

# Reverse proxy port
#
# Optional
# Default: ":80"
#
# port = ":80"

# Timeout in seconds.
# Duration to give active requests a chance to finish during hot-reloads
#
# Optional
# Default: 10
#
# graceTimeOut = 10

# Traefik logs file
# If not defined, logs to stdout
#
# Optional
#
# traefikLogsFile = "log/traefik.log"

# Access logs file
#
# Optional
#
# accessLogsFile = "log/access.log"

# Log level
#
# Optional
# Default: "ERROR"
#
# logLevel = "ERROR"

# SSL certificates and keys
# You may add several certificate/key pairs to terminate HTTPS for multiple domain names using TLS SNI
#
# Optional
#
# [[certificates]]
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"

# Backends throttle duration: minimum duration between 2 events from providers
# before applying a new configuration. It avoids unnecessary reloads if multiples events
# are sent in a short amount of time.
#
# Optional
# Default: "2s"
#
# ProvidersThrottleDuration = "5s"

```


## <a id="file"></a> File backend

Like any other reverse proxy, Træfɪk can be configured with a file. You have two choices:

- simply add your configuration at the end of the global configuration file `traefik.toml` :

```toml
# traefik.toml
port = ":80"
graceTimeOut = 10
logLevel = "DEBUG"

[file]

# rules
[backends]
  [backends.backend1]
    [backends.backend1.circuitbreaker]
      expression = "NetworkErrorRatio() > 0.5"
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend2.LoadBalancer]
      method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2

[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host"
    value = "test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
    [frontends.frontend2.routes.test_2]
    rule = "Path"
    value = "/test"
```

- or put your rules in a separate file, for example `rules.tml`:

```toml
# traefik.toml
port = ":80"
graceTimeOut = 10
logLevel = "DEBUG"

[file]
filename = "rules.toml"
```

```toml
# rules.toml
[backends]
  [backends.backend1]
    [backends.backend1.circuitbreaker]
      expression = "NetworkErrorRatio() > 0.5"
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend2.LoadBalancer]
      method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2

[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host"
    value = "test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
    [frontends.frontend2.routes.test_2]
    rule = "Path"
    value = "/test"

```

If you want Træfɪk to watch file changes automatically, just add:

```toml
[file]
watch = true
```

## <a id="api"></a> API backend

Træfik can be configured using a restful api.
To enable it:

```toml
[web]
address = ":8080"

# SSL certificate and key used
#
# Optional
#
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"
#
# Set REST API to read-only mode
#
# Optional
# ReadOnly = false
```

- `/`: provides a simple HTML frontend of Træfik

![Web UI Providers](img/web.frontend.png)
![Web UI Health](img/traefik-health.png)

- `/health`: `GET` json metrics

```sh
$ curl -s "http://localhost:8080/health" | jq .
{
  // Træfɪk PID
  "pid": 2458,
  // Træfɪk server uptime (formated time)
  "uptime": "39m6.885931127s",
  //  Træfɪk server uptime in seconds
  "uptime_sec": 2346.885931127,
  // current server date
  "time": "2015-10-07 18:32:24.362238909 +0200 CEST",
  // current server date in seconds
  "unixtime": 1444235544,
  // count HTTP response status code in realtime
  "status_code_count": {
    "502": 1
  },
  // count HTTP response status code since Træfɪk started
  "total_status_code_count": {
    "200": 7,
    "404": 21,
    "502": 13
  },
  // count HTTP response
  "count": 1,
  // count HTTP response
  "total_count": 41,
  // sum of all response time (formated time)
  "total_response_time": "35.456865605s",
  // sum of all response time in seconds
  "total_response_time_sec": 35.456865605,
  // average response time (formated time)
  "average_response_time": "864.8016ms",
  // average response time in seconds
  "average_response_time_sec": 0.8648016000000001
}
```

- `/api`: `GET` configuration for all providers

```sh
$ curl -s "http://localhost:8080/api" | jq .
{
  "file": {
    "frontends": {
      "frontend2": {
        "routes": {
          "test_2": {
            "value": "/test",
            "rule": "Path"
          }
        },
        "backend": "backend1"
      },
      "frontend1": {
        "routes": {
          "test_1": {
            "value": "test.localhost",
            "rule": "Host"
          }
        },
        "backend": "backend2"
      }
    },
    "backends": {
      "backend2": {
        "loadBalancer": {
          "method": "drr"
        },
        "servers": {
          "server2": {
            "weight": 2,
            "URL": "http://172.17.0.5:80"
          },
          "server1": {
            "weight": 1,
            "url": "http://172.17.0.4:80"
          }
        }
      },
      "backend1": {
        "loadBalancer": {
          "method": "wrr"
        },
        "circuitBreaker": {
          "expression": "NetworkErrorRatio() > 0.5"
        },
        "servers": {
          "server2": {
            "weight": 1,
            "url": "http://172.17.0.3:80"
          },
          "server1": {
            "weight": 10,
            "url": "http://172.17.0.2:80"
          }
        }
      }
    }
  }
}
```

- `/api/providers`: `GET` providers
- `/api/providers/{provider}`: `GET` or `PUT` provider
- `/api/providers/{provider}/backends`: `GET` backends
- `/api/providers/{provider}/backends/{backend}`: `GET` a backend
- `/api/providers/{provider}/backends/{backend}/servers`: `GET` servers in a backend
- `/api/providers/{provider}/backends/{backend}/servers/{server}`: `GET` a server in a backend
- `/api/providers/{provider}/frontends`: `GET` frontends
- `/api/providers/{provider}/frontends/{frontend}`: `GET` a frontend
- `/api/providers/{provider}/frontends/{frontend}/routes`: `GET` routes in a frontend
- `/api/providers/{provider}/frontends/{frontend}/routes/{route}`: `GET` a route in a frontend


## <a id="docker"></a> Docker backend

Træfɪk can be configured to use Docker as a backend configuration:

```toml
################################################################
# Docker configuration backend
################################################################

# Enable Docker configuration backend
#
# Optional
#
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

# Enable docker TLS connection
#
#  [docker.tls]
#  ca = "/etc/ssl/ca.crt"
#  cert = "/etc/ssl/docker.crt"
#  key = "/etc/ssl/docker.key"
#  insecureskipverify = true
```

Labels can be used on containers to override default behaviour:

- `traefik.backend=foo`: assign the container to `foo` backend
- `traefik.port=80`: register this port. Useful when the container exposes multiples ports.
- `traefik.protocol=https`: override the default `http` protocol
- `traefik.weight=10`: assign this weight to the container
- `traefik.enable=false`: disable this container in Træfɪk
- `traefik.frontend.rule=Host`: override the default frontend rule (Default: Host). See [frontends](#frontends).
- `traefik.frontend.value=test.example.com`: override the default frontend value (Default: `{containerName}.{domain}`) See [frontends](#frontends). Must be associated with label traefik.frontend.rule.
- `traefik.frontend.passHostHeader=true`: forward client `Host` header to the backend.
* `traefik.domain=traefik.localhost`: override the default domain


## <a id="marathon"></a> Marathon backend

Træfɪk can be configured to use Marathon as a backend configuration:


```toml
################################################################
# Mesos/Marathon configuration backend
################################################################

# Enable Marathon configuration backend
#
# Optional
#
[marathon]

# Marathon server endpoint.
# You can also specify multiple endpoint for Marathon:
# endpoint := "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
#
# Required
#
endpoint = "http://127.0.0.1:8080"

# Network interface used to call Marathon web services
# Optional
# Default: "eth0"
#
# networkInterface = "eth0"

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
```

Labels can be used on containers to override default behaviour:

- `traefik.backend=foo`: assign the application to `foo` backend
- `traefik.portIndex=1`: register port by index in the application's ports array. Useful when the application exposes multiple ports.
- `traefik.port=80`: register the explicit application port value. Cannot be used alongside `traefik.portIndex`.
- `traefik.protocol=https`: override the default `http` protocol
- `traefik.weight=10`: assign this weight to the application
- `traefik.enable=false`: disable this application in Træfɪk
- `traefik.frontend.rule=Host`: override the default frontend rule (Default: Host). See [frontends](#frontends).
- `traefik.frontend.value=test.example.com`: override the default frontend value (Default: `{appName}.{domain}`) See [frontends](#frontends). Must be associated with label traefik.frontend.rule.
- `traefik.frontend.passHostHeader=true`: forward client `Host` header to the backend.
* `traefik.domain=traefik.localhost`: override the default domain

## <a id="consul"></a> Consul backend

Træfɪk can be configured to use Consul as a backend configuration:

```toml
################################################################
# Consul KV configuration backend
################################################################

# Enable Consul KV configuration backend
#
# Optional
#
[consul]

# Consul server endpoint
#
# Required
#
endpoint = "127.0.0.1:8500"

# Enable watch Consul changes
#
# Optional
#
watch = true

# Prefix used for KV store.
#
# Optional
#
prefix = "traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "consul.tmpl"
```

The Keys-Values structure should look (using `prefix = "/traefik"`):

- backend 1

| Key                                                    | Value                       |
|--------------------------------------------------------|-----------------------------|
| `/traefik/backends/backend1/circuitbreaker/expression` | `NetworkErrorRatio() > 0.5` |
| `/traefik/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik/backends/backend1/servers/server2/weight`    | `1`                         |

- backend 2

| Key                                                 | Value                  |
|-----------------------------------------------------|------------------------|
| `/traefik/backends/backend2/loadbalancer/method`    | `drr`                  |
| `/traefik/backends/backend2/servers/server1/url`    | `http://172.17.0.4:80` |
| `/traefik/backends/backend2/servers/server1/weight` | `1`                    |
| `/traefik/backends/backend2/servers/server2/url`    | `http://172.17.0.5:80` |
| `/traefik/backends/backend2/servers/server2/weight` | `2`                    |

- frontend 1

| Key                                                | Value            |
|----------------------------------------------------|------------------|
| `/traefik/frontends/frontend1/backend`             | `backend2`       |
| `/traefik/frontends/frontend1/routes/test_1/rule`  | `Host`           |
| `/traefik/frontends/frontend1/routes/test_1/value` | `test.localhost` |

- frontend 2

| Key                                                | Value      |
|----------------------------------------------------|------------|
| `/traefik/frontends/frontend2/backend`             | `backend1` |
| `/traefik/frontends/frontend2/passHostHeader`      | `true`     |
| `/traefik/frontends/frontend2/routes/test_2/rule`  | `Path`     |
| `/traefik/frontends/frontend2/routes/test_2/value` | `/test`    |


## <a id="etcd"></a> Etcd backend

Træfɪk can be configured to use Etcd as a backend configuration:

```toml
################################################################
# Etcd configuration backend
################################################################

# Enable Etcd configuration backend
#
# Optional
#
# [etcd]

# Etcd server endpoint
#
# Required
#
# endpoint = "127.0.0.1:4001"

# Enable watch Etcd changes
#
# Optional
#
# watch = true

# Prefix used for KV store.
#
# Optional
#
# prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "etcd.tmpl"
```

The Keys-Values structure should look (using `prefix = "/traefik"`):

- backend 1

| Key                                                    | Value                       |
|--------------------------------------------------------|-----------------------------|
| `/traefik/backends/backend1/circuitbreaker/expression` | `NetworkErrorRatio() > 0.5` |
| `/traefik/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik/backends/backend1/servers/server2/weight`    | `1`                         |

- backend 2

| Key                                                 | Value                  |
|-----------------------------------------------------|------------------------|
| `/traefik/backends/backend2/loadbalancer/method`    | `drr`                  |
| `/traefik/backends/backend2/servers/server1/url`    | `http://172.17.0.4:80` |
| `/traefik/backends/backend2/servers/server1/weight` | `1`                    |
| `/traefik/backends/backend2/servers/server2/url`    | `http://172.17.0.5:80` |
| `/traefik/backends/backend2/servers/server2/weight` | `2`                    |

- frontend 1

| Key                                                | Value            |
|----------------------------------------------------|------------------|
| `/traefik/frontends/frontend1/backend`             | `backend2`       |
| `/traefik/frontends/frontend1/routes/test_1/rule`  | `Host`           |
| `/traefik/frontends/frontend1/routes/test_1/value` | `test.localhost` |

- frontend 2

| Key                                                | Value      |
|----------------------------------------------------|------------|
| `/traefik/frontends/frontend2/backend`             | `backend1` |
| `/traefik/frontends/frontend2/passHostHeader`      | `true`     |
| `/traefik/frontends/frontend2/routes/test_2/rule`  | `Path`     |
| `/traefik/frontends/frontend2/routes/test_2/value` | `/test`    |


## <a id="zk"></a> Zookeeper backend

Træfɪk can be configured to use Zookeeper as a backend configuration:

```toml
################################################################
# Zookeeper configuration backend
################################################################

# Enable Zookeeperconfiguration backend
#
# Optional
#
# [zookeeper]

# Zookeeper server endpoint
#
# Required
#
# endpoint = "127.0.0.1:2181"

# Enable watch Zookeeper changes
#
# Optional
#
# watch = true

# Prefix used for KV store.
#
# Optional
#
# prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "zookeeper.tmpl"
```
The Keys-Values structure should look (using `prefix = "/traefik"`):

- backend 1

| Key                                                    | Value                       |
|--------------------------------------------------------|-----------------------------|
| `/traefik/backends/backend1/circuitbreaker/expression` | `NetworkErrorRatio() > 0.5` |
| `/traefik/backends/backend1/servers/server1/url`       | `http://172.17.0.2:80`      |
| `/traefik/backends/backend1/servers/server1/weight`    | `10`                        |
| `/traefik/backends/backend1/servers/server2/url`       | `http://172.17.0.3:80`      |
| `/traefik/backends/backend1/servers/server2/weight`    | `1`                         |

- backend 2

| Key                                                 | Value                  |
|-----------------------------------------------------|------------------------|
| `/traefik/backends/backend2/loadbalancer/method`    | `drr`                  |
| `/traefik/backends/backend2/servers/server1/url`    | `http://172.17.0.4:80` |
| `/traefik/backends/backend2/servers/server1/weight` | `1`                    |
| `/traefik/backends/backend2/servers/server2/url`    | `http://172.17.0.5:80` |
| `/traefik/backends/backend2/servers/server2/weight` | `2`                    |

- frontend 1

| Key                                               | Value            |
|---------------------------------------------------|------------------|
| `/traefik/frontends/frontend1/backend             | `backend2`       |
| `/traefik/frontends/frontend1/routes/test_1/rule  | `Host`           |
| `/traefik/frontends/frontend1/routes/test_1/value | `test.localhost` |

- frontend 2

| Key                                                | Value      |
|----------------------------------------------------|------------|
| `/traefik/frontends/frontend2/backend`             | `backend1` |
| `/traefik/frontends/frontend2/passHostHeader`      | `true`     |
| `/traefik/frontends/frontend2/routes/test_2/rule`  | `Path`     |
| `/traefik/frontends/frontend2/routes/test_2/value` | `/test`    |


## <a id="boltdb"></a> BoltDB backend

Træfɪk can be configured to use BoltDB as a backend configuration:

```toml
################################################################
# BoltDB configuration backend
################################################################

# Enable BoltDB configuration backend
#
# Optional
#
# [boltdb]

# BoltDB file
#
# Required
#
# endpoint = "/my.db"

# Enable watch BoltDB changes
#
# Optional
#
# watch = true

# Prefix used for KV store.
#
# Optional
#
# prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "boltdb.tmpl"
```


## <a id="benchmarks"></a> Benchmarks

Here are some early Benchmarks between Nginx and Træfɪk acting as simple load balancers between two servers.

- Nginx:

```sh
$ docker run -d -e VIRTUAL_HOST=test1.localhost emilevauge/whoami
$ docker run -d -e VIRTUAL_HOST=test1.localhost emilevauge/whoami
$ docker run --log-driver=none -d -p 80:80 -v /var/run/docker.sock:/tmp/docker.sock:ro jwilder/nginx-proxy
$ ab -n 20000 -c 20  -r http://test1.localhost/
This is ApacheBench, Version 2.3 <$Revision: 1528965 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking test1.localhost (be patient)
Completed 2000 requests
Completed 4000 requests
Completed 6000 requests
Completed 8000 requests
Completed 10000 requests
Completed 12000 requests
Completed 14000 requests
Completed 16000 requests
Completed 18000 requests
Completed 20000 requests
Finished 20000 requests


Server Software:        nginx/1.9.2
Server Hostname:        test1.localhost
Server Port:            80

Document Path:          /
Document Length:        287 bytes

Concurrency Level:      20
Time taken for tests:   5.874 seconds
Complete requests:      20000
Failed requests:        0
Total transferred:      8900000 bytes
HTML transferred:       5740000 bytes
Requests per second:    3404.97 [#/sec] (mean)
Time per request:       5.874 [ms] (mean)
Time per request:       0.294 [ms] (mean, across all concurrent requests)
Transfer rate:          1479.70 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.1      0       2
Processing:     0    6   2.4      6      35
Waiting:        0    5   2.3      5      33
Total:          0    6   2.4      6      36

Percentage of the requests served within a certain time (ms)
  50%      6
  66%      6
  75%      7
  80%      7
  90%      9
  95%     10
  98%     12
  99%     13
 100%     36 (longest request)
```

- Træfɪk:

```sh
docker run -d -l traefik.backend=test1 -l traefik.frontend.rule=Host -l traefik.frontend.value=test1.docker.localhost emilevauge/whoami
docker run -d -l traefik.backend=test1 -l traefik.frontend.rule=Host -l traefik.frontend.value=test1.docker.localhost emilevauge/whoami
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/traefik.toml -v /var/run/docker.sock:/var/run/docker.sock emilevauge/traefik
$ ab -n 20000 -c 20  -r http://test1.docker.localhost/
This is ApacheBench, Version 2.3 <$Revision: 1528965 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking test1.docker.localhost (be patient)
Completed 2000 requests
Completed 4000 requests
Completed 6000 requests
Completed 8000 requests
Completed 10000 requests
Completed 12000 requests
Completed 14000 requests
Completed 16000 requests
Completed 18000 requests
Completed 20000 requests
Finished 20000 requests


Server Software:        .
Server Hostname:        test1.docker.localhost
Server Port:            80

Document Path:          /
Document Length:        312 bytes

Concurrency Level:      20
Time taken for tests:   6.545 seconds
Complete requests:      20000
Failed requests:        0
Total transferred:      8600000 bytes
HTML transferred:       6240000 bytes
Requests per second:    3055.60 [#/sec] (mean)
Time per request:       6.545 [ms] (mean)
Time per request:       0.327 [ms] (mean, across all concurrent requests)
Transfer rate:          1283.11 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.2      0       7
Processing:     1    6   2.2      6      22
Waiting:        1    6   2.1      6      21
Total:          1    7   2.2      6      22

Percentage of the requests served within a certain time (ms)
  50%      6
  66%      7
  75%      8
  80%      8
  90%      9
  95%     10
  98%     11
  99%     13
 100%     22 (longest request)
```
