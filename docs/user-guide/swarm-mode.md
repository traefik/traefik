# Docker Swarm (mode) cluster

This section explains how to create a multi-host docker cluster with swarm mode using [docker-machine](https://docs.docker.com/machine) and how to deploy Traefik on it.

The cluster consists of:

- 3 servers
- 1 manager
- 2 workers
- 1 [overlay](https://docs.docker.com/network/overlay/) network (multi-host networking)


## Prerequisites

1. You will need to install [docker-machine](https://docs.docker.com/machine/)
2. You will need the latest [VirtualBox](https://www.virtualbox.org/wiki/Downloads)


## Cluster provisioning

First, let's create all the required nodes.
It's a shorter version of the [swarm tutorial](https://docs.docker.com/engine/swarm/swarm-tutorial/).

```shell
docker-machine create -d virtualbox manager
docker-machine create -d virtualbox worker1
docker-machine create -d virtualbox worker2
```

Then, let's setup the cluster, in order:

1. initialize the cluster
1. get the token for other host to join
1. on both workers, join the cluster with the token

```shell
docker-machine ssh manager "docker swarm init \
	--listen-addr $(docker-machine ip manager) \
	--advertise-addr $(docker-machine ip manager)"

export worker_token=$(docker-machine ssh manager "docker swarm \
join-token worker -q")

docker-machine ssh worker1 "docker swarm join \
	--token=${worker_token} \
	--listen-addr $(docker-machine ip worker1) \
	--advertise-addr $(docker-machine ip worker1) \
	$(docker-machine ip manager)"

docker-machine ssh worker2 "docker swarm join \
	--token=${worker_token} \
	--listen-addr $(docker-machine ip worker2) \
	--advertise-addr $(docker-machine ip worker2) \
	$(docker-machine ip manager)"
```

Let's validate the cluster is up and running.

```shell
docker-machine ssh manager docker node ls
```
```
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
013v16l1sbuwjqcn7ucbu4jwt    worker1   Ready   Active
8buzkquycd17jqjber0mo2gn8    worker2   Ready   Active
fnpj8ozfc85zvahx2r540xfcf *  manager   Ready   Active        Leader
```

Finally, let's create a network for Traefik to use.

```shell
docker-machine ssh manager "docker network create --driver=overlay traefik-net"
```


## Deploy Traefik

Let's deploy Traefik as a docker service in our cluster.
The only requirement for Traefik to work with swarm mode is that it needs to run on a manager node - we are going to use a [constraint](https://docs.docker.com/engine/reference/commandline/service_create/#specify-service-constraints---constraint) for that.

```shell
docker-machine ssh manager "docker service create \
	--name traefik \
	--constraint=node.role==manager \
	--publish 80:80 --publish 8080:8080 \
	--mount	type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
	--network traefik-net \
	traefik \
	--docker \
	--docker.swarmMode \
	--docker.domain=traefik \
	--docker.watch \
	--api"
```

Let's explain this command:

| Option                                                                      | Description                                                                                    |
|-----------------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `--publish 80:80 --publish 8080:8080`                                       | we publish port `80` and `8080` on the cluster.                                                |
| `--constraint=node.role==manager`                                           | we ask docker to schedule Traefik on a manager node.                                            |
| `--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock` | we bind mount the docker socket where Traefik is scheduled to be able to speak to the daemon.   |
| `--network traefik-net`                                                     | we attach the Traefik service (and thus the underlying container) to the `traefik-net` network. |
| `--docker`                                                                  | enable docker provider, and `--docker.swarmMode` to enable the swarm mode on Traefik.            |
| `--api`                                                                      | activate the webUI on port 8080                                                                |


## Deploy your apps

We can now deploy our app on the cluster, here [whoami](https://github.com/containous/whoami), a simple web server in Go.
We start 2 services, on the `traefik-net` network.

```shell
docker-machine ssh manager "docker service create \
	--name whoami0 \
	--label traefik.port=80 \
	--network traefik-net \
	containous/whoami"

docker-machine ssh manager "docker service create \
	--name whoami1 \
	--label traefik.port=80 \
	--network traefik-net \
	--label traefik.backend.loadbalancer.sticky=true \
	containous/whoami"
```

!!! note
    We set `whoami1` to use sticky sessions (`--label traefik.backend.loadbalancer.stickiness=true`).
    We'll demonstrate that later.

!!! note
    If using `docker stack deploy`, there is [a specific way that the labels must be defined in the docker-compose file](https://github.com/containous/traefik/issues/994#issuecomment-269095109).

Check that everything is scheduled and started:

```shell
docker-machine ssh manager "docker service ls"
```
```
ID            NAME     MODE        REPLICAS  IMAGE                     PORTS
moq3dq4xqv6t  traefik  replicated  1/1       traefik:latest            *:80->80/tcp,*:8080->8080/tcp
ysil6oto1wim  whoami0  replicated  1/1       containous/whoami:latest
z9re2mnl34k4  whoami1  replicated  1/1       containous/whoami:latest
```


## Access to your apps through Traefik

```shell
curl -H Host:whoami0.traefik http://$(docker-machine ip manager)
```
```yaml
Hostname: 5b0b3d148359
IP: 127.0.0.1
IP: 10.0.0.8
IP: 10.0.0.4
IP: 172.18.0.5
GET / HTTP/1.1
Host: whoami0.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.2
X-Forwarded-Host: whoami0.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```
```shell
curl -H Host:whoami1.traefik http://$(docker-machine ip manager)
```
```yaml
Hostname: 3633163970f6
IP: 127.0.0.1
IP: 10.0.0.14
IP: 10.0.0.6
IP: 172.18.0.5
GET / HTTP/1.1
Host: whoami1.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.2
X-Forwarded-Host: whoami1.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```

!!! note
    As Traefik is published, you can access it from any machine and not only the manager.

```shell
curl -H Host:whoami0.traefik http://$(docker-machine ip worker1)
```
```yaml
Hostname: 5b0b3d148359
IP: 127.0.0.1
IP: 10.0.0.8
IP: 10.0.0.4
IP: 172.18.0.5
GET / HTTP/1.1
Host: whoami0.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.3
X-Forwarded-Host: whoami0.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```
```shell
curl -H Host:whoami1.traefik http://$(docker-machine ip worker2)
```
```yaml
Hostname: 3633163970f6
IP: 127.0.0.1
IP: 10.0.0.14
IP: 10.0.0.6
IP: 172.18.0.5
GET / HTTP/1.1
Host: whoami1.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.4
X-Forwarded-Host: whoami1.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```

## Scale both services

```shell
docker-machine ssh manager "docker service scale whoami0=5"
docker-machine ssh manager "docker service scale whoami1=5"
```

Check that we now have 5 replicas of each `whoami` service:

```shell
docker-machine ssh manager "docker service ls"
```
```
ID            NAME     MODE        REPLICAS  IMAGE                     PORTS
moq3dq4xqv6t  traefik  replicated  1/1       traefik:latest            *:80->80/tcp,*:8080->8080/tcp
ysil6oto1wim  whoami0  replicated  5/5       containous/whoami:latest
z9re2mnl34k4  whoami1  replicated  5/5       containous/whoami:latest
```

## Access to your `whoami0` through Traefik multiple times.

Repeat the following command multiple times and note that the Hostname changes each time as Traefik load balances each request against the 5 tasks:

```shell
curl -H Host:whoami0.traefik http://$(docker-machine ip manager)
```
```yaml
Hostname: f3138d15b567
IP: 127.0.0.1
IP: 10.0.0.5
IP: 10.0.0.4
IP: 172.18.0.3
GET / HTTP/1.1
Host: whoami0.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.2
X-Forwarded-Host: whoami0.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```

Do the same against `whoami1`:

```shell
curl -c cookies.txt -H Host:whoami1.traefik http://$(docker-machine ip manager)
```
```yaml
Hostname: 348e2f7bf432
IP: 127.0.0.1
IP: 10.0.0.15
IP: 10.0.0.6
IP: 172.18.0.6
GET / HTTP/1.1
Host: whoami1.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.255.0.2
X-Forwarded-Host: whoami1.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```

Because the sticky sessions require cookies to work, we used the `-c cookies.txt` option to store the cookie into a file.
The cookie contains the IP of the container to which the session sticks:

```shell
cat ./cookies.txt
```
```
# Netscape HTTP Cookie File
# https://curl.haxx.se/docs/http-cookies.html
# This file was generated by libcurl! Edit at your own risk.

whoami1.traefik FALSE  /  FALSE  0  _TRAEFIK_BACKEND  http://10.0.0.15:80
```

If you load the cookies file (`-b cookies.txt`) for the next request, you will see that stickiness is maintained:

```shell
curl -b cookies.txt -H Host:whoami1.traefik http://$(docker-machine ip manager)
```
```yaml
Hostname: 348e2f7bf432
IP: 127.0.0.1
IP: 10.0.0.15
IP: 10.0.0.6
IP: 172.18.0.6
GET / HTTP/1.1
Host: whoami1.traefik
User-Agent: curl/7.55.1
Accept: */*
Accept-Encoding: gzip
Cookie: _TRAEFIK_BACKEND=http://10.0.0.15:80
X-Forwarded-For: 10.255.0.2
X-Forwarded-Host: whoami1.traefik
X-Forwarded-Proto: http
X-Forwarded-Server: 77fc29c69fe4
```

![GIF Magica](https://i.giphy.com/ujUdrdpX7Ok5W.gif)
