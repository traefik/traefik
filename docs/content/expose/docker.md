# Exposing Services with Traefik on Docker

This guide will help you expose your services securely through Traefik Proxy using Docker. We'll cover routing HTTP and HTTPS traffic, implementing TLS, adding middlewares, Let's Encrypt integration, and sticky sessions.

## Prerequisites

- Docker and Docker Compose installed
- Basic understanding of Docker concepts
- Traefik deployed using the Traefik Docker Setup guide

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

Now we'll enhance our routing by directing traffic to different services based on [URL paths](../reference/routing-configuration/http/router/rules-and-priority.md#path-pathprefix-and-pathregexp). This is useful for API versioning, frontend/backend separation, or organizing microservices.

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

## Add Middlewares

Middlewares allow you to modify requests or responses as they pass through Traefik. Let's add two useful middlewares: [Headers](../reference/routing-configuration/http/middlewares/headers.md) for security and [IP allowlisting](../reference/routing-configuration/http/middlewares/ipallowlist.md) for access control.

Add the following labels to your whoami service in `docker-compose.yml`:

```yaml
labels:
  
  # Secure Headers Middleware
  - "traefik.http.middlewares.secure-headers.headers.frameDeny=true"
  - "traefik.http.middlewares.secure-headers.headers.sslRedirect=true"
  - "traefik.http.middlewares.secure-headers.headers.browserXssFilter=true"
  - "traefik.http.middlewares.secure-headers.headers.contentTypeNosniff=true"
  - "traefik.http.middlewares.secure-headers.headers.stsIncludeSubdomains=true"
  - "traefik.http.middlewares.secure-headers.headers.stsPreload=true"
  - "traefik.http.middlewares.secure-headers.headers.stsSeconds=31536000"
  
  # IP Allowlist Middleware
  - "traefik.http.middlewares.ip-allowlist.ipallowlist.sourceRange=127.0.0.1/32,192.168.0.0/16,10.0.0.0/8"
```

Add the same middleware to your whoami-api service:

```yaml
labels:
  - "traefik.http.routers.whoami-api.middlewares=secure-headers,ip-allowlist"
```

Apply the changes:

```bash
docker compose up -d
```

### Test the Middlewares

Now let's verify that our middlewares are working correctly:

Test the Secure Headers middleware:

```bash
curl -k -I -H "Host: whoami.docker.localhost" https://localhost/
```

In the response headers, you should see security headers set by the middleware:

- `X-Frame-Options: DENY` 
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` with the appropriate settings

Test the IP Allowlist middleware:

If your request comes from an IP that's in the allow list (e.g., 127.0.0.1), it should succeed:

```bash
curl -k -I -H "Host: whoami.docker.localhost" https://localhost/
```

If you try to access from an IP not in the allow list, the request will be rejected with a `403` Forbidden response. To simulate this in a local environment, you can modify the middleware configuration temporarily to exclude your IP address, then test again.

## Generate Certificates with Let's Encrypt

Let's Encrypt provides free, automated TLS certificates. Let's configure Traefik to automatically obtain and renew certificates for our services.

Instead of using self-signed certificates, update your existing `docker-compose.yml` file with the following changes:

Add the Let's Encrypt certificate resolver to the Traefik service command section:

```yaml
command:
  - "--api.insecure=false"
  - "--api.dashboard=true"
  - "--providers.docker=true"
  - "--providers.docker.exposedbydefault=false"
  - "--providers.docker.network=proxy"
  - "--entryPoints.web.address=:80"
  - "--entryPoints.websecure.address=:443"
  - "--entryPoints.websecure.http.tls=true"
  - "--entryPoints.web.http.redirections.entryPoint.to=websecure"
  - "--entryPoints.web.http.redirections.entryPoint.scheme=https"
  # Let's Encrypt configuration
  - "--certificatesresolvers.le.acme.email=your-email@example.com" # replace with your actual email
  - "--certificatesresolvers.le.acme.storage=/letsencrypt/acme.json"
  - "--certificatesresolvers.le.acme.httpchallenge.entrypoint=web"
```

Add a volume for Let's Encrypt certificates:

```yaml
volumes:
  # ...Existing volumes...
  - "./letsencrypt:/letsencrypt"
```

Update your service labels to use the certificate resolver:

```yaml
labels:
  - "traefik.http.routers.whoami.tls.certresolver=le"
