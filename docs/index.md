<p align="center">
<img src="img/traefik.logo.png" alt="Træfik" title="Træfik" />
</p>

[![Build Status SemaphoreCI](https://semaphoreci.com/api/v1/containous/traefik/branches/master/shields_badge.svg)](https://semaphoreci.com/containous/traefik)
[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](https://docs.traefik.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/containous/traefik)](https://goreportcard.com/report/github.com/containous/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/containous/traefik/blob/master/LICENSE.md)
[![Join the chat at https://slack.traefik.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.traefik.io)
[![Twitter](https://img.shields.io/twitter/follow/traefikproxy.svg?style=social)](https://twitter.com/intent/follow?screen_name=traefikproxy)


Træfik is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy.
Træfik integrates with your existing infrastructure components ([Docker](https://www.docker.com/), [Swarm mode](https://docs.docker.com/engine/swarm/), [Kubernetes](https://kubernetes.io), [Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Rancher](https://rancher.com), [Amazon ECS](https://aws.amazon.com/ecs), ...) and configures itself automatically and dynamically.
Pointing Træfik at your orchestrator should be the _only_ configuration step you need.

## Overview

Imagine that you have deployed a bunch of microservices with the help of an orchestrator (like Swarm or Kubernetes) or a service registry (like etcd or consul).
Now you want users to access these microservices, and you need a reverse proxy.

Traditional reverse-proxies require that you configure _each_ route that will connect paths and subdomains to _each_ microservice.
In an environment where you add, remove, kill, upgrade, or scale your services _many_ times a day, the task of keeping the routes up to date becomes tedious.

**This is when Træfik can help you!**

Træfik listens to your service registry/orchestrator API and instantly generates the routes so your microservices are connected to the outside world -- without further intervention from your part.

**Run Træfik and let it do the work for you!**
_(But if you'd rather configure some of your routes manually, Træfik supports that too!)_

![Architecture](img/architecture.png)

## Features

- Continuously updates its configuration (No restarts!)
- Supports multiple load balancing algorithms
- Provides HTTPS to your microservices by leveraging [Let's Encrypt](https://letsencrypt.org) (wildcard certificates support)
- Circuit breakers, retry
- High Availability with cluster mode (beta)
- See the magic through its clean web UI
- Websocket, HTTP/2, GRPC ready
- Provides metrics (Rest, Prometheus, Datadog, Statsd, InfluxDB)
- Keeps access logs (JSON, CLF)
- Fast
- Exposes a Rest API
- Packaged as a single binary file (made with :heart: with go) and available as a [tiny](https://microbadger.com/images/traefik) [official](https://hub.docker.com/r/_/traefik/) docker image


## Supported Providers

- [Docker](/configuration/backends/docker/) / [Swarm mode](/configuration/backends/docker/#docker-swarm-mode)
- [Kubernetes](/configuration/backends/kubernetes/)
- [Mesos](/configuration/backends/mesos/) / [Marathon](/configuration/backends/marathon/)
- [Rancher](/configuration/backends/rancher/) (API, Metadata)
- [Azure Service Fabric](/configuration/backends/servicefabric/)
- [Consul Catalog](/configuration/backends/consulcatalog/)
- [Consul](/configuration/backends/consul/) / [Etcd](/configuration/backends/etcd/) / [Zookeeper](/configuration/backends/zookeeper/) / [BoltDB](/configuration/backends/boltdb/)
- [Eureka](/configuration/backends/eureka/)
- [Amazon ECS](/configuration/backends/ecs/)
- [Amazon DynamoDB](/configuration/backends/dynamodb/)
- [File](/configuration/backends/file/)
- [Rest](/configuration/backends/rest/)

## The Træfik Quickstart (Using Docker)

In this quickstart, we'll use [Docker compose](https://docs.docker.com/compose) to create our demo infrastructure.

To save some time, you can clone [Træfik's repository](https://github.com/containous/traefik) and use the quickstart files located in the [examples/quickstart](https://github.com/containous/traefik/tree/master/examples/quickstart/) directory.

### 1 — Launch Træfik — Tell It to Listen to Docker

Create a `docker-compose.yml` file where you will define a `reverse-proxy` service that uses the official Træfik image:

```yaml
version: '3'

services:
  reverse-proxy:
    image: traefik # The official Traefik docker image
    command: --api --docker # Enables the web UI and tells Træfik to listen to docker
    ports:
      - "80:80"     # The HTTP port
      - "8080:8080" # The Web UI (enabled by --api)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock # So that Traefik can listen to the Docker events
```

**That's it. Now you can launch Træfik!**

Start your `reverse-proxy` with the following command:

```shell
docker-compose up -d reverse-proxy
```

You can open a browser and go to [http://localhost:8080](http://localhost:8080) to see Træfik's dashboard (we'll go back there once we have launched a service in step 2).

### 2 — Launch a Service — Træfik Detects It and Creates a Route for You

Now that we have a Træfik instance up and running, we will deploy new services.

Edit your `docker-compose.yml` file and add the following at the end of your file.

```yaml
# ...
  whoami:
    image: emilevauge/whoami # A container that exposes an API to show its IP address
    labels:
      - "traefik.frontend.rule=Host:whoami.docker.localhost"
```

The above defines `whoami`: a simple web service that outputs information about the machine it is deployed on (its IP address, host, and so on).

Start the `whoami` service with the following command:

```shell
docker-compose up -d whoami
```

Go back to your browser ([http://localhost:8080](http://localhost:8080)) and see that Træfik has automatically detected the new container and updated its own configuration.

When Traefik detects new services, it creates the corresponding routes so you can call them ... _let's see!_  (Here, we're using curl)

```shell
curl -H Host:whoami.docker.localhost http://127.0.0.1
```

_Shows the following output:_
```yaml
Hostname: 8656c8ddca6c
IP: 172.27.0.3
#...
```

### 3 — Launch More Instances — Traefik Load Balances Them

Run more instances of your `whoami` service with the following command:

```shell
docker-compose up -d --scale whoami=2
```

Go back to your browser ([http://localhost:8080](http://localhost:8080)) and see that Træfik has automatically detected the new instance of the container.

Finally, see that Træfik load-balances between the two instances of your services by running twice the following command:

```shell
curl -H Host:whoami.docker.localhost http://127.0.0.1
```

The output will show alternatively one of the followings:

```yaml
Hostname: 8656c8ddca6c
IP: 172.27.0.3
#...
```

```yaml
Hostname: 8458f154e1f1
IP: 172.27.0.4
# ...
```

### 4 — Enjoy Træfik's Magic

Now that you have a basic understanding of how Træfik can automatically create the routes to your services and load balance them, it might be time to dive into [the documentation](/) and let Træfik work for you!
Whatever your infrastructure is, there is probably [an available Træfik provider](/#supported-providers) that will do the job.

Our recommendation would be to see for yourself how simple it is to enable HTTPS with [Træfik's let's encrypt integration](/user-guide/examples/#lets-encrypt-support) using the dedicated [user guide](/user-guide/docker-and-lets-encrypt/).

## Resources

Here is a talk given by [Emile Vauge](https://github.com/emilevauge) at [GopherCon 2017](https://gophercon.com).
You will learn Træfik basics in less than 10 minutes.

[![Traefik GopherCon 2017](https://img.youtube.com/vi/RgudiksfL-k/0.jpg)](https://www.youtube.com/watch?v=RgudiksfL-k)

Here is a talk given by [Ed Robinson](https://github.com/errm) at [ContainerCamp UK](https://container.camp) conference.
You will learn fundamental Træfik features and see some demos with Kubernetes.

[![Traefik ContainerCamp UK](https://img.youtube.com/vi/aFtpIShV60I/0.jpg)](https://www.youtube.com/watch?v=aFtpIShV60I)

## Downloads

### The Official Binary File

You can grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/master/traefik.sample.toml):

```shell
./traefik -c traefik.toml
```

### The Official Docker Image

Using the tiny Docker image:

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik
```
