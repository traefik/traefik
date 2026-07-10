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
      respondingTimeouts:
        roundTrip: "30s"
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

    [http.routers.my-router.respondingTimeouts]
      roundTrip = "30s"
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
  - "traefik.http.routers.my-router.respondingtimeouts.roundtrip=30s"
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
    "traefik.http.routers.my-router.respondingtimeouts.roundtrip=30s",
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
| <a id="opt-observability-traceVerbosity" href="#opt-observability-traceVerbosity" title="#opt-observability-traceVerbosity">`observability.traceVerbosity`</a> | Defines the verbosity level of tracing for this router. Accepted values are `minimal` and `detailed`.                                                                                                                                                                                                               | `minimal`                   | No       |
| <a id="opt-respondingTimeouts-roundTrip" href="#opt-respondingTimeouts-roundTrip" title="#opt-respondingTimeouts-roundTrip">`respondingTimeouts.roundTrip`</a> | The maximum duration for the whole client transaction (client → proxy → backend → proxy → client) on this router. On expiry before the response has started, the client receives a `504 Gateway Timeout`; afterwards, the connection is closed. `0` (or unset) means no timeout. See [Responding Timeouts](#responding-timeouts) for details. | 0 (no timeout)              | No       |
| <a id="opt-parentRefs" href="#opt-parentRefs" title="#opt-parentRefs">`parentRefs`</a> | References to parent router names for multi-layer routing. When specified, this router becomes a child router that processes requests after parent routers have applied their middlewares. See [Multi-Layer Routing](../routing/multi-layer-routing.md) for details.                                        |                             | No       |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The name of the service that will handle the matched requests. Services can be load balancer services, weighted round robin, mirroring, or failover services. See [Service](../load-balancing/service.md) for details.                                                                                               |                             | Yes      |

## Responding Timeouts

The `respondingTimeouts.roundTrip` option bounds the **whole transaction** on a router: the timer starts when the router receives the request (right after the request headers are parsed), and covers the request body upload, the backend processing (including all retry attempts), and the response delivery.
It bounds slow clients too: a slow upload or a slow response read is interrupted at the deadline.

Unlike the entry point [`respondingTimeouts`](../../../install-configuration/entrypoints.md#opt-transport-respondingTimeouts-readTimeout) (which carries `readTimeout`/`writeTimeout`/`idleTimeout`), the router-level `respondingTimeouts` exposes only `roundTrip`.

When the deadline expires **before** the response has started, the client receives a `504 Gateway Timeout`.
When it expires **after** the response has started (streaming), the transaction is torn down and the connection is closed.

**Interaction with entry point timeouts.** For requests matched by the router, `roundTrip` **replaces** the connection deadlines armed from the entry point [`respondingTimeouts`](../../../install-configuration/entrypoints.md#opt-transport-respondingTimeouts-readTimeout) (`readTimeout`/`writeTimeout`), in both directions: a `roundTrip` longer than the entry point timeouts extends the window for requests matched by that router, while a shorter one tightens it.

!!! warning "A long `roundTrip` relaxes the entry point's slow-client protection"

    The entry point `readTimeout` (default `60s`) caps how long a client may take to send its request, which limits slow-client attacks such as Slowloris. A `roundTrip` longer than `readTimeout` lifts that cap for the matched requests: a slow client on that route can then hold a connection open for the whole `roundTrip` duration. Because connection limits are enforced per entry point, that cost is borne by everything served by the entry point, not by the router alone. Set a long `roundTrip` only on routes that need it.

**Protocol upgrades (WebSocket, `kubectl exec`, ...).** For upgrade requests (any `Connection: Upgrade` protocol), the deadline bounds the **handshake only** and is disarmed at the protocol switch: an established tunnel is never torn down by this timeout, while a backend that never answers the handshake still yields a `504`.

**Streaming responses (SSE, gRPC streaming, ...).** Response-streaming routes are not detectable at request time, so the timeout applies to them as well: omit the option (or set it to `0`) on such routers.

**Multi-layer routing.** When a parent router and a child router both define `respondingTimeouts`, the most restrictive deadline wins: a child router cannot extend the budget set by its parent.

!!! warning "Experimental `fast` proxy"

    When the experimental [`fast` proxy](../../../install-configuration/experimental/fastproxy.md) is enabled, the timeout is not enforced mid-flight on the backend leg.

## Router Naming

- The character `@` is not authorized in the router name
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names and rules

{% include-markdown "includes/traefik-for-business-applications.md" %}
