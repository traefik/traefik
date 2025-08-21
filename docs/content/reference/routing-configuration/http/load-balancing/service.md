---
title: "Traefik HTTP Services Documentation"
description: "A service is in charge of connecting incoming requests to the Servers that can handle them. Read the technical documentation."
---

## Service Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs.

Each service has a load-balancer, even if there is only one server to forward traffic to.

## Configuration Example

```yaml tab="Structured (YAML)"
http:
  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://private-ip-server-1/"
            weight: 2
            preservePath: true
        sticky:
          cookie:
            name: "sticky-cookie"
        healthcheck:
          path: "/health"
          interval: "10s"
          timeout: "3s"
        passiveHealthcheck:
          failureWindow: "3s"
          maxFailedAttempts: "3"
        passHostHeader: true
        serversTransport: "customTransport@file"
        responseForwarding:
          flushInterval: "150ms"
```

```toml tab="Structured (TOML)"
[http.services]
  [http.services.my-service.loadBalancer]
    [[http.services.my-service.loadBalancer.servers]]
      url = "http://private-ip-server-1/"
    
    [http.services.my-service.loadBalancer.sticky.cookie]
      name = "sticky-cookie"

    [http.services.my-service.loadBalancer.healthcheck]
      path = "/health"
      interval = "10s"
      timeout = "3s"

    [http.services.my-service.loadBalancer.passiveHealthcheck]
      failureWindow = "3s"
      maxFailedAttempts = "3"
    
    passHostHeader = true
    serversTransport = "customTransport@file"

    [http.services.my-service.loadBalancer.responseForwarding]
      flushInterval = "150ms"
```

```yaml tab="Labels"
labels:
  - "traefik.http.services.my-service.loadBalancer.servers[0].url=http://private-ip-server-1/"
  - "traefik.http.services.my-service.loadBalancer.servers[0].weight=2"
  - "traefik.http.services.my-service.loadBalancer.servers[0].preservePath=true"
  - "traefik.http.services.my-service.loadBalancer.sticky.cookie.name=sticky-cookie"
  - "traefik.http.services.my-service.loadBalancer.healthcheck.path=/health"
  - "traefik.http.services.my-service.loadBalancer.healthcheck.interval=10s"
  - "traefik.http.services.my-service.loadBalancer.healthcheck.timeout=3s"
  - "traefik.http.services.my-service.loadBalancer.passiveHealthcheck.failureWindow=3s"
  - "traefik.http.services.my-service.loadBalancer.passiveHealthcheck.maxFailedAttempts=3"
  - "traefik.http.services.my-service.loadBalancer.passHostHeader=true"
  - "traefik.http.services.my-service.loadBalancer.serversTransport=customTransport@file"
  - "traefik.http.services.my-service.loadBalancer.responseForwarding.flushInterval=150ms"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.http.services.my-service.loadBalancer.servers[0].url=http://private-ip-server-1/",
    "traefik.http.services.my-service.loadBalancer.servers[0].weight=2",
    "traefik.http.services.my-service.loadBalancer.servers[0].preservePath=true",
    "traefik.http.services.my-service.loadBalancer.sticky.cookie.name=sticky-cookie",
    "traefik.http.services.my-service.loadBalancer.healthcheck.path=/health",
    "traefik.http.services.my-service.loadBalancer.healthcheck.interval=10s",
    "traefik.http.services.my-service.loadBalancer.healthcheck.timeout=3s",
    "traefik.http.services.my-service.loadBalancer.passiveHealthcheck.failureWindow=3s",
    "traefik.http.services.my-service.loadBalancer.passiveHealthcheck.maxFailedAttempts=3",
    "traefik.http.services.my-service.loadBalancer.passHostHeader=true",
    "traefik.http.services.my-service.loadBalancer.serversTransport=customTransport@file",
    "traefik.http.services.my-service.loadBalancer.responseForwarding.flushInterval=150ms"
  ]
}
```

### Configuration Options

| Field                              | Description                                                                                                                                                                                                                                                                                                                                                                                   | Required |
|------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `servers`                          | Represents individual backend instances for your service                                                                                                                                                                                                                                                                                                                                      | Yes      |
| `sticky`                           | Defines a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.                                                                                                                                                                                                                                                                  | No       |
| `healthcheck`                      | Configures health check to remove unhealthy servers from the load balancing rotation.                                                                                                                                                                                                                                                                                                         | No       |
| `passiveHealthcheck`               | Configures the passive health check to remove unhealthy servers from the load balancing rotation.                                                                                                                                                                                                                                                                                             | No       |
| `passHostHeader`                   | Allows forwarding of the client Host header to server. By default, `passHostHeader` is true.                                                                                                                                                                                                                                                                                                  | No       |
| `serversTransport`                 | Allows to reference an [HTTP ServersTransport](./serverstransport.md) configuration for the communication between Traefik and your servers. If no `serversTransport` is specified, the `default@internal` will be used.                                                                                                                                                                       | No       |
| `responseForwarding`               | Configures how Traefik forwards the response from the backend server to the client.                                                                                                                                                                                                                                                                                                           | No       |
| `responseForwarding.FlushInterval` | Specifies the interval in between flushes to the client while copying the response body. It is a duration in milliseconds, defaulting to 100ms. A negative value means to flush immediately after each write to the client. The `FlushInterval` is ignored when ReverseProxy recognizes a response as a streaming response; for such responses, writes are flushed to the client immediately. | No       |

