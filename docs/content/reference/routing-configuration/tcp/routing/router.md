---
title: "Traefik TCP Routers Documentation"
description: "TCP routers are responsible for connecting incoming TCP connections to the services that can handle them. Read the technical documentation."
---

## TCP Router

A TCP router is in charge of connecting incoming TCP connections to the services that can handle them. TCP routers analyze incoming connections based on rules, and when a match is found, forward the connection through any configured middlewares to the appropriate service.

!!! note "TCP vs HTTP Routing"
    If both HTTP routers and TCP routers listen to the same EntryPoint, the TCP routers will apply before the HTTP routers. If no matching route is found for the TCP routers, then the HTTP routers will take over.

## Configuration Example

```yaml tab="Structured (YAML)"
tcp:
  routers:
    my-tcp-router:
      entryPoints:
        - "tcp-ep"
        - "websecure"
      rule: "HostSNI(`example.com`)"
      priority: 10
      middlewares:
        - "tcp-ipallowlist"
      tls:
        passthrough: false
        certResolver: "letsencrypt"
        options: "modern-tls"
        domains:
          - main: "example.com"
            sans:
              - "www.example.com"
      service: my-tcp-service
```

```toml tab="Structured (TOML)"
[tcp.routers]
  [tcp.routers.my-tcp-router]
    entryPoints = ["tcp-ep", "websecure"]
    rule = "HostSNI(`example.com`)"
    priority = 10
    middlewares = ["tcp-ipallowlist"]
    service = "my-tcp-service"

    [tcp.routers.my-tcp-router.tls]
      passthrough = false
      certResolver = "letsencrypt"
      options = "modern-tls"

      [[tcp.routers.my-tcp-router.tls.domains]]
        main = "example.com"
        sans = ["www.example.com"]
```

```yaml tab="Labels"
labels:
  - "traefik.tcp.routers.my-tcp-router.entrypoints=tcp-ep,websecure"
  - "traefik.tcp.routers.my-tcp-router.rule=HostSNI(`example.com`)"
  - "traefik.tcp.routers.my-tcp-router.priority=10"
  - "traefik.tcp.routers.my-tcp-router.middlewares=tcp-ipallowlist"
  - "traefik.tcp.routers.my-tcp-router.tls.certresolver=letsencrypt"
  - "traefik.tcp.routers.my-tcp-router.tls.passthrough=false"
  - "traefik.tcp.routers.my-tcp-router.tls.options=modern-tls"
  - "traefik.tcp.routers.my-tcp-router.tls.domains[0].main=example.com"
  - "traefik.tcp.routers.my-tcp-router.tls.domains[0].sans=www.example.com"
  - "traefik.tcp.routers.my-tcp-router.service=my-tcp-service"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.tcp.routers.my-tcp-router.entrypoints=tcp-ep,websecure",
    "traefik.tcp.routers.my-tcp-router.rule=HostSNI(`example.com`)",
    "traefik.tcp.routers.my-tcp-router.priority=10",
    "traefik.tcp.routers.my-tcp-router.middlewares=tcp-ipallowlist",
    "traefik.tcp.routers.my-tcp-router.tls.certresolver=letsencrypt",
    "traefik.tcp.routers.my-tcp-router.tls.passthrough=false",
    "traefik.tcp.routers.my-tcp-router.tls.options=modern-tls",
    "traefik.tcp.routers.my-tcp-router.tls.domains[0].main=example.com",
    "traefik.tcp.routers.my-tcp-router.tls.domains[0].sans=www.example.com",
    "traefik.tcp.routers.my-tcp-router.service=my-tcp-service"
  ]
}
```

## Configuration Options

| Field                                                                          | Description                                                                                                                                                                                                                                                                                                          | Default              | Required |
|--------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------|----------|
| <a id="opt-entryPoints" href="#opt-entryPoints" title="#opt-entryPoints">`entryPoints`</a> | The list of entry points to which the router is attached. If not specified, TCP routers are attached to all TCP entry points.                                                                                                                                                                                        | All TCP entry points | No       |
| <a id="opt-rule" href="#opt-rule" title="#opt-rule">`rule`</a> | Rules are a set of matchers configured with values, that determine if a particular connection matches specific criteria. If the rule is verified, the router becomes active, calls middlewares, and then forwards the connection to the service. See [Rules & Priority](./rules-and-priority.md) for details.        |                      | Yes      |
| <a id="opt-priority" href="#opt-priority" title="#opt-priority">`priority`</a> | To avoid rule overlap, routes are sorted, by default, in descending order using rules length. The priority is directly equal to the length of the rule, and so the longest length has the highest priority. A value of `0` for the priority is ignored. See [Rules & Priority](./rules-and-priority.md) for details. | Rule length          | No       |
| <a id="opt-middlewares" href="#opt-middlewares" title="#opt-middlewares">`middlewares`</a> | The list of middlewares that are applied to the router. Middlewares are applied in the order they are declared. See [TCP Middlewares overview](../middlewares/overview.md) for available TCP middlewares.                                                                                                            |                      | No       |
| <a id="opt-tls" href="#opt-tls" title="#opt-tls">`tls`</a> | TLS configuration for the router. When specified, the router will only handle TLS connections. See [TLS configuration](../tls.md) for detailed TLS options.                                                                                                                                                          |                      | No       |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The name of the service that will handle the matched connections. Services can be load balancer services or weighted round robin services. See [TCP Service](../service.md) for details.                                                                                                                             |                      | Yes      |

## Router Naming

- The character `@` is not authorized in the router name
- Router names should be descriptive and follow your naming conventions
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names and rules

{!traefik-for-business-applications.md!}
