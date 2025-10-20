---
title: "Multi-Layer Routing"
description: "Learn how to use Traefik's multi-layer routing to create hierarchical router relationships where parent routers can apply middleware before child routers make routing decisions."
---

# Multi-Layer Routing

Hierarchical Router Relationships for Advanced Routing Scenarios.

## Overview

Multi-layer routing enables you to create hierarchical relationships between routers,
where parent routers can process requests through middleware before child routers make final routing decisions.

This feature allows middleware at the parent level to modify requests (adding headers, performing authentication, etc.) that influence how child routers evaluate their rules and route traffic to services.

Multi-layer routing is particularly useful for progressive request enrichment, where each layer adds context to the request, enabling increasingly specific routing decisions:

- **Authentication-Based Routing**: Parent router authenticates requests and adds user context (roles, permissions) as headers, child routers route based on these headers
- **Staged Middleware Application**: Apply common middleware (rate limiting, CORS) at parent level (for a given domain/path), but specific middleware at child level

!!! info "Provider Support"

    Multi-layer routing is supported by the following providers:

    - **File provider** (YAML, TOML, JSON)
    - **KV stores** (Consul, etcd, Redis, ZooKeeper)
    - **Kubernetes CRD** (IngressRoute)

    Multi-layer routing is not available for other providers (Docker, Kubernetes Ingress, Gateway API, etc.).


## How It Works

```
Request → EntryPoint → Parent Router → Middleware → Child Router A → Service A
                                          ↓       → Child Router B → Service B
                                     Modify Request
                                  (e.g., add headers)
```

1. **Request arrives** at an entrypoint
2. **Parent router matches** based on its rule (e.g., ```Host(`example.com`)```)
3. **Parent middleware executes**, potentially modifying the request
4. **One child router matches** based on its rule (which may use modified request attributes)
5. **Request is forwarded** to the matching child router's service

## Building a Router Hierarchy

### Root Routers

- Have no `parentRefs` (top of the hierarchy)
- **Can** have `tls`, `observability`, and `entryPoints` configuration
- Can be either parent routers (with children) or standalone routers (with service)
- **Can** have models applied (non-root routers cannot have models)

### Intermediate Routers

- Reference their parent router(s) via `parentRefs`
- Have one or more child routers
- **Must not** have a `service` defined
- **Must not** have `entryPoints`, `tls`, or `observability` configuration

### Leaf Routers

- Reference their parent router(s) via `parentRefs`
- **Must** have a `service` defined
- **Must not** have `entryPoints`, `tls`, or `observability` configuration

## Configuration Example

??? example "Authentication-Based Routing"

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

    ```txt tab="KV (Consul/etcd/Redis/ZK)"
    | Key                                                                    | Value                           |
    |------------------------------------------------------------------------|---------------------------------|
    | `traefik/http/routers/api-parent/rule`                                 | `PathPrefix(\`/api\`)`          |
    | `traefik/http/routers/api-parent/middlewares/0`                        | `auth-middleware`               |
    | `traefik/http/routers/api-parent/entrypoints/0`                        | `websecure`                     |
    | `traefik/http/routers/api-parent/tls`                                  | `true`                          |
    | `traefik/http/routers/api-admin/rule`                                  | `HeadersRegexp(\`X-User-Role\`, \`admin\`)` |
    | `traefik/http/routers/api-admin/service`                               | `admin-service`                 |
    | `traefik/http/routers/api-admin/parentrefs/0`                          | `api-parent`                    |
    | `traefik/http/routers/api-user/rule`                                   | `HeadersRegexp(\`X-User-Role\`, \`user\`)` |
    | `traefik/http/routers/api-user/service`                                | `user-service`                  |
    | `traefik/http/routers/api-user/parentrefs/0`                           | `api-parent`                    |
    | `traefik/http/middlewares/auth-middleware/forwardauth/address`         | `http://auth-service:8080/auth` |
    | `traefik/http/middlewares/auth-middleware/forwardauth/authresponseheaders/0` | `X-User-Role`         |
    | `traefik/http/middlewares/auth-middleware/forwardauth/authresponseheaders/1` | `X-User-Name`         |
    | `traefik/http/services/admin-service/loadbalancer/servers/0/url`       | `http://admin-backend:8080`     |
    | `traefik/http/services/user-service/loadbalancer/servers/0/url`        | `http://user-backend:8080`      |
    ```

    **How it works:**

    1. Request to `/api/endpoint` matches `api-parent` router
    2. `auth-middleware` (ForwardAuth) validates the request and adds `X-User-Role` header
    3. Modified request is evaluated by child routers
    4. If `X-User-Role: admin`, `api-admin` router matches and forwards to `admin-service`
    5. If `X-User-Role: user`, `api-user` router matches and forwards to `user-service`

{!traefik-for-business-applications.md!}
