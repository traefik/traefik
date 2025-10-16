---
title: "Traefik HTTP Services Documentation"
description: "A service is in charge of connecting incoming requests to the Servers that can handle them. Read the technical documentation."
---

Traefik services define how to distribute incoming traffic across your backend servers.
Each service implements one of the load balancing strategies detailed on this page to ensure optimal traffic distribution and high availability.

## Service Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs.

Each service has a load-balancer, even if there is only one server to forward traffic to.

### Configuration Example

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
| <a id="opt-servers" href="#opt-servers" title="#opt-servers">`servers`</a> | Represents individual backend instances for your service                                                                                                                                                                                                                                                                                                                                      | Yes      |
| <a id="opt-sticky" href="#opt-sticky" title="#opt-sticky">`sticky`</a> | Defines a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.                                                                                                                                                                                                                                                                  | No       |
| <a id="opt-healthcheck" href="#opt-healthcheck" title="#opt-healthcheck">`healthcheck`</a> | Configures health check to remove unhealthy servers from the load balancing rotation.                                                                                                                                                                                                                                                                                                         | No       |
| <a id="opt-passiveHealthcheck" href="#opt-passiveHealthcheck" title="#opt-passiveHealthcheck">`passiveHealthcheck`</a> | Configures the passive health check to remove unhealthy servers from the load balancing rotation.                                                                                                                                                                                                                                                                                             | No       |
| <a id="opt-passHostHeader" href="#opt-passHostHeader" title="#opt-passHostHeader">`passHostHeader`</a> | Allows forwarding of the client Host header to server. By default, `passHostHeader` is true.                                                                                                                                                                                                                                                                                                  | No       |
| <a id="opt-serversTransport" href="#opt-serversTransport" title="#opt-serversTransport">`serversTransport`</a> | Allows to reference an [HTTP ServersTransport](./serverstransport.md) configuration for the communication between Traefik and your servers. If no `serversTransport` is specified, the `default@internal` will be used.                                                                                                                                                                       | No       |
| <a id="opt-responseForwarding" href="#opt-responseForwarding" title="#opt-responseForwarding">`responseForwarding`</a> | Configures how Traefik forwards the response from the backend server to the client.                                                                                                                                                                                                                                                                                                           | No       |
| <a id="opt-responseForwarding-FlushInterval" href="#opt-responseForwarding-FlushInterval" title="#opt-responseForwarding-FlushInterval">`responseForwarding.FlushInterval`</a> | Specifies the interval in between flushes to the client while copying the response body. It is a duration in milliseconds, defaulting to 100ms. A negative value means to flush immediately after each write to the client. The `FlushInterval` is ignored when ReverseProxy recognizes a response as a streaming response; for such responses, writes are flushed to the client immediately. | No       |

#### Servers

