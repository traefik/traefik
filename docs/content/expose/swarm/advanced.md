# Exposing Services with Traefik on Docker Swarm - Advanced

This guide builds on the concepts and setup from the [Basic Guide](basic.md). Make sure you've completed the basic guide and have a working Traefik setup with Docker Swarm before proceeding.

In this advanced guide, you'll learn how to enhance your Traefik deployment with:

- **Middlewares** for security headers and access control
- **Let's Encrypt** for automated certificate management
- **Sticky sessions** for stateful applications
- **Multi-layer routing** for complex authentication scenarios

## Prerequisites

- Completed the [Basic Guide](basic.md)
- Docker Swarm cluster initialized
- Working Traefik setup from the basic guide

## Add Middlewares

Middlewares allow you to modify requests or responses as they pass through Traefik. Let's add two useful middlewares: [Headers](../../reference/routing-configuration/http/middlewares/headers.md) for security and [IP allowlisting](../../reference/routing-configuration/http/middlewares/ipallowlist.md) for access control.

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

For more advanced configuration options, see the [reference documentation](../../reference/routing-configuration/http/load-balancing/service.md).

## Multi-Layer Routing

Multi-layer routing enables hierarchical relationships between routers, where parent routers can process requests through middleware before child routers make final routing decisions. This is particularly useful for authentication-based routing or staged middleware application.

!!! info "Provider Requirement"
    Multi-layer routing requires the File provider, as Docker Swarm labels do not support the `parentRefs` field. However, you can use **both Docker Swarm and File providers together** - Swarm labels for service discovery and File configuration for multi-layer routing.

### Setup Multi-Layer Routing with Docker Swarm

To use multi-layer routing with Docker Swarm, you need to enable the File provider alongside the Docker provider.

Update your Traefik service in `docker-compose.yml`:

```yaml
services:
  traefik:
    image: traefik:v3.4
    command:
      - "--api.dashboard=true"
      - "--providers.docker.swarmMode=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.docker.network=traefik_proxy"
      - "--providers.file.directory=/etc/traefik/dynamic"  # Enable File provider
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--entrypoints.websecure.http.tls=true"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    configs:
      - source: mlr-config
        target: /etc/traefik/dynamic/mlr.yml
    networks:
      - traefik_proxy
    deploy:
      placement:
        constraints:
          - node.role == manager

configs:
  mlr-config:
    file: ./dynamic/mlr.yml

networks:
  traefik_proxy:
    external: true
```

### Authentication-Based Routing Example

Let's create a multi-layer routing setup where a parent router authenticates requests, and child routers direct traffic based on user roles.

First, keep your Docker Swarm services defined with labels as usual:

```yaml
# In docker-compose.yml
services:
  # ... traefik service from above ...

  # Admin backend service
  admin-backend:
    image: traefik/whoami
    networks:
      - traefik_proxy
    environment:
      - WHOAMI_NAME=Admin Backend
    deploy:
      replicas: 2
      labels:
        - "traefik.enable=true"
        - "traefik.http.services.admin-backend.loadbalancer.server.port=80"

  # User backend service
  user-backend:
    image: traefik/whoami
    networks:
      - traefik_proxy
    environment:
      - WHOAMI_NAME=User Backend
    deploy:
      replicas: 2
      labels:
        - "traefik.enable=true"
        - "traefik.http.services.user-backend.loadbalancer.server.port=80"
```

Now create the multi-layer routing configuration in a file. Create `dynamic/mlr.yml`:

```yaml
http:
  routers:
    # Parent router with authentication middleware
    api-parent:
      rule: "Host(`api.swarm.localhost`) && PathPrefix(`/api`)"
      middlewares:
        - auth-middleware
      entryPoints:
        - websecure
      # Note: No service and no TLS config - this is a parent router

    # Child router for admin users
    api-admin:
      rule: "HeadersRegexp(`X-Auth-User`, `admin`)"
      service: admin-backend@swarm  # Reference Swarm service
      parentRefs:
        - api-parent@file  # Explicit reference to parent in file provider

    # Child router for regular users
    api-user:
      rule: "HeadersRegexp(`X-Auth-User`, `user`)"
      service: user-backend@swarm  # Reference Swarm service
      parentRefs:
        - api-parent@file  # Explicit reference to parent in file provider

  middlewares:
    auth-middleware:
      basicAuth:
        users:
          - "admin:$apr1$DmXR3Add$wfdbGw6RWIhFb0ffXMM4d0"
          - "user:$apr1$GJtcIY1o$mSLdsWYeXpPHVsxGDqadI."
        headerField: X-Auth-User
```

!!! note "Generating Password Hashes"
    The password hashes above are generated using `htpasswd`. To create your own user credentials:

    ```bash
    # Using htpasswd (Apache utils)
    htpasswd -nb admin yourpassword
    ```

!!! important "Cross-Provider References"
    Notice the `@swarm` suffix on service names and the `@file` suffix in `parentRefs`. When using the File provider to orchestrate multi-layer routing with Swarm services:

    - Use `service-name@swarm` to reference Swarm services
    - Use `parent-name@file` in `parentRefs` to reference the parent router in the File provider

    The `@provider` suffix tells Traefik which provider namespace to look in for the resource.

Deploy the stack:

```bash
docker stack deploy -c docker-compose.yml traefik
```

### Test Multi-Layer Routing

Test the routing behavior:

```bash
# Request goes through parent router → auth middleware → admin child router
curl -k -u admin:test -H "Host: api.swarm.localhost" https://localhost/api
```

You should see the response from the admin-backend service when authenticating as `admin`. Try with `user:test` credentials to reach the user-backend service instead.

### How It Works

1. **Request arrives** at `api.swarm.localhost/api`
2. **Parent router** (`api-parent`) matches based on host and path
3. **BasicAuth middleware** authenticates the user and sets the `X-Auth-User` header with the username
4. **Child router** (`api-admin` or `api-user`) matches based on the header value
5. **Request forwarded** to the appropriate Swarm service

For more details about multi-layer routing, see the [Multi-Layer Routing documentation](../../reference/routing-configuration/http/routing/multi-layer-routing.md).

## Conclusion

In this advanced guide, you've learned how to:

- Add security with middlewares like secure headers and IP allow listing
- Automate certificate management with Let's Encrypt
- Implement sticky sessions for stateful applications
- Setup multi-layer routing for authentication-based routing

These advanced capabilities allow you to build production-ready Traefik deployments with Docker Swarm. Each of these can be further customized to meet your specific requirements.

### Next Steps

Now that you've mastered both basic and advanced Traefik features with Docker Swarm, you might want to explore:

- [Advanced routing options](../../reference/routing-configuration/http/routing/rules-and-priority.md) like query parameter matching, header-based routing, and more
- [Additional middlewares](../../reference/routing-configuration/http/middlewares/overview.md) for authentication, rate limiting, and request modifications
- [Observability features](../../reference/install-configuration/observability/metrics.md) for monitoring and debugging your Traefik deployment
- [TCP services](../../reference/routing-configuration/tcp/service.md) for exposing TCP services
- [UDP services](../../reference/routing-configuration/udp/service.md) for exposing UDP services
- [Docker provider documentation](../../reference/install-configuration/providers/docker.md) for more details about the Docker integration
