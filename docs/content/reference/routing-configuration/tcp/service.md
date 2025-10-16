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

### Configuration Options

| Field | Description                                 | Default |
|----------|------------------------------------------|--------- |
| <a id="opt-servers" href="#opt-servers" title="#opt-servers">`servers`</a> |  Servers declare a single instance of your program.  | "" |
| <a id="opt-servers-address" href="#opt-servers-address" title="#opt-servers-address">`servers.address`</a> |   The address option (IP:Port) point to a specific instance. | "" |
| <a id="opt-servers-tls" href="#opt-servers-tls" title="#opt-servers-tls">`servers.tls`</a> | The `tls` option determines whether to use TLS when dialing with the backend. | false |
| <a id="opt-servers-serversTransport" href="#opt-servers-serversTransport" title="#opt-servers-serversTransport">`servers.serversTransport`</a> | `serversTransport` allows to reference a TCP [ServersTransport](./serverstransport.md configuration for the communication between Traefik and your servers. If no serversTransport is specified, the default@internal will be used. |  "" |
| <a id="healthcheck" href="#healthcheck" title="#healthcheck">`healthCheck`</a> | Configures health check to remove unhealthy servers from the load balancing rotation. | | No |


### Health Check

The `healthCheck` option configures health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider TCP servers healthy as long as the connection to the target server succeeds.
For advanced health checks, you can configure TCP payload exchange by specifying `send` and `expect` parameters.

To propagate status changes (e.g. all servers of this service are down) upwards, HealthCheck must also be enabled on the parent(s) of this service.

Below are the available options for the health check mechanism:

| Field | Description | Default | Required |
|-------|-------------|---------|----------|
| <a id="port" href="#port" title="#port">`port`</a> | Replaces the server URL port for the health check endpoint. | | No |
| <a id="send" href="#send" title="#send">`send`</a> | Defines the payload to send to the server during the health check. | "" | No |
| <a id="expect" href="#expect" title="#expect">`expect`</a> | Defines the expected response payload from the server. Only used when `send` is specified. | "" | No |
| <a id="interval" href="#interval" title="#interval">`interval`</a> | Defines the frequency of the health check calls for healthy targets. | 30s | No |
| <a id="unhealthyInterval" href="#unhealthyInterval" title="#unhealthyInterval">`unhealthyInterval`</a> | Defines the frequency of the health check calls for unhealthy targets. When not defined, it defaults to the `interval` value. | 30s | No |
| <a id="timeout" href="#timeout" title="#timeout">`timeout`</a> | Defines the maximum duration Traefik will wait for a health check request before considering the server unhealthy. | 5s | No |

#### Health Check Types

**Connection-only health check**

When no `send` parameter is specified, Traefik will perform a simple TCP connection test.
The server is considered healthy if the TCP connection succeeds.

**Payload-based health check**

When `send` parameter is specified, Traefik will:

1. Establish a TCP connection to the server
2. Send the specified payload
3. If `expect` is specified, wait for the expected response
4. Consider the server healthy if all steps succeed

#### TLS Support

TCP health checks fully support TLS connections.

When the target server uses TLS (`tls: true`), the health check will:

1. Establish a TCP connection
2. Perform TLS handshake
3. Exchange payload over the encrypted connection (if configured)

#### Port Override

You can configure health checks to use a different port than the main service by setting the `port` option. This is useful when your application exposes a dedicated health check endpoint on a separate port.

??? example "Basic Connection Health Check -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            healthCheck:
              interval: "10s"
              timeout: "3s"
            servers:
            - address: "192.168.1.10:8080"
            - address: "192.168.1.11:8080"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [tcp.services.my-service.loadBalancer.healthCheck]
          interval = "10s"
          timeout = "3s"
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "192.168.1.10:8080"
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "192.168.1.11:8080"
    ```

??? example "Payload-based Health Check -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        redis-service:
          loadBalancer:
            healthCheck:
              send: "PING"
              expect: "PONG"
              interval: "10s"
              timeout: "3s"
            servers:
            - address: "192.168.1.10:6379"
            - address: "192.168.1.11:6379"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.redis-service.loadBalancer]
        [tcp.services.redis-service.loadBalancer.healthCheck]
          send = "PING"
          expect = "PONG"
          interval = "10s"
          timeout = "3s"
        [[tcp.services.redis-service.loadBalancer.servers]]
          address = "192.168.1.10:6379"
        [[tcp.services.redis-service.loadBalancer.servers]]
          address = "192.168.1.11:6379"
    ```

??? example "TLS Health Check -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        secure-service:
          loadBalancer:
            healthCheck:
              send: "HELLO"
              expect: "WORLD"
              interval: "10s"
              timeout: "5s"
            servers:
            - address: "192.168.1.10:8443"
              tls: true
            - address: "192.168.1.11:8443"
              tls: true
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.secure-service.loadBalancer]
        [tcp.services.secure-service.loadBalancer.healthCheck]
          send = "HELLO"
          expect = "WORLD"
          interval = "10s"
          timeout = "5s"
        [[tcp.services.secure-service.loadBalancer.servers]]
          address = "192.168.1.10:8443"
          tls = true
        [[tcp.services.secure-service.loadBalancer.servers]]
          address = "192.168.1.11:8443"
          tls = true
    ```

??? example "Health Check with Port Override -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        app-service:
          loadBalancer:
            healthCheck:
              port: 8090
              send: "STATUS"
              expect: "OK"
              interval: "15s"
              timeout: "3s"
            servers:
            - address: "192.168.1.10:8080"
            - address: "192.168.1.11:8080"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.app-service.loadBalancer]
        [tcp.services.app-service.loadBalancer.healthCheck]
          port = 8090
          send = "STATUS"
          expect = "OK"
          interval = "15s"
          timeout = "3s"
        [[tcp.services.app-service.loadBalancer.servers]]
          address = "192.168.1.10:8080"
        [[tcp.services.app-service.loadBalancer.servers]]
          address = "192.168.1.11:8080"
    ```

## Servers

Servers represent individual backend instances for your service. The [service loadBalancer](#servers-load-balancer) `servers` option lets you configure the list of instances that will handle incoming requests.

### Configuration Options

| Field | Description | Default | Required |
|-------|-------------|---------|----------|
| <a id="servers-address" href="#servers-address" title="#servers-address">`address`</a> | The address option (IP:Port) point to a specific instance. | | Yes |
| <a id="servers-tls" href="#servers-tls" title="#servers-tls">`tls`</a> | The `tls` option determines whether to use TLS when dialing with the backend. | false | No |

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

### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. whenever one of its children is reported as down, this service becomes aware of it, and takes it into account (i.e. it ignores the down child) when running the load-balancing algorithm. In addition, if the parent of this service also has HealthCheck enabled, this service reports to its parent any status change.

!!! note "Behavior"

    If HealthCheck is enabled for a given service and any of its descendants does not have it enabled, the creation of the service will fail.

    HealthCheck on Weighted services can be defined currently only with the [File provider](../../../install-configuration/providers/others/file.md).

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

