---
title: "Traefik HTTP Services Documentation"
description: "A service is in charge of connecting incoming requests to the Servers that can handle them. Read the technical documentation."
--- 

## Servers

Servers declare a single instance of your program.

### Configuration Options

| Field | Description                                 | Required |
|----------|------------------------------------------|----------|
|`url`| Points to a specific instance. | Yes for File provider, No for [Docker provider](../../other-providers/docker.md) |
|`weight`| Allows for weighted load balancing on the servers. | No |
|`preservePath`| Allows to preserve the URL path. | No |

### Configuration Examples

```yaml tab="A Service with One Server"
## Dynamic configuration
http:
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://private-ip-server-1/"
```

```yaml tab="A Service with Two Servers with Weight"
## Dynamic configuration
http:
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://private-ip-server-1/"
            weight: 2
          - url: "http://private-ip-server-2/"
            weight: 1
```

```yaml tab="A Service with One Server and PreservePath"
## Dynamic configuration
http:
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://private-ip-server-1/base"
            preservePath: true
```

## Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs.

The example below declares a service with two servers (with load balancing):

```yaml tab="YAML"
## Dynamic configuration
http:
  services:
    my-service:
      loadBalancer:
        servers:
        - url: "http://private-ip-server-1/"
        - url: "http://private-ip-server-2/"
```

```toml tab="TOML"
## Dynamic configuration
[http.services]
  [http.services.my-service.loadBalancer]
    [[http.services.my-service.loadBalancer.servers]]
      url = "http://private-ip-server-1/"
    [[http.services.my-service.loadBalancer.servers]]
      url = "http://private-ip-server-2/"
```

## Weighted Round Robin (WRR)

The WRR is able to load balance the requests between multiple services based on weights.

This strategy is only available to load balance between services and not between servers.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-ingress.md) providers. To load balance between servers based on weights, the Load Balancer service should be used instead.

```yaml tab="File (YAML)"

## Dynamic configuration
http:
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
        - url: "http://private-ip-server-1/"

    appv2:
      loadBalancer:
        servers:
        - url: "http://private-ip-server-2/"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.services]
  [http.services.app]
    [[http.services.app.weighted.services]]
      name = "appv1"
      weight = 3
    [[http.services.app.weighted.services]]
      name = "appv2"
      weight = 1

  [http.services.appv1]
    [http.services.appv1.loadBalancer]
      [[http.services.appv1.loadBalancer.servers]]
        url = "http://private-ip-server-1/"

  [http.services.appv2]
    [http.services.appv2.loadBalancer]
      [[http.services.appv2.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```

### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. whenever one of its children is reported as down, this service becomes aware of it, and takes it into account (i.e. it ignores the down child) when running the load-balancing algorithm. In addition, if the parent of this service also has HealthCheck enabled, this service reports to its parent any status change.

