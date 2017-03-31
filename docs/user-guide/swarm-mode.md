# Docker Swarm (mode) cluster

This section explains how to create a multi-host docker cluster with
swarm mode using [docker-machine](https://docs.docker.com/machine) and
how to deploy Træfik on it.

The cluster consists of:

- 3 servers
- 1 manager
- 2 workers
- 1 [overlay](https://docs.docker.com/engine/userguide/networking/dockernetworks/#an-overlay-network) network
(multi-host networking)

## Prerequisites

1. You will need to install [docker-machine](https://docs.docker.com/machine/)
2. You will need the latest [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

## Cluster provisioning

First, let's create all the required nodes. It's a shorter version of
the [swarm tutorial](https://docs.docker.com/engine/swarm/swarm-tutorial/).

```sh
docker-machine create -d virtualbox manager
docker-machine create -d virtualbox worker1
docker-machine create -d virtualbox worker2
```

Then, let's setup the cluster, in order :

1. initialize the cluster
2. get the token for other host to join
3. on both workers, join the cluster with the token

```sh
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

```sh
docker-machine ssh manager docker node ls
ID                           HOSTNAME  STATUS  AVAILABILITY  MANAGER STATUS
2a770ov9vixeadep674265u1n    worker1   Ready   Active
dbi3or4q8ii8elbws70g4hkdh *  manager   Ready   Active        Leader
esbhhy6vnqv90xomjaomdgy46    worker2   Ready   Active
```

Finally, let's create a network for Træfik to use.

```sh
docker-machine ssh manager "docker network create --driver=overlay traefik-net"
```

## Deploy Træfik

Let's deploy Træfik as a docker service in our cluster. The only
requirement for Træfik to work with swarm mode is that it needs to run
on a manager node — we are going to use a
[constraint](https://docs.docker.com/engine/reference/commandline/service_create/#/specify-service-constraints-constraint) for
that.

```
docker-machine ssh manager "docker service create \
	--name traefik \
	--constraint=node.role==manager \
	--publish 80:80 --publish 8080:8080 \
	--mount	type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
	--network traefik-net \
	traefik \
	--docker \
	--docker.swarmmode \
	--docker.domain=traefik \
	--docker.watch \
	--web"
```

Let's explain this command:

- `--publish 80:80 --publish 8080:8080`: we publish port `80` and
  `8080` on the cluster.
- `--constraint=node.role==manager`: we ask docker to schedule Træfik
  on a manager node.
- `--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock`:
  we bind mount the docker socket where Træfik is scheduled to be able
  to speak to the daemon.
- `--network traefik-net`: we attach the Træfik service (and thus
  the underlying container) to the `traefik-net` network.
- `--docker`: enable docker backend, and `--docker.swarmmode` to
  enable the swarm mode on Træfik.
- `--web`: activate the webUI on port 8080

## Deploy your apps

We can now deploy our app on the cluster,
here [whoami](https://github.com/emilevauge/whoami), a simple web
server in Go. We start 2 services, on the `traefik-net` network.

```sh
docker-machine ssh manager "docker service create \
	--name whoami0 \
	--label traefik.port=80 \
	--network traefik-net \
	emilevauge/whoami"

docker-machine ssh manager "docker service create \
	--name whoami1 \
	--label traefik.port=80 \
	--network traefik-net \
	--label traefik.backend.loadbalancer.sticky=true \
	emilevauge/whoami"
```

Note that we set whoami1 to use sticky sessions (`--label traefik.backend.loadbalancer.sticky=true`).  We'll demonstrate that later.
If using `docker stack deploy`, there is [a specific way that the labels must be defined in the docker-compose file](https://github.com/containous/traefik/issues/994#issuecomment-269095109).

Check that everything is scheduled and started:

```sh
docker-machine ssh manager "docker service ls"
ID            NAME     REPLICAS  IMAGE              COMMAND
ab046gpaqtln  whoami0  1/1       emilevauge/whoami
cgfg5ifzrpgm  whoami1  1/1       emilevauge/whoami
dtpl249tfghc  traefik  1/1       traefik            --docker --docker.swarmmode --docker.domain=traefik --docker.watch --web
```

## Access to your apps through Træfik

```sh
curl -H Host:whoami0.traefik http://$(docker-machine ip manager)
Hostname: 8147a7746e7a
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.3
IP: fe80::42:aff:fe00:903
IP: 172.18.0.3
IP: fe80::42:acff:fe12:3
GET / HTTP/1.1
Host: 10.0.9.3:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.3:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c

curl -H Host:whoami1.traefik http://$(docker-machine ip manager)
Hostname: ba2c21488299
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.4
IP: fe80::42:aff:fe00:904
IP: 172.18.0.2
IP: fe80::42:acff:fe12:2
GET / HTTP/1.1
Host: 10.0.9.4:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.4:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c
```

Note that as Træfik is published, you can access it from any machine
and not only the manager.

```sh
curl -H Host:whoami0.traefik http://$(docker-machine ip worker1)
Hostname: 8147a7746e7a
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.3
IP: fe80::42:aff:fe00:903
IP: 172.18.0.3
IP: fe80::42:acff:fe12:3
GET / HTTP/1.1
Host: 10.0.9.3:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.3:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c

curl -H Host:whoami1.traefik http://$(docker-machine ip worker2)
Hostname: ba2c21488299
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.4
IP: fe80::42:aff:fe00:904
IP: 172.18.0.2
IP: fe80::42:acff:fe12:2
GET / HTTP/1.1
Host: 10.0.9.4:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.4:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c
```

## Scale both services

```sh
docker-machine ssh manager "docker service scale whoami0=5"

docker-machine ssh manager "docker service scale whoami1=5"
```


Check that we now have 5 replicas of each `whoami` service:

```sh
docker-machine ssh manager "docker service ls"
ID            NAME     REPLICAS  IMAGE              COMMAND
ab046gpaqtln  whoami0  5/5       emilevauge/whoami
cgfg5ifzrpgm  whoami1  5/5       emilevauge/whoami
dtpl249tfghc  traefik  1/1       traefik            --docker --docker.swarmmode --docker.domain=traefik --docker.watch --web
```
## Access to your whoami0 through Træfik multiple times.

Repeat the following command multiple times and note that the Hostname changes each time as Traefik load balances each request against the 5 tasks.
```sh
curl -H Host:whoami0.traefik http://$(docker-machine ip manager)
Hostname: 8147a7746e7a
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.3
IP: fe80::42:aff:fe00:903
IP: 172.18.0.3
IP: fe80::42:acff:fe12:3
GET / HTTP/1.1
Host: 10.0.9.3:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.3:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c
```

Do the same against whoami1.  
```sh
curl -H Host:whoami1.traefik http://$(docker-machine ip manager)
Hostname: ba2c21488299
IP: 127.0.0.1
IP: ::1
IP: 10.0.9.4
IP: fe80::42:aff:fe00:904
IP: 172.18.0.2
IP: fe80::42:acff:fe12:2
GET / HTTP/1.1
Host: 10.0.9.4:80
User-Agent: curl/7.35.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 192.168.99.1
X-Forwarded-Host: 10.0.9.4:80
X-Forwarded-Proto: http
X-Forwarded-Server: 8fbc39271b4c
```
Wait, I thought we added the sticky flag to whoami1?  Traefik relies on a cookie to maintain stickyness so you'll need to test this with a browser.

First you need to add whoami1.traefik to your hosts file:
```ssh
if [ -n "$(grep whoami1.traefik /etc/hosts)" ];  
then 
echo "whoami1.traefik already exists (make sure the ip is current)"; 
else 
sudo -- sh -c -e "echo '$(docker-machine ip manager)\twhoami1.traefik' 
>> /etc/hosts"; 
fi
```

Now open your browser and go to http://whoami1.traefik/

You will now see that stickyness is maintained.

![](http://i.giphy.com/ujUdrdpX7Ok5W.gif)


