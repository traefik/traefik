---
title: "Traefik UDP Routers Documentation"
description: "UDP routers are responsible for connecting incoming UDP packets to the services that can handle them. Read the technical documentation."
---

## UDP Router

A UDP router is in charge of connecting incoming UDP packets to the services that can handle them. Unlike HTTP and TCP routers, UDP routers operate at the transport layer and have unique characteristics due to the connectionless nature of UDP.

!!! important "UDP Router Characteristics"
    - UDP is connectionless, so there is no concept of a request URL path or Host SNI to match against
    - UDP routers are essentially load-balancers that distribute packets to backend services
    - UDP routers can only target UDP services (not HTTP or TCP services)
    - Sessions are tracked with configurable timeouts to maintain state between client and backend

## Configuration Example

```yaml tab="Structured (YAML)"
udp:
  routers:
    my-udp-router:
      entryPoints:
        - "udp-ep"
        - "dns"
      service: my-udp-service
```

```toml tab="Structured (TOML)"
[udp.routers]
  [udp.routers.my-udp-router]
    entryPoints = ["udp-ep", "dns"]
    service = "my-udp-service"
```

```yaml tab="Labels"
labels:
  - "traefik.udp.routers.my-udp-router.entrypoints=udp-ep,dns"
  - "traefik.udp.routers.my-udp-router.service=my-udp-service"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.udp.routers.my-udp-router.entrypoints=udp-ep,dns",
    "traefik.udp.routers.my-udp-router.service=my-udp-service"
  ]
}
```

## Configuration Options

| Field                              | Description                                                                                                                                                                                                                                                                                                                                                                                | Default | Required |
|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="entryPoints" href="#entryPoints" title="#entryPoints">`entryPoints`</a> | The list of entry points to which the router is attached. If not specified, UDP routers are attached to all UDP entry points. | All UDP entry points | No |
| <a id="service" href="#service" title="#service">`service`</a> | The name of the service that will handle the matched UDP packets. UDP services are typically load balancer services that distribute packets to multiple backend servers. See [UDP Service](../service.md) for details. | | Yes |

## Sessions and Timeout

Even though UDP is connectionless, Traefik's UDP router implementation relies on sessions to maintain state about ongoing communication between clients and backends. This allows the proxy to know where to forward response packets from backends.

Each session has an associated timeout that cleans up inactive sessions after a specified duration of inactivity.

Session timeout can be configured using the `entryPoints.name.udp.timeout` option in the static configuration. See [EntryPoints documentation](../../install-configuration/entrypoints.md) for details.

## Router Naming

- The character `@` is not authorized in the router name
- Router names should be descriptive and follow your naming conventions
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names
