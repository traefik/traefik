# Branching

Decisions must be made
{: .subtitle }

The Branching middleware allows you to define an alternative middleware chain based on conditional evaluation of request values.

## Configuration Example

Below is an example of Branching based on a `Header` matching rule.

```yaml tab="Docker"
labels:
  # Middleware Declaration
  - "traefik.http.middlewares.resp-header.headers.customResponseHeaders.X-Foo=hit"
  - "traefik.http.middlewares.header-branch.branching.condition=Header[`Foo`].0 == `bar`"
  - "traefik.http.middlewares.header-branch.branching.chain.middlewares=resp-header"
  # Router Declaration
  - "traefik.http.routers.router1.service=service1"
  - "traefik.http.routers.router1.middlewares=header-branch"
  - "traefik.http.routers.router1.rule=Host(`mydomain`)"
  - "traefik.http.services.service1.loadbalancer.server.port=80"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
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
        - name: header-branch
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: header-branch
spec:
  branching:
    condition: Header[`Foo`].0 == `bar`
    chain:
      middlewares:
      - name: resp-header
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: resp-header
spec:
  headers:
    customResponseHeaders:
      X-Foo: hit
```

```yaml tab="Consul Catalog"
# Middleware Declaration
- "traefik.http.middlewares.resp-header.headers.customResponseHeaders.X-Foo=hit"
- "traefik.http.middlewares.header-branch.branching.condition=Header[`Foo`].0 == `bar`"
- "traefik.http.middlewares.header-branch.branching.chain.middlewares=resp-header"
# Router Declaration
- "traefik.http.routers.router1.service=service1"
- "traefik.http.routers.router1.middlewares=header-branch"
- "traefik.http.routers.router1.rule=Host(`mydomain`)"
- "traefik.http.services.service1.loadbalancer.server.port=80"
```

```json tab="Marathon"
"labels": {  
  "traefik.http.middlewares.resp-header.headers.customResponseHeaders.X-Foo": "hit",
  "traefik.http.middlewares.header-branch.branching.condition": "Header[`Foo`].0 == `bar`",
  "traefik.http.middlewares.header-branch.branching.chain.middlewares": "resp-header",

  "traefik.http.routers.router1.service": "service1",
  "traefik.http.routers.router1.middlewares": "header-branch",
  "traefik.http.routers.router1.rule": "Host(`mydomain`)",  
  "traefik.http.services.service1.loadbalancer.server.port": "80"
}
```

```yaml tab="Rancher"
labels:
  # Middleware Declaration
  - "traefik.http.middlewares.resp-header.headers.customResponseHeaders.X-Foo=hit"
  - "traefik.http.middlewares.header-branch.branching.condition=Header[`Foo`].0 == `bar`"
  - "traefik.http.middlewares.header-branch.branching.chain.middlewares=resp-header"
  # Router Declaration
  - "traefik.http.routers.router1.service=service1"
  - "traefik.http.routers.router1.middlewares=header-branch"
  - "traefik.http.routers.router1.rule=Host(`mydomain`)"
  - "traefik.http.services.service1.loadbalancer.server.port=80"
```

```yaml tab="File (YAML)"
# ...
http:
  routers:
    router1:
      service: service1
      middlewares:
        - header-branch
      rule: "Host(`mydomain`)"

  middlewares:
    resp-header:
      headers:
        customResponseHeaders:
          X-Foo: hit
    header-branch:
      branching:
        chain:
          condition: Header[`Foo`].0 == `bar`
          middlewares:
            - resp-header

  services:
    service1:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:80"
```

```toml tab="File (TOML)"
# ...
[http.routers]
  [http.routers.router1]
    service = "service1"
    middlewares = ["header-branch"]
    rule = "Host(`mydomain`)"

[http.middlewares]
  [http.middlewares.resp-header.headers]
    [http.middlewares.resp-header.headers.customResponseHeaders]
        X-Foo = "hit"

  [http.middlewares.header-branch.branching]
    condition = "Header[`Foo`].0 == `bar`"
    [http.middlewares.header-branch.branching.chain]
      middlewares = ["resp-header"]

[http.services]
  [http.services.service1]
    [http.services.service1.loadBalancer]
      [[http.services.service1.loadBalancer.servers]]
        url = "http://127.0.0.1:80"
```

## Configuration Options

### `Condition`

The branch condition expression follows the standard [JSON pointer syntax](https://datatracker.ietf.org/doc/html/rfc6901). It is evaluated against the [Request object](https://pkg.go.dev/net/http@go1.17#Request) as defined in the Go standard library.

!!! important Request Object
    Only exported (capitalized) fields can be used for evaluation.

### `Chain`

Specifies a list of middleware that will be activated when the condition is matched. It works exactly as the [chain middleware](chain.md).
