
<p align="center">
<img src="docs/img/traefik.logo.png" alt="Traefik" title="Traefik" />
</p>

[![Build Status SemaphoreCI](https://semaphoreci.com/api/v1/containous/traefik/branches/master/shields_badge.svg)](https://semaphoreci.com/containous/traefik)
[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](https://docs.traefik.io)
[![Go Report Card](https://goreportcard.com/badge/containous/traefik)](http://goreportcard.com/report/containous/traefik)
[![](https://images.microbadger.com/badges/image/traefik.svg)](https://microbadger.com/images/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/containous/traefik/blob/master/LICENSE.md)
[![Join the chat at https://slack.traefik.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.traefik.io)
[![Twitter](https://img.shields.io/twitter/follow/traefik.svg?style=social)](https://twitter.com/intent/follow?screen_name=traefik)


Traefik is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy.
Traefik integrates with your existing infrastructure components ([Docker](https://www.docker.com/), [Swarm mode](https://docs.docker.com/engine/swarm/), [Kubernetes](https://kubernetes.io), [Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Rancher](https://rancher.com), [Amazon ECS](https://aws.amazon.com/ecs), ...) and configures itself automatically and dynamically.
Pointing Traefik at your orchestrator should be the _only_ configuration step you need.

---

. **[Overview](#overview)** .
**[Features](#features)** .
**[Supported backends](#supported-backends)** .
**[Quickstart](#quickstart)** .
**[Web UI](#web-ui)** .
**[Test it](#test-it)** .
**[Documentation](#documentation)** .

. **[Support](#support)** .
**[Release cycle](#release-cycle)** .
**[Contributing](#contributing)** .
**[Maintainers](#maintainers)** .
**[Plumbing](#plumbing)** .
**[Credits](#credits)** .

---

## Overview

Imagine that you have deployed a bunch of microservices with the help of an orchestrator (like Swarm or Kubernetes) or a service registry (like etcd or consul).
Now you want users to access these microservices, and you need a reverse proxy.

Traditional reverse-proxies require that you configure _each_ route that will connect paths and subdomains to _each_ microservice. 
In an environment where you add, remove, kill, upgrade, or scale your services _many_ times a day, the task of keeping the routes up to date becomes tedious. 

**This is when Traefik can help you!**

Traefik listens to your service registry/orchestrator API and instantly generates the routes so your microservices are connected to the outside world -- without further intervention from your part. 

**Run Traefik and let it do the work for you!** 
_(But if you'd rather configure some of your routes manually, Traefik supports that too!)_

![Architecture](docs/img/architecture.png)

## Features

- Continuously updates its configuration (No restarts!)
- Supports multiple load balancing algorithms
- Provides HTTPS to your microservices by leveraging [Let's Encrypt](https://letsencrypt.org)  (wildcard certificates support)
- Circuit breakers, retry
- High Availability with cluster mode (beta)
- See the magic through its clean web UI
- Websocket, HTTP/2, GRPC ready
- Provides metrics (Rest, Prometheus, Datadog, Statsd, InfluxDB)
- Keeps access logs (JSON, CLF)
- Fast
- Exposes a Rest API
- Packaged as a single binary file (made with :heart: with go) and available as a [tiny](https://microbadger.com/images/traefik) [official](https://hub.docker.com/r/_/traefik/) docker image


## Supported Backends

- [Docker](https://docs.traefik.io/configuration/backends/docker) / [Swarm mode](https://docs.traefik.io/configuration/backends/docker#docker-swarm-mode)
- [Kubernetes](https://docs.traefik.io/configuration/backends/kubernetes)
- [Mesos](https://docs.traefik.io/configuration/backends/mesos) / [Marathon](https://docs.traefik.io/configuration/backends/marathon)
- [Rancher](https://docs.traefik.io/configuration/backends/rancher) (API, Metadata)
- [Azure Service Fabric](https://docs.traefik.io/configuration/backends/servicefabric)
- [Consul Catalog](https://docs.traefik.io/configuration/backends/consulcatalog)
- [Consul](https://docs.traefik.io/configuration/backends/consul) / [Etcd](https://docs.traefik.io/configuration/backends/etcd) / [Zookeeper](https://docs.traefik.io/configuration/backends/zookeeper) / [BoltDB](https://docs.traefik.io/configuration/backends/boltdb)
- [Eureka](https://docs.traefik.io/configuration/backends/eureka)
- [Amazon ECS](https://docs.traefik.io/configuration/backends/ecs)
- [Amazon DynamoDB](https://docs.traefik.io/configuration/backends/dynamodb)
- [File](https://docs.traefik.io/configuration/backends/file)
- [Rest](https://docs.traefik.io/configuration/backends/rest)

## Quickstart

To get your hands on Traefik, you can use the [5-Minute Quickstart](http://docs.traefik.io/#the-traefik-quickstart-using-docker) in our documentation (you will need Docker).

Alternatively, if you don't want to install anything on your computer, you can try Traefik online in this great [Katacoda tutorial](https://www.katacoda.com/courses/traefik/deploy-load-balancer) that shows how to load balance requests between multiple Docker containers. 

If you are looking for a more comprehensive and real use-case example, you can also check [Play-With-Docker](http://training.play-with-docker.com/traefik-load-balancing/) to see how to load balance between multiple nodes.

## Web UI

You can access the simple HTML frontend of Traefik.

![Web UI Providers](docs/img/web.frontend.png)
![Web UI Health](docs/img/traefik-health.png)

## Documentation

You can find the complete documentation at [https://docs.traefik.io](https://docs.traefik.io).
A collection of contributions around Traefik can be found at [https://awesome.traefik.io](https://awesome.traefik.io).

## Support

To get community support, you can:
- join the Traefik community Slack channel: [![Join the chat at https://slack.traefik.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.traefik.io)
- use [Stack Overflow](https://stackoverflow.com/questions/tagged/traefik) (using the `traefik` tag)

If you need commercial support, please contact [Containo.us](https://containo.us) by mail: <mailto:support@containo.us>.

## Download

- Grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/v1.7/traefik.sample.toml):

```shell
./traefik --configFile=traefik.toml
```

- Or use the official tiny Docker image and run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/v1.7/traefik.sample.toml):

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik
```

- Or get the sources:

```shell
git clone https://github.com/containous/traefik
```

## Introductory Videos

Here is a talk given by [Emile Vauge](https://github.com/emilevauge) at GopherCon 2017.
You will learn Traefik basics in less than 10 minutes.

[![Traefik GopherCon 2017](https://img.youtube.com/vi/RgudiksfL-k/0.jpg)](https://www.youtube.com/watch?v=RgudiksfL-k)

Here is a talk given by [Ed Robinson](https://github.com/errm) at [ContainerCamp UK](https://container.camp) conference.
You will learn fundamental Traefik features and see some demos with Kubernetes.

[![Traefik ContainerCamp UK](https://img.youtube.com/vi/aFtpIShV60I/0.jpg)](https://www.youtube.com/watch?v=aFtpIShV60I)

## Maintainers

[Information about process and maintainers](MAINTAINER.md)

## Contributing

If you'd like to contribute to the project, refer to the [contributing documentation](CONTRIBUTING.md).

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md).
By participating in this project, you agree to abide by its terms.

## Release Cycle

- We release a new version (e.g. 1.1.0, 1.2.0, 1.3.0) every other month.
- Release Candidates are available before the release (e.g. 1.1.0-rc1, 1.1.0-rc2, 1.1.0-rc3, 1.1.0-rc4, before 1.1.0)
- Bug-fixes (e.g. 1.1.1, 1.1.2, 1.2.1, 1.2.3) are released as needed (no additional features are delivered in those versions, bug-fixes only)

Each version is supported until the next one is released (e.g. 1.1.x will be supported until 1.2.0 is out)

We use [Semantic Versioning](http://semver.org/)

## Mailing lists

- General announcements, new releases: mail at news+subscribe@traefik.io or on [the online viewer](https://groups.google.com/a/traefik.io/forum/#!forum/news)
- Security announcements: mail at security+subscribe@traefik.io or on [the online viewer](https://groups.google.com/a/traefik.io/forum/#!forum/security).

## Credits

Kudos to [Peka](http://peka.byethost11.com/photoblog/) for his awesome work on the logo ![logo](docs/img/traefik.icon.png).

Traefik's logo is licensed under the Creative Commons 3.0 Attributions license.

Traefik's logo was inspired by the gopher stickers made by Takuya Ueda (https://twitter.com/tenntenn).
The original Go gopher was designed by Renee French (http://reneefrench.blogspot.com/).