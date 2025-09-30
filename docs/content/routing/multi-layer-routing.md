---
title: "Multi-Layer Routing"
description: "Learn how to use Traefik's multi-layer routing to create hierarchical router relationships where parent routers can apply middleware before child routers make routing decisions."
---

# Multi-Layer Routing

Hierarchical Router Relationships for Advanced Routing Scenarios
{: .subtitle }

## Overview

Multi-layer routing enables you to create hierarchical relationships between routers, where parent routers can process requests through middleware before child routers make final routing decisions.

This feature allows middleware at the parent level to modify requests (adding headers, performing authentication, etc.) that influence how child routers evaluate their rules and route traffic to services.

## Use Cases

Multi-layer routing is particularly useful for:

- **Authentication-Based Routing**: Parent router authenticates requests and adds user context (roles, permissions) as headers, child routers route based on these headers
- **Staged Middleware Application**: Apply common middleware (rate limiting, CORS) at parent level, specific middleware at child level
- **Tenant Isolation**: Parent router identifies tenant from domain/path, adds tenant header, child routers route to tenant-specific services
- **Progressive Request Enrichment**: Each layer adds context to the request, enabling increasingly specific routing decisions

## How It Works

```
Request → EntryPoint → Parent Router → Middleware → Child Muxer → Child Router → Service
                                          ↓
                                  Modified Request
                                  (e.g., added headers)
```

1. **Request arrives** at an entrypoint
2. **Parent router matches** based on its rule (e.g., `PathPrefix(/api)`)
3. **Parent middleware executes**, potentially modifying the request
4. **Child muxer evaluates** the modified request against child router rules
5. **Child router matches** based on its rule (which may use modified request attributes)
6. **Request is forwarded** to the child router's service

## Key Concepts

### Parent Routers

Parent routers:
- Have one or more child routers (via `childRefs` computed at runtime)
- Apply middleware to modify requests before child evaluation
- **Must not** have a `service` defined
- **Must not** have `tls` or `observability` configuration

### Child Routers

Child routers:
- Reference their parent router(s) via `parentRefs`
- Evaluate rules on the potentially modified request
- **Must** have a `service` defined
- **Must not** have `tls` or `observability` configuration

### Root Routers

Root routers:
- Have no `parentRefs` (top of the hierarchy)
- **Can** have `tls`, `observability`, and `entryPoints` configuration
- Can be either parent routers (with children) or standalone routers (with service)

## Configuration Example

### File Provider

??? example "Authentication-Based Routing with File Provider"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        # Parent router with authentication
        api-parent:
          rule: "PathPrefix(`/api`)"
          middlewares:
            - auth-middleware
          entryPoints:
            - websecure
          tls: {}
          # Note: No service defined - this is a parent router

        # Child router for admin users
        api-admin:
          rule: "HeadersRegexp(`X-User-Role`, `admin`)"
          service: admin-service
          parentRefs:
            - api-parent

        # Child router for regular users
        api-user:
          rule: "HeadersRegexp(`X-User-Role`, `user`)"
          service: user-service
          parentRefs:
            - api-parent

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
              - url: "http://admin-backend:8080"

        user-service:
          loadBalancer:
            servers:
              - url: "http://user-backend:8080"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      # Parent router with authentication
      [http.routers.api-parent]
        rule = "PathPrefix(`/api`)"
        middlewares = ["auth-middleware"]
        entryPoints = ["websecure"]
        [http.routers.api-parent.tls]
        # Note: No service defined - this is a parent router

      # Child router for admin users
      [http.routers.api-admin]
        rule = "HeadersRegexp(`X-User-Role`, `admin`)"
        service = "admin-service"
        parentRefs = ["api-parent"]

      # Child router for regular users
      [http.routers.api-user]
        rule = "HeadersRegexp(`X-User-Role`, `user`)"
        service = "user-service"
        parentRefs = ["api-parent"]

    [http.middlewares]
      [http.middlewares.auth-middleware.forwardAuth]
        address = "http://auth-service:8080/auth"
        authResponseHeaders = ["X-User-Role", "X-User-Name"]

    [http.services]
      [http.services.admin-service.loadBalancer]
        [[http.services.admin-service.loadBalancer.servers]]
          url = "http://admin-backend:8080"

      [http.services.user-service.loadBalancer]
        [[http.services.user-service.loadBalancer.servers]]
          url = "http://user-backend:8080"
    ```

    **How it works:**

    1. Request to `/api/endpoint` matches `api-parent` router
    2. `auth-middleware` (ForwardAuth) validates the request and adds `X-User-Role` header
    3. Modified request is evaluated by child routers
    4. If `X-User-Role: admin`, `api-admin` router matches and forwards to `admin-service`
    5. If `X-User-Role: user`, `api-user` router matches and forwards to `user-service`

### Kubernetes CRD Provider

For detailed Kubernetes CRD configuration, see the [Kubernetes CRD Multi-Layer Routing](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md#multi-layer-routing-with-ingressroutes) section.

## Configuration Attributes

### `parentRefs`

_Optional, Default=[]_

List of parent router names (for File provider) or IngressRoute references (for Kubernetes CRD).

When a router has `parentRefs`, it becomes a non-root router and inherits restrictions:
- Cannot have `tls` configuration
- Cannot have `observability` configuration
- Must have a `service` defined

```yaml tab="File (YAML)"
http:
  routers:
    child-router:
      rule: "Path(`/child`)"
      service: my-service
      parentRefs:
        - parent-router
```

```toml tab="File (TOML)"
[http.routers]
  [http.routers.child-router]
    rule = "Path(`/child`)"
    service = "my-service"
    parentRefs = ["parent-router"]
```

### `childRefs`

_Computed at runtime, not user-configurable_

The `childRefs` field is automatically computed by Traefik based on the `parentRefs` declarations of other routers. You should never manually configure this field.

## Validation Rules

Traefik enforces the following validation rules for multi-layer routing:

!!! important "Root Router Configuration"

    Root routers (routers with no `parentRefs`) can have:
    - `tls` configuration
    - `observability` configuration
    - `entryPoints` configuration
    - Either a `service` OR child routers (via `childRefs`)

!!! important "Non-Root Router Configuration"

    Non-root routers (routers with `parentRefs`) **must not** have:
    - `tls` configuration
    - `observability` configuration

    Non-root routers **must** have:
    - A `service` defined

!!! important "Parent Router Configuration"

    Parent routers (routers with `childRefs`) **must not** have:
    - A `service` defined

    Parent routers can have:
    - `middlewares` (applied before child evaluation)
    - `entryPoints` (if they are root routers)
    - `tls` (if they are root routers)

!!! warning "Circular Dependencies"

    Traefik automatically detects and prevents circular dependencies in router hierarchies. If a circular dependency is detected, the affected routers will be marked as disabled.

!!! warning "Unreachable Routers"

    If a child router references a parent that doesn't exist or is disabled, the child router will be marked as unreachable and disabled.

## Common Patterns

### Pattern 1: Authentication Gateway

Use a parent router to authenticate all requests to a specific path, with child routers handling role-based routing.

```yaml
# Parent: Authenticate all /api requests
api-gateway:
  rule: "PathPrefix(`/api`)"
  middlewares: [auth]

# Children: Route based on roles
api-admin:
  parentRefs: [api-gateway]
  rule: "HeadersRegexp(`X-Role`, `admin`)"
  service: admin-svc

api-user:
  parentRefs: [api-gateway]
  rule: "HeadersRegexp(`X-Role`, `user`)"
  service: user-svc
```

### Pattern 2: Multi-Tenant Routing

Parent router identifies tenant, child routers route to tenant-specific services.

```yaml
# Parent: Extract tenant from subdomain
tenant-gateway:
  rule: "HostRegexp(`^(?P<tenant>[a-z]+)\\.example\\.com$`)"
  middlewares: [tenant-extractor]  # Adds X-Tenant header

# Children: Route to tenant services
tenant-api:
  parentRefs: [tenant-gateway]
  rule: "PathPrefix(`/api`)"
  service: api-svc  # Service uses X-Tenant header for routing

tenant-web:
  parentRefs: [tenant-gateway]
  rule: "PathPrefix(`/`)"
  service: web-svc
```

### Pattern 3: Staged Middleware Application

Apply common middleware at parent level, specific middleware at child level.

```yaml
# Parent: Apply rate limiting and CORS
api-common:
  rule: "Host(`api.example.com`)"
  middlewares: [rate-limit, cors]

# Children: Apply endpoint-specific middleware
api-public:
  parentRefs: [api-common]
  rule: "PathPrefix(`/public`)"
  middlewares: [cache]  # Additional caching for public endpoints
  service: public-svc

api-private:
  parentRefs: [api-common]
  rule: "PathPrefix(`/private`)"
  middlewares: [auth]  # Additional auth for private endpoints
  service: private-svc
```

## Troubleshooting

### Child Router Not Matching

If a child router is not matching requests:

1. **Verify parent router matches**: The parent router must match first
2. **Check middleware modifications**: Ensure parent middleware adds expected headers/context
3. **Inspect child rule**: Child rule must match the **modified** request (after parent middleware)
4. **Check observability**: Enable tracing to see the full request path

### Router Marked as Disabled

If a router is marked as disabled:

1. **Check for circular dependencies**: Ensure no router circular references exist
2. **Verify parent exists**: Ensure all referenced parents exist and are enabled
3. **Validate configuration**: Ensure router follows validation rules (service, TLS, observability)

### Unexpected Routing Behavior

If requests are routed unexpectedly:

1. **Check rule evaluation order**: Parent rules are evaluated before child rules
2. **Verify middleware order**: Middleware order affects request modification
3. **Review priority settings**: Higher priority routers are evaluated first
4. **Enable debug logging**: Use `--log.level=DEBUG` to see routing decisions

## Best Practices

!!! tip "Keep Hierarchies Shallow"

    While Traefik supports multiple levels of hierarchy, keeping hierarchies shallow (1-2 levels) improves maintainability and debugging.

!!! tip "Use Descriptive Router Names"

    Use clear, descriptive names for routers that indicate their role in the hierarchy (e.g., `api-parent`, `api-admin-child`).

!!! tip "Document Middleware Behavior"

    Document what headers or modifications each parent middleware applies, as this affects child router matching.

!!! tip "Test with Observability"

    Use Traefik's tracing and access logs to understand request flow through router hierarchies.

!!! tip "Validate Early"

    Test your multi-layer routing configuration in a development environment before deploying to production.

## Further Reading

- [HTTP Routers](../reference/routing-configuration/http/router/rules-and-priority.md) - Detailed router configuration
- [Kubernetes CRD Multi-Layer Routing](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md#multi-layer-routing-with-ingressroutes) - Kubernetes-specific implementation
- [User Guide: Authentication-Based Routing](../user-guides/multi-layer-routing-authentication.md) - Complete tutorial
- [Middlewares](../reference/routing-configuration/http/middlewares/overview.md) - Available middleware

{!traefik-for-business-applications.md!}