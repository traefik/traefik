---
title: "Traefik Proxy TCP Middleware Overview"
description: "Read the official Traefik Proxy documentation for an overview of the available TCP middleware."
---
# TCP Middleware Overview

Attached to the routers, pieces of middleware are a means of tweaking the requests before they are sent to your service (or before the answer from the services are sent to the clients).

There are several available middlewares in Traefik, some can modify the request, the headers, some are in charge of redirections, some add authentication, and so on.

Middlewares that use the same protocol can be combined into chains to fit every scenario.

## Configuration Example

```yaml tab="Structured (YAML)"
# As YAML Configuration File
tcp:
  routers:
    router1:
      service: myService
      middlewares:
        - "foo-ip-allowlist"
      rule: "Host(`example.com`)"

  middlewares:
    foo-ip-allowlist:
      ipAllowList:
        sourceRange:
          - "127.0.0.1/32"
          - "192.168.1.7"

  services:
    service1:
      loadBalancer:
        servers:
        - address: "10.0.0.10:4000"
        - address: "10.0.0.11:4000"
```

```toml tab="Structured (TOML)"
# As TOML Configuration File
[tcp.routers]
  [tcp.routers.router1]
    service = "myService"
    middlewares = ["foo-ip-allowlist"]
    rule = "Host(`example.com`)"

[tcp.middlewares]
  [tcp.middlewares.foo-ip-allowlist.ipAllowList]
    sourceRange = ["127.0.0.1/32", "192.168.1.7"]

[tcp.services]
  [tcp.services.service1]
    [tcp.services.service1.loadBalancer]
    [[tcp.services.service1.loadBalancer.servers]]
      address = "10.0.0.10:4000"
    [[tcp.services.service1.loadBalancer.servers]]
      address = "10.0.0.11:4000"
```

```yaml tab="Labels"
labels:
  # Create a middleware named `foo-ip-allowlist`
  - "traefik.tcp.middlewares.foo-ip-allowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7"
  # Apply the middleware named `foo-ip-allowlist` to the router named `router1`
  - "traefik.tcp.routers.router1.middlewares=foo-ip-allowlist@docker"
```

```json tab="Consul Catalog" 
{
  //...
  "Tags" : [
    // Create a middleware named `foo-ip-allowlist`
    "traefik.tcp.middlewares.foo-ip-allowlist.ipallowlist.sourcerange=127.0.0.1/32, 192.168.1.7",
    // Apply the middleware named `foo-ip-allowlist` to the router named `router1`
    "traefik.tcp.routers.router1.middlewares=foo-ip-allowlist@consulcatalog"
  ]
}

```

```yaml tab="Kubernetes"
---
apiVersion: traefik.io/v1alpha1
kind: MiddlewareTCP
metadata:
  name: foo-ip-allowlist
spec:
  ipAllowList:
    sourcerange:
      - 127.0.0.1/32
      - 192.168.1.7

---
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: ingressroute
spec:
# more fields...
  routes:
    # more fields...
    middlewares:
      - name: foo-ip-allowlist
```

## Available TCP Middlewares

| Middleware                                | Purpose                                           | Area                        |
|-------------------------------------------|---------------------------------------------------|-----------------------------|
| [InFlightConn](inflightconn.md)           | Limits the number of simultaneous connections.    | Security, Request lifecycle |
| [IPAllowList](ipallowlist.md)             | Limit the allowed client IPs.                     | Security, Request lifecycle |
