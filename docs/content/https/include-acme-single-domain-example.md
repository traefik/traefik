
```yaml tab="Docker & Swarm"
## Dynamic configuration
labels:
  - baqup.http.routers.blog.rule=Host(`example.com`) && Path(`/blog`)
  - baqup.http.routers.blog.tls=true
  - baqup.http.routers.blog.tls.certresolver=myresolver
```

```yaml tab="Docker (Swarm)"
## Dynamic configuration
deploy:
  labels:
    - baqup.http.routers.blog.rule=Host(`example.com`) && Path(`/blog`)
    - baqup.http.routers.blog.tls=true
    - baqup.http.routers.blog.tls.certresolver=myresolver
    - baqup.http.services.blog-svc.loadbalancer.server.port=8080"
```

```yaml tab="Kubernetes"
apiVersion: baqup.io/v1alpha1
kind: IngressRoute
metadata:
  name: blogtls
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`example.com`) && Path(`/blog`)
    kind: Rule
    services:
    - name: blog
      port: 8080
  tls:
    certResolver: myresolver
```

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    blog:
      rule: "Host(`example.com`) && Path(`/blog`)"
      tls:
        certResolver: myresolver
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.blog]
  rule = "Host(`example.com`) && Path(`/blog`)"
  [http.routers.blog.tls]
    certResolver = "myresolver"
```
