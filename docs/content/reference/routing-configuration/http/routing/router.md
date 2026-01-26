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
        options: "modern"
        domains:
          - main: "example.com"
            sans:
              - "www.example.com"
      observability:
        metrics: true
        accessLogs: true
        tracing: true
      parentRefs:
        - "parent-router-1"
        - "parent-router-2"
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
    parentRefs = ["parent-router-1", "parent-router-2"]

    [http.routers.my-router.tls]
      certResolver = "letsencrypt"
      options = "modern"

      [[http.routers.my-router.tls.domains]]
        main = "example.com"
        sans = ["www.example.com"]

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
  - "traefik.http.routers.my-router.tls.options=modern"
  - "traefik.http.routers.my-router.tls.domains[0].main=example.com"
  - "traefik.http.routers.my-router.tls.domains[0].sans=www.example.com"
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
    "traefik.http.routers.my-router.tls.options=modern",
    "traefik.http.routers.my-router.tls.domains[0].main=example.com",
    "traefik.http.routers.my-router.tls.domains[0].sans=www.example.com",
    "traefik.http.routers.my-router.observability.metrics=true",
    "traefik.http.routers.my-router.observability.accessLogs=true",
    "traefik.http.routers.my-router.observability.tracing=true",
  ]
}
```

## Configuration Options

| Field                                                                                                          | Description                                                                                                                                                                                                                                                                                                          | Default                     | Required |
|----------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------|----------|
| <a id="opt-entryPoints" href="#opt-entryPoints" title="#opt-entryPoints">`entryPoints`</a> | The list of entry points to which the router is attached. If not specified, HTTP routers are attached to all entry points.                                                                                                                                                                                           | All entry points            | No       |
| <a id="opt-rule" href="#opt-rule" title="#opt-rule">`rule`</a> | Rules are a set of matchers configured with values, that determine if a particular request matches specific criteria. If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service. See [Rules & Priority](./rules-and-priority.md) for details.              |                             | Yes      |
| <a id="opt-priority" href="#opt-priority" title="#opt-priority">`priority`</a> | To avoid path overlap, routes are sorted, by default, in descending order using rules length. The priority is directly equal to the length of the rule, and so the longest length has the highest priority. A value of `0` for the priority is ignored. Negative values are supported. See [Rules & Priority](./rules-and-priority.md) for details. | Rule length                 | No       |
| <a id="opt-middlewares" href="#opt-middlewares" title="#opt-middlewares">`middlewares`</a> | The list of middlewares that are applied to the router. Middlewares are applied in the order they are declared. See [Middlewares overview](../middlewares/overview.md) for available middlewares.                                                                                                                    |                             | No       |
| <a id="opt-tls" href="#opt-tls" title="#opt-tls">`tls`</a> | TLS configuration for the router. When specified, the router will only handle HTTPS requests.                                                                                                                                                                                                                        |                             | No       |
| <a id="opt-tls-certResolver" href="#opt-tls-certResolver" title="#opt-tls-certResolver">`tls.certResolver`</a> | The name of the certificate resolver to use for automatic certificate generation. See [Certificate Resolver](../tls/overview.md#certificate-resolver) for details.                                                                                                                                                   |                             | No       |
| <a id="opt-tls-options" href="#opt-tls-options" title="#opt-tls-options">`tls.options`</a> | The name of the TLS options to use for configuring TLS parameters (cipher suites, min/max TLS version, client authentication, etc.). See [TLS Options](../tls/tls-options.md) for detailed configuration.                                                                                                            | `default`                   | No       |
| <a id="opt-tls-domains" href="#opt-tls-domains" title="#opt-tls-domains">`tls.domains`</a> | List of domains and Subject Alternative Names (SANs) for explicit certificate domain specification. When using ACME certificate resolvers, domains are automatically extracted from router rules, making this option optional.                                                                                       |                             | No       |
| <a id="opt-observability" href="#opt-observability" title="#opt-observability">`observability`</a> | Observability configuration for the router. Allows fine-grained control over access logs, metrics, and tracing per router. See [Observability](./observability.md) for details.                                                                                                                                      | Inherited from entry points | No       |
| <a id="opt-parentRefs" href="#opt-parentRefs" title="#opt-parentRefs">`parentRefs`</a> | References to parent router names for multi-layer routing. When specified, this router becomes a child router that processes requests after parent routers have applied their middlewares. See [Multi-Layer Routing](../routing/multi-layer-routing.md) for details.                                        |                             | No       |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The name of the service that will handle the matched requests. Services can be load balancer services, weighted round robin, mirroring, or failover services. See [Service](../load-balancing/service.md) for details.                                                                                               |                             | Yes      |

## Router Naming

- The character `@` is not authorized in the router name
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names and rules

{% include-markdown "includes/traefik-for-business-applications.md" %}
