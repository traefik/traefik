<p align="center">
<img src="img/traefik.logo.png" alt="Træfik" title="Træfik" />
</p>

[![Build Status](https://travis-ci.org/containous/traefik.svg?branch=master)](https://travis-ci.org/containous/traefik)
[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](https://docs.traefik.io)
[![Go Report Card](https://goreportcard.com/badge/kubernetes/helm)](http://goreportcard.com/report/containous/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/containous/traefik/blob/master/LICENSE.md)
[![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com)
[![Twitter](https://img.shields.io/twitter/follow/traefikproxy.svg?style=social)](https://twitter.com/intent/follow?screen_name=traefikproxy)


Træfik (pronounced like [traffic](https://speak-ipa.bearbin.net/speak.cgi?speak=%CB%88tr%C3%A6f%C9%AAk)) is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends ([Docker](https://www.docker.com/), [Swarm](https://docs.docker.com/swarm), [Mesos/Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Zookeeper](https://zookeeper.apache.org), [BoltDB](https://github.com/boltdb/bolt), [Amazon ECS](https://aws.amazon.com/ecs/), [Amazon DynamoDB](https://aws.amazon.com/dynamodb/), Rest API, file...) to manage its configuration automatically and dynamically.

## Overview

Imagine that you have deployed a bunch of microservices on your infrastructure. You probably used a service registry (like etcd or consul) and/or an orchestrator (swarm, Mesos/Marathon) to manage all these services.
If you want your users to access some of your microservices from the Internet, you will have to use a reverse proxy and configure it using virtual hosts or prefix paths:

- domain `api.domain.com` will point the microservice `api` in your private network
- path `domain.com/web` will point the microservice `web` in your private network
- domain `backoffice.domain.com` will point the microservices `backoffice` in your private network, load-balancing between your multiple instances

But a microservices architecture is dynamic... Services are added, removed, killed or upgraded often, eventually several times a day.

Traditional reverse-proxies are not natively dynamic. You can't change their configuration and hot-reload easily.

Here enters Træfik.

![Architecture](img/architecture.png)

Træfik can listen to your service registry/orchestrator API, and knows each time a microservice is added, removed, killed or upgraded, and can generate its configuration automatically.
Routes to your services will be created instantly.

Run it and forget it!


## Quickstart

You can have a quick look at Træfik in this [Katacoda tutorial](https://www.katacoda.com/courses/traefik/deploy-load-balancer) that shows how to load balance requests between multiple Docker containers.

Here is a talk given by [Ed Robinson](https://github.com/errm) at the [ContainerCamp UK](https://container.camp) conference.
You will learn fundamental Træfik features and see some demos with Kubernetes.

[![Traefik ContainerCamp UK](http://img.youtube.com/vi/aFtpIShV60I/0.jpg)](https://www.youtube.com/watch?v=aFtpIShV60I)

Here is a talk (in French) given by [Emile Vauge](https://github.com/emilevauge) at the [Devoxx France 2016](http://www.devoxx.fr) conference. 
You will learn fundamental Træfik features and see some demos with Docker, Mesos/Marathon and Let's Encrypt. 

[![Traefik Devoxx France](http://img.youtube.com/vi/QvAz9mVx5TI/0.jpg)](http://www.youtube.com/watch?v=QvAz9mVx5TI)

## Get it

### Binary

You can grab the latest binary from the [releases](https://github.com/containous/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/containous/traefik/master/traefik.sample.toml):

```shell
./traefik -c traefik.toml
```

### Docker

Using the tiny Docker image:

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml traefik
```

## Test it

You can test Træfik easily using [Docker compose](https://docs.docker.com/compose), with this `docker-compose.yml` file in a folder named `traefik`:

```yaml
version: '2'

services:
  proxy:
    image: traefik
    command: --web --docker --docker.domain=docker.localhost --logLevel=DEBUG
    networks:
      - webgateway
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /dev/null:/traefik.toml

networks:
  webgateway:
    driver: bridge
```

Start it from within the `traefik` folder:

    docker-compose up -d

In a browser you may open `http://localhost:8080` to access Træfik's dashboard and observe the following magic.

Now, create a folder named `test` and create a `docker-compose.yml` in it with this content:

```yaml
version: '2'

services:
  whoami:
    image: emilevauge/whoami
    networks:
      - web
    labels:
      - "traefik.backend=whoami"
      - "traefik.frontend.rule=Host:whoami.docker.localhost"

networks:
  web:
    external:
      name: traefik_webgateway
```

Then, start and scale it in the `test` folder:

```
docker-compose up -d
docker-compose scale whoami=2
```

Finally, test load-balancing between the two services `test_whoami_1` and `test_whoami_2`:

```bash
$ curl -H Host:whoami.docker.localhost http://127.0.0.1
Hostname: ef194d07634a
IP: 127.0.0.1
IP: ::1
IP: 172.17.0.4
IP: fe80::42:acff:fe11:4
GET / HTTP/1.1
Host: 172.17.0.4:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 172.17.0.1
X-Forwarded-Host: 172.17.0.4:80
X-Forwarded-Proto: http
X-Forwarded-Server: dbb60406010d

$ curl -H Host:whoami.docker.localhost http://127.0.0.1
Hostname: 6c3c5df0c79a
IP: 127.0.0.1
IP: ::1
IP: 172.17.0.3
IP: fe80::42:acff:fe11:3
GET / HTTP/1.1
Host: 172.17.0.3:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 172.17.0.1
X-Forwarded-Host: 172.17.0.3:80
X-Forwarded-Proto: http
X-Forwarded-Server: dbb60406010d
```