```

Do the same for any other services you want to secure:

```yaml
labels:
  - "traefik.http.routers.whoami-api.tls.certresolver=le"
```

Create a directory for storing Let's Encrypt certificates:

```bash
mkdir -p letsencrypt
```

Apply the changes:

```bash
docker compose up -d
```

!!! important "Public DNS Required"
    Let's Encrypt may require a publicly accessible domain to validate domain ownership. For testing with local domains like `whoami.docker.localhost`, the certificate will remain self-signed. In production, replace it with a real domain that has a publicly accessible DNS record pointing to your Traefik instance.

Once the certificate is issued, you can verify it:

```bash
# Verify the certificate chain
curl -v https://whoami.docker.localhost/ 2>&1 | grep -i "server certificate"
```

You should see that your certificate is issued by Let's Encrypt.

## Configure Sticky Sessions

Sticky sessions ensure that a user's requests always go to the same backend server, which is essential for applications that maintain session state. Let's implement sticky sessions for our whoami service.

### First, Add Sticky Session Labels

Add the following labels to your whoami service in the `docker-compose.yml` file:

```yaml
labels:
  - "traefik.http.services.whoami.loadbalancer.sticky.cookie=true"
  - "traefik.http.services.whoami.loadbalancer.sticky.cookie.name=sticky_cookie"
  - "traefik.http.services.whoami.loadbalancer.sticky.cookie.secure=true"
  - "traefik.http.services.whoami.loadbalancer.sticky.cookie.httpOnly=true"
```

Apply the changes:

```bash
docker compose up -d
```

### Then, Scale Up the Service

To demonstrate sticky sessions with Docker, use Docker Compose's scale feature:

```bash
docker compose up -d --scale whoami=3
```

This creates multiple instances of the whoami service.

!!! important "Scaling After Configuration Changes"
    If you run `docker compose up -d` after scaling, it will reset the number of whoami instances back to 1. Always scale after applying configuration changes and starting the services.

### Test Sticky Sessions

You can test the sticky sessions by making multiple requests and observing that they all go to the same backend container:

```bash
# First request - save cookies to a file
curl -k -c cookies.txt -H "Host: whoami.docker.localhost" https://localhost/

# Subsequent requests - use the cookies
curl -k -b cookies.txt -H "Host: whoami.docker.localhost" https://localhost/
curl -k -b cookies.txt -H "Host: whoami.docker.localhost" https://localhost/
```

Pay attention to the `Hostname` field in each response - it should remain the same across all requests when using the cookie file, confirming that sticky sessions are working.

For comparison, try making requests without the cookie:

```bash
# Requests without cookies should be load-balanced across different containers
curl -k -H "Host: whoami.docker.localhost" https://localhost/
curl -k -H "Host: whoami.docker.localhost" https://localhost/
```

You should see different `Hostname` values in these responses, as each request is load-balanced to a different container.

!!! important "Browser Testing"
    When testing in browsers, you need to use the same browser session to maintain the cookie. The cookie is set with `httpOnly` and `secure` flags for security, so it will only be sent over HTTPS connections and won't be accessible via JavaScript.

For more advanced configuration options, see the [reference documentation](../reference/routing-configuration/http/load-balancing/service.md).

## Conclusion

In this guide, you've learned how to:

- Expose HTTP services through Traefik in Docker
- Set up path-based routing to direct traffic to different backend services
- Secure your services with TLS using self-signed certificates
- Add security with middlewares like secure headers and IP allow listing
- Automate certificate management with Let's Encrypt
- Implement sticky sessions for stateful applications

These fundamental capabilities provide a solid foundation for exposing any application through Traefik Proxy in Docker. Each of these can be further customized to meet your specific requirements.

### Next Steps

Now that you understand the basics of exposing services with Traefik Proxy, you might want to explore:

- [Advanced routing options](../reference/routing-configuration/http/router/rules-and-priority.md) like query parameter matching, header-based routing, and more
- [Additional middlewares](../reference/routing-configuration/http/middlewares/overview.md) for authentication, rate limiting, and request modifications
- [Observability features](../reference/install-configuration/observability/metrics.md) for monitoring and debugging your Traefik deployment
- [TCP services](../reference/routing-configuration/tcp/service.md) for exposing TCP services
- [UDP services](../reference/routing-configuration/udp/service.md) for exposing UDP services
- [Docker provider documentation](../reference/install-configuration/providers/docker.md) for more details about the Docker integration
