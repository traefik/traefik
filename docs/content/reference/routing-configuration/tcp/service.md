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
| <a id="servers" href="#servers" title="#servers">`servers`</a> |  Servers declare a single instance of your program.  | "" |
| <a id="servers-address" href="#servers-address" title="#servers-address">`servers.address`</a> |   The address option (IP:Port) point to a specific instance. | "" |
| <a id="servers-tls" href="#servers-tls" title="#servers-tls">`servers.tls`</a> | The `tls` option determines whether to use TLS when dialing with the backend. | false |
| <a id="servers-serversTransport" href="#servers-serversTransport" title="#servers-serversTransport">`servers.serversTransport`</a> | `serversTransport` allows to reference a TCP [ServersTransport](./serverstransport.md configuration for the communication between Traefik and your servers. If no serversTransport is specified, the default@internal will be used. |  "" |

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
 
