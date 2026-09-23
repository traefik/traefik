# Exposing Services with Traefik on Docker - Basic

This guide will help you get started with exposing your services through Traefik Proxy using Docker. You'll learn the fundamentals of routing HTTP traffic, setting up path-based routing, and securing your services with TLS.

## Prerequisites

- Docker and Docker Compose installed
- Basic understanding of Docker concepts
- Traefik deployed using the [Traefik Docker Setup guide](../../setup/docker.md)

## Expose Your First HTTP Service

Let's expose a simple HTTP service using the [whoami](https://hub.docker.com/r/traefik/whoami) application. This will demonstrate basic routing to a backend service.

First, create a `docker-compose.yml` file:

```yaml
services:
  traefik:
    image: "traefik:v3.4"
    container_name: "traefik"
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    networks:
      - proxy
    command:
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.docker.network=proxy"
      - "--entryPoints.web.address=:80"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"

  whoami:
    image: "traefik/whoami"
    restart: unless-stopped
    networks:
      - proxy
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.whoami.rule=Host(`whoami.docker.localhost`)"
      - "traefik.http.routers.whoami.entrypoints=web"

networks:
  proxy:
    name: proxy
```

Save this as `docker-compose.yml` and start the services:

```bash
docker compose up -d
```

### Verify Your Service

Your service is now available at http://whoami.docker.localhost/. Test that it works:

```bash
curl -H "Host: whoami.docker.localhost" http://localhost/
```

You should see output similar to:

```bash
Hostname: whoami
IP: 127.0.0.1
IP: ::1
IP: 172.18.0.3
IP: fe80::215:5dff:fe00:c9e
RemoteAddr: 172.18.0.2:55108
GET / HTTP/1.1
Host: whoami.docker.localhost
User-Agent: curl/7.68.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 172.18.0.1
X-Forwarded-Host: whoami.docker.localhost
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: 5789f594e7d5
X-Real-Ip: 172.18.0.1
```

This confirms that Traefik is successfully routing requests to your whoami application.

## Add Routing Rules

Now we'll enhance our routing by directing traffic to different services based on [URL paths](../../reference/routing-configuration/http/routing/rules-and-priority.md#path-pathprefix-and-pathregexp). This is useful for API versioning, frontend/backend separation, or organizing microservices.

Update your `docker-compose.yml` to add another service:

```yaml
# ...

# New service
  whoami-api:
    image: "traefik/whoami"
    networks:
      - proxy
    container_name: "whoami-api"
    environment:
      - WHOAMI_NAME=API Service
    labels:
      - "traefik.enable=true"
      # Path-based routing
      - "traefik.http.routers.whoami-api.rule=Host(`whoami.docker.localhost`) && PathPrefix(`/api`)"
      - "traefik.http.routers.whoami-api.entrypoints=web"
```

Apply the changes:

```bash
docker compose up -d
```

### Test the Path-Based Routing

Verify that different paths route to different services:

```bash
# Root path should go to the main whoami service
curl -H "Host: whoami.docker.localhost" http://localhost/

# /api path should go to the whoami-api service
curl -H "Host: whoami.docker.localhost" http://localhost/api
```

For the `/api` requests, you should see the response showing "API Service" in the environment variables section, confirming that your path-based routing is working correctly.

## Enable TLS

Let's secure our service with HTTPS by adding TLS. We'll start with a self-signed certificate for local development.

### Create a Self-Signed Certificate

Generate a self-signed certificate:

```bash
mkdir -p certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/local.key -out certs/local.crt \
  -subj "/CN=*.docker.localhost"
```

Create a directory for dynamic configuration and add a TLS configuration file:

```bash
mkdir -p dynamic
cat > dynamic/tls.yml << EOF
tls:
  certificates:
    - certFile: /certs/local.crt
      keyFile: /certs/local.key
EOF
```

Update your `docker-compose.yml` file with the following changes:

```yaml
services:
  traefik:
    image: "traefik:v3.4"
    container_name: "traefik"
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    networks:
      - proxy
    command:
      - "--api.insecure=false"
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.docker.network=proxy"
      - "--providers.file.directory=/etc/traefik/dynamic"
      - "--entryPoints.web.address=:80"
      - "--entryPoints.websecure.address=:443"
      - "--entryPoints.websecure.http.tls=true"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      # Add the following volumes
      - "./certs:/certs:ro"
      - "./dynamic:/etc/traefik/dynamic:ro"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dashboard.rule=Host(`dashboard.docker.localhost`)"
      - "traefik.http.routers.dashboard.entrypoints=websecure"
      - "traefik.http.routers.dashboard.service=api@internal"
      # Add the following label
      - "traefik.http.routers.dashboard.tls=true"

  whoami:
    image: "traefik/whoami"
    restart: unless-stopped
    networks:
      - proxy
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.whoami.rule=Host(`whoami.docker.localhost`)"
      - "traefik.http.routers.whoami.entrypoints=websecure"
      # Add the following label
      - "traefik.http.routers.whoami.tls=true"

  whoami-api:
    image: "traefik/whoami"
    container_name: "whoami-api"
    restart: unless-stopped
    networks:
      - proxy
    environment:
      - WHOAMI_NAME=API Service
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.whoami-api.rule=Host(`whoami.docker.localhost`) && PathPrefix(`/api`)"
      - "traefik.http.routers.whoami-api.entrypoints=websecure"
      # Add the following label
      - "traefik.http.routers.whoami-api.tls=true"

networks:
  proxy:
    name: proxy
```

Apply the changes:

```bash
docker compose up -d
```

Your browser can access https://whoami.docker.localhost/ for the service. You'll need to accept the security warning for the self-signed certificate.

## Next Steps

Now that you've mastered the basics of exposing services with Traefik on Docker, you're ready to explore more advanced features like middlewares, Let's Encrypt certificates, sticky sessions, and multi-layer routing.

Continue to the [Advanced Guide](advanced.md) to learn about:

- Adding middlewares for security and access control
- Generating certificates with Let's Encrypt
- Configuring sticky sessions for stateful applications
- Setting up multi-layer routing for authentication-based routing
