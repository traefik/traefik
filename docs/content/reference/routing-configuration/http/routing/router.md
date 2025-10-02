---
title: "Traefik HTTP Routers Documentation"
description: "HTTP routers are responsible for connecting incoming requests to the services that can handle them. Read the technical documentation."
---

## HTTP Router

An HTTP router is in charge of connecting incoming requests to the services that can handle them. Routers analyze incoming requests based on rules, and when a match is found, forward the request through any configured middlewares to the appropriate service.

## Configuration Example

```yaml tab="Structured (YAML)"
http:
  routers:
    my-router:
      entryPoints:
        - "web"
        - "websecure"
      rule: "Host(`example.com`) && Path(`/api`)"
      priority: 10
      middlewares:
        - "auth"
        - "ratelimit"
      tls:
        certResolver: "letsencrypt"
      observability:
        metrics: true
        accessLogs: true
        tracing: true
      service: my-service
```

```toml tab="Structured (TOML)"
[http.routers]
  [http.routers.my-router]
    entryPoints = ["web", "websecure"]
    rule = "Host(`example.com`) && Path(`/api`)"
    priority = 10
    middlewares = ["auth", "ratelimit"]
    service = "my-service"

    [http.routers.my-router.tls]
      certResolver = "letsencrypt"

    [http.routers.my-router.observability]
      metrics = true
      accessLogs = true
      tracing = true
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.my-router.entrypoints=web,websecure"
  - "traefik.http.routers.my-router.rule=Host(`example.com`) && Path(`/api`)"
  - "traefik.http.routers.my-router.priority=10"
  - "traefik.http.routers.my-router.middlewares=auth,ratelimit"
  - "traefik.http.routers.my-router.service=my-service"
  - "traefik.http.routers.my-router.tls.certresolver=letsencrypt"
  - "traefik.http.routers.my-router.observability.metrics=true"
  - "traefik.http.routers.my-router.observability.accessLogs=true"
  - "traefik.http.routers.my-router.observability.tracing=true"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.http.routers.my-router.entrypoints=web,websecure",
    "traefik.http.routers.my-router.rule=Host(`example.com`) && Path(`/api`)",
    "traefik.http.routers.my-router.priority=10",
    "traefik.http.routers.my-router.middlewares=auth,ratelimit",
    "traefik.http.routers.my-router.service=my-service",
    "traefik.http.routers.my-router.tls.certresolver=letsencrypt",
    "traefik.http.routers.my-router.observability.metrics=true",
    "traefik.http.routers.my-router.observability.accessLogs=true",
    "traefik.http.routers.my-router.observability.tracing=true"
  ]
}
```

## Configuration Options

| Field                              | Description                                                                                                                                                                                                                                                                                                                                                                                | Default | Required |
|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="entryPoints" href="#entryPoints" title="#entryPoints">`entryPoints`</a> | The list of entry points to which the router is attached. If not specified, HTTP routers are attached to all entry points. | All entry points | No |
| <a id="rule" href="#rule" title="#rule">`rule`</a> | Rules are a set of matchers configured with values, that determine if a particular request matches specific criteria. If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service. See [Rules & Priority](./rules-and-priority.md) for details. | | Yes |
| <a id="priority" href="#priority" title="#priority">`priority`</a> | To avoid path overlap, routes are sorted, by default, in descending order using rules length. The priority is directly equal to the length of the rule, and so the longest length has the highest priority. A value of `0` for the priority is ignored. See [Rules & Priority](./rules-and-priority.md) for details. | Rule length | No |
| <a id="middlewares" href="#middlewares" title="#middlewares">`middlewares`</a> | The list of middlewares that are applied to the router. Middlewares are applied in the order they are declared. See [Middlewares overview](../middlewares/overview.md) for available middlewares. | | No |
| <a id="tls" href="#tls" title="#tls">`tls`</a> | TLS configuration for the router. When specified, the router will only handle HTTPS requests. See [TLS overview](../tls/overview.md) for detailed TLS configuration. | | No |
| <a id="observability" href="#observability" title="#observability">`observability`</a> | Observability configuration for the router. Allows fine-grained control over access logs, metrics, and tracing per router. See [Observability](./observability.md) for details. | Inherited from entry points | No |
| <a id="service" href="#service" title="#service">`service`</a> | The name of the service that will handle the matched requests. Services can be load balancer services, weighted round robin, mirroring, or failover services. See [Service](../load-balancing/service.md) for details.| | Yes |


## Router Naming

- The character `@` is not authorized in the router name
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names and rules
