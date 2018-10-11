# Cluster setup with Docker Swarm Mode and Traefik

This guide shows you how to create a cluster of Linux servers managed with Docker Swarm mode to deploy your projects.

It also shows how to set up an integrated main Traefik load balancer / proxy to receive incoming connections, re-transmit communication to Docker containers based on the domains, generate TLS / SSL certificates (for HTTPS) with Let's Encrypt and handle HTTPS.

And it shows how to expose the internal web UI through the same Traefik itself, with a HTTPS certificate and HTTP Basic Auth.

The guide is designed for you to be able to have a main load balancer that covers all the Docker Swarm cluster, handles HTTPS certificates and requests for each domain. But in a way that allows you to have other Traefik services inside each stack without interfering each other, to redirect based on path in the same stack (e.g. one container handles `/` for a web frontend and another handles `/api` for an API under the same domain), or to redirect from HTTP to HTTPS selectively.

## Install a new Linux server with Docker

* Create a new remote server (VPS).
* If you can create a `swap` disk partition, do it based on the [Ubuntu FAQ for swap partitions](https://help.ubuntu.com/community/SwapFaq#How_much_swap_do_I_need.3F).
* Deploy the latest Ubuntu LTS version image.
* Connect to it via SSH, e.g.:

```bash
ssh root@172.173.174.175
```

* Define a server name using a subdomain of a domain you own, for example `dog.example.com`.
* Create a temporal environment variable with the name of the host to be used later, e.g.:

```bash
export USE_HOSTNAME=dog.example.com
```

* Set up the server `hostname`:

```bash
# Set up the server hostname
echo $USE_HOSTNAME > /etc/hostname
hostname -F /etc/hostname
```

* Update packages:

```bash
# Install the latest updates
apt-get update
apt-get upgrade -y
```

* Install Docker following the official guide: https://docs.docker.com/install/
* Or alternatively, run the official convenience script:

```bash
# Download Docker
curl -fsSL get.docker.com -o get-docker.sh
# Install Docker using the stable channel (instead of the default "edge")
CHANNEL=stable sh get-docker.sh
# Remove Docker install script
rm get-docker.sh
```

## Set up swarm mode

In Docker Swarm Mode you have one or more "manager" nodes and one or more "worker" nodes (that can be the same manager nodes).

The first step, is to configure one (or more) manager nodes.

* On the main manager node, run:

```bash
docker swarm init
```

* On the main manager node, for each additional manager node you want to set up, run:

```bash
docker swarm join-token manager
```

* Copy the result and paste it in the additional manager node's terminal, it will be something like:

```bash
 docker swarm join --token SWMTKN-1-5tl7yaasdfd9qt9j0easdfnml4lqbosbasf14p13-f3hem9ckmkhasdf3idrzk5gz 172.173.174.175:2377
```

* On the main manager node, for each additional worker node you want to set up, run:

```bash
docker swarm join-token worker
```

* Copy the result and paste it in the additional worker node's terminal, it will be something like:

```bash
docker swarm join --token SWMTKN-1-5tl7ya98erd9qtasdfml4lqbosbhfqv3asdf4p13-dzw6ugasdfk0arn0 172.173.174.175:2377
```

## Traefik

Set up a main load balancer with Traefik that handles the public connections and Let's encrypt HTTPS certificates. 

* Connect via SSH to a manager node in your cluster (you might have only one node) that will have the Traefik service.
* Create a network that will be shared with Traefik and the containers that should be accessible from the outside, with:

```bash
docker network create --driver=overlay traefik-public
```

* Create a volume in where Traefik will store HTTPS certificates:

```bash
docker volume create traefik-public-certificates
```

**Note**: you can store certificates in Consul and deploy Traefik in each node, there's another guide covering that, for this guide we will keep it simple and deploy it to a single node.

* Get the Swarm node ID of this node and store it in an environment variable:

```bash
export NODE_ID=$(docker info -f '{{.Swarm.NodeID}}')
```

* Create a tag in this node, so that Traefik is always deployed to the same node and uses the existing volume:

```bash
docker node update --label-add traefik-public.traefik-public-certificates=true $NODE_ID
```

* Create an environment variable with your email, to be used for the generation of Let's Encrypt certificates:

```bash
export EMAIL=admin@example.com
```

* Create an environment variable with the name of the host (you might have created it already), e.g.:

```bash
export USE_HOSTNAME=dog.example.com
# or if you have your $HOSTNAME variable configured:
export USE_HOSTNAME=$HOSTNAME
```

* You will access the Traefik dashboard at `traefik.<your hostname>`, e.g. `traefik.dog.example.com`. So, make sure that your DNS records point `traefik.<your hostname>` to one of the IPs of the cluster. Better if it is the IP where the Traefik service runs (the manager node you are currently connected to).

* Create an environment variable with a username (you will use it for the HTTP Basic Auth), for example:

```bash
export USERNAME=admin
```

* Create an environment variable with the password, e.g.:

```bash
export PASSWORD=changethis
```

* Use `openssl` to generate the "hashed" version of the password and store it in an environment variable:

```bash
export HASHED_PASSWORD=$(openssl passwd -apr1 $PASSWORD)
```

* Create an environment variable with the user name and password in "`htpasswd`" format:

```bash
export USERNAME_PASSWORD=$USERNAME:$HASHED_PASSWORD
```

* You can check the contents with:

```bash
echo $USERNAME_PASSWORD
```

It will look like:

```
admin:$apr1$89eqM5Ro$CxaFELthUKV21DpI3UTQO.
```

* Create a Traefik service, copy this long command in the terminal:

```bash
docker service create \
    --name traefik \
    --constraint=node.labels.traefik-public.traefik-public-certificates==true \
    --publish 80:80 \
    --publish 443:443 \
    --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
    --mount type=volume,source=traefik-public-certificates,target=/certificates \
    --network traefik-public \
    --label "traefik.frontend.rule=Host:traefik.$USE_HOSTNAME" \
    --label "traefik.enable=true" \
    --label "traefik.port=8080" \
    --label "traefik.tags=traefik-public" \
    --label "traefik.docker.network=traefik-public" \
    --label "traefik.redirectorservice.frontend.entryPoints=http" \
    --label "traefik.redirectorservice.frontend.redirect.entryPoint=https" \
    --label "traefik.webservice.frontend.entryPoints=https" \
    --label "traefik.frontend.auth.basic=$USERNAME_PASSWORD" \
    traefik:v1.6 \
    --docker \
    --docker.swarmmode \
    --docker.watch \
    --docker.exposedbydefault=false \
    --constraints=tag==traefik-public \
    --entrypoints='Name:http Address::80' \
    --entrypoints='Name:https Address::443 TLS' \
    --acme \
    --acme.email=$EMAIL \
    --acme.storage=/certificates/acme.json \
    --acme.entryPoint=https \
    --acme.httpChallenge.entryPoint=http\
    --acme.onhostrule=true \
    --acme.acmelogging=true \
    --logLevel=INFO \
    --accessLog \
    --api
```

You will be able to securely access the web UI at `https://traefik.<your domain>` using the created username and password.

The previous command explained:

* `docker service create`: create a Docker Swarm mode service
* `--name traefik`: name the service "traefik"
* `--constraint=node.labels.traefik-public.traefik-public-certificates==true` make it run on a specific node, to be able to use the certificates stored in a volume in that node
* `--publish 80:80`: listen on ports 80 - HTTP
* `--publish 443:443`: listen on port 443 - HTTPS
* `--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock`: communicate with Docker, to read labels, etc.
* `--mount type=volume,source=traefik-public-certificates,target=/certificates`: create a volume to store TLS certificates
* `--network traefik-public`: listen to the specific network traefik-public
* `--label "traefik.frontend.rule=Host:traefik.$USE_HOSTNAME"`: enable the Traefik API and dashboard in the host `traefik.$USE_HOSTNAME`, using the `$USE_HOSTNAME` environment variable created above
* `--label "traefik.enable=true"`: make Traefik expose "itself" as a Docker service, this is what makes the Traefik dashboard available with HTTPS and basic auth
* `--label "traefik.port=8080"`: when Traefik exposes itself as a service (for the dashboard), use the internal service port `8080`
* `--label "traefik.tags=traefik-public"`: as the main Traefik proxy will only expose services with the `traefik-public` tag (using a parameter below), make the dashboard service have this tag too, so that the Traefik public (itself) can find it and expose it
* `--label "traefik.docker.network=traefik-public"`: make the dashboard service use the `traefik-public` network to expose itself
* `--label "traefik.redirectorservice.frontend.entryPoints=http"`: make the web dashboard listen to HTTP, so that it can redirect to HTTPS
* `--label "traefik.redirectorservice.frontend.redirect.entryPoint=https"`: make Traefik redirect HTTP trafic to HTTPS for the web dashboard
* `--label "traefik.webservice.frontend.entryPoints=https"`: make the web dashboard listen and serve on HTTPS
* `--label "traefik.frontend.auth.basic=$USERNAME_PASSWORD"`: enable basic auth, so that not every one can access your Traefik web dashboard, it uses the username and password created above
* `traefik:v1.6`: use the image `traefik:v1.6`
* `--docker`: enable Docker
* `--docker.swarmmode`: enable Docker Swarm Mode
* `--docker.watch`: enable "watch", so it reloads its config based on new stacks and labels
* `--docker.exposedbydefault=false`: don't expose all the services, only services with traefik.enable=true
* `--constraints=tag==traefik-public`: only show services with traefik.tag=traefik-public, to isolate from possible intra-stack traefik instances
* `--entrypoints='Name:http Address::80'`: create an entrypoint http, on port 80
* `--entrypoints='Name:https Address::443 TLS'`: create an entrypoint https, on port 443 with TLS enabled
* `--acme`: enable Let's encrypt
* `--acme.email=$EMAIL`: let's encrypt email, using the environment variable
* `--acme.storage=/certificates/acme.json`: where to store the Let's encrypt TLS certificates - in the mapped volume
* `--acme.entryPoint=https`: the entrypoint for Let's encrypt - created above
* `--acme.httpChallenge.entryPoint=http`: use HTTP for the ACME (Let's Encrypt HTTPS certificates) challenge, as HTTPS was disabled after a security issue
* `--acme.onhostrule=true`: get new certificates automatically with host rules: "traefik.frontend.rule=Host:web.example.com"
* `--acme.acmelogging=true`: log Let's encrypt activity - to debug when and if it gets certificates
* `--logLevel=INFO`: default logging, if the web UI is not enough to debug configurations and hosts detected, or you want to see more of the logs, set it to `DEBUG`. Have in mind that after some time it might affect performance.
* `--accessLog`: enable the access log, to see and debug HTTP traffic
* `--api`: enable the API, which includes the dashboard


To check if it worked, check the logs:

```bash
docker service logs traefik
# To make it scrollable with `less`, run:
# docker service logs traefik | less
```
