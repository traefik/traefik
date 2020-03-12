
```yaml tab="Docker"
## Dynamic configuration
labels:
  - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
  - traefik.http.routers.blog.tls=true
  - traefik.http.routers.blog.tls.certresolver=myresolver
  - traefik.http.routers.blog.tls.domains[0].main=company.org
  - traefik.http.routers.blog.tls.domains[0].sans=*.company.org
```

```yaml tab="Docker (Swarm)"
## Dynamic configuration
deploy:
  labels:
    - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
    - traefik.http.services.blog-svc.loadbalancer.server.port=8080"
    - traefik.http.routers.blog.tls=true
    - traefik.http.routers.blog.tls.certresolver=myresolver
    - traefik.http.routers.blog.tls.domains[0].main=company.org
    - traefik.http.routers.blog.tls.domains[0].sans=*.company.org
```

```yaml tab="Kubernetes"
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
  tls:
    certResolver: myresolver
    domains:
    - main: company.org
      sans:
      - *.company.org
```

```json tab="Marathon"
labels: {
  "traefik.http.routers.blog.rule": "Host(`company.com`) && Path(`/blog`)",
  "traefik.http.routers.blog.tls": "true",
  "traefik.http.routers.blog.tls.certresolver": "myresolver",
  "traefik.http.routers.blog.tls.domains[0].main": "company.com",
  "traefik.http.routers.blog.tls.domains[0].sans": "*.company.com",
  "traefik.http.services.blog-svc.loadbalancer.server.port": "8080"
}
```

```yaml tab="Rancher"
## Dynamic configuration
labels:
  - traefik.http.routers.blog.rule=Host(`company.com`) && Path(`/blog`)
  - traefik.http.routers.blog.tls=true
  - traefik.http.routers.blog.tls.certresolver=myresolver
  - traefik.http.routers.blog.tls.domains[0].main=company.org
  - traefik.http.routers.blog.tls.domains[0].sans=*.company.org
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.blog]
    rule = "Host(`company.com`) && Path(`/blog`)"
    [http.routers.blog.tls]
      certResolver = "myresolver" # From static configuration
      [[http.routers.blog.tls.domains]]
        main = "company.org"
        sans = ["*.company.org"]
```

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    blog:
      rule: "Host(`company.com`) && Path(`/blog`)"
      tls:
        certResolver: myresolver
        domains:
          - main: "company.org"
            sans:
              - "*.company.org"
```