#### Servers

Servers represent individual backend instances for your service. The [service loadBalancer](#service-load-balancer) `servers` option lets you configure the list of instances that will handle incoming requests.

##### Configuration Options

| Field          | Description                                        | Required                                                                         |
|----------------|----------------------------------------------------|----------------------------------------------------------------------------------|
| `url`          | Points to a specific instance.                     | Yes for File provider, No for [Docker provider](../../other-providers/docker.md) |
| `weight`       | Allows for weighted load balancing on the servers. | No                                                                               |
| `preservePath` | Allows to preserve the URL path.                   | No                                                                               |

#### Health Check

The `healthcheck` option configures health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider HTTP(s) servers healthy as long as they return a status code to the health check request (carried out every interval) between `2XX` and `3XX`, or matching the configured status.
For gRPC servers, Traefik will consider them healthy as long as they return SERVING to [gRPC health check v1 requests](https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

To propagate status changes (e.g. all servers of this service are down) upwards, HealthCheck must also be enabled on the parent(s) of this service.

Below are the available options for the health check mechanism:

| Field               | Description                                                                                                                   | Default | Required |
|---------------------|-------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| `path`              | Defines the server URL path for the health check endpoint.                                                                    | ""      | Yes      |
| `scheme`            | Replaces the server URL scheme for the health check endpoint.                                                                 |         | No       |
| `mode`              | If defined to `grpc`, will use the gRPC health check protocol to probe the server.                                            | http    | No       |
| `hostname`          | Defines the value of hostname in the Host header of the health check request.                                                 | ""      | No       |
| `port`              | Replaces the server URL port for the health check endpoint.                                                                   |         | No       |
| `interval`          | Defines the frequency of the health check calls for healthy targets.                                                          | 30s     | No       |
| `unhealthyInterval` | Defines the frequency of the health check calls for unhealthy targets. When not defined, it defaults to the `interval` value. | 30s     | No       |
| `timeout`           | Defines the maximum duration Traefik will wait for a health check request before considering the server unhealthy.            | 5s      | No       |
| `headers`           | Defines custom headers to be sent to the health check endpoint.                                                               |         | No       |
| `followRedirects`   | Defines whether redirects should be followed during the health check calls.                                                   | true    | No       |
| `hostname`          | Defines the value of hostname in the Host header of the health check request.                                                 | ""      | No       |
| `method`            | Defines the HTTP method that will be used while connecting to the endpoint.                                                   | GET     | No       |
| `status`            | Defines the expected HTTP status code of the response to the health check request.                                            |         | No       |

#### Passive Health Check

The `passiveHealthcheck` option configures passive health check to remove unhealthy servers from the load balancing rotation.

Passive health checks rely on real traffic to assess server health.
Traefik forwards requests as usual and evaluates each response or timeout,
incrementing a failure counter whenever a request fails.
If the number of successive failures within a specified time window exceeds the configured threshold,
Traefik will automatically stop routing traffic to that server until it recovers.
A server will be considered healthy again after the configured failure window has passed.

Below are the available options for the passive health check mechanism:

| Field               | Description                                                                                                                                                                         | Default | Required |
|---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| `failureWindow`     | Defines the time window during which the failed attempts must occur for the server to be marked as unhealthy. It also defines for how long the server will be considered unhealthy. | 10s     | No       |
| `maxFailedAttempts` | Defines the number of consecutive failed attempts allowed within the failure window before marking the server as unhealthy.                                                         | 1       | No       |

## Weighted Round Robin (WRR)

The WRR is able to load balance the requests between multiple services based on weights.

This strategy is only available to load balance between services and not between servers.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-ingress.md) providers. To load balance between servers based on weights, the Load Balancer service should be used instead.

```yaml tab="Structured (YAML)"
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

```toml tab="Structured (TOML)"
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

```yaml tab="Structured (YAML)"
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

```toml tab="Structured (TOML)"
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

The mirroring is able to mirror requests sent to a service to other services. Please note that by default the whole request is buffered in memory while it is being mirrored. See the `maxBodySize` option in the example below for how to modify this behaviour. You can also omit the request body by setting the `mirrorBody` option to false.

!!! warning "Default behavior of `percent`"

    When configuring a `mirror` service, if the `percent` field is not set, it defaults to `0`, meaning **no traffic will be sent to the mirror**.
    
!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-ingress.md) providers.
    
```yaml tab="Structured (YAML)"
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
          # Percent defines the percentage of requests that should be mirrored.
          # Default value is 0, which means no traffic will be sent to the mirror.
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

```toml tab="Structured (TOML)"
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

```yaml tab="Structured (YAML)"
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

```toml tab="Structured (TOML)"
## Dynamic configuration
[http.services]
  [http.services.mirrored-api]
    [http.services.mirrored-api.mirroring]
      service = "appv1"
      [http.services.mirrored-api.mirroring.healthCheck]
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

```yaml tab="Structured (YAML)"
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

```toml tab="Structured (TOML)"
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
