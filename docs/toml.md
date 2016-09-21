
# Global configuration

## Main section

```toml
# traefik.toml
################################################################
# Global configuration
################################################################

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
# Accepted values, in order of severity: "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"
# Messages at and above the selected level will be logged.
#
# logLevel = "ERROR"

# Backends throttle duration: minimum duration between 2 events from providers
# before applying a new configuration. It avoids unnecessary reloads if multiples events
# are sent in a short amount of time.
#
# Optional
# Default: "2s"
#
# ProvidersThrottleDuration = "5s"

# If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used.
# If you encounter 'too many open files' errors, you can either change this value, or change `ulimit` value.
#
# Optional
# Default: http.DefaultMaxIdleConnsPerHost
#
# MaxIdleConnsPerHost = 200

# If set to true invalid SSL certificates are accepted for backends.
# Note: This disables detection of man-in-the-middle attacks so should only be used on secure backend networks.
# Optional
# Default: false
#
# InsecureSkipVerify = true

# Entrypoints to be used by frontends that do not specify any entrypoint.
# Each frontend can specify its own entrypoints.
#
# Optional
# Default: ["http"]
#
# defaultEntryPoints = ["http", "https"]
```

### Constraints

In a micro-service architecture, with a central service discovery, setting constraints limits Træfɪk scope to a smaller number of routes.

Træfɪk filters services according to service attributes/tags set in your configuration backends.

Supported backends:

- Docker
- Consul K/V
- BoltDB
- Zookeeper
- Etcd
- Consul Catalog

Supported filters:

- ```tag```

```
# Constraints definition

#
# Optional
#

# Simple matching constraint
# constraints = ["tag==api"]

# Simple mismatching constraint
# constraints = ["tag!=api"]

# Globbing
# constraints = ["tag==us-*"]

# Backend-specific constraint
# [consulCatalog]
#   endpoint = 127.0.0.1:8500
#   constraints = ["tag==api"]

# Multiple constraints
#   - "tag==" must match with at least one tag
#   - "tag!=" must match with none of tags
# constraints = ["tag!=us-*", "tag!=asia-*"]
# [consulCatalog]
#   endpoint = 127.0.0.1:8500
#   constraints = ["tag==api", "tag!=v*-beta"]
```

## Entrypoints definition

```toml
# Entrypoints definition
#
# Optional
# Default:
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#
# To redirect an http entrypoint to an https entrypoint (with SNI support):
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#     [entryPoints.http.redirect]
#       entryPoint = "https"
#   [entryPoints.https]
#   address = ":443"
#     [entryPoints.https.tls]
#       [[entryPoints.https.tls.certificates]]
#       CertFile = "integration/fixtures/https/snitest.com.cert"
#       KeyFile = "integration/fixtures/https/snitest.com.key"
#       [[entryPoints.https.tls.certificates]]
#       CertFile = "integration/fixtures/https/snitest.org.cert"
#       KeyFile = "integration/fixtures/https/snitest.org.key"
#
# To redirect an entrypoint rewriting the URL:
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#     [entryPoints.http.redirect]
#       regex = "^http://localhost/(.*)"
#       replacement = "http://mydomain/$1"
#
# Only accept clients that present a certificate signed by a specified
# Certificate Authority (CA)
# ClientCAFiles can be configured with multiple CA:s in the same file or
# use multiple files containing one or several CA:s. The CA:s has to be in PEM format.
# All clients will be required to present a valid cert.
# The requirement will apply to all server certs in the entrypoint
# In the example below both snitest.com and snitest.org will require client certs
#
# [entryPoints]
#   [entryPoints.https]
#   address = ":443"
#   [entryPoints.https.tls]
#   ClientCAFiles = ["tests/clientca1.crt", "tests/clientca2.crt"]
#     [[entryPoints.https.tls.certificates]]
#     CertFile = "integration/fixtures/https/snitest.com.cert"
#     KeyFile = "integration/fixtures/https/snitest.com.key"
#     [[entryPoints.https.tls.certificates]]
#     CertFile = "integration/fixtures/https/snitest.org.cert"
#     KeyFile = "integration/fixtures/https/snitest.org.key"
#
# To enable basic auth on an entrypoint
# with 2 user/pass: test:test and test2:test2
# Passwords can be encoded in MD5, SHA1 and BCrypt: you can use htpasswd to generate those ones
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#   [entryPoints.http.auth.basic]
#   users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
#
# To enable digest auth on an entrypoint
# with 2 user/realm/pass: test:traefik:test and test2:traefik:test2
# You can use htdigest to generate those ones
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#   [entryPoints.http.auth.basic]
#   users = ["test:traefik:a2688e031edb4be6a3797f3882655c05 ", "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
```

