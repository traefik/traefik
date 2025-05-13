---
title: "Traefik UDP Routers Rules & Priority Documentation"
description: "In Traefik Proxy, a router is in charge of connecting incoming requests to the Services that can handle them. Read the technical documentation."
---

A router is in charge of connecting incoming requests to the services that can handle them.
In the process, routers may use pieces of [middleware](../../http/middlewares/overview.md) to update the request,
or act before forwarding the request to the service.

Similarly to TCP, as UDP is the transport layer, there is no concept of a request,
so there is no notion of an URL path prefix to match an incoming UDP packet with.
Furthermore, as there is no good TLS support at the moment for multiple hosts,
there is no Host SNI notion to match against either.
Therefore, there is no criterion that could be used as a rule to match incoming packets in order to route them.
So UDP _routers_ at this time are pretty much only load-balancers in one form or another.

!!! tip
    UDP routers can only target UDP services (and not HTTP or TCP services).

## Sessions and timeout

Even though UDP is connectionless (and because of that),
the implementation of an UDP router in Traefik relies on what we (and a couple of other implementations) call a `session`.
It means that some state is kept about an ongoing communication between a client and a backend,
notably so that the proxy knows where to forward a response packet from a backend.

As expected, a `timeout` is associated to each of these sessions,
so that they get cleaned out if they go through a period of inactivity longer than a given duration.

Timeout can be configured using the `entryPoints.name.udp.timeout` option as described under [EntryPoints](../../../install-configuration/entrypoints.md)

## EntryPoints

If not specified, UDP routers will accept packets from all defined (UDP) EntryPoints. If one wants to limit the router scope to a set of EntryPoints, one should set the `entryPoints` option.

## Configuration Example

Listens to Every Entry Point

```yaml tab="Structured (YAML)"
udp:
  routers:
    Router-1:
      # By default, routers listen to all UDP entrypoints
      # i.e. "other", and "streaming".
      service: "service-1"
```

```toml tab="Structured (TOML)"
[udp.routers]
  [udp.routers.Router-1]
    # By default, routers listen to all UDP entrypoints,
    # i.e. "other", and "streaming".
    service = "service-1"
```

```yaml tab="Labels"
labels:
  - "traefik.udp.routers.Router-1.service=service-1"
```

```json tab="Tags"
{
  //...
  "Tags": [
    "traefik.udp.routers.Router-1.service=service-1"
  ]
}
```

Listens to Specific EntryPoints

```yaml tab="Structured (YAML)"
udp:
  routers:
    Router-1:
      # does not listen on "other" entry point
      entryPoints:
        - "streaming"
      service: "service-1"
```

```toml tab="Structured (TOML)"
[udp.routers]
  [udp.routers.Router-1]
    # does not listen on "other" entry point
    entryPoints = ["streaming"]
    service = "service-1"
```

```yaml tab="Labels"
labels:
  - "traefik.udp.routers.Router-1.entryPoints=streaming"
  - "traefik.udp.routers.Router-1.service=service-1"
```

```json tab="Tags"
{
  //...
  "Tags": [
    "traefik.udp.routers.Router-1.entryPoints=streaming",
    "traefik.udp.routers.Router-1.service=service-1"
  ]
}
```

!!! info "Service"

    There must be one (and only one) UDP [service](../service.md) referenced per UDP router.
    Services are the target for the router.

{!traefik-for-business-applications.md!}
