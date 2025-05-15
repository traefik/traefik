# Exposing Services with Traefik on Docker Swarm

This guide will help you expose your services securely through Traefik Proxy using Docker Swarm. We'll cover routing HTTP and HTTPS traffic, implementing TLS, adding middlewares, Let's Encrypt integration, and sticky sessions.

## Prerequisites

- Docker Swarm cluster initialized
- Basic understanding of Docker Swarm concepts
- Traefik deployed using the Traefik Docker Swarm Setup guide

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

Now we'll enhance our routing by directing traffic to different services based on [URL paths](../reference/routing-configuration/http/router/rules-and-priority.md#path-pathprefix-and-pathregexp). This is useful for API versioning, frontend/backend separation, or organizing microservices.

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

## Add Middlewares

Middlewares allow you to modify requests or responses as they pass through Traefik. Let's add two useful middlewares: [Headers](../reference/routing-configuration/http/middlewares/headers.md) for security and [IP allowlisting](../reference/routing-configuration/http/middlewares/ipallowlist.md) for access control.

Add the following labels to your whoami service deployment section in `docker-compose.yml`:

```yaml
deploy:
  # ... existing configuration ...
  labels:
    # ... existing labels ...
    
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
    
    # Apply the middlewares
    - "traefik.http.routers.whoami.middlewares=secure-headers,ip-allowlist"
```

Add the same middleware to your whoami-api service:

```yaml
deploy:
  # ... existing configuration ...
  labels:
    # ... existing labels ...
    - "traefik.http.routers.whoami-api.middlewares=secure-headers,ip-allowlist"
```

Apply the changes:

```bash
docker stack deploy -c docker-compose.yml traefik
```

### Test the Middlewares

Now let's verify that our middlewares are working correctly:

Test the Secure Headers middleware:

```bash
curl -k -I -H "Host: whoami.swarm.localhost" https://localhost/
```

In the response headers, you should see security headers set by the middleware:

- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` with the appropriate settings

Test the IP Allowlist middleware:

If your request comes from an IP that's in the allow list (e.g., 127.0.0.1), it should succeed:

```bash
curl -k -I -H "Host: whoami.swarm.localhost" https://localhost/
```

If you try to access from an IP not in the allow list, the request will be rejected with a `403` Forbidden response. To simulate this in a local environment, you can modify the middleware configuration temporarily to exclude your IP address, then test again.

## Generate Certificates with Let's Encrypt

Let's Encrypt provides free, automated TLS certificates. Let's configure Traefik to automatically obtain and renew certificates for our services.

Instead of using self-signed certificates, update your existing `docker-compose.yml` file with the following changes:

Add the Let's Encrypt certificate resolver to the Traefik service command section:

```yaml
command:
  # ... existing commands ...
  # Let's Encrypt configuration
  - "--certificatesresolvers.le.acme.email=your-email@example.com" # replace with your actual email
  - "--certificatesresolvers.le.acme.storage=/letsencrypt/acme.json"
  - "--certificatesresolvers.le.acme.httpchallenge.entrypoint=web"
```

Add a volume for Let's Encrypt certificates:

```yaml
volumes:
  # ...Existing volumes...
  - letsencrypt:/letsencrypt
```

Update your service labels to use the certificate resolver:

```yaml
labels:
  # ... existing labels ...
  - "traefik.http.routers.whoami.tls.certresolver=le"
```

Do the same for any other services you want to secure:

```yaml
labels:
  # ... existing labels ...
  - "traefik.http.routers.whoami-api.tls.certresolver=le"
```

Create a named volume for storing Let's Encrypt certificates by adding to the volumes section:

```yaml
volumes:
  # ... existing volumes ...
  letsencrypt:
    driver: local
