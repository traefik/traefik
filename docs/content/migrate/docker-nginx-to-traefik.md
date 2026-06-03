---
title: "Migrate from NGINX as your reverse proxy to Traefik - Docker Container Routing Guide"
description: "Step-by-step guide to migrate from NGINX to Traefik in your containerized setup"
---

# Docker Container Routing Migration Guide: Switching Your Reverse Proxy from NGINX to Traefik

Simplify reverse proxy in your containerized setup by replacing NGINX with Traefik.
{: .subtitle }

---

## What You Will Achieve

By completing this migration, you will move from NGINX reverse proxy to traefik.

### Setup Details
1. Automatically redirects all web requests to websecure(443).
2. Automatic certificate management (letsencrypt certificates).
3. Provides access logs and traefik logs.
4. Has a header middleware setup which can be extended or changed as per your need.
5. Can be extended to any number of services running on your machine.
6. An all container setup.

### Network Details (If you intend to use this structure)
```text
docker network create proxy_mother
proxy/compose.yml uses "proxy_mother" network
├── Traefik container (entry point on ports 80, 443)
├── Service 1 container (connected to proxy_mother)
├── Service 2 container (connected to proxy_mother)
└── Service N container (connected to proxy_mother)
```

```yaml tab="Your Existing NGINX based configuration"
# ~/services/proxy
services:
  proxy:
    image: nginx:latest
    container_name: proxy
    ports:
      - "80:80"
      - "443:443"
    restart: always
    networks:
      - mother
    volumes:
      - "./nginx/mime.types:/etc/nginx/mime.types"
      # Assuming you have a working nginx.conf file.
      - "./nginx/nginx.conf:/etc/nginx/nginx.conf"
      - "/etc/letsencrypt:/etc/letsencrypt:ro"
      - "./var/log/nginx:/var/log/nginx:rw"
networks:
  mother:
```

```yaml tab="Intended Traefik based configuration"
# ~/services/proxy
services:
  proxy:
    image: traefik:latest
    container_name: proxy
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      # Mount the docker socket so that traefik communicate with your services.	
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      # named volume mount for traefik to store the certificates.
      - traefik-certs:/letsencrypt
      # Logging (If required)
      - "./var/:/var/log/traefik/"
    command:
      # Providers
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.docker.network=proxy_mother"

      # EntryPoints
      - "--entrypoints.web.address=:80"
      - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
      - "--entrypoints.web.http.redirections.entrypoint.scheme=https"
      - "--entrypoints.web.http.redirections.entrypoint.permanent=true"
      - "--entrypoints.websecure.address=:443"
      - "--entrypoints.websecure.http.tls=true"

      # API & Dashboard
      - "--api.dashboard=true"
      # Observability & logging
      - "--log.level=INFO"
      - "--log.filePath=/var/log/traefik/traefik.log"
      - "--accesslog=true"
      - "--accesslog.filePath=/var/log/traefik/access.log"

        # Certificates
      - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=yourmail@example.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"

    # Traefik Dynamic configuration via Docker labels
    labels:
      # Enable self‑routing (Dashboard)
      - "traefik.enable=true"

      # Replace proxy.docker.invalid to a domain your own.
      - "traefik.http.routers.dashboard.rule=Host(`proxy.docker.invalid`)"
      - "traefik.http.routers.dashboard.entrypoints=websecure"
      - "traefik.http.routers.dashboard.service=api@internal"
      - "traefik.http.routers.dashboard.tls.certresolver=letsencrypt"
      - "traefik.http.routers.dashboard.tls=true"

      # Basic‑auth middleware
      - "traefik.http.middlewares.dashboard-auth.basicauth.users=<PASTE_YOUR_HASH_HERE>"
      # Example Response Header Middleware
      - "traefik.http.middlewares.headers-dashboard.headers.stsSeconds=31536000"
      - "traefik.http.middlewares.headers-dashboard.headers.stsIncludeSubdomains=true"
      - "traefik.http.middlewares.headers-dashboard.headers.stsPreload=true"
      - "traefik.http.middlewares.headers-dashboard.headers.frameDeny=true"
      - "traefik.http.middlewares.headers-dashboard.headers.browserXssFilter=true"
      - "traefik.http.middlewares.headers-dashboard.headers.contentTypeNosniff=true"
      - "traefik.http.middlewares.headers-dashboard.headers.referrerPolicy=no-referrer"

      # Use the middlewares in this router.
      - "traefik.http.routers.dashboard.middlewares=dashboard-auth@docker,headers-dashboard@docker"

    networks:
      - mother

volumes:
  traefik-certs:

networks:
  mother:
```

```yaml tab="Sample app service"
services:
  my_app:
    image: my_app:v1
    container_name: my_app
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.app.tls=true"
      - "traefik.http.routers.app.entrypoints=websecure"
      - "traefik.http.routers.app.tls=true" 
      - "traefik.http.routers.app.tls.certresolver=letsencrypt"
      - "traefik.http.services.app.loadbalancer.server.port=<PORT HERE>"
      - "traefik.http.routers.app.rule=Host(`app.domain.invalid`)"
      # Use the header middleware for dashboard as reference to write them for this docker service.

    networks:
      - proxy_mother		 	

networks:
    proxy_mother:
        external:true
```

