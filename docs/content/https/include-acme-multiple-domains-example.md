
```yaml tab="Docker & Swarm"
## Dynamic configuration
labels:
  - traefik.http.routers.blog.rule=Host(`example.com`) && Path(`/blog`)
  - traefik.http.routers.blog.tls=true
  - traefik.http.routers.blog.tls.certresolver=myresolver
  - traefik.http.routers.blog.tls.domains[0].main=example.com
  - traefik.http.routers.blog.tls.domains[0].sans=*.example.org
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
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
    domains:
    - main: example.com
      sans:
      - '*.example.org'
```

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    blog:
      rule: "Host(`example.com`) && Path(`/blog`)"
      tls:
        certResolver: myresolver
        domains:
          - main: "example.com"
            sans:
              - "*.example.org"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.blog]
    rule = "Host(`example.com`) && Path(`/blog`)"
    [http.routers.blog.tls]
      certResolver = "myresolver" # From static configuration
      [[http.routers.blog.tls.domains]]
        main = "example.com"
        sans = ["*.example.org"]
```
