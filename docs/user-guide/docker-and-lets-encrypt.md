# Let's Encrypt & Docker

In this use case, we want to use Traefik as a _layer-7_ load balancer with SSL termination for a set of micro-services used to run a web application.

We also want to automatically _discover any services_ on the Docker host and let Traefik reconfigure itself automatically when containers get created (or shut down) so HTTP traffic can be routed accordingly.

In addition, we want to use Let's Encrypt to automatically generate and renew SSL certificates per hostname.

## Setting Up

In order for this to work, you'll need a server with a public IP address, with Docker and docker-compose installed on it.

In this example, we're using the fictitious domain _my-awesome-app.org_.

In real-life, you'll want to use your own domain and have the DNS configured accordingly so the hostname records you'll want to use point to the aforementioned public IP address.

## Networking

Docker containers can only communicate with each other over TCP when they share at least one network.
This makes sense from a topological point of view in the context of networking, since Docker under the hood creates IPTable rules so containers can't reach other containers _unless you'd want to_.

In this example, we're going to use a single network called `web` where all containers that are handling HTTP traffic (including Traefik) will reside in.

On the Docker host, run the following command:

```shell
docker network create web
```

Now, let's create a directory on the server where we will configure the rest of Traefik:

```shell
mkdir -p /opt/traefik
```

Within this directory, we're going to create 3 empty files:

```shell
touch /opt/traefik/docker-compose.yml
touch /opt/traefik/acme.json && chmod 600 /opt/traefik/acme.json
touch /opt/traefik/traefik.toml
```

The `docker-compose.yml` file will provide us with a simple, consistent and more importantly, a deterministic way to create Traefik.

The contents of the file is as follows:

```yaml
version: '2'

services:
  traefik:
    image: traefik:1.5.4
    restart: always
    ports:
      - 80:80
      - 443:443
    networks:
      - web
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /opt/traefik/traefik.toml:/traefik.toml
      - /opt/traefik/acme.json:/acme.json
    container_name: traefik

networks:
  web:
    external: true
```

As you can see, we're mounting the `traefik.toml` file as well as the (empty) `acme.json` file in the container.  
Also, we're mounting the `/var/run/docker.sock` Docker socket in the container as well, so Traefik can listen to Docker events and reconfigure its own internal configuration when containers are created (or shut down).  
Also, we're making sure the container is automatically restarted by the Docker engine in case of problems (or: if the server is rebooted).
We're publishing the default HTTP ports `80` and `443` on the host, and making sure the container is placed within the `web` network we've created earlier on.  
Finally, we're giving this container a static name called `traefik`.

Let's take a look at a simple `traefik.toml` configuration as well before we'll create the Traefik container:

```toml
debug = false

logLevel = "ERROR"
defaultEntryPoints = ["https","http"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
  [entryPoints.https]
  address = ":443"
  [entryPoints.https.tls]

[retry]

[docker]
endpoint = "unix:///var/run/docker.sock"
domain = "my-awesome-app.org"
watch = true
exposedByDefault = false

[acme]
email = "your-email-here@my-awesome-app.org"
storage = "acme.json"
entryPoint = "https"
onHostRule = true
[acme.httpChallenge]
entryPoint = "http"
```

This is the minimum configuration required to do the following:

- Log `ERROR`-level messages (or more severe) to the console, but silence `DEBUG`-level messages
- Check for new versions of Traefik periodically
- Create two entry points, namely an `HTTP` endpoint on port `80`, and an `HTTPS` endpoint on port `443` where all incoming traffic on port `80` will immediately get redirected to `HTTPS`.
- Enable the Docker provider and listen for container events on the Docker unix socket we've mounted earlier. However, **new containers will not be exposed by Traefik by default, we'll get into this in a bit!**
- Enable automatic request and configuration of SSL certificates using Let's Encrypt.
    These certificates will be stored in the `acme.json` file, which you can back-up yourself and store off-premises.

Alright, let's boot the container. From the `/opt/traefik` directory, run `docker-compose up -d` which will create and start the Traefik container.

## Exposing Web Services to the Outside World

Now that we've fully configured and started Traefik, it's time to get our applications running!

Let's take a simple example of a micro-service project consisting of various services, where some will be exposed to the outside world and some will not.

The `docker-compose.yml` of our project looks like this:

```yaml
version: "2.1"

services:
  app:
    image: my-docker-registry.com/my-awesome-app/app:latest
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: always
    networks:
      - web
      - default
    expose:
      - "9000"
    labels:
      - "traefik.docker.network=web"
      - "traefik.enable=true"
      - "traefik.basic.frontend.rule=Host:app.my-awesome-app.org"
      - "traefik.basic.port=9000"
      - "traefik.basic.protocol=http"
      - "traefik.admin.frontend.rule=Host:admin-app.my-awesome-app.org"
      - "traefik.admin.protocol=https"
      - "traefik.admin.port=9443"

  db:
    image: my-docker-registry.com/back-end/5.7
    restart: always

  redis:
    image: my-docker-registry.com/back-end/redis:4-alpine
    restart: always

  events:
    image: my-docker-registry.com/my-awesome-app/events:latest
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: always
    networks:
      - web
      - default
    expose:
      - "3000"
    labels:
      - "traefik.backend=my-awesome-app-events"
      - "traefik.docker.network=web"
      - "traefik.frontend.rule=Host:events.my-awesome-app.org"
      - "traefik.enable=true"
      - "traefik.port=3000"

networks:
  web:
    external: true
```

Here, we can see a set of services with two applications that we're actually exposing to the outside world.  
Notice how there isn't a single container that has any published ports to the host -- everything is routed through Docker networks.  
Also, only the containers that we want traffic to get routed to are attached to the `web` network we created at the start of this document.

Since the `traefik` container we've created and started earlier is also attached to this network, HTTP requests can now get routed to these containers.

### Labels

As mentioned earlier, we don't want containers exposed automatically by Traefik.

The reason behind this is simple: we want to have control over this process ourselves.
Thanks to Docker labels, we can tell Traefik how to create its internal routing configuration.

Let's take a look at the labels themselves for the `app` service, which is a HTTP webservice listing on port 9000:

```yaml
- "traefik.docker.network=web"
- "traefik.enable=true"
- "traefik.basic.frontend.rule=Host:app.my-awesome-app.org"
- "traefik.basic.port=9000"
- "traefik.basic.protocol=http"
- "traefik.admin.frontend.rule=Host:admin-app.my-awesome-app.org"
- "traefik.admin.protocol=https"
- "traefik.admin.port=9443"
```

We use both `container labels` and `service labels`.

#### Container labels

First, we specify the `backend` name which corresponds to the actual service we're routing **to**.

We also tell Traefik to use the `web` network to route HTTP traffic to this container.
With the `traefik.enable` label, we tell Traefik to include this container in its internal configuration.

With the `frontend.rule` label, we tell Traefik that we want to route to this container if the incoming HTTP request contains the `Host` `app.my-awesome-app.org`.
Essentially, this is the actual rule used for Layer-7 load balancing.

Finally but not unimportantly, we tell Traefik to route **to** port `9000`, since that is the actual TCP/IP port the container actually listens on.

### Service labels

`Service labels` allow managing many routes for the same container.

When both `container labels` and `service labels` are defined, `container labels` are just used as default values for missing `service labels` but no frontend/backend are going to be defined only with these labels.
Obviously, labels `traefik.frontend.rule` and `traefik.port` described above, will only be used to complete information set in `service labels` during the container frontends/backends creation.

In the example, two service names are defined : `basic` and `admin`.
They allow creating two frontends and two backends.

- `basic` has only one `service label` : `traefik.basic.protocol`.
Traefik will use values set in `traefik.frontend.rule` and `traefik.port` to create the `basic` frontend and backend.
The frontend listens to incoming HTTP requests which contain the `Host` `app.my-awesome-app.org` and redirect them in `HTTP` to the port `9000` of the backend.
- `admin` has all the `services labels` needed to create the `admin` frontend and backend (`traefik.admin.frontend.rule`, `traefik.admin.protocol`, `traefik.admin.port`).
Traefik will create a frontend to listen to incoming HTTP requests which contain the `Host` `admin-app.my-awesome-app.org` and redirect them in `HTTPS` to the port `9443` of the backend.

#### Gotchas and tips

- Always specify the correct port where the container expects HTTP traffic using `traefik.port` label.  
    If a container exposes multiple ports, Traefik may forward traffic to the wrong port.
    Even if a container only exposes one port, you should always write configuration defensively and explicitly.
- Should you choose to enable the `exposedByDefault` flag in the `traefik.toml` configuration, be aware that all containers that are placed in the same network as Traefik will automatically be reachable from the outside world, for everyone and everyone to see.
    Usually, this is a bad idea.
- With the `traefik.frontend.auth.basic` label, it's possible for Traefik to provide a HTTP basic-auth challenge for the endpoints you provide the label for.
- Traefik has built-in support to automatically export [Prometheus](https://prometheus.io) metrics
- Traefik supports websockets out of the box. In the example above, the `events`-service could be a NodeJS-based application which allows clients to connect using websocket protocol.
    Thanks to the fact that HTTPS in our example is enforced, these websockets are automatically secure as well (WSS)

### Final thoughts

Using Traefik as a Layer-7 load balancer in combination with both Docker and Let's Encrypt provides you with an extremely flexible, powerful and self-configuring solution for your projects.

With Let's Encrypt, your endpoints are automatically secured with production-ready SSL certificates that are renewed automatically as well.
