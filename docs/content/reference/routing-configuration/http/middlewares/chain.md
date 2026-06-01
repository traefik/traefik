---
title: "Traefik Chain Middleware Documentation"
description: "The HTTP chain middleware lets you define reusable combinations of other middleware, to reuse the same groups. Read the technical documentation."
---

The `chain` middleware enables you to define reusable combinations of other pieces of middleware.
It makes it effortless to reuse the same groups.

## Configuration Example

Below is an example of a Chain containing `AllowList`, `BasicAuth`, and `RedirectScheme`.

```yaml tab="Structured (YAML)"
# ...
http:
  routers:
    router1:
      service: service1
      middlewares:
        - secured
      rule: "Host(`mydomain`)"

  middlewares:
    secured:
      chain:
        middlewares:
          - https-only
          - known-ips
          - auth-users

    auth-users:
      basicAuth:
        users:
          - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"

    https-only:
      redirectScheme:
        scheme: https

    known-ips:
      ipAllowList:
        sourceRange:
          - "192.168.1.7"
          - "127.0.0.1/32"

  services:
    service1:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:80"
```

```toml tab="Structured (TOML)"
# ...
[http.routers]
  [http.routers.router1]
    service = "service1"
    middlewares = ["secured"]
    rule = "Host(`mydomain`)"

[http.middlewares]
  [http.middlewares.secured.chain]
    middlewares = ["https-only", "known-ips", "auth-users"]

  [http.middlewares.auth-users.basicAuth]
    users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]

  [http.middlewares.https-only.redirectScheme]
    scheme = "https"

  [http.middlewares.known-ips.ipAllowList]
    sourceRange = ["192.168.1.7", "127.0.0.1/32"]

[http.services]
  [http.services.service1]
    [http.services.service1.loadBalancer]
      [[http.services.service1.loadBalancer.servers]]
        url = "http://127.0.0.1:80"
``` 

```yaml tab="Labels"
labels:
  - "traefik.http.routers.router1.service=service1"
  - "traefik.http.routers.router1.middlewares=secured"
  - "traefik.http.routers.router1.rule=Host(`mydomain`)"
  - "traefik.http.middlewares.secured.chain.middlewares=https-only,known-ips,auth-users"
  - "traefik.http.middlewares.auth-users.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"
  - "traefik.http.middlewares.https-only.redirectscheme.scheme=https"
  - "traefik.http.middlewares.known-ips.ipallowlist.sourceRange=192.168.1.7,127.0.0.1/32"
  - "traefik.http.services.service1.loadbalancer.server.port=80"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.router1.service=service1",
    "traefik.http.routers.router1.middlewares=secured",
    "traefik.http.routers.router1.rule=Host(`mydomain`)",
    "traefik.http.middlewares.secured.chain.middlewares=https-only,known-ips,auth-users",
    "traefik.http.middlewares.auth-users.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
    "traefik.http.middlewares.https-only.redirectscheme.scheme=https",
    "traefik.http.middlewares.known-ips.ipallowlist.sourceRange=192.168.1.7,127.0.0.1/32",
    "traefik.http.services.service1.loadbalancer.server.port=80"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: test
  namespace: default
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`mydomain`)
      kind: Rule
      services:
        - name: whoami
          port: 80
      middlewares:
        - name: secured
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: secured
spec:
  chain:
    middlewares:
    - name: https-only
    - name: known-ips
    - name: auth-users
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: auth-users
spec:
  basicAuth:
    users:
    - test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: https-only
spec:
  redirectScheme:
    scheme: https
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: known-ips
spec:
  ipAllowList:
    sourceRange:
    - 192.168.1.7
    - 127.0.0.1/32
```


## Real-World Example: API Gateway Middleware Stack

The most common production pattern for API gateways behind Traefik is combining **authentication**, **rate limiting**, and **CORS headers** into a single reusable middleware stack. The `chain` middleware makes this straightforward.

