# Services

Configuring How to Reach the Services
{: .subtitle }

![services](../../assets/img/services.png)

The `Services` are responsible for configuring how to reach the actual services that will eventually handle the incoming requests. 

## Configuration Example

??? example "Declaring an HTTP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [http.services]
      [http.services.my-service.LoadBalancer]
         method = "wrr" # Load Balancing based on weights
         
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-1/"
            weight = 30 # 30% of the requests will go to that instance
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-2/"
            weight = 70 # 70% of the requests will go to that instance         
    ```

??? example "Declaring a TCP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [tcp.services]
      [tcp.services.my-service.LoadBalancer]         
         [[tcp.services.my-service.LoadBalancer.servers]]
            address = "xx.xx.xx.xx:xx"
         [[tcp.services.my-service.LoadBalancer.servers]]
            address = "xx.xx.xx.xx:xx"
    ```

## Configuring HTTP Services

### General

Currently, `LoadBalancer` is the only supported kind of HTTP `Service` (see below).
However, since Traefik is an ever evolving project, other kind of HTTP Services will be available in the future,
reason why you have to specify it. 

### Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs. 

??? example "Declaring a Service with Two Servers (with Load Balancing) -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [http.services]
      [http.services.my-service.LoadBalancer]
         method = "wrr" # Load Balancing based on weights
         
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-1/"
            weight = 50 # 50% of the requests will go to that instance
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-2/"
            weight = 50 # 50% of the requests will go to that instance         
    ```

#### Servers

Servers declare a single instance of your program.
The `url` option point to a specific instance. 
The `weight` option defines the weight of the server for the load balancing algorithm.

!!! note
    Paths in the servers' `url` have no effet. 
    If you want the requests to be sent to a specific path on your servers,
    configure your [`routers`](../routers/index.md) to use a corresponding [middleware](../../middlewares/overview.md) (e.g. the [AddPrefix](../../middlewares/addprefix.md) or [ReplacePath](../../middlewares/replacepath.md)) middlewares.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [http.services]
      [http.services.my-service.LoadBalancer]
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-1/"
            weight = 1
    ```

#### Load-balancing

Various methods of load balancing are supported:

- `wrr`: Weighted Round Robin.
- `drr`: Dynamic Round Robin: increases weights on servers that perform better than others (rolls back to original weights when the server list is updated)

??? example "Load Balancing Using DRR -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [http.services]
      [http.services.my-service.LoadBalancer]
         method = "drr"
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-1/"
            weight = 1
         [[http.services.my-service.LoadBalancer.servers]]
            url = "http://private-ip-server-1/"
            weight = 1
    ```

#### Sticky sessions
  
When sticky sessions are enabled, a cookie is set on the initial request to track which server handles the first response.
On subsequent requests, the client is forwarded to the same server.

!!! note "Stickiness & Unhealthy Servers"
   
    If the server specified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).

!!! note "Cookie Name" 
    
    The default cookie name is an abbreviation of a sha1 (ex: `_1d52e`).

??? example "Adding Stickiness"

    ```toml
    [http.services]
      [http.services.my-service]
        [http.services.my-service.LoadBalancer.stickiness]
    ```

??? example "Adding Stickiness with a Custom Cookie Name"

    ```toml
    [http.services]
      [http.services.my-service]
        [http.services.my-service.LoadBalancer.stickiness]
           cookieName = "my_stickiness_cookie_name"
    ```

#### Health Check

Configure healthcheck to remove unhealthy servers from the load balancing rotation.
Traefik will consider your servers healthy as long as they return status codes between `2XX` and `3XX` to the health check requests (carried out every `interval`).

Below are the available options for the health check mechanism:

- `path` is appended to the server URL to set the healcheck endpoint.
- `scheme`, if defined, will replace the server URL `scheme` for the healthcheck endpoint
- `hostname`, if defined, will replace the server URL `hostname` for the healthcheck endpoint.
- `port`, if defined, will replace the server URL `port` for the healthcheck endpoint.
- `interval` defines the frequency of the healthcheck calls.
- `timeout` defines the maximum duration Traefik will wait for a healthcheck request before considering the server failed (unhealthy).
- `headers` defines custom headers to be sent to the healthcheck endpoint.

!!! note "Interval & Timeout Format"

    Interval and timeout are to be given in a format understood by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).
    The interval must be greater than the timeout. If configuration doesn't reflect this, the interval will be set to timeout + 1 second.

!!! note "Recovering Servers"
   
    Traefik keeps monitoring the health of unhealthy servers. 
    If a server has recovered (returning `2xx` -> `3xx` responses again), it will be added back to the load balacer rotation pool.

??? example "Custom Interval & Timeout -- Using the File Provider"

    ```toml
    [http.services]
      [http.servicess.Service-1]
        [http.services.Service-1.healthcheck]
            path = "/health"
            interval = "10s"
            timeout = "3s"
    ```

??? example "Custom Port -- Using the File Provider"

    ```toml
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.healthcheck]
            path = "/health"
            port = 8080
    ```

??? example "Custom Scheme -- Using the File Provider"

    ```toml
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.healthcheck]
            path = "/health"
            scheme = "http"
    ```

??? example "Additional HTTP Headers -- Using the File Provider"

    ```toml
    [http.services]
        [http.services.Service-1]
            [http.servicess.Service-1.healthcheck]
                path = "/health"

                [Service.Service-1.healthcheck.headers]
                    My-Custom-Header = "foo"
                    My-Header = "bar"
    ```
    
## Configuring TCP Services

### General

Currently, `LoadBalancer` is the only supported kind of TCP `Service`.
However, since Traefik is an ever evolving project, other kind of TCP Services will be available in the future,
reason why you have to specify it. 

### Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs. 

??? example "Declaring a Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [tcp.services]
      [tcp.services.my-service.LoadBalancer]
         [[tcp.services.my-service.LoadBalancer.servers]]
            address = "xx.xx.xx.xx:xx"
         [[tcp.services.my-service.LoadBalancer.servers]]
            address = "xx.xx.xx.xx:xx"
    ```

#### Servers

Servers declare a single instance of your program.
The `address` option (IP:Port) point to a specific instance.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```toml
    [tcp.services]
      [tcp.services.my-service.LoadBalancer]
         [[tcp.services.my-service.LoadBalancer.servers]]
            address = "xx.xx.xx.xx:xx"
    ```

!!! note "Weight"
    
    The TCP LoadBalancer is currently a round robin only implementation and doesn't yet support weights.