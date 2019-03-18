<p align="center">
<img src="img/traefik.logo.png" alt="Traefik" title="Traefik" />
</p>

[![Build Status SemaphoreCI](https://semaphoreci.com/api/v1/containous/traefik/branches/master/shields_badge.svg)](https://semaphoreci.com/containous/traefik)
[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](/)
[![Go Report Card](https://goreportcard.com/badge/github.com/containous/traefik)](https://goreportcard.com/report/github.com/containous/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/containous/traefik/blob/master/LICENSE.md)
[![Join the chat at https://slack.traefik.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.traefik.io)
[![Twitter](https://img.shields.io/twitter/follow/traefik.svg?style=social)](https://twitter.com/intent/follow?screen_name=traefik)


Traefik is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy.
Traefik integrates with your existing infrastructure components ([Docker](https://www.docker.com/), [Swarm mode](https://docs.docker.com/engine/swarm/), [Kubernetes](https://kubernetes.io), [Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Rancher](https://rancher.com), [Amazon ECS](https://aws.amazon.com/ecs), ...) and configures itself automatically and dynamically.
Pointing Traefik at your orchestrator should be the _only_ configuration step you need.

## Overview

Imagine that you have deployed a bunch of microservices with the help of an orchestrator (like Swarm or Kubernetes) or a service registry (like etcd or consul).
Now you want users to access these microservices, and you need a reverse proxy.

Traditional reverse-proxies require that you configure _each_ route that will connect paths and subdomains to _each_ microservice.
In an environment where you add, remove, kill, upgrade, or scale your services _many_ times a day, the task of keeping the routes up to date becomes tedious.

**This is when Traefik can help you!**

Traefik listens to your service registry/orchestrator API and instantly generates the routes so your microservices are connected to the outside world -- without further intervention from your part.

**Run Traefik and let it do the work for you!**
_(But if you'd rather configure some of your routes manually, Traefik supports that too!)_

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
- Packaged as a single binary file (made with ❤️ with go) and available as a [tiny](https://microbadger.com/images/traefik) [official](https://hub.docker.com/r/_/traefik/) docker image


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

## The Traefik Quickstart (Using Docker)

In this quickstart, we'll use [Docker compose](https://docs.docker.com/compose) to create our demo infrastructure.

To save some time, you can clone [Traefik's repository](https://github.com/containous/traefik) and use the quickstart files located in the [examples/quickstart](https://github.com/containous/traefik/tree/v1.7/examples/quickstart/) directory.

### 1 — Launch Traefik — Tell It to Listen to Docker

Create a `docker-compose.yml` file where you will define a `reverse-proxy` service that uses the official Traefik image:

```yaml
version: '3'

services:
  reverse-proxy:
    image: traefik # The official Traefik docker image
    command: --api --docker # Enables the web UI and tells Traefik to listen to docker
    ports:
      - "80:80"     # The HTTP port
      - "8080:8080" # The Web UI (enabled by --api)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock # So that Traefik can listen to the Docker events
```

!!! warning
    Enabling the Web UI with the `--api` flag might expose configuration elements. You can read more about this on the [API/Dashboard's Security section](/configuration/api#security).


**That's it. Now you can launch Traefik!**

Start your `reverse-proxy` with the following command:

```shell
docker-compose up -d reverse-proxy
```

You can open a browser and go to [http://localhost:8080](http://localhost:8080) to see Traefik's dashboard (we'll go back there once we have launched a service in step 2).

### 2 — Launch a Service — Traefik Detects It and Creates a Route for You

Now that we have a Traefik instance up and running, we will deploy new services.

Edit your `docker-compose.yml` file and add the following at the end of your file.

```yaml
# ...
  whoami:
    image: containous/whoami # A container that exposes an API to show its IP address
    labels:
      - "traefik.frontend.rule=Host:whoami.docker.localhost"
```

The above defines `whoami`: a simple web service that outputs information about the machine it is deployed on (its IP address, host, and so on).

Start the `whoami` service with the following command:

```shell
docker-compose up -d whoami
```

Go back to your browser ([http://localhost:8080](http://localhost:8080)) and see that Traefik has automatically detected the new container and updated its own configuration.

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
docker-compose scale whoami=2
```

Go back to your browser ([http://localhost:8080](http://localhost:8080)) and see that Traefik has automatically detected the new instance of the container.

Finally, see that Traefik load-balances between the two instances of your services by running twice the following command:

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

### 4 — Enjoy Traefik's Magic

Now that you have a basic understanding of how Traefik can automatically create the routes to your services and load balance them, it might be time to dive into [the documentation](/) and let Traefik work for you!
Whatever your infrastructure is, there is probably [an available Traefik provider](/#supported-providers) that will do the job.

Our recommendation would be to see for yourself how simple it is to enable HTTPS with [Traefik's let's encrypt integration](/user-guide/examples/#lets-encrypt-support) using the dedicated [user guide](/user-guide/docker-and-lets-encrypt/).

## Resources

Here is a talk given by [Emile Vauge](https://github.com/emilevauge) at GopherCon 2017.
You will learn Traefik basics in less than 10 minutes.

[![Traefik GopherCon 2017](https://img.youtube.com/vi/RgudiksfL-k/0.jpg)](https://www.youtube.com/watch?v=RgudiksfL-k)

Here is a talk given by [Ed Robinson](https://github.com/errm) at [ContainerCamp UK](https://container.camp) conference.
You will learn fundamental Traefik features and see some demos with Kubernetes.

[![Traefik ContainerCamp UK](https://img.youtube.com/vi/aFtpIShV60I/0.jpg)](https://www.youtube.com/watch?v=aFtpIShV60I)

## Downloads

### The Official Binary File

You can grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/v1.7/traefik.sample.toml):

```shell
./traefik -c traefik.toml
```

### The Official Docker Image

Using the tiny Docker image:

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik
```
 
## Security

### Security Advisories

We strongly advise you to join our mailing list to be aware of the latest announcements from our security team. You can subscribe sending a mail to security+subscribe@traefik.io or on [the online viewer](https://groups.google.com/a/traefik.io/forum/#!forum/security).

### CVE

Reported vulnerabilities can be found on 
[cve.mitre.org](https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=traefik).

### Report a Vulnerability

We want to keep Traefik safe for everyone.
If you've discovered a security vulnerability in Traefik, we appreciate your help in disclosing it to us in a responsible manner, using [this form](https://security.traefik.io).
