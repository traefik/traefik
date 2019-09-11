# Services

Configuring How to Reach the Services
{: .subtitle }

![services](../../assets/img/services.png)

The `Services` are responsible for configuring how to reach the actual services that will eventually handle the incoming requests. 

## Configuration Example

??? example "Declaring an HTTP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service.loadBalancer]

        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-2/"
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            servers:
            - url: "http://private-ip-server-1/"
            - url: "http://private-ip-server-2/"
    ```

??? example "Declaring a TCP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
         [[tcp.services.my-service.loadBalancer.servers]]
           address = "xx.xx.xx.xx:xx"
         [[tcp.services.my-service.loadBalancer.servers]]
           address = "xx.xx.xx.xx:xx"
    ```
    
    ```yaml tab="YAML"
    tcp:
      services:
        my-service:
          loadBalancer:         
            servers:
            - address: "xx.xx.xx.xx:xx"
            - address: "xx.xx.xx.xx:xx"
    ```

## Configuring HTTP Services

### Servers Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs. 

??? example "Declaring a Service with Two Servers (with Load Balancing) -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service.loadBalancer]

        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-2/"
    ```

    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            servers:
            - url: "http://private-ip-server-1/"
            - url: "http://private-ip-server-2/"
    ```

#### Servers

Servers declare a single instance of your program.
The `url` option point to a specific instance. 

!!! note
    Paths in the servers' `url` have no effet. 
    If you want the requests to be sent to a specific path on your servers,
    configure your [`routers`](../routers/index.md) to use a corresponding [middleware](../../middlewares/overview.md) (e.g. the [AddPrefix](../../middlewares/addprefix.md) or [ReplacePath](../../middlewares/replacepath.md)) middlewares.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service.loadBalancer]
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            servers:
              - url: "http://private-ip-server-1/"
    ```

#### Load-balancing

For now, only round robin load balancing is supported:

??? example "Load Balancing -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service.loadBalancer]
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-2/"
    ```

    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            servers:
            - url: "http://private-ip-server-1/"
            - url: "http://private-ip-server-2/"
    ```

#### Sticky sessions
  
When sticky sessions are enabled, a cookie is set on the initial request to track which server handles the first response.
On subsequent requests, the client is forwarded to the same server.

!!! note "Stickiness & Unhealthy Servers"
   
    If the server specified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).

!!! note "Cookie Name" 
    
    The default cookie name is an abbreviation of a sha1 (ex: `_1d52e`).

!!! note "Secure & HTTPOnly flags"

    By default, the affinity cookie is created without those flags. One however can change that through configuration. 

??? example "Adding Stickiness"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service]
        [http.services.my-service.loadBalancer.sticky.cookie]
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            sticky:
             cookie: {}
    ```

??? example "Adding Stickiness with custom Options"

    ```toml tab="TOML"
    [http.services]
      [http.services.my-service]
        [http.services.my-service.loadBalancer.sticky.cookie]
          name = "my_sticky_cookie_name"
          secure = true
          httpOnly = true
    ```

    ```yaml tab="YAML"
    http:
      services:
        my-service:
          loadBalancer:
            sticky:
              cookie:
                name: my_sticky_cookie_name
                secure: true
                httpOnly: true
    ```

#### Health Check

Configure health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider your servers healthy as long as they return status codes between `2XX` and `3XX` to the health check requests (carried out every `interval`).

Below are the available options for the health check mechanism:

- `path` is appended to the server URL to set the health check endpoint.
- `scheme`, if defined, will replace the server URL `scheme` for the health check endpoint
- `hostname`, if defined, will replace the server URL `hostname` for the health check endpoint.
- `port`, if defined, will replace the server URL `port` for the health check endpoint.
- `interval` defines the frequency of the health check calls.
- `timeout` defines the maximum duration Traefik will wait for a health check request before considering the server failed (unhealthy).
- `headers` defines custom headers to be sent to the health check endpoint.

!!! note "Interval & Timeout Format"

    Interval and timeout are to be given in a format understood by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).
    The interval must be greater than the timeout. If configuration doesn't reflect this, the interval will be set to timeout + 1 second.

!!! note "Recovering Servers"
   
    Traefik keeps monitoring the health of unhealthy servers. 
    If a server has recovered (returning `2xx` -> `3xx` responses again), it will be added back to the load balacer rotation pool.

??? example "Custom Interval & Timeout -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.servicess.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          interval = "10s"
          timeout = "3s"
    ```

    ```yaml tab="YAML"
    http:
      servicess:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              interval: "10s"
              timeout: "3s"
    ```

??? example "Custom Port -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          port = 8080
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              port: 8080
    ```

??? example "Custom Scheme -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          scheme = "http"
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              scheme: http
    ```

??? example "Additional HTTP Headers -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"

          [http.services.Service-1.loadBalancer.healthCheck.headers]
            My-Custom-Header = "foo"
            My-Header = "bar"
    ```
    
    ```yaml tab="YAML"
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              headers:
                My-Custom-Header: foo
                My-Header: bar
    ```

### Weighted Round Robin (service)

The WRR is able to load balance the requests between multiple services based on weights.

This strategy is only available to load balance between [services](./index.md) and not between [servers](./index.md#servers).

This strategy can be defined only with [File](../../providers/file.md).

```toml tab="TOML"
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

```yaml tab="YAML"
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

### Mirroring (service)

The mirroring is able to mirror requests sent to a service to other services.

This strategy can be defined only with [File](../../providers/file.md).

```toml tab="TOML"
[http.services]
  [http.services.mirrored-api]
    [http.services.mirrored-api.mirroring]
      service = "appv1"
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

```yaml tab="YAML"
http:
  services:
    mirrored-api:
      mirroring:
        service: appv1
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
        - url: "http://private-ip-server-2/"
```

## Configuring TCP Services

### General

Currently, `LoadBalancer` is the only supported kind of TCP `Service`.
However, since Traefik is an ever evolving project, other kind of TCP Services will be available in the future,
reason why you have to specify it. 

### Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs. 

??? example "Declaring a Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
        [[tcp.services.my-service.loadBalancer.servers]]
           address = "xx.xx.xx.xx:xx"
    ```

    ```yaml tab="YAML"
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
            - address: "xx.xx.xx.xx:xx"
            - address: "xx.xx.xx.xx:xx"
    ```

#### Servers

Servers declare a single instance of your program.
The `address` option (IP:Port) point to a specific instance.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
    ```

    ```yaml tab="YAML"
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
              address: "xx.xx.xx.xx:xx"
    ```

#### Termination Delay

As a proxy between a client and a server, it can happen that either side (e.g. client side) decides to terminate its writing capability on the connnection (i.e. issuance of a FIN packet).
The proxy needs to propagate that intent to the other side, and so when that happens, it also does the same on its connection with the other side (e.g. backend side).

However, if for some reason (bad implementation, or malicious itent) the other side does not eventually do the same as well,
the connection would stay half-open, which would lock resources for however long.
To that end, as soon as the proxy enters this termination sequence, it sets a deadline on fully terminating the connections on both sides.
The termination delay controls that deadline. It is a duration in milliseconds, defaulting to 100. A negative value means an infinite deadline (i.e. the connection is never fully terminated by the proxy itself).

??? example "A Service with a termination delay -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer]]
          terminationDelay = 200
    ```

    ```yaml tab="YAML"
    tcp:
      services:
        my-service:
          loadBalancer:
            terminationDelay: 200
    ```