Servers represent individual backend instances for your service. The [service loadBalancer](#service-load-balancer) `servers` option lets you configure the list of instances that will handle incoming requests.

##### Configuration Options

| Field          | Description                                        | Required                                                                         |
|----------------|----------------------------------------------------|----------------------------------------------------------------------------------|
| <a id="opt-url" href="#opt-url" title="#opt-url">`url`</a> | Points to a specific instance.                     | Yes for File provider, No for [Docker provider](../../other-providers/docker.md) |
| <a id="opt-weight" href="#opt-weight" title="#opt-weight">`weight`</a> | Allows for weighted load balancing on the servers. | No                                                                               |
| <a id="opt-preservePath" href="#opt-preservePath" title="#opt-preservePath">`preservePath`</a> | Allows to preserve the URL path.                   | No                                                                               |

#### Health Check

The `healthcheck` option configures health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider HTTP(s) servers healthy as long as they return a status code to the health check request (carried out every interval) between `2XX` and `3XX`, or matching the configured status.
For gRPC servers, Traefik will consider them healthy as long as they return SERVING to [gRPC health check v1 requests](https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

To propagate status changes (e.g. all servers of this service are down) upwards, HealthCheck must also be enabled on the parent(s) of this service.

Below are the available options for the health check mechanism:

| Field               | Description                                                                                                                   | Default | Required |
|---------------------|-------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="opt-path" href="#opt-path" title="#opt-path">`path`</a> | Defines the server URL path for the health check endpoint.                                                                    | ""      | Yes      |
| <a id="opt-scheme" href="#opt-scheme" title="#opt-scheme">`scheme`</a> | Replaces the server URL scheme for the health check endpoint.                                                                 |         | No       |
| <a id="opt-mode" href="#opt-mode" title="#opt-mode">`mode`</a> | If defined to `grpc`, will use the gRPC health check protocol to probe the server.                                            | http    | No       |
| <a id="opt-hostname" href="#opt-hostname" title="#opt-hostname">`hostname`</a> | Defines the value of hostname in the Host header of the health check request.                                                 | ""      | No       |
| <a id="opt-port" href="#opt-port" title="#opt-port">`port`</a> | Replaces the server URL port for the health check endpoint.                                                                   |         | No       |
| <a id="opt-interval" href="#opt-interval" title="#opt-interval">`interval`</a> | Defines the frequency of the health check calls for healthy targets.                                                          | 30s     | No       |
| <a id="opt-unhealthyInterval" href="#opt-unhealthyInterval" title="#opt-unhealthyInterval">`unhealthyInterval`</a> | Defines the frequency of the health check calls for unhealthy targets. When not defined, it defaults to the `interval` value. | 30s     | No       |
| <a id="opt-timeout" href="#opt-timeout" title="#opt-timeout">`timeout`</a> | Defines the maximum duration Traefik will wait for a health check request before considering the server unhealthy.            | 5s      | No       |
| <a id="opt-headers" href="#opt-headers" title="#opt-headers">`headers`</a> | Defines custom headers to be sent to the health check endpoint.                                                               |         | No       |
| <a id="opt-followRedirects" href="#opt-followRedirects" title="#opt-followRedirects">`followRedirects`</a> | Defines whether redirects should be followed during the health check calls.                                                   | true    | No       |
| <a id="opt-hostname-2" href="#opt-hostname-2" title="#opt-hostname-2">`hostname`</a> | Defines the value of hostname in the Host header of the health check request.                                                 | ""      | No       |
| <a id="opt-method" href="#opt-method" title="#opt-method">`method`</a> | Defines the HTTP method that will be used while connecting to the endpoint.                                                   | GET     | No       |
| <a id="opt-status" href="#opt-status" title="#opt-status">`status`</a> | Defines the expected HTTP status code of the response to the health check request.                                            |         | No       |

#### Sticky sessions

When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.
On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.

##### Stickiness on multiple levels

    When chaining or mixing load-balancers (e.g. a load-balancer of servers is one of the "children" of a load-balancer of services), for stickiness to work all the way, the option needs to be specified at all required levels. Which means the client needs to send a cookie with as many key/value pairs as there are sticky levels.

##### Stickiness & Unhealthy Servers

    If the server specified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).

##### Cookie Name

    The default cookie name is an abbreviation of a sha1 (ex: `_1d52e`).

##### MaxAge

    By default, the affinity cookie will never expire as the `MaxAge` option is set to zero.

    This option indicates the number of seconds until the cookie expires.  
    When set to a negative number, the cookie expires immediately.
    
##### Secure & HTTPOnly & SameSite flags

    By default, the affinity cookie is created without those flags.
    One however can change that through configuration.

    `SameSite` can be `none`, `lax`, `strict` or empty.

##### Domain

    The Domain attribute of a cookie specifies the domain for which the cookie is valid. 
    
    By setting the Domain attribute, the cookie can be shared across subdomains (for example, a cookie set for example.com would be accessible to www.example.com, api.example.com, etc.). This is particularly useful in cases where sticky sessions span multiple subdomains, ensuring that the session is maintained even when the client interacts with different parts of the infrastructure.

??? example "Adding Stickiness -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            sticky:
             cookie: {}
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service]
        [http.services.my-service.loadBalancer.sticky.cookie]
    ```

??? example "Adding Stickiness with custom Options -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            sticky:
              cookie:
                name: my_sticky_cookie_name
                secure: true
                domain: mysite.site
                httpOnly: true
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service]
        [http.services.my-service.loadBalancer.sticky.cookie]
          name = "my_sticky_cookie_name"
          secure = true
          httpOnly = true
          domain = "mysite.site"
          sameSite = "none"
    ```

