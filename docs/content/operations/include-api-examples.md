```yaml tab="Docker (Standalone)"
# Dynamic Configuration
labels:
  - "traefik.http.routers.api.rule=Host(`traefik.domain.com`)"
  - "traefik.http.routers.api.service=api@internal"
  - "traefik.http.routers.api.middlewares=auth"
  - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="Docker (Swarm)"
# Dynamic Configuration
labels:
  - "traefik.http.routers.api.rule=Host(`traefik.domain.com`)"
  - "traefik.http.routers.api.service=api@internal"
  - "traefik.http.services.api-dummy-svc.loadbalancer.server.port=9999" # Dummy service for Swarm port detection. The port can be any valid integer value.
  - "traefik.http.routers.api.middlewares=auth"
  - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```json tab="Marathon"
"labels": {
  "traefik.http.routers.api.rule": "Host(`traefik.domain.com`)",
  "traefik.http.routers.api.service": "api@internal",
  "traefik.http.routers.api.middlewares": "auth",
  "traefik.http.middlewares.auth.basicauth.users": "test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
}
```

```yaml tab="Rancher"
# Dynamic Configuration
labels:
  - "traefik.http.routers.api.rule=Host(`traefik.domain.com`)"
  - "traefik.http.routers.api.service=api@internal"
  - "traefik.http.routers.api.middlewares=auth"
  - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="File (TOML)"
# Dynamic Configuration
[http.routers.my-api]
  rule = "Host(`traefik.domain.com`)"
  service = "api@internal"
  middlewares = ["auth"]

[http.middlewares.auth.basicAuth]
  users = [
    "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
    "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
  ]
```

```yaml tab="File (YAML)"
# Dynamic Configuration
http:
  routers:
    api:
      rule: Host(`traefik.domain.com`)
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
