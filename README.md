![Træfɪk](docs/img/traefik.logo.png "Træfɪk")

Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends (Docker, Mesos/Marathon, Consul, Etcd, Rest API, file...) to manage its configuration automatically and dynamically (hot-reload).

![Backends](docs/img/backends.png "Backends")

# Features

* No dependency hell, single binary made with go
* Simple json Rest API
* Simple TOML file configuration
* Multiple backends supported: Docker, Mesos/Marathon, Consul, Etcd, and more to come
* Watchers for backends, can listen change in backends to apply a new configuration automatically
* Hot-reloading of configuration. No need to restart the process
* Graceful shutdown http connections during hot-reloads
* Rest Metrics
* Tiny docker image included
* SSL backends support
* SSL frontend support

# Plumbing

* [Oxy](https://github.com/mailgun/oxy/): an awsome proxy library made by Mailgun guys
* [Gorilla mux](https://github.com/gorilla/mux): famous request router
* [Negroni](https://github.com/codegangsta/negroni): web middlewares made simple
* [Graceful](https://github.com/tylerb/graceful): graceful shutdown of http.Handler servers

# Quick start

* The simple way: go to the [releases](https://github.com/emilevauge/traefik/releases) page and get a binary.
* Or simply execute:

```
go get github.com/emilevauge/traefik
```
* Just run it!

```
./traefik traefik.toml
```

# Configuration

## Global configuration

```toml
# traefik.toml
port = ":80"
graceTimeOut = 10
logLevel = "DEBUG"
traefikLogsStdout = true
# traefikLogsFile = "log/traefik.log"
# accessLogsFile = "log/access.log"
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"
```

## File backend

Like any other reverse proxy, Træfɪk can be configured with a file. You have two choices:

* simply add your configuration at the end of the global configuration file ```traefik.tml``` :

```toml
# traefik.toml
port = ":80"
graceTimeOut = 10
logLevel = "DEBUG"
traefikLogsStdout = true
# traefikLogsFile = "log/traefik.log"
# accessLogsFile = "log/access.log"
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"

[file]

[backends]
  [backends.backend1]
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1

[routes]
  [routes.route1]
  backend = "backend2"
    [routes.route1.rules.test1]
    category = "Host"
    value = "test.localhost"
  [routes.route2]
  backend = "backend1"
    [routes.route2.rules.test2]
    category = "Path"
    value = "/test"

```

* or put your configuration in a separate file, for example ```rules.tml```:

```toml
# traefik.toml
port = ":80"
graceTimeOut = 10
logLevel = "DEBUG"
traefikLogsStdout = true
# traefikLogsFile = "log/traefik.log"
# accessLogsFile = "log/access.log"
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"

[file]
filename = "rules.toml"
```

```toml
# rules.toml
[backends]
  [backends.backend1]
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1

[routes]
  [routes.route1]
  backend = "backend2"
    [routes.route1.rules.test1]
    category = "Host"
    value = "test.localhost"
  [routes.route2]
  backend = "backend1"
    [routes.route2.rules.test2]
    category = "Path"
    value = "/test"
```

If you want Træfɪk to watch file changes automatically, just add:

```toml
[file]
watch = true
```
