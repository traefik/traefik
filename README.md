# /Træfɪk/

/Træfɪk/ is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends (Docker, Mesos/Marathon, Consul, Etcd, Rest API, file...) to manage its configuration automatically and dynamically (hot-reload).

## Features

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

## Plumbing

* [Oxy](github.com/mailgun/oxy/): an awsome proxy librarymade by Mailgun guys
* [Gorilla mux](github.com/gorilla/mux): famous request router
* [Negroni](github.com/codegangsta/negroni): web middlewares made simple
* [Graceful](github.com/tylerb/graceful): graceful shutdown of http.Handler servers

## Quick start

* The simple way: go to the [releases](https://github.com/emilevauge/traefik/releases) page and get a binary.
* Or simply execute:

```
go get github.com/emilevauge/traefik
```
* Just run it!

```
./traefik traefik.toml
```

## Configuration

Here is a sample configuration TOML file:

```toml
port = ":80"
graceTimeOut = 10
traefikLogsFile = "log/traefik.log"
traefikLogsStdout = true
accessLogsFile = "log/access.log"
logLevel = "DEBUG"
# CertFile = "traefik.crt"
# KeyFile = "traefik.key"

[docker]
endpoint = "unix:///var/run/docker.sock"
watch = true
domain = "localhost"
# filename = "docker.tmpl"

# [marathon]
# endpoint = "http://127.0.0.1:8080"
# networkInterface = "eth0"
# watch = true
# domain = "localhost"
# filename = "marathon.tmpl"

[web]
address = ":8080"

# [file]
# filename = "rules.toml"
# watch = true

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
