---
title: "Multi-Layer Routing with Authentication Tutorial"
description: "Learn how to implement authentication-based routing using Traefik's multi-layer routing feature. A practical tutorial with complete examples."
---

# Multi-Layer Routing with Authentication

Build an Authentication Gateway with Role-Based Routing
{: .subtitle }

## Overview

This tutorial demonstrates how to use Traefik's multi-layer routing feature to build an authentication gateway that:
1. Authenticates incoming requests using ForwardAuth middleware
2. Injects user context (roles) as headers
3. Routes requests to different backend services based on those roles

This pattern is particularly useful for:
- Multi-tenant applications with role-based access control
- API gateways with authentication and authorization
- Microservices architectures with centralized authentication

## Architecture

```
Request: GET /whoami + Bearer token
│
├── Parent Router: PathPrefix(`/whoami`)
│   ├── ForwardAuth Middleware → Auth Service
│   │   ├── Validates Bearer token
│   │   └── Adds X-User-Role header (admin|developer)
│   │
│   └── Child Muxer (receives modified request with X-User-Role header)
│       ├── Admin Router: HeadersRegexp(`X-User-Role`, `admin`) → admin-service
│       └── Developer Router: HeadersRegexp(`X-User-Role`, `developer`) → developer-service
```

## Components

This tutorial uses the following components:

- **Traefik**: Main reverse proxy with multi-layer routing
- **Auth Service**: Authentication service that validates tokens and adds role headers
- **Admin Service**: Backend service for admin users
- **Developer Service**: Backend service for developer users

## Example 1: File Provider Configuration

This example uses Traefik's File provider for easy local testing.

### Directory Structure

```
.
├── traefik.yml           # Static configuration
├── dynamic.yml           # Dynamic configuration with multi-layer routing
├── auth-service/         # Simple authentication service
│   └── main.go
└── docker-compose.yml    # Complete environment
```

### Static Configuration

```yaml title="traefik.yml"
## Static configuration
entryPoints:
  web:
    address: ":80"
  traefik:
    address: ":8080"

api:
  dashboard: true
  insecure: true

providers:
  file:
    filename: /etc/traefik/dynamic.yml
    watch: true

log:
  level: INFO

accessLog:
  format: json
  fields:
    headers:
      names:
        X-User-Role: keep
        X-User-Name: keep
```

### Dynamic Configuration

```yaml title="dynamic.yml"
## Dynamic configuration
http:
  routers:
    # Parent router (root) with authentication
    whoami-parent:
      rule: "PathPrefix(`/whoami`)"
      entryPoints:
        - web
      middlewares:
        - auth-middleware
      # No service - this is a parent router

    # Child router for admin users
    whoami-admin:
      rule: "HeadersRegexp(`X-User-Role`, `admin`)"
      service: admin-service
      parentRefs:
        - whoami-parent

    # Child router for developer users
    whoami-developer:
      rule: "HeadersRegexp(`X-User-Role`, `developer`)"
      service: developer-service
      parentRefs:
        - whoami-parent

  middlewares:
    auth-middleware:
      forwardAuth:
        address: "http://auth-service:8080/auth"
        authResponseHeaders:
          - X-User-Role
          - X-User-Name

  services:
    admin-service:
      loadBalancer:
        servers:
          - url: "http://admin-service:80"

    developer-service:
      loadBalancer:
        servers:
          - url: "http://developer-service:80"
```

### Authentication Service

For this tutorial, you'll need an authentication service that implements the ForwardAuth protocol. The service should:

- Listen on port 8080 at the `/auth` endpoint
- Accept incoming authentication requests from Traefik
- Validate the `Authorization` header (Bearer token format)
- Return appropriate responses:
  - **401 Unauthorized**: For missing or invalid tokens
  - **200 OK**: For valid tokens, with response headers:
    - `X-User-Role`: User's role (`admin` or `developer`)
    - `X-User-Name`: User identifier

For testing purposes, the service can use these token-to-role mappings:
- `bob-token` → `admin` role
- `jack-token` → `developer` role

You can implement this service in any language or use an existing authentication solution that supports the ForwardAuth pattern.

### Docker Compose

!!! note "Authentication Service"

    Replace `your-auth-service:latest` with your authentication service image that implements the ForwardAuth protocol as described above. You can also use existing solutions like Authelia, OAuth2 Proxy, or Keycloak.

```yaml title="docker-compose.yml"
version: '3.8'

services:
  traefik:
    image: traefik:v3.5
    command:
      - "--configFile=/etc/traefik/traefik.yml"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./dynamic.yml:/etc/traefik/dynamic.yml:ro
    networks:
      - app

  auth-service:
    image: your-auth-service:latest
    networks:
      - app

  admin-service:
    image: traefik/whoami
    environment:
      - WHOAMI_NAME=AdminService
    networks:
      - app

  developer-service:
    image: traefik/whoami
    environment:
      - WHOAMI_NAME=DeveloperService
    networks:
      - app

networks:
  app:
```