## Retry configuration

```toml
# Enable retry sending request if network error
#
# Optional
#
[retry]

# Number of attempts
#
# Optional
# Default: (number servers in backend) -1
#
# attempts = 3
```

## ACME (Let's Encrypt) configuration

```toml
# Sample entrypoint configuration when using ACME
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

# Enable ACME (Let's Encrypt): automatic SSL
#
# Optional
#
[acme]

# Email address used for registration
#
# Required
#
email = "test@traefik.io"

# File used for certificates storage.
# WARNING, if you use Traefik in Docker, you have 2 options:
#  - create a file on your host and mount it as a volume
#      storageFile = "acme.json"
#      $ docker run -v "/my/host/acme.json:acme.json" traefik
#  - mount the folder containing the file as a volume
#      storageFile = "/etc/traefik/acme/acme.json"
#      $ docker run -v "/my/host/acme:/etc/traefik/acme" traefik
#
# Required
#
storageFile = "acme.json"

# Entrypoint to proxy acme challenge to.
# WARNING, must point to an entrypoint on port 443
#
# Required
#
entryPoint = "https"

# Enable on demand certificate. This will request a certificate from Let's Encrypt during the first TLS handshake for a hostname that does not yet have a certificate.
# WARNING, TLS handshakes will be slow when requesting a hostname certificate for the first time, this can leads to DoS attacks.
# WARNING, Take note that Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits
#
# Optional
#
# onDemand = true

# Enable certificate generation on frontends Host rules. This will request a certificate from Let's Encrypt for each frontend with a Host rule.
# For example, a rule Host:test1.traefik.io,test2.traefik.io will request a certificate with main domain test1.traefik.io and SAN test2.traefik.io.
#
# Optional
#
# OnHostRule = true

# CA server to use
# Uncomment the line to run on the staging let's encrypt server
# Leave comment to go to prod
#
# Optional
#
# caServer = "https://acme-staging.api.letsencrypt.org/directory"

# Domains list
# You can provide SANs (alternative domains) to each main domain
# All domains must have A/AAAA records pointing to Traefik
# WARNING, Take note that Let's Encrypt have rate limiting: https://letsencrypt.org/docs/rate-limits
# Each domain & SANs will lead to a certificate request.
#
# [[acme.domains]]
#   main = "local1.com"
#   sans = ["test1.local1.com", "test2.local1.com"]
# [[acme.domains]]
#   main = "local2.com"
#   sans = ["test1.local2.com", "test2x.local2.com"]
# [[acme.domains]]
#   main = "local3.com"
# [[acme.domains]]
#   main = "local4.com"
[[acme.domains]]
   main = "local1.com"
   sans = ["test1.local1.com", "test2.local1.com"]
[[acme.domains]]
   main = "local3.com"
[[acme.domains]]
   main = "local4.com"
```

# Configuration backends

## File backend

Like any other reverse proxy, Træfɪk can be configured with a file. You have two choices:

- simply add your configuration at the end of the global configuration file `traefik.toml`:

```toml
# traefik.toml
logLevel = "DEBUG"
defaultEntryPoints = ["http", "https"]
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
      entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"

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
    [backends.backend1.maxconn]
      amount = 10
      extractorfunc = "request.host"
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
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  priority = 10
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
    rule = "Path:/test"
```

- or put your rules in a separate file, for example `rules.toml`:

```toml
# traefik.toml
logLevel = "DEBUG"
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
      entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"

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
    [backends.backend1.maxconn]
      amount = 10
      extractorfunc = "request.host"
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
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  priority = 10
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
    rule = "Path:/test"
```

If you want Træfɪk to watch file changes automatically, just add:

```toml
[file]
watch = true
```

## API backend

Træfik can be configured using a RESTful api.
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
#
# To enable basic auth on the webui
# with 2 user/pass: test:test and test2:test2
# Passwords can be encoded in MD5, SHA1 and BCrypt: you can use htpasswd to generate those ones
#   [web.auth.basic] 
#     users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
# To enable digest auth on the webui
# with 2 user/realm/pass: test:traefik:test and test2:traefik:test2
# You can use htdigest to generate those ones
#   [web.auth.basic] 
#     users = ["test:traefik:a2688e031edb4be6a3797f3882655c05 ", "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"]

