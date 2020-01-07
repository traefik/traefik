
```yaml tab="Docker"
## Dynamic configuration
labels:
  - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
  - traefik.http.routers.blog.tls=true
  - traefik.http.routers.blog.tls.certresolver=le
```

```yaml tab="Docker (Swarm)"
## Dynamic configuration
deploy:
  labels:
    - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
    - traefik.http.routers.blog.tls=true
    - traefik.http.routers.blog.tls.certresolver=le
    - traefik.http.services.blog-svc.loadbalancer.server.port=8080"
```

```yaml tab="Kubernetes"
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: blogtls
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`company.com`) && Path(`/blog`)
    kind: Rule
    services:
    - name: blog
      port: 8080
  tls: {}
```

```json tab="Marathon"
labels: {
  "traefik.http.routers.blog.rule": "Host(`company.com`) && Path(`/blog`)",
  "traefik.http.routers.blog.tls": "true",
  "traefik.http.routers.blog.tls.certresolver": "le",
  "traefik.http.services.blog-svc.loadbalancer.server.port": "8080"
}
```

```yaml tab="Rancher"
## Dynamic configuration
labels:
  - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
  - traefik.http.routers.blog.tls=true
  - traefik.http.routers.blog.tls.certresolver=le
```

```toml tab="Single Domain"
## Dynamic configuration
[http.routers]
    [http.routers.blog]
    rule = "Host(`company.com`) && Path(`/blog`)"
    [http.routers.blog.tls]
        certResolver = "le" # From static configuration
```

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    blog:
      rule: "Host(`company.com`) && Path(`/blog`)"
      tls:
        certResolver: le
```