---

## Prerequisites

!!! Assumptions "Docker based setup"

This page has been written keeping in mind a container based setup where
NGINX handles every request on port 80, 443 and proxy passes them to containers running
on the same docker network which provide different services.

!!! tip "Backup Recommendations"

    ```bash
    # Backup your NGINX compose files and config files.
    # Assuming you are inside the directory that has your NGINX config directory.
    cp -r proxy/ nginx_bak/
    ```
---

## Migration Strategy Overview

This migration achieves the shift from NGINX as your reverse proxy to traefik.

```text
Current:     Request → NGINX → Your Services

Final:       Request → Traefik → Your Services
```

**Migration Flow:**

- **Step 0** - Review Traefik Fundamentals.
- **Step 1** - Setup Traefik as your reverse proxy.
- **Step 2** - Stop NGINX and start Traefik.
- **Step 3** - Check your setup.

---

## Step 0: Review Traefik Fundamentals.
Unlike NGINX which uses configuration files, Traefik auto-discovers services via Docker labels and handles certificate management natively.
Learn how Traefik differs from NGINX.
Below are some pages which you can refer.

### Reference Pages

- [Traefik Architecture](../getting-started/configuration-overview.md)
- [Traefik Docker](../getting-started/docker.md)
- [Traefik TLS Options](../reference/routing-configuration/http/tls/tls-options.md)
- [Traefik Headers Middleware](../reference/routing-configuration/http/middlewares/headers.md)
- [Traefik EntryPoints configuration](../reference/install-configuration/entrypoints.md)

---

## Step 1: Setup Traefik as your reverse proxy.
1. Copy the compose.yml file given above.
2. !!! note "Change the networks and other user specific details."
3. Do not forget to generate the credentials for your dashboard.
    If you intend to just test it, add
        "--api.insecure=true" under command in your compose file.(Not to be used in production)
```bash
# Generate basic auth credentials
htpasswd -nb admin "P@ssw0rd" | sed -e 's/\$/\$\$/g'
# Output: admin:$$2y$05$...
# Use this in your docker labels

```
4. Configure your docker services to use traefik by adding labels to your compose file.
5. Use the Sample app compose file as reference.

---

## Step 2: Stop NGINX and start Traefik.
```bash
# Assuming your have performed a backup before rewriting contents at proxy/
cd services/nginx_bak;
docker compose down;

cd ../proxy
docker compose up -d;
# If you wish to see the logs
# docker compose logs -f
```

---

## Step 3: Check your setup

1. Login to your dashboard and view the http routers and middlewares. If you don't see them,
review your setup and troubleshoot
2. Check if traefik is able to pass requests to your service.

---

## Troubleshooting

1. There is a dashboard available in Traefik that can help to understand what's going on.
Refer to the [dedicated documentation](../reference/install-configuration/api-dashboard.md#configuration-example) to enable it.
2. Since we have enabled it, you can visit the dashboard on the domain you configured it to run; login and troubleshoot.
3. Check for logs at ./var (If you have setup logging, else see the compose file)

---

!!! success "Migration Complete"

    Congratulations! You have successfully migrated from NGINX to Traefik.

## Next Steps

**Learn More About Traefik:**

- [HTTP Middlewares](../reference/routing-configuration/http/middlewares/overview.md) - Learn more about HTTP middlewares (security headers, custom headers etc).
- [TLS Configuration](../reference/routing-configuration/http/tls/overview.md) - Advanced TLS and certificate management
- Explore [Traefik Middlewares](../reference/routing-configuration/http/middlewares/overview.md) for advanced traffic management.

**Enhance Your Setup:**

- Learn more about [metrics](../reference/install-configuration/observability/metrics.md) and [tracing](../reference/install-configuration/observability/tracing.md).
- Configure [access logs](../reference/install-configuration/observability/logs-and-accesslogs.md) for observability
- Consider [Traefik Hub](https://traefik.io/traefik-hub/) for enterprise features like AI & API Gateway, API Management, and advanced security.

---

## Found an Issue?
**See a typo, outdated/incorrect information, or have a suggestion?** [Edit this page on GitHub](https://github.com/traefik/traefik/blob/master/docs/content/migrate/nginx-to-traefik.md)

---

## Feedback and Support

If you encounter issues during migration or have suggestions for improving this guide:

- **Report Issues:** [GitHub Issues](https://github.com/traefik/traefik/issues)
- **Community Support:** [Traefik Community Forum](https://community.traefik.io/)
- **Enterprise Support:** [Traefik Labs Commercial Support](https://traefik.io/pricing/)

We welcome contributions to improve this migration guide. See our [contribution guidelines](../contributing/submitting-pull-requests.md) to get started.