```

Apply the changes:

```bash
docker stack deploy -c docker-compose.yml traefik
```

!!! important "Public DNS Required"
    Let's Encrypt may require a publicly accessible domain to validate domain ownership. For testing with local domains like `whoami.swarm.localhost`, the certificate will remain self-signed. In production, replace it with a real domain that has a publicly accessible DNS record pointing to your Traefik instance.

Once the certificate is issued, you can verify it:

```bash
# Verify the certificate chain
curl -v https://whoami.swarm.localhost/ 2>&1 | grep -i "server certificate"
```

You should see that your certificate is issued by Let's Encrypt.

## Configure Sticky Sessions

Sticky sessions ensure that a user's requests always go to the same backend server, which is essential for applications that maintain session state. Let's implement sticky sessions for our whoami service.

Docker Swarm already has multiple replicas running; we'll now add sticky session configuration. Update your whoami service in the `docker-compose.yml` file:

### Add Sticky Session Configuration

Add the following labels to your whoami service in the `docker-compose.yml` file:

```yaml
deploy:
  # ... existing configuration ...
  labels:
    # ... existing labels ...
    
    # Sticky Sessions Configuration
    - "traefik.http.services.whoami.loadbalancer.sticky.cookie=true"
    - "traefik.http.services.whoami.loadbalancer.sticky.cookie.name=sticky_cookie"
    - "traefik.http.services.whoami.loadbalancer.sticky.cookie.secure=true"
    - "traefik.http.services.whoami.loadbalancer.sticky.cookie.httpOnly=true"
```

Apply the changes:

```bash
docker stack deploy -c docker-compose.yml traefik
```

### Test Sticky Sessions

You can test the sticky sessions by making multiple requests and observing that they all go to the same backend container:

```bash
# First request - save cookies to a file
curl -k -c cookies.txt -H "Host: whoami.swarm.localhost" https://localhost/

# Subsequent requests - use the cookies
curl -k -b cookies.txt -H "Host: whoami.swarm.localhost" https://localhost/
curl -k -b cookies.txt -H "Host: whoami.swarm.localhost" https://localhost/
```

Pay attention to the `Hostname` field in each response - it should remain the same across all requests when using the cookie file, confirming that sticky sessions are working.

For comparison, try making requests without the cookie:

```bash
# Requests without cookies should be load-balanced across different containers
curl -k -H "Host: whoami.swarm.localhost" https://localhost/
curl -k -H "Host: whoami.swarm.localhost" https://localhost/
```

You should see different `Hostname` values in these responses, as each request is load-balanced to a different container.

!!! important "Browser Testing"
    When testing in browsers, you need to use the same browser session to maintain the cookie. The cookie is set with `httpOnly` and `secure` flags for security, so it will only be sent over HTTPS connections and won't be accessible via JavaScript.

For more advanced configuration options, see the [reference documentation](../reference/routing-configuration/http/load-balancing/service.md).

## Conclusion

In this guide, you've learned how to:

- Expose HTTP services through Traefik in Docker Swarm
- Set up path-based routing to direct traffic to different backend services
- Secure your services with TLS using self-signed certificates
- Add security with middlewares like secure headers and IP allow listing
- Automate certificate management with Let's Encrypt
- Implement sticky sessions for stateful applications

These fundamental capabilities provide a solid foundation for exposing any application through Traefik Proxy in Docker Swarm. Each of these can be further customized to meet your specific requirements.

### Next Steps

Now that you understand the basics of exposing services with Traefik Proxy, you might want to explore:

- [Advanced routing options](../reference/routing-configuration/http/router/rules-and-priority.md) like query parameter matching, header-based routing, and more
- [Additional middlewares](../reference/routing-configuration/http/middlewares/overview.md) for authentication, rate limiting, and request modifications
- [Observability features](../reference/install-configuration/observability/metrics.md) for monitoring and debugging your Traefik deployment
- [TCP services](../reference/routing-configuration/tcp/service.md) for exposing TCP services
- [UDP services](../reference/routing-configuration/udp/service.md) for exposing UDP services
- [Docker provider documentation](../reference/install-configuration/providers/docker.md) for more details about the Docker integration
