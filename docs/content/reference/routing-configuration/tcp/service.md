---
title: "Traefik TCP Services Documentation"
description: "A service is in charge of connecting incoming requests to the Servers that can handle them. Read the technical documentation."
--- 

## General

Each of the fields of the service section represents a kind of service. Which means, that for each specified service, one of the fields, and only one, has to be enabled to define what kind of service is created. Currently, the two available kinds are `LoadBalancer`, and `Weighted`.

## Servers Load Balancer

The servers load balancer is in charge of balancing the requests between the servers of the same service.

### Configuration Examples

Declaring a Service with Two Servers -- Using the [File Provider](../../install-configuration/providers/others/file.md)

```yaml tab="Structured (YAML)"
tcp:
  services:
    my-service:
      loadBalancer:
        servers:
        - address: "xx.xx.xx.xx:xx"
        - address: "xx.xx.xx.xx:xx"
```

```toml tab="Structured (TOML)"
[tcp.services]
  [tcp.services.my-service.loadBalancer]
    [[tcp.services.my-service.loadBalancer.servers]]
      address = "xx.xx.xx.xx:xx"
    [[tcp.services.my-service.loadBalancer.servers]]
        address = "xx.xx.xx.xx:xx"
```

## Configuration Options

| Field | Description                                 | Default |
|----------|------------------------------------------|--------- |
| `servers` |  Servers declare a single instance of your program.  | "" |
| `servers.address` |   The address option (IP:Port) point to a specific instance. | "" |
| `servers.tls` | The `tls` option determines whether to use TLS when dialing with the backend. | false |
| `servers.serversTransport` | `serversTransport` allows to reference a TCP [ServersTransport](./serverstransport.md configuration for the communication between Traefik and your servers. If no serversTransport is specified, the default@internal will be used. |  "" |
| `servers.proxyProtocol.version` | Traefik supports PROXY Protocol version 1 and 2 on TCP Services. More Information [here](#serversproxyprotocolversion) |  2 |

### servers.proxyProtocol.version

Traefik supports [PROXY Protocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2 on TCP Services. It can be enabled by setting `proxyProtocol` on the load balancer.
The option specifies the version of the protocol to be used. Either 1 or 2.

## Weighted Round Robin

The Weighted Round Robin (alias `WRR`) load-balancer of services is in charge of balancing the requests between multiple services based on provided weights.

This strategy is only available to load balance between [services](./service.md) and not between servers.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../install-configuration/providers/others/file.md) or [IngressRoute](../../install-configuration/providers/kubernetes/kubernetes-crd.md) providers.

```yaml tab="Structured (YAML)"
tcp:
  services:
    app:
      weighted:
        services:
        - name: appv1
          weight: 3
        - name: appv2
          weight: 1

    appv1:
      loadBalancer:
        servers:
        - address: "xxx.xxx.xxx.xxx:8080"

    appv2:
      loadBalancer:
        servers:
        - address: "xxx.xxx.xxx.xxx:8080"
```

```toml tab="Structured (TOML)"
[tcp.services]
  [tcp.services.app]
    [[tcp.services.app.weighted.services]]
      name = "appv1"
      weight = 3
    [[tcp.services.app.weighted.services]]
      name = "appv2"
      weight = 1

  [tcp.services.appv1]
    [tcp.services.appv1.loadBalancer]
      [[tcp.services.appv1.loadBalancer.servers]]
        address = "private-ip-server-1:8080/"

  [tcp.services.appv2]
    [tcp.services.appv2.loadBalancer]
      [[tcp.services.appv2.loadBalancer.servers]]
        address = "private-ip-server-2:8080/"
```
 