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
      - /var/run/docker.sock:/var/run/docker.sock #So that Traefik can listen to the Docker events
```

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
    image: emilevauge/whoami # A container that exposes an API to show its IP address
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
docker-compose up -d --scale whoami=2 
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

Now that you have a basic understanding of how Traefik can automatically create the routes to your services and load balance them, it might be time to dive into [the documentation](https://docs.traefik.io/) and let Traefik work for you!
Whatever your infrastructure is, there is probably [an available Traefik backend](https://docs.traefik.io/#supported-backends) that will do the job.

Our recommendation would be to see for yourself how simple it is to enable HTTPS with [Traefik's let's encrypt integration](https://docs.traefik.io/user-guide/examples/#lets-encrypt-support) using the dedicated [user guide](https://docs.traefik.io/user-guide/docker-and-lets-encrypt/).
