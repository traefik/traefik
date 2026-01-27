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
        healthCheck:
          send: "PING"
          expect: "PONG"
          interval: "10s"
          timeout: "3s"
        serversTransport: "customTransport@file"
```

```toml tab="Structured (TOML)"
[tcp.services]
  [tcp.services.my-service.loadBalancer]
    [[tcp.services.my-service.loadBalancer.servers]]
      address = "xx.xx.xx.xx:xx"
    [[tcp.services.my-service.loadBalancer.servers]]
        address = "xx.xx.xx.xx:xx"

    [tcp.services.my-service.loadBalancer.healthCheck]
      send = "PING"
      expect = "PONG"
      interval = "10s"
      timeout = "3s"

    serversTransport = "customTransport@file"
```

```yaml tab="Labels"
labels:
  - "traefik.tcp.services.my-service.loadBalancer.servers[0].address=xx.xx.xx.xx:xx"
  - "traefik.tcp.services.my-service.loadBalancer.servers[1].address=xx.xx.xx.xx:xx"
  - "traefik.tcp.services.my-service.loadBalancer.healthCheck.send=PING"
  - "traefik.tcp.services.my-service.loadBalancer.healthCheck.expect=PONG"
  - "traefik.tcp.services.my-service.loadBalancer.healthCheck.interval=10s"
  - "traefik.tcp.services.my-service.loadBalancer.healthCheck.timeout=3s"
  - "traefik.tcp.services.my-service.loadBalancer.serversTransport=customTransport@file"
```

```json tab="Tags"
{
  "Tags": [
    "traefik.tcp.services.my-service.loadBalancer.servers[0].address=xx.xx.xx.xx:xx",
    "traefik.tcp.services.my-service.loadBalancer.servers[1].address=xx.xx.xx.xx:xx",
    "traefik.tcp.services.my-service.loadBalancer.healthCheck.send=PING",
    "traefik.tcp.services.my-service.loadBalancer.healthCheck.expect=PONG",
    "traefik.tcp.services.my-service.loadBalancer.healthCheck.interval=10s",
    "traefik.tcp.services.my-service.loadBalancer.healthCheck.timeout=3s",
    "traefik.tcp.services.my-service.loadBalancer.serversTransport=customTransport@file"
  ]
}
```

#### HTTP Health Check Configuration Example

You can also configure health checks using HTTP mode:

```yaml tab="Structured (YAML)"
tcp:
  services:
    my-service:
      loadBalancer:
        servers:
        - address: "xx.xx.xx.xx:8080"
        healthCheck:
          mode: "HTTP"
          httpHealthCheck:
            path: "/health"
            method: "GET"
            status: 200
          interval: "10s"
          timeout: "3s"
```

```toml tab="Structured (TOML)"
[tcp.services]
  [tcp.services.my-service.loadBalancer]
    [[tcp.services.my-service.loadBalancer.servers]]
      address = "xx.xx.xx.xx:8080"

    [tcp.services.my-service.loadBalancer.healthCheck]
      mode = "HTTP"
      interval = "10s"
      timeout = "3s"

      [tcp.services.my-service.loadBalancer.healthCheck.httpHealthCheck]
        path = "/health"
        method = "GET"
        status = 200
