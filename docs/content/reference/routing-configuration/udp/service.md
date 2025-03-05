---
title: "Traefik UDP Services Documentation"
description: "A service is in charge of connecting incoming requests to the Servers that can handle them. Read the technical documentation. Read the technical documentation."
--- 

### General

Each of the fields of the service section represents a kind of service.
Which means, that for each specified service, one of the fields, and only one,
has to be enabled to define what kind of service is created.
Currently, the available kind is `LoadBalancer`.

## Servers Load Balancer

The servers load balancer is in charge of balancing the requests between the servers of the same service.

### Servers

The Servers field defines all the servers that are part of this load-balancing group,
i.e. each address (IP:Port) on which an instance of the service's program is deployed.

#### Configuration Example

A Service with One Server -- Using the [File Provider](../../install-configuration/providers/others/file.md)

```yaml tab="Structured (YAML)"
## Dynamic configuration
udp:
  services:
    my-service:
      loadBalancer:
        servers:
          - address: "xx.xx.xx.xx:xx"
```

```toml tab="Structured (TOML)"
## Dynamic configuration
[udp.services]
  [udp.services.my-service.loadBalancer]
    [[udp.services.my-service.loadBalancer.servers]]
      address = "xx.xx.xx.xx:xx"
```

{!traefik-for-business-applications.md!}