```yaml tab="Structured (YAML)"
# Common API gateway middleware stack: auth + rate-limit + CORS
http:
  routers:
    api-router:
      service: api-service
      middlewares:
        - api-stack
      rule: "Host(`api.example.com`)"

  middlewares:
    api-stack:
      chain:
        middlewares:
          - api-auth
          - api-ratelimit
          - api-cors

    api-auth:
      basicAuth:
        users:
          - "admin:$2y$10$..."  # bcrypt hash

    api-ratelimit:
      rateLimit:
        average: 100
        burst: 50

    api-cors:
      headers:
        accessControlAllowMethods:
          - GET
          - POST
          - OPTIONS
        accessControlAllowOriginList:
          - "https://app.example.com"

  services:
    api-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:8080"
```

```toml tab="Structured (TOML)"
# Common API gateway middleware stack: auth + rate-limit + CORS
[http.routers]
  [http.routers.api-router]
    service = "api-service"
    middlewares = ["api-stack"]
    rule = "Host(`api.example.com`)"

[http.middlewares]
  [http.middlewares.api-stack.chain]
    middlewares = ["api-auth", "api-ratelimit", "api-cors"]

  [http.middlewares.api-auth.basicAuth]
    users = ["admin:$2y$10$..."]

  [http.middlewares.api-ratelimit.rateLimit]
    average = 100
    burst = 50

  [http.middlewares.api-cors.headers]
    accessControlAllowMethods = ["GET", "POST", "OPTIONS"]
    accessControlAllowOriginList = ["https://app.example.com"]

[http.services]
  [http.services.api-service]
    [http.services.api-service.loadBalancer]
      [[http.services.api-service.loadBalancer.servers]]
        url = "http://127.0.0.1:8080"
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.api-router.service=api-service"
  - "traefik.http.routers.api-router.middlewares=api-stack"
  - "traefik.http.routers.api-router.rule=Host(`api.example.com`)"
  - "traefik.http.middlewares.api-stack.chain.middlewares=api-auth,api-ratelimit,api-cors"
  - "traefik.http.middlewares.api-auth.basicauth.users=admin:$2y$10$..."
  - "traefik.http.middlewares.api-ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.api-ratelimit.ratelimit.burst=50"
  - "traefik.http.middlewares.api-cors.headers.accesscontrolallowmethods=GET,POST,OPTIONS"
  - "traefik.http.middlewares.api-cors.headers.accesscontrolalloworiginlist=https://app.example.com"
  - "traefik.http.services.api-service.loadbalancer.server.port=8080"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.api-router.service=api-service",
    "traefik.http.routers.api-router.middlewares=api-stack",
    "traefik.http.routers.api-router.rule=Host(`api.example.com`)",
    "traefik.http.middlewares.api-stack.chain.middlewares=api-auth,api-ratelimit,api-cors",
    "traefik.http.middlewares.api-auth.basicauth.users=admin:$2y$10$...",
    "traefik.http.middlewares.api-ratelimit.ratelimit.average=100",
    "traefik.http.middlewares.api-ratelimit.ratelimit.burst=50",
    "traefik.http.middlewares.api-cors.headers.accesscontrolallowmethods=GET,POST,OPTIONS",
    "traefik.http.middlewares.api-cors.headers.accesscontrolalloworiginlist=https://app.example.com",
    "traefik.http.services.api-service.loadbalancer.server.port=8080"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-ingress
  namespace: default
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      services:
        - name: api-service
          port: 8080
      middlewares:
        - name: api-stack
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: api-stack
spec:
  chain:
    middlewares:
    - name: api-auth
    - name: api-ratelimit
    - name: api-cors
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: api-auth
spec:
  basicAuth:
    users:
    - admin:$2y$10$...
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: api-ratelimit
spec:
  rateLimit:
    average: 100
    burst: 50
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: api-cors
spec:
  headers:
    accessControlAllowMethods:
      - GET
      - POST
      - OPTIONS
    accessControlAllowOriginList:
      - https://app.example.com
```

## Configuration Options

| Field | Description | Default | Required |
|:------|:------------|:--------|:---------|
| <a id="opt-middlewares" href="#opt-middlewares" title="#opt-middlewares">`middlewares`</a> | List of middlewares to chain.<br /> The middlewares have to be in the same namespace as the `chain` middleware. | [] | Yes |