??? example "Setting Stickiness on all the required levels -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        wrr1:
          weighted:
            sticky:
              cookie:
                name: lvl1
            services:
              - name: whoami1
                weight: 1
              - name: whoami2
                weight: 1

        whoami1:
          loadBalancer:
            sticky:
              cookie:
                name: lvl2
            servers:
              - url: http://127.0.0.1:8081
              - url: http://127.0.0.1:8082

        whoami2:
          loadBalancer:
            sticky:
              cookie:
                name: lvl2
            servers:
              - url: http://127.0.0.1:8083
              - url: http://127.0.0.1:8084
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.wrr1]
        [http.services.wrr1.weighted.sticky.cookie]
          name = "lvl1"
        [[http.services.wrr1.weighted.services]]
          name = "whoami1"
          weight = 1
        [[http.services.wrr1.weighted.services]]
          name = "whoami2"
          weight = 1

      [http.services.whoami1]
        [http.services.whoami1.loadBalancer]
          [http.services.whoami1.loadBalancer.sticky.cookie]
            name = "lvl2"
          [[http.services.whoami1.loadBalancer.servers]]
            url = "http://127.0.0.1:8081"
          [[http.services.whoami1.loadBalancer.servers]]
            url = "http://127.0.0.1:8082"

      [http.services.whoami2]
        [http.services.whoami2.loadBalancer]
          [http.services.whoami2.loadBalancer.sticky.cookie]
            name = "lvl2"
          [[http.services.whoami2.loadBalancer.servers]]
            url = "http://127.0.0.1:8083"
          [[http.services.whoami2.loadBalancer.servers]]
            url = "http://127.0.0.1:8084"
    ```

    To keep a session open with the same server, the client would then need to specify the two levels within the cookie for each request, e.g. with curl:

    ```
    curl -b "lvl1=whoami1; lvl2=http://127.0.0.1:8081" http://localhost:8000
    ```

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
| <a id="opt-failureWindow" href="#opt-failureWindow" title="#opt-failureWindow">`failureWindow`</a> | Defines the time window during which the failed attempts must occur for the server to be marked as unhealthy. It also defines for how long the server will be considered unhealthy. | 10s     | No       |
| <a id="opt-maxFailedAttempts" href="#opt-maxFailedAttempts" title="#opt-maxFailedAttempts">`maxFailedAttempts`</a> | Defines the number of consecutive failed attempts allowed within the failure window before marking the server as unhealthy.                                                         | 1       | No       |

## Weighted Round Robin (WRR)

The WRR is able to load balance the requests between multiple services based on weights.

This strategy is only available to load balance between services and not between servers.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-crd.md) providers. To load balance between servers based on weights, the Load Balancer service should be used instead.

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
## P2C

Power of two choices algorithm is a load balancing strategy that selects two servers at random and chooses the one with the least number of active requests.

??? example "P2C Load Balancing -- Using the [File Provider](../../../install-configuration/providers/others/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            strategy: "p2c"
            servers:
            - url: "http://private-ip-server-1/"
            - url: "http://private-ip-server-2/"
            - url: "http://private-ip-server-3/"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service.loadBalancer]
        strategy = "p2c"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-2/"       
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-3/"
    ```

## Mirroring

The mirroring is able to mirror requests sent to a service to other services. Please note that by default the whole request is buffered in memory while it is being mirrored. See the `maxBodySize` option in the example below for how to modify this behaviour. You can also omit the request body by setting the `mirrorBody` option to false.

!!! warning "Default behavior of `percent`"

    When configuring a `mirror` service, if the `percent` field is not set, it defaults to `0`, meaning **no traffic will be sent to the mirror**.
    
!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../../install-configuration/providers/others/file.md) or [IngressRoute](../../../install-configuration/providers/kubernetes/kubernetes-crd.md) providers.
    
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
