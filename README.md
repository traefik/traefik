![Træfɪk](http://traefik.github.io/traefik.logo.svg  "Træfɪk")
___

[![Build Status](https://travis-ci.org/containous/traefik.svg?branch=master)](https://travis-ci.org/containous/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/containous/traefik/blob/master/LICENSE.md)
[![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com)
[![Twitter](https://img.shields.io/twitter/follow/traefikproxy.svg?style=social)](https://twitter.com/intent/follow?screen_name=traefikproxy)



Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends ([Docker :whale:](https://www.docker.com/), [Mesos/Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Zookeeper](https://zookeeper.apache.org), [BoltDB](https://github.com/boltdb/bolt), Rest API, file...) to manage its configuration automatically and dynamically.


## Features

- No dependency hell, single binary made with go
- Simple json Rest API
- Simple TOML file configuration
- Multiple backends supported: Docker, Mesos/Marathon, Consul, Etcd, and more to come
- Watchers for backends, can listen change in backends to apply a new configuration automatically
- Hot-reloading of configuration. No need to restart the process
- Graceful shutdown http connections during hot-reloads
- Circuit breakers on backends
- Round Robin, rebalancer load-balancers
- Rest Metrics
- Tiny docker image included [![Image Layers](https://badge.imagelayers.io/containous/traefik:latest.svg)](https://imagelayers.io/?images=containous/traefik:latest)
- SSL backends support
- SSL frontend support
- Clean AngularJS Web UI
- Websocket support
- HTTP/2 support

## Demo

Here is a demo of Træfɪk using Docker backend, showing a load-balancing between two servers, hot reloading of configuration, and graceful shutdown.

[![asciicast](https://asciinema.org/a/4tcyde7riou5vxulo6my3mtko.png)](https://asciinema.org/a/4tcyde7riou5vxulo6my3mtko)

## Web UI

You can access to a simple HTML frontend of Træfik.

![Web UI Providers](docs/img/web.frontend.png)
![Web UI Health](docs/img/traefik-health.png)

## Plumbing

- [Oxy](https://github.com/vulcand/oxy): an awsome proxy library made by Mailgun guys
- [Gorilla mux](https://github.com/gorilla/mux): famous request router
- [Negroni](https://github.com/codegangsta/negroni): web middlewares made simple
- [Manners](https://github.com/mailgun/manners): graceful shutdown of http.Handler servers

## Quick start

- The simple way: grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/master/traefik.sample.toml):

```shell
./traefik -c traefik.toml
```

- Use the tiny Docker image:

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml containous/traefik
```

- From sources:

```shell
git clone https://github.com/containous/traefik
```

## Documentation

You can find the complete documentation [here](docs/index.md).

## Contributing

Please refer to [this section](.github/CONTRIBUTING.md).

## Træfɪk here and there

These projects use Træfɪk internally. If your company uses Træfɪk, we would be glad to get your feedback :) Contact us on [![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com)

- Project [Mantl](https://mantl.io/) from Cisco

![Web UI Providers](docs/img/mantl-logo.png)
> Mantl is a modern platform for rapidly deploying globally distributed services. A container orchestrator, docker, a network stack, something to pool your logs, something to monitor health, a sprinkle of service discovery and some automation.

- Project [Apollo](http://capgemini.github.io/devops/apollo/) from Cap Gemini

![Web UI Providers](docs/img/apollo-logo.png)
> Apollo is an open source project to aid with building and deploying IAAS and PAAS services. It is particularly geared towards managing containerized applications across multiple hosts, and big data type workloads. Apollo leverages other open source components to provide basic mechanisms for deployment, maintenance, and scaling of infrastructure and applications.

## Partners

[![Zenika](docs/img/zenika.logo.png)](https://zenika.com)

Zenika is one of the leading providers of professional Open Source services and agile methodologies in
Europe. We provide consulting, development, training and support for the world’s leading Open Source
software products.



[![Asteris](docs/img/asteris.logo.png)](https://aster.is)

Founded in 2014, Asteris creates next-generation infrastructure software for the modern datacenter. Asteris writes software that makes it easy for companies to implement continuous delivery and realtime data pipelines. We support the HashiCorp stack, along with Kubernetes, Apache Mesos, Spark and Kafka. We're core committers on mantl.io, consul-cli and mesos-consul.
.
