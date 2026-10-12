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
        readTimeout: "30s"
        writeTimeout: "30s"
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
      readTimeout = "30s"
      writeTimeout = "30s"
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
  - "traefik.http.routers.my-router.respondingtimeouts.readtimeout=30s"
  - "traefik.http.routers.my-router.respondingtimeouts.writetimeout=30s"
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
    "traefik.http.routers.my-router.respondingtimeouts.readtimeout=30s",
    "traefik.http.routers.my-router.respondingtimeouts.writetimeout=30s",
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
| <a id="opt-tls-options" href="#opt-tls-options" title="#opt-tls-options">`tls.options`</a> | The name of the TLS options to use for configuring TLS parameters (cipher suites, min/max TLS version, client authentication, etc.). They are mapped to the host names of the router rule, and not to the router itself, so all the routers serving the same host name on the same entry point must reference the same TLS options. See [TLS Options](../tls/tls-options.md) for detailed configuration, and [Conflicting TLS Options](../tls/tls-options.md#conflicting-tls-options) for the consequences of not doing so.                                                                                                            | `default`                   | No       |
| <a id="opt-tls-domains" href="#opt-tls-domains" title="#opt-tls-domains">`tls.domains`</a> | List of domains and Subject Alternative Names (SANs) for explicit certificate domain specification. When using ACME certificate resolvers, domains are automatically extracted from router rules, making this option optional.                                                                                       |                             | No       |
| <a id="opt-observability" href="#opt-observability" title="#opt-observability">`observability`</a> | Observability configuration for the router. Allows fine-grained control over access logs, metrics, and tracing per router. See [Observability](./observability.md) for details.                                                                                                                                      | Inherited from entry points | No       |
| <a id="opt-observability-traceVerbosity" href="#opt-observability-traceVerbosity" title="#opt-observability-traceVerbosity">`observability.traceVerbosity`</a> | Defines the verbosity level of tracing for this router. Accepted values are `minimal` and `detailed`.                                                                                                                                                                                                               | `minimal`                   | No       |
| <a id="opt-respondingTimeouts-readTimeout" href="#opt-respondingTimeouts-readTimeout" title="#opt-respondingTimeouts-readTimeout">`respondingTimeouts.readTimeout`</a> | The maximum duration for reading the request body, from when the router receives the request. `0` means no timeout, and lifts the entry point `readTimeout` for the router. When unset, the entry point `readTimeout` applies. See [Responding Timeouts](#responding-timeouts) for details. | Entry point `readTimeout`   | No       |
| <a id="opt-respondingTimeouts-writeTimeout" href="#opt-respondingTimeouts-writeTimeout" title="#opt-respondingTimeouts-writeTimeout">`respondingTimeouts.writeTimeout`</a> | The maximum duration before timing out writes of the response, from when the router receives the request. It includes the wait for the backend: on expiry before the response has started, the client receives a `504 Gateway Timeout`; afterwards, the connection is closed. `0` means no timeout, and lifts the entry point `writeTimeout` for the router. When unset, the entry point `writeTimeout` applies. See [Responding Timeouts](#responding-timeouts) for details. | Entry point `writeTimeout`  | No       |
| <a id="opt-parentRefs" href="#opt-parentRefs" title="#opt-parentRefs">`parentRefs`</a> | References to parent router names for multi-layer routing. When specified, this router becomes a child router that processes requests after parent routers have applied their middlewares. See [Multi-Layer Routing](../routing/multi-layer-routing.md) for details.                                        |                             | No       |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The name of the service that will handle the matched requests. Services can be load balancer services, weighted round robin, mirroring, or failover services. See [Service](../load-balancing/service.md) for details.                                                                                               |                             | Yes      |

## Responding Timeouts

The router `respondingTimeouts` mirror the `readTimeout` and `writeTimeout` of the entry point [`respondingTimeouts`](../../../install-configuration/entrypoints.md#opt-transport-respondingTimeouts-readTimeout), for the requests matched by the router.
Both count from when the router receives the request, right after the request headers are parsed. A router only exists once the headers are read, so the time spent reading them stays bounded by the entry point timeouts only.

- `readTimeout` bounds the request body upload. Unlike the entry point one, it does not cover the request headers.
- `writeTimeout` bounds the backend processing (including all retry attempts) and the response delivery.
  When it expires **before** the response has started, the client receives a `504 Gateway Timeout`, where the entry point `writeTimeout` would close the connection.
  When it expires **after** the response has started (streaming), the transaction is torn down and the connection is closed.

Both timeouts count from the same instant, so setting them to the same value bounds the whole transaction (client → proxy → backend → proxy → client) with a single deadline.
This is how the Kubernetes Gateway API HTTPRoute `timeouts.request` is mapped.

**Interaction with entry point timeouts.** For requests matched by the router, each router timeout **replaces** the matching entry point one, in both directions: a longer timeout extends the window for requests matched by that router, while a shorter one tightens it.
A timeout of `0` lifts the matching entry point timeout for requests matched by the router, without setting any timeout of its own.
An unset timeout leaves the entry point one in force.

!!! warning "A long `readTimeout` relaxes the entry point's slow-client protection"

    The entry point `readTimeout` (default `60s`) caps how long a client may take to send its request, which limits slow-client attacks such as Slowloris. A router `readTimeout` longer than the entry point one lifts that cap for the matched requests: a slow client on that route can then hold a connection open for the whole `readTimeout` duration. Because connection limits are enforced per entry point, that cost is borne by everything served by the entry point, not by the router alone. A `readTimeout` of `0` lifts that cap entirely. Set a long or zero `readTimeout` only on routes that need it.

**Protocol upgrades (WebSocket, `kubectl exec`, ...).** The router `writeTimeout` does not apply to upgrade requests (any `Connection: Upgrade` protocol): an established tunnel is never torn down by this timeout, and the entry point `writeTimeout` still applies to the handshake.

**Streaming responses (SSE, gRPC streaming, ...).** Response-streaming routes are not detectable at request time, so the `writeTimeout` applies to them as well: omit it (or set it to `0`) on such routers.

**Multi-layer routing.** When a parent router and a child router both define `respondingTimeouts`, the most restrictive deadline wins, for each timeout: a child router cannot extend the budget set by its parent.

!!! warning "Experimental `fast` proxy"

    When the experimental [`fast` proxy](../../../install-configuration/experimental/fastproxy.md) is enabled, the timeout is not enforced mid-flight on the backend leg.

## Router Naming

- The character `@` is not authorized in the router name
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names and rules

{% include-markdown "includes/traefik-for-business-applications.md" %}
