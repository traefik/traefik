# Exposing Services with Traefik on Docker Swarm - Basic

This guide will help you get started with exposing your services through Traefik Proxy using Docker Swarm. You'll learn the fundamentals of routing HTTP traffic, setting up path-based routing, and securing your services with TLS.

## Prerequisites

- Docker Swarm cluster initialized
- Basic understanding of Docker Swarm concepts
- Traefik deployed using the [Traefik Docker Swarm Setup guide](../../setup/swarm.md)


## Expose Your First HTTP Service

Let's expose a simple HTTP service using the [whoami](https://hub.docker.com/r/traefik/whoami) application. This will demonstrate basic routing to a backend service.

First, update your existing `docker-compose.yml` file if you haven't already:

```yaml
services:
  whoami:
    image: traefik/whoami
    networks:
      - traefik_proxy
    deploy:
      replicas: 3
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.whoami.rule=Host(`whoami.swarm.localhost`)"
        - "traefik.http.routers.whoami.entrypoints=web,websecure"
```

Save this as `docker-compose.yml` and deploy the stack:

```bash
docker stack deploy -c docker-compose.yml traefik
```

### Verify Your Service

Your service is now available at http://whoami.swarm.localhost/. Test that it works:

```bash
curl -H "Host: whoami.swarm.localhost" http://localhost/
```

You should see output similar to:

```bash
Hostname: whoami.1.7c8f7tr56q3p949rscxrkp80e
IP: 127.0.0.1
IP: ::1
IP: 10.0.1.8
IP: fe80::215:5dff:fe00:c9e
RemoteAddr: 10.0.1.2:45098
GET / HTTP/1.1
Host: whoami.swarm.localhost
User-Agent: curl/7.68.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.0.1.1
X-Forwarded-Host: whoami.swarm.localhost
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: 5789f594e7d5
X-Real-Ip: 10.0.1.1
```

This confirms that Traefik is successfully routing requests to your whoami application.

## Add Routing Rules

Now we'll enhance our routing by directing traffic to different services based on [URL paths](../../reference/routing-configuration/http/routing/rules-and-priority.md#path-pathprefix-and-pathregexp). This is useful for API versioning, frontend/backend separation, or organizing microservices.

Update your `docker-compose.yml` to add another service:

```yaml
# ...

# New service
  whoami-api:
    image: traefik/whoami
    networks:
      - traefik_proxy
    environment:
      - WHOAMI_NAME=API Service
    deploy:
      replicas: 2
      labels:
        - "traefik.enable=true"
        # Path-based routing
        - "traefik.http.routers.whoami-api.rule=Host(`whoami.swarm.localhost`) && PathPrefix(`/api`)"
        - "traefik.http.routers.whoami-api.entrypoints=web,websecure"
        - "traefik.http.routers.whoami-api.service=whoami-api-svc"
        - "traefik.http.services.whoami-api-svc.loadbalancer.server.port=80"

# ...
```

Apply the changes:

```bash
docker stack deploy -c docker-compose.yml traefik
```

### Test the Path-Based Routing

Verify that different paths route to different services:

```bash
# Root path should go to the main whoami service
curl -H "Host: whoami.swarm.localhost" http://localhost/

# /api path should go to the whoami-api service
curl -H "Host: whoami.swarm.localhost" http://localhost/api
```

For the `/api` requests, you should see the response showing "API Service" in the environment variables section, confirming that your path-based routing is working correctly.

## Enable TLS

Let's secure our service with HTTPS by adding TLS. We'll start with a self-signed certificate for local development.

### Create a Self-Signed Certificate

Generate a self-signed certificate and dynamic config file to tell Traefik where the cert lives:

```bash
mkdir -p certs

# key + cert (valid for one year)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/local.key -out certs/local.crt \
  -subj "/CN=*.swarm.localhost"

# dynamic config that tells Traefik where the cert lives
cat > certs/tls.yml <<'EOF'
tls:
  certificates:
    - certFile: /certificates/local.crt
      keyFile:  /certificates/local.key
EOF
```

Create a Docker config for the certificate files:

```bash
docker config create swarm-cert.crt certs/local.crt
docker config create swarm-cert.key certs/local.key
docker config create swarm-tls.yml certs/tls.yml
```

Update your `docker-compose.yml` file with the following changes:

```yaml
# Add to the Traefik command section:
command:
  # ... existing commands ...
  - "--entryPoints.websecure.address=:443"
  - "--entryPoints.websecure.http.tls=true"
  - "--providers.file.directory=/etc/traefik/dynamic"
```

```yaml
# Add to the root of your docker-compose.yml file:
configs:
  swarm-cert.crt:
    file: ./certs/local.crt
  swarm-cert.key:
    file: ./certs/local.key
  swarm-tls.yml:
    file: ./certs/tls.yml
```

Deploy the stack:

```bash
docker stack deploy -c docker-compose.yml traefik
```

Your browser can access https://whoami.swarm.localhost/ for the service. You'll need to accept the security warning for the self-signed certificate.

## Next Steps

Now that you've mastered the basics of exposing services with Traefik on Docker Swarm, you're ready to explore more advanced features like middlewares, Let's Encrypt certificates, sticky sessions, and multi-layer routing.

Continue to the [Advanced Guide](advanced.md) to learn about:

- Adding middlewares for security and access control
- Generating certificates with Let's Encrypt
- Configuring sticky sessions for stateful applications
- Setting up multi-layer routing for authentication-based routing
