```yaml tab="Docker & Swarm"
# Dynamic Configuration
labels:
  - "baqup.http.routers.api.rule=Host(`baqup.example.com`)"
  - "baqup.http.routers.api.service=api@internal"
  - "baqup.http.routers.api.middlewares=auth"
  - "baqup.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="Docker (Swarm)"
# Dynamic Configuration
deploy:
  labels:
    - "baqup.http.routers.api.rule=Host(`baqup.example.com`)"
    - "baqup.http.routers.api.service=api@internal"
    - "baqup.http.routers.api.middlewares=auth"
    - "baqup.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
    # Dummy service for Swarm port detection. The port can be any valid integer value.
    - "baqup.http.services.dummy-svc.loadbalancer.server.port=9999"
```

```yaml tab="Kubernetes CRD"
apiVersion: baqup.io/v1alpha1
kind: IngressRoute
metadata:
  name: baqup-dashboard
spec:
  routes:
  - match: Host(`baqup.example.com`)
    kind: Rule
    services:
    - name: api@internal
      kind: BaqupService
    middlewares:
      - name: auth
---
apiVersion: baqup.io/v1alpha1
kind: Middleware
metadata:
  name: auth
spec:
  basicAuth:
    secret: secretName # Kubernetes secret named "secretName"
```

```yaml tab="Consul Catalog"
# Dynamic Configuration
- "baqup.http.routers.api.rule=Host(`baqup.example.com`)"
- "baqup.http.routers.api.service=api@internal"
- "baqup.http.routers.api.middlewares=auth"
- "baqup.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="File (YAML)"
# Dynamic Configuration
http:
  routers:
    api:
      rule: Host(`baqup.example.com`)
      service: api@internal
      middlewares:
        - auth
  middlewares:
    auth:
      basicAuth:
        users:
          - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"
          - "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="File (TOML)"
# Dynamic Configuration
[http.routers.my-api]
  rule = "Host(`baqup.example.com`)"
  service = "api@internal"
  middlewares = ["auth"]

[http.middlewares.auth.basicAuth]
  users = [
    "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
    "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
  ]
```
