# Middlewares

Tweaking the Request
{: .subtitle }

![Overview](../assets/img/middleware/overview.png)

Attached to the routers, pieces of middleware are a means of tweaking the requests before they are sent to your [service](../routing/services/index.md) (or before the answer from the services are sent to the clients).

There are several available middleware in Traefik, some can modify the request, the headers, some are in charge of redirections, some add authentication, and so on.

Pieces of middleware can be combined in chains to fit every scenario.

!!! warning "Provider Namespace"

    Be aware of the concept of Providers Namespace described in the [Configuration Discovery](../providers/overview.md#provider-namespace) section.
    It also applies to Middlewares.

## Configuration Example

```yaml tab="Docker"
# As a Docker Label
whoami:
  #  A container that exposes an API to show its IP address
  image: traefik/whoami
  labels:
    # Create a middleware named `foo-add-prefix`
    - "traefik.http.middlewares.foo-add-prefix.addprefix.prefix=/foo"
    # Apply the middleware named `foo-add-prefix` to the router named `router1`
    - "traefik.http.routers.router1.middlewares=foo-add-prefix@docker"
```

```yaml tab="Kubernetes IngressRoute"
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: stripprefix
spec:
  stripPrefix:
    prefixes:
      - /stripit

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroute
spec:
# more fields...
  routes:
    # more fields...
    middlewares:
      - name: stripprefix
```

```yaml tab="Consul Catalog"
# Create a middleware named `foo-add-prefix`
- "traefik.http.middlewares.foo-add-prefix.addprefix.prefix=/foo"
# Apply the middleware named `foo-add-prefix` to the router named `router1`
- "traefik.http.routers.router1.middlewares=foo-add-prefix@consulcatalog"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.foo-add-prefix.addprefix.prefix": "/foo",
  "traefik.http.routers.router1.middlewares": "foo-add-prefix@marathon"
}
```

```yaml tab="Rancher"
# As a Rancher Label
labels:
  # Create a middleware named `foo-add-prefix`
  - "traefik.http.middlewares.foo-add-prefix.addprefix.prefix=/foo"
  # Apply the middleware named `foo-add-prefix` to the router named `router1`
  - "traefik.http.routers.router1.middlewares=foo-add-prefix@rancher"
```

```toml tab="File (TOML)"
# As TOML Configuration File
[http.routers]
  [http.routers.router1]
    service = "myService"
    middlewares = ["foo-add-prefix"]
    rule = "Host(`example.com`)"

[http.middlewares]
  [http.middlewares.foo-add-prefix.addPrefix]
    prefix = "/foo"

[http.services]
  [http.services.service1]
    [http.services.service1.loadBalancer]

      [[http.services.service1.loadBalancer.servers]]
        url = "http://127.0.0.1:80"
```

```yaml tab="File (YAML)"
# As YAML Configuration File
http:
  routers:
    router1:
      service: myService
      middlewares:
        - "foo-add-prefix"
      rule: "Host(`example.com`)"

  middlewares:
    foo-add-prefix:
      addPrefix:
        prefix: "/foo"

  services:
    service1:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:80"
```

## Available Middlewares

| Middleware                                | Purpose                                           | Area                        |
|-------------------------------------------|---------------------------------------------------|-----------------------------|
| [AddPrefix](addprefix.md)                 | Add a Path Prefix                                 | Path Modifier               |
| [BasicAuth](basicauth.md)                 | Basic auth mechanism                              | Security, Authentication    |
| [Buffering](buffering.md)                 | Buffers the request/response                      | Request Lifecycle           |
| [Chain](chain.md)                         | Combine multiple pieces of middleware             | Middleware tool             |
| [CircuitBreaker](circuitbreaker.md)       | Stop calling unhealthy services                   | Request Lifecycle           |
| [Compress](compress.md)                   | Compress the response                             | Content Modifier            |
| [DigestAuth](digestauth.md)               | Adds Digest Authentication                        | Security, Authentication    |
| [Errors](errorpages.md)                   | Define custom error pages                         | Request Lifecycle           |
| [ForwardAuth](forwardauth.md)             | Authentication delegation                         | Security, Authentication    |
| [Headers](headers.md)                     | Add / Update headers                              | Security                    |
| [IPWhiteList](ipwhitelist.md)             | Limit the allowed client IPs                      | Security, Request lifecycle |
| [InFlightReq](inflightreq.md)             | Limit the number of simultaneous connections      | Security, Request lifecycle |
| [PassTLSClientCert](passtlsclientcert.md) | Adding Client Certificates in a Header            | Security                    |
| [RateLimit](ratelimit.md)                 | Limit the call frequency                          | Security, Request lifecycle |
| [RedirectScheme](redirectscheme.md)       | Redirect easily the client elsewhere              | Request lifecycle           |
| [RedirectRegex](redirectregex.md)         | Redirect the client elsewhere                     | Request lifecycle           |
| [ReplacePath](replacepath.md)             | Change the path of the request                    | Path Modifier               |
| [ReplacePathRegex](replacepathregex.md)   | Change the path of the request                    | Path Modifier               |
| [Retry](retry.md)                         | Automatically retry the request in case of errors | Request lifecycle           |
| [StripPrefix](stripprefix.md)             | Change the path of the request                    | Path Modifier               |
| [StripPrefixRegex](stripprefixregex.md)   | Change the path of the request                    | Path Modifier               |