### Testing

Start the environment:

```bash
docker-compose up -d
```

Test with admin user:

```bash
curl -H "Authorization: Bearer bob-token" http://localhost/whoami
```

Expected response:

```
Hostname: admin-service-xxxx
Name: AdminService
...
```

Test with developer user:

```bash
curl -H "Authorization: Bearer jack-token" http://localhost/whoami
```

Expected response:

```
Hostname: developer-service-xxxx
Name: DeveloperService
...
```

Test with invalid token:

```bash
curl -H "Authorization: Bearer invalid" http://localhost/whoami
```

Expected response:

```
401 Unauthorized
```

## Example 2: Kubernetes CRD Configuration

This example shows how to implement the same authentication-based routing in Kubernetes using IngressRoute CRDs.

### Prerequisites

- Kubernetes cluster (minikube, kind, or any cluster)
- Traefik installed with Kubernetes CRD provider enabled
- `kubectl` configured

### Authentication Service Deployment

Deploy your authentication service that implements the ForwardAuth protocol as described in Example 1.

```yaml title="auth-service.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
        - name: auth-service
          # Replace with your authentication service image
          # The service must implement the ForwardAuth protocol:
          # - Listen on port 8080 at /auth endpoint
          # - Validate Authorization header
          # - Return X-User-Role and X-User-Name headers on success
          image: your-auth-service:latest
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: default
spec:
  selector:
    app: auth-service
  ports:
    - port: 8080
      targetPort: 8080
```

!!! tip "Authentication Service Options"

    You can use:
    - Your own custom authentication service
    - Existing solutions like Authelia, OAuth2 Proxy, Keycloak
    - Any service that implements the ForwardAuth protocol

### Backend Services

```yaml title="backend-services.yaml"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: admin-service
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: admin-service
  template:
    metadata:
      labels:
        app: admin-service
    spec:
      containers:
        - name: whoami
          image: traefik/whoami
          env:
            - name: WHOAMI_NAME
              value: "AdminService"
---
apiVersion: v1
kind: Service
metadata:
  name: admin-service
  namespace: default
spec:
  selector:
    app: admin-service
  ports:
    - port: 80
      targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: developer-service
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: developer-service
  template:
    metadata:
      labels:
        app: developer-service
    spec:
      containers:
        - name: whoami
          image: traefik/whoami
          env:
            - name: WHOAMI_NAME
              value: "DeveloperService"
---
apiVersion: v1
kind: Service
metadata:
  name: developer-service
  namespace: default
spec:
  selector:
    app: developer-service
  ports:
    - port: 80
      targetPort: 80
```

### IngressRoutes

```yaml title="ingressroutes.yaml"
# ForwardAuth Middleware
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: auth-middleware
  namespace: default
spec:
  forwardAuth:
    address: "http://auth-service.default.svc.cluster.local:8080/auth"
    authResponseHeaders:
      - X-User-Role
      - X-User-Name
---
# Parent IngressRoute with authentication
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami-parent
  namespace: default
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`localhost`) && PathPrefix(`/whoami`)
      kind: Rule
      middlewares:
        - name: auth-middleware
---
# Child IngressRoute for admin users
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami-admin
  namespace: default
spec:
  parentRefs:
    - name: whoami-parent
  routes:
    - match: HeadersRegexp(`X-User-Role`, `admin`)
      kind: Rule
      services:
        - name: admin-service
          port: 80
---
# Child IngressRoute for developer users
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami-developer
  namespace: default
spec:
  parentRefs:
    - name: whoami-parent
  routes:
    - match: HeadersRegexp(`X-User-Role`, `developer`)
      kind: Rule
      services:
        - name: developer-service
          port: 80
```

### Deploy and Test

Deploy all resources:

```bash
kubectl apply -f auth-service.yaml
kubectl apply -f backend-services.yaml
kubectl apply -f ingressroutes.yaml
```

Wait for pods to be ready:

```bash
kubectl get pods -w
```

Get Traefik service endpoint (adjust based on your cluster):

```bash
# For minikube
minikube service traefik --url

# For LoadBalancer service
kubectl get svc traefik -n traefik
```

Test with admin user:

```bash
curl -H "Host: localhost" -H "Authorization: Bearer bob-token" http://<TRAEFIK_IP>/whoami
```

Test with developer user:

```bash
curl -H "Host: localhost" -H "Authorization: Bearer jack-token" http://<TRAEFIK_IP>/whoami
```