!!! note "Behavior"

    If HealthCheck is enabled for a given service and any of its descendants does not have it enabled, the creation of the service will fail.

    HealthCheck on Weighted services can be defined currently only with the [File provider](../../../install-configuration/providers/others/file.md).  

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  services:
    app:
      weighted:
        healthCheck: {}
        services:
        - name: appv1
          weight: 3
        - name: appv2
          weight: 1

    appv1:
      loadBalancer:
        healthCheck:
          path: /status
          interval: 10s
          timeout: 3s
        servers:
        - url: "http://private-ip-server-1/"

    appv2:
      loadBalancer:
        healthCheck:
          path: /status
          interval: 10s
          timeout: 3s
        servers:
        - url: "http://private-ip-server-2/"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.services]
  [http.services.app]
    [http.services.app.weighted.healthCheck]
    [[http.services.app.weighted.services]]
      name = "appv1"
      weight = 3
    [[http.services.app.weighted.services]]
      name = "appv2"
      weight = 1

  [http.services.appv1]
    [http.services.appv1.loadBalancer]
      [http.services.appv1.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.appv1.loadBalancer.servers]]
        url = "http://private-ip-server-1/"

  [http.services.appv2]
    [http.services.appv2.loadBalancer]
      [http.services.appv2.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.appv2.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```

## Mirroring

The mirroring is able to mirror requests sent to a service to other services. Please note that by default the whole request is buffered in memory while it is being mirrored. See the maxBodySize option in the example below for how to modify this behaviour. You can also omit the request body by setting the mirrorBody option to false.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-ingress.md) providers.
    
```yaml tab="File (YAML)"
## Dynamic configuration
http:
  services:
    mirrored-api:
      mirroring:
        service: appv1
        # mirrorBody defines whether the request body should be mirrored.
        # Default value is true.
        mirrorBody: false
        # maxBodySize is the maximum size allowed for the body of the request.
        # If the body is larger, the request is not mirrored.
        # Default value is -1, which means unlimited size.
        maxBodySize: 1024
        mirrors:
        - name: appv2
          percent: 10

    appv1:
      loadBalancer:
        servers:
        - url: "http://private-ip-server-1/"

    appv2:
      loadBalancer:
        servers:
        - url: "http://private-ip-server-2/
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.services]
  [http.services.mirrored-api]
    [http.services.mirrored-api.mirroring]
      service = "appv1"
      # maxBodySize is the maximum size in bytes allowed for the body of the request.
      # If the body is larger, the request is not mirrored.
      # Default value is -1, which means unlimited size.
      maxBodySize = 1024
      # mirrorBody defines whether the request body should be mirrored.
      # Default value is true.
      mirrorBody = false
    [[http.services.mirrored-api.mirroring.mirrors]]
      name = "appv2"
      percent = 10

  [http.services.appv1]
    [http.services.appv1.loadBalancer]
      [[http.services.appv1.loadBalancer.servers]]
        url = "http://private-ip-server-1/"

  [http.services.appv2]
    [http.services.appv2.loadBalancer]
      [[http.services.appv2.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```

### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. if the main handler of the service becomes unreachable, the information is propagated upwards to its parent.

!!! note "Behavior"

    If HealthCheck is enabled for a given service and any of its descendants does not have it enabled, the creation of the service will fail.

    HealthCheck on Mirroring services can be defined currently only with the [File provider](../../../install-configuration/providers/others/file.md).  

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  services:
    mirrored-api:
      mirroring:
        healthCheck: {}
        service: appv1
        mirrors:
        - name: appv2
          percent: 10

    appv1:
      loadBalancer:
        healthCheck:
          path: /status
          interval: 10s
          timeout: 3s
        servers:
        - url: "http://private-ip-server-1/"

    appv2:
      loadBalancer:
        servers:
        - url: "http://private-ip-server-2/"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.services]
  [http.services.mirrored-api]
    [http.services.mirrored-api.mirroring]
      [http.services.mirrored-api.mirroring.healthCheck]
      service = "appv1"
    [[http.services.mirrored-api.mirroring.mirrors]]
      name = "appv2"
      percent = 10

  [http.services.appv1]
    [http.services.appv1.loadBalancer]
      [http.services.appv1.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.appv1.loadBalancer.servers]]
        url = "http://private-ip-server-1/"

  [http.services.appv2]
    [http.services.appv2.loadBalancer]
      [http.services.appv1.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.appv2.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```

## Failover 

A failover service job is to forward all requests to a fallback service when the main service becomes unreachable.

!!! info "Relation to HealthCheck"
    The failover service relies on the HealthCheck system to get notified when its main service becomes unreachable, which means HealthCheck needs to be enabled and functional on the main service. However, HealthCheck does not need to be enabled on the failover service itself for it to be functional. It is only required in order to propagate upwards the information when the failover itself becomes down (i.e. both its main and its fallback are down too).

!!! info "Supported Provider"
    This strategy can currently only be defined with the [File](../../../install-configuration/providers/others/file.md) provider.

### HealthCheck

HealthCheck enables automatic self-healthcheck for this service, i.e. if the main and the fallback services become unreachable, the information is propagated upwards to its parent.

!!! note "Behavior"

    If HealthCheck is enabled for a given service and any of its descendants does not have it enabled, the creation of the service will fail.

    HealthCheck on a Failover service can be defined currently only with the [File provider](../../../install-configuration/providers/others/file.md).  

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  services:
    app:
      failover:
        healthCheck: {}
        service: main
        fallback: backup

    main:
      loadBalancer:
        healthCheck:
          path: /status
          interval: 10s
          timeout: 3s
        servers:
        - url: "http://private-ip-server-1/"

    backup:
      loadBalancer:
        healthCheck:
          path: /status
          interval: 10s
          timeout: 3s
        servers:
        - url: "http://private-ip-server-2/"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.services]
  [http.services.app]
    [http.services.app.failover.healthCheck]
    [http.services.app.failover]
      service = "main"
      fallback = "backup"

  [http.services.main]
    [http.services.main.loadBalancer]
      [http.services.main.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.main.loadBalancer.servers]]
        url = "http://private-ip-server-1/"

  [http.services.backup]
    [http.services.backup.loadBalancer]
      [http.services.backup.loadBalancer.healthCheck]
        path = "/health"
        interval = "10s"
        timeout = "3s"
      [[http.services.backup.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```