```

- `/`: provides a simple HTML frontend of Træfik

![Web UI Providers](img/web.frontend.png)
![Web UI Health](img/traefik-health.png)

- `/ping`: `GET` simple endpoint to check for Træfik process liveness.

```sh
$ curl -sv "http://localhost:8080/ping"
*   Trying ::1...
* Connected to localhost (::1) port 8080 (#0)
> GET /ping HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Thu, 25 Aug 2016 01:35:36 GMT
< Content-Length: 2
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
OK
```

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
            "rule": "Path:/test"
          }
        },
        "backend": "backend1"
      },
      "frontend1": {
        "routes": {
          "test_1": {
            "rule": "Host:test.localhost"
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


## Docker backend

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

# Expose containers by default in traefik
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
- `traefik.backend.maxconn.amount=10`: set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.
- `traefik.backend.maxconn.extractorfunc=client.ip`: set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect.
- `traefik.backend.loadbalancer.method=drr`: override the default `wrr` load balancer algorithm
- `traefik.backend.circuitbreaker.expression=NetworkErrorRatio() > 0.5`: create a [circuit breaker](/basics/#backends) to be used against the backend
- `traefik.port=80`: register this port. Useful when the container exposes multiples ports.
- `traefik.protocol=https`: override the default `http` protocol
- `traefik.weight=10`: assign this weight to the container
- `traefik.enable=false`: disable this container in Træfɪk
- `traefik.frontend.rule=Host:test.traefik.io`: override the default frontend rule (Default: `Host:{containerName}.{domain}`).
- `traefik.frontend.passHostHeader=true`: forward client `Host` header to the backend.
- `traefik.frontend.priority=10`: override default frontend priority
- `traefik.frontend.entryPoints=http,https`: assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`.
- `traefik.docker.network`: Set the docker network to use for connections to this container

NB: when running inside a container, Træfɪk will need network access through `docker network connect <network> <traefik-container>`

## Marathon backend

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

# Enable watch Marathon changes
#
# Optional
#
watch = true

# Default domain used.
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
# Default: false
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
# InsecureSkipVerify = true

# DCOSToken for DCOS environment, This will override the Authorization header
#
# Optional
#
# dcosToken = "xxxxxx"
```

Labels can be used on containers to override default behaviour:

- `traefik.backend=foo`: assign the application to `foo` backend
- `traefik.backend.maxconn.amount=10`: set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.
- `traefik.backend.maxconn.extractorfunc=client.ip`: set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect.
- `traefik.backend.loadbalancer.method=drr`: override the default `wrr` load balancer algorithm
- `traefik.backend.circuitbreaker.expression=NetworkErrorRatio() > 0.5`: create a [circuit breaker](/basics/#backends) to be used against the backend
- `traefik.portIndex=1`: register port by index in the application's ports array. Useful when the application exposes multiple ports.
- `traefik.port=80`: register the explicit application port value. Cannot be used alongside `traefik.portIndex`.
- `traefik.protocol=https`: override the default `http` protocol
- `traefik.weight=10`: assign this weight to the application
- `traefik.enable=false`: disable this application in Træfɪk
- `traefik.frontend.rule=Host:test.traefik.io`: override the default frontend rule (Default: `Host:{containerName}.{domain}`).
- `traefik.frontend.passHostHeader=true`: forward client `Host` header to the backend.
- `traefik.frontend.priority=10`: override default frontend priority
- `traefik.frontend.entryPoints=http,https`: assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`.


## Kubernetes Ingress backend


Træfɪk can be configured to use Kubernetes Ingress as a backend configuration:

```toml
################################################################
# Kubernetes Ingress configuration backend
################################################################
# Enable Kubernetes Ingress configuration backend
#
# Optional
#
[kubernetes]

# Kubernetes server endpoint
#
# When deployed as a replication controller in Kubernetes,
# Traefik will use env variable KUBERNETES_SERVICE_HOST
# and KUBERNETES_SERVICE_PORT_HTTPS as endpoint
# Secure token will be found in /var/run/secrets/kubernetes.io/serviceaccount/token
# and SSL CA cert in /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
#
# Optional
#
# endpoint = "http://localhost:8080"
# namespaces = ["default","production"]
#
# See: http://kubernetes.io/docs/user-guide/labels/#list-and-watch-filtering
# labelselector = "A and not B"
#
```

Annotations can be used on containers to override default behaviour for the whole Ingress resource:

- `traefik.frontend.rule.type: PathPrefixStrip`: override the default frontend rule type (Default: `PathPrefix`).

You can find here an example [ingress](https://raw.githubusercontent.com/containous/traefik/master/examples/k8s.ingress.yaml) and [replication controller](https://raw.githubusercontent.com/containous/traefik/master/examples/k8s.rc.yaml).

## Consul backend

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

# Enable consul TLS connection
#
# Optional
#
# [consul.tls]
# ca = "/etc/ssl/ca.crt"
# cert = "/etc/ssl/consul.crt"
# key = "/etc/ssl/consul.key"
# insecureskipverify = true

# Constraint on ConsulKV tags
#
# Optional
#
# constraints = ["tag==api", "tag==he*ld"]
# Matching with containers having a key "/traefik/backends/backend1/servers/server1/tags" equal to "api,helloworld"
```

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on traefik KV structure.

## Consul catalog backend

Træfɪk can be configured to use service discovery catalog of Consul as a backend configuration:

```toml
################################################################
# Consul Catalog configuration backend
################################################################

# Enable Consul Catalog configuration backend
#
# Optional
#
[consulCatalog]

# Consul server endpoint
#
# Required
#
endpoint = "127.0.0.1:8500"

# Default domain used.
#
# Optional
#
domain = "consul.localhost"

# Prefix for Consul catalog tags
#
# Optional
#
prefix = "traefik"
```

This backend will create routes matching on hostname based on the service name
used in consul.

Additional settings can be defined using Consul Catalog tags:

- `traefik.enable=false`: disable this container in Træfɪk
- `traefik.protocol=https`: override the default `http` protocol
- `traefik.backend.weight=10`: assign this weight to the container
- `traefik.backend.circuitbreaker=NetworkErrorRatio() > 0.5`
- `traefik.backend.loadbalancer=drr`: override the default load balancing mode
- `traefik.backend.maxconn.amount=10`: set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.
- `traefik.backend.maxconn.extractorfunc=client.ip`: set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect.
- `traefik.frontend.rule=Host:test.traefik.io`: override the default frontend rule (Default: `Host:{containerName}.{domain}`).
- `traefik.frontend.passHostHeader=true`: forward client `Host` header to the backend.
- `traefik.frontend.priority=10`: override default frontend priority
- `traefik.frontend.entryPoints=http,https`: assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`.

## Etcd backend

Træfɪk can be configured to use Etcd as a backend configuration:

```toml
################################################################
# Etcd configuration backend
################################################################

# Enable Etcd configuration backend
#
# Optional
#
[etcd]

# Etcd server endpoint
#
# Required
#
endpoint = "127.0.0.1:2379"

# Enable watch Etcd changes
#
# Optional
#
watch = true

# Prefix used for KV store.
#
# Optional
#
prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "etcd.tmpl"

# Enable etcd TLS connection
#
# Optional
#
# [etcd.tls]
# ca = "/etc/ssl/ca.crt"
# cert = "/etc/ssl/etcd.crt"
# key = "/etc/ssl/etcd.key"
# insecureskipverify = true

# Constraint on Etcd tags
#
# Optional
#
# constraints = ["tag==api", "tag==he*ld"]
# Matching with containers having a key "/traefik/backends/backend1/servers/server1/tags" equal to "api,helloworld"
```

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on traefik KV structure.


## Zookeeper backend

Træfɪk can be configured to use Zookeeper as a backend configuration:

```toml
################################################################
# Zookeeper configuration backend
################################################################

# Enable Zookeeperconfiguration backend
#
# Optional
#
[zookeeper]

# Zookeeper server endpoint
#
# Required
#
endpoint = "127.0.0.1:2181"

# Enable watch Zookeeper changes
#
# Optional
#
watch = true

# Prefix used for KV store.
#
# Optional
#
prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "zookeeper.tmpl"

# Constraint on Zookeeper tags
#
# Optional
#
# constraints = ["tag==api", "tag==he*ld"]
# Matching with containers having a key "/traefik/backends/backend1/servers/server1/tags" equal to "api,helloworld"
```

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on traefik KV structure.

## BoltDB backend

Træfɪk can be configured to use BoltDB as a backend configuration:

```toml
################################################################
# BoltDB configuration backend
################################################################

# Enable BoltDB configuration backend
#
# Optional
#
[boltdb]

# BoltDB file
#
# Required
#
endpoint = "/my.db"

# Enable watch BoltDB changes
#
# Optional
#
watch = true

# Prefix used for KV store.
#
# Optional
#
prefix = "/traefik"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "boltdb.tmpl"

# Constraint on BoltDB tags
#
# Optional
#
# constraints = ["tag==api", "tag==he*ld"]
# Matching with containers having a key "/traefik/backends/backend1/servers/server1/tags" equal to "api,helloworld"
```

Please refer to the [Key Value storage structure](/user-guide/kv-config/#key-value-storage-structure) section to get documentation on traefik KV structure.