```

### Configuration Options

| Field | Description                                 | Default |
|----------|------------------------------------------|--------- |
| <a id="opt-servers" href="#opt-servers" title="#opt-servers">`servers`</a> |  Servers declare a single instance of your program.  | "" |
| <a id="opt-servers-address" href="#opt-servers-address" title="#opt-servers-address">`servers.address`</a> |   The address option (IP:Port) point to a specific instance. | "" |
| <a id="opt-servers-tls" href="#opt-servers-tls" title="#opt-servers-tls">`servers.tls`</a> | The `tls` option determines whether to use TLS when dialing with the backend. | false |
| <a id="opt-serversTransport" href="#opt-serversTransport" title="#opt-serversTransport">`serversTransport`</a> | `serversTransport` allows to reference a TCP [ServersTransport](./serverstransport.md) configuration for the communication between Traefik and your servers. If no serversTransport is specified, the default@internal will be used. |  "" |
| <a id="opt-healthCheck" href="#opt-healthCheck" title="#opt-healthCheck">`healthCheck`</a> | Configures health check to remove unhealthy servers from the load balancing rotation. See [HealthCheck](#health-check) for details. | | No |

### Health Check

The `healthCheck` option configures health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider TCP servers healthy as long as the connection to the target server succeeds.
For advanced health checks, you can configure TCP payload exchange by specifying `send` and `expect` parameters.

To propagate status changes (e.g. all servers of this service are down) upwards, HealthCheck must also be enabled on the parent(s) of this service.

Below are the available options for the health check mechanism:

| Field | Description | Default | Required |
|-------|-------------|---------|----------|
| <a id="opt-port" href="#opt-port" title="#opt-port">`port`</a> | Replaces the server address port for the health check endpoint. | | No |
| <a id="opt-send" href="#opt-send" title="#opt-send">`send`</a> | Defines the payload to send to the server during the health check. | "" | No |
| <a id="opt-expect" href="#opt-expect" title="#opt-expect">`expect`</a> | Defines the expected response payload from the server. | "" | No |
| <a id="opt-interval" href="#opt-interval" title="#opt-interval">`interval`</a> | Defines the frequency of the health check calls for healthy targets. | 30s | No |
| <a id="opt-unhealthyInterval" href="#opt-unhealthyInterval" title="#opt-unhealthyInterval">`unhealthyInterval`</a> | Defines the frequency of the health check calls for unhealthy targets. When not defined, it defaults to the `interval` value. | 30s | No |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | Defines the maximum duration Traefik will wait for a health check connection before considering the server unhealthy. | 5s | No |
| <a id="opt-mode" href="#opt-mode" title="#opt-mode">`mode`</a> | Defines the health check mode. Can be `TCP` (default) for TCP payload exchange, or `HTTP` for HTTP health checks. | TCP | No |

#### HTTP Mode

When the health check `mode` is set to `HTTP`, you can configure HTTP-specific parameters using the `httpHealthCheck` field:

| Field | Description | Default | Required |
|-------|-------------|---------|----------|
| <a id="opt-httpHealthCheck" href="#opt-httpHealthCheck" title="#opt-httpHealthCheck">`httpHealthCheck`</a> | Configures HTTP health check options. Required when mode is set to `http`. | | No |
| <a id="opt-httpHealthCheck-path" href="#opt-httpHealthCheck-path" title="#opt-httpHealthCheck-path">`httpHealthCheck.path`</a> | Defines the HTTP request path for the health check. | / | No |
| <a id="opt-httpHealthCheck-port" href="#opt-httpHealthCheck-port" title="#opt-httpHealthCheck-port">`httpHealthCheck.port`</a> | Overrides the server address port for the HTTP health check endpoint. | | No |
| <a id="opt-httpHealthCheck-method" href="#opt-httpHealthCheck-method" title="#opt-httpHealthCheck-method">`httpHealthCheck.method`</a> | Defines the HTTP method used for the health check request. | GET | No |
| <a id="opt-httpHealthCheck-status" href="#opt-httpHealthCheck-status" title="#opt-httpHealthCheck-status">`httpHealthCheck.status`</a> | Defines the expected HTTP status code for a successful health check. | 200 | No |

## Weighted Round Robin

The Weighted Round Robin (alias `WRR`) load-balancer of services is in charge of balancing the connections between multiple services based on provided weights.

This strategy is only available to load balance between [services](./service.md) and not between servers.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File provider](../../install-configuration/providers/others/file.md).

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

### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. whenever one of its children is reported as down, 
this service becomes aware of it, and takes it into account (i.e. it ignores the down child) when running the load-balancing algorithm.
In addition, if the parent of this service also has HealthCheck enabled, this service reports to its parent any status change.

!!! note "Behavior"

    If HealthCheck is enabled for a given service and any of its descendants does not have it enabled, the creation of the service will fail.

    HealthCheck on Weighted services can be defined currently only with the [File provider](../../install-configuration/providers/others/file.md).

```yaml tab="Structured (YAML)"
## Dynamic configuration
tcp:
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
          send: "PING"
          expect: "PONG"
          interval: 10s
          timeout: 3s
        servers:
        - address: "192.168.1.10:6379"

    appv2:
      loadBalancer:
        healthCheck:
          send: "PING"
          expect: "PONG"
          interval: 10s
          timeout: 3s
        servers:
        - address: "192.168.1.11:6379"
```

```toml tab="Structured (TOML)"
## Dynamic configuration
[tcp.services]
  [tcp.services.app]
    [tcp.services.app.weighted.healthCheck]
    [[tcp.services.app.weighted.services]]
      name = "appv1"
      weight = 3
    [[tcp.services.app.weighted.services]]
      name = "appv2"
      weight = 1

  [tcp.services.appv1]
    [tcp.services.appv1.loadBalancer]
      [tcp.services.appv1.loadBalancer.healthCheck]
        send = "PING"
        expect = "PONG"
        interval = "10s"
        timeout = "3s"
      [[tcp.services.appv1.loadBalancer.servers]]
        address = "192.168.1.10:6379"

  [tcp.services.appv2]
    [tcp.services.appv2.loadBalancer]
      [tcp.services.appv2.loadBalancer.healthCheck]
        send = "PING"
        expect = "PONG"
        interval = "10s"
        timeout = "3s"
      [[tcp.services.appv2.loadBalancer.servers]]
        address = "192.168.1.11:6379"
```

