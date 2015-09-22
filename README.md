![Træfɪk](http://traefik.github.io/traefik.logo.svg  "Træfɪk")
___

[![Circle CI](https://img.shields.io/circleci/project/EmileVauge/traefik.svg)](https://circleci.com/gh/EmileVauge/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/EmileVauge/traefik/blob/master/LICENSE.md)

Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends ([Docker :whale:](https://www.docker.com/), [Mesos/Marathon](https://mesosphere.github.io/marathon/), [Consul](https://consul.io/), [Etcd](https://coreos.com/etcd/), Rest API, file...) to manage its configuration automatically and dynamically.


# Features

* No dependency hell, single binary made with go
* Simple json Rest API
* Simple TOML file configuration
* Multiple backends supported: Docker, Mesos/Marathon, Consul, Etcd, and more to come
* Watchers for backends, can listen change in backends to apply a new configuration automatically
* Hot-reloading of configuration. No need to restart the process
* Graceful shutdown http connections during hot-reloads
* Circuit breakers on backends
* Round Robin, rebalancer load-balancers
* Rest Metrics
* Tiny docker image included
* SSL backends support
* SSL frontend support
* WebUI

# Plumbing

* [Oxy](https://github.com/mailgun/oxy/): an awsome proxy library made by Mailgun guys
* [Gorilla mux](https://github.com/gorilla/mux): famous request router
* [Negroni](https://github.com/codegangsta/negroni): web middlewares made simple
* [Graceful](https://github.com/tylerb/graceful): graceful shutdown of http.Handler servers

# Quick start

* The simple way: grab the latest binary from the [releases](https://github.com/emilevauge/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/EmileVauge/traefik/master/traefik.sample.toml):

```
./traefik traefik.toml
```

* Use the tiny Docker image:

```
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/traefik.toml emilevauge/traefik
```

* From sources:

```
git clone https://github.com/EmileVauge/traefik
```

# Documentation

You can find the complete documentation [here](docs/index.md).