## Observability and Debugging

### View Traefik Dashboard

Access the Traefik dashboard to see routers and their relationships:

```
http://localhost:8080/dashboard/
```

You should see:
- `whoami-parent` router with no service
- `whoami-admin` router with `parentRefs` pointing to parent
- `whoami-developer` router with `parentRefs` pointing to parent

### Enable Distributed Tracing

Add to Traefik static configuration:

```yaml
tracing:
  serviceName: traefik
  otlp:
    http:
      endpoint: http://tempo:4318
```

Traces will show the request flow:
```
Trace: Request to /whoami
├── Span: entrypoint.web
    ├── Span: router.whoami-parent
    │   ├── Span: middleware.auth-middleware
    │   └── Span: router.whoami-admin  (or whoami-developer)
    │       └── Span: service.admin-service
```

### View Access Logs

For File provider (docker-compose):

```bash
docker-compose logs traefik | grep "X-User-Role"
```

For Kubernetes:

```bash
kubectl logs -n traefik deployment/traefik | grep "X-User-Role"
```

### Common Issues

**401 Unauthorized for valid tokens:**
- Check auth-service logs: `docker-compose logs auth-service` or `kubectl logs deployment/auth-service`
- Verify Authorization header format: `Bearer <token>`
- Ensure auth-service is reachable from Traefik

**404 Not Found:**
- Verify parent router rule matches the request path
- Check Traefik dashboard for router configuration
- Ensure request includes required Host header (for Kubernetes)

**Request goes to wrong service:**
- Check auth-service is setting correct `X-User-Role` header
- Verify child router rules match the role values
- Enable debug logging: `--log.level=DEBUG`

## Advanced Patterns

### Multiple Authentication Methods

Use different parent routers for different authentication methods:

```yaml
http:
  routers:
    # JWT authentication parent
    api-jwt-parent:
      rule: "PathPrefix(`/api/jwt`)"
      middlewares: [jwt-auth]

    # Basic auth parent
    api-basic-parent:
      rule: "PathPrefix(`/api/basic`)"
      middlewares: [basic-auth]

    # Shared child router
    api-admin:
      rule: "HeadersRegexp(`X-User-Role`, `admin`)"
      service: admin-service
      parentRefs:
        - api-jwt-parent
        - api-basic-parent
```

### Tenant Isolation

Use parent router to extract tenant ID, child routers for per-tenant services:

```yaml
http:
  routers:
    # Parent extracts tenant
    tenant-parent:
      rule: "HostRegexp(`^(?P<tenant>[a-z]+)\\.example\\.com$`)"
      middlewares: [tenant-extractor]  # Adds X-Tenant-ID header

    # Child routes to tenant-specific services
    tenant-api:
      rule: "PathPrefix(`/api`)"
      service: api-service  # Uses X-Tenant-ID for routing
      parentRefs: [tenant-parent]
```

### Progressive Rate Limiting

Apply different rate limits at different levels:

```yaml
http:
  routers:
    # Parent: Global rate limit
    api-parent:
      rule: "PathPrefix(`/api`)"
      middlewares: [global-rate-limit]  # 1000 req/s

    # Child: Per-endpoint limits
    api-expensive:
      rule: "Path(`/api/expensive`)"
      middlewares: [strict-rate-limit]  # 10 req/s
      service: expensive-service
      parentRefs: [api-parent]
```

## Best Practices

!!! tip "Security"

    - Always validate tokens in the auth service
    - Use HTTPS in production (TLS termination at parent router)
    - Implement token expiration and refresh logic
    - Log authentication failures for security monitoring

!!! tip "Performance"

    - Cache authentication results when possible
    - Use connection pooling for auth service
    - Monitor auth service latency - it's in the critical path
    - Consider async token validation for high-throughput scenarios

!!! tip "Maintainability"

    - Keep parent router rules simple and broad
    - Use descriptive names for routers and services
    - Document what headers each middleware adds
    - Test auth service independently before integration

## Summary

Multi-layer routing with authentication provides a powerful pattern for:
- Centralizing authentication logic
- Implementing role-based access control
- Building flexible API gateways
- Separating concerns between authentication and routing

The key benefits are:
- **Clean separation**: Authentication logic separate from business routing
- **Reusability**: One authentication layer, multiple routing strategies
- **Flexibility**: Easy to add new roles or routes without changing auth logic
- **Observability**: Clear visibility into authentication and routing decisions

For more information, see:
- [Multi-Layer Routing Concepts](../routing/multi-layer-routing.md)
- [ForwardAuth Middleware](../middlewares/http/forwardauth.md)
- [Kubernetes CRD Provider](../providers/kubernetes-crd.md)

{!traefik-for-business-applications.md!}