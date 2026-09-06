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
      tls:
        options: "my-tls-options"
```

```toml tab="Structured (TOML)"
[udp.routers]
  [udp.routers.my-udp-router]
    entryPoints = ["udp-ep", "dns"]
    service = "my-udp-service"

    [udp.routers.my-udp-router.tls]
      options = "my-tls-options"
```

```yaml tab="Labels"
labels:
  - "traefik.udp.routers.my-udp-router.entrypoints=udp-ep,dns"
  - "traefik.udp.routers.my-udp-router.service=my-udp-service"
  - "traefik.udp.routers.my-udp-router.tls.options=my-tls-options"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.udp.routers.my-udp-router.entrypoints=udp-ep,dns",
    "traefik.udp.routers.my-udp-router.service=my-udp-service",
    "traefik.udp.routers.my-udp-router.tls.options=my-tls-options"
  ]
}
```

## Configuration Options

| Field                              | Description                                                                                                                                                                                                                                                                                                                                                                                | Default | Required |
|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="opt-entryPoints" href="#opt-entryPoints" title="#opt-entryPoints">`entryPoints`</a> | The list of entry points to which the router is attached. If not specified, UDP routers are attached to all UDP entry points. | All UDP entry points | No |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | The name of the service that will handle the matched UDP packets. UDP services are typically load balancer services that distribute packets to multiple backend servers. See [UDP Service](../service.md) for details. | | Yes |
| <a id="opt-tls" href="#opt-tls" title="#opt-tls">`tls`</a> | Enables DTLS termination for the router. When specified, incoming UDP packets are decrypted before being forwarded to the service. Set `tls.options` to select a named [TLS Options](../../http/tls/tls-options.md) set (minimum protocol version, cipher suites, etc.); an empty `tls: {}` block uses the default TLS Options. Unlike TCP's `tls` field, there is no `passthrough`, `certResolver`, or `domains` support, since UDP routing has no HostSNI-based rule matching yet. | | No |

## Sessions and Timeout

Even though UDP is connectionless, Traefik's UDP router implementation relies on sessions to maintain state about ongoing communication between clients and backends. This allows the proxy to know where to forward response packets from backends.

Each session has an associated timeout that cleans up inactive sessions after a specified duration of inactivity.

Session timeout can be configured using the `entryPoints.name.udp.timeout` option in the static configuration. See [EntryPoints documentation](../../../install-configuration/entrypoints.md) for details.

## TLS Termination

Setting `tls` on a UDP router enables DTLS termination: Traefik performs the DTLS handshake and forwards decrypted UDP payloads to the service.

DTLS termination on UDP routers is scoped down compared to TCP's `tls` configuration:

- There is no `passthrough` mode.
- There is no per-domain certificate selection (`domains`) or `certResolver`, since UDP routing has no HostSNI-based rule matching yet — only one router per entry point is supported.
- Certificate selection follows the same [TLS Options](../../http/tls/tls-options.md) and default-store mechanism used by TCP/HTTP, including the fallback behavior for clients that don't send SNI (common for non-browser DTLS clients such as WebRTC or IoT devices).

## Router Naming

- The character `@` is not authorized in the router name
- Router names should be descriptive and follow your naming conventions
- In provider-specific configurations (Docker, Kubernetes), router names are often auto-generated based on service names
