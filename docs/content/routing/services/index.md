---
title: "Traefik Services Documentation"
description: "Learn how to configure routing and load balancing in Traefik Proxy to reach Services, which handle incoming requests. Read the technical documentation."
---

# Services

Configuring How to Reach the Services
{: .subtitle }

![services](../../assets/img/services.png)

The `Services` are responsible for configuring how to reach the actual services that will eventually handle the incoming requests.

## Configuration Examples

??? example "Declaring an HTTP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            servers:
            - url: "http://<private-ip-server-1>:<private-port-server-1>/"
            - url: "http://<private-ip-server-2>:<private-port-server-2>/"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service.loadBalancer]

        [[http.services.my-service.loadBalancer.servers]]
          url = "http://<private-ip-server-1>:<private-port-server-1>/"
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://<private-ip-server-2>:<private-port-server-2>/"
    ```

??? example "Declaring a TCP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
            - address: "<private-ip-server-1>:<private-port-server-1>"
            - address: "<private-ip-server-2>:<private-port-server-2>"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
         [[tcp.services.my-service.loadBalancer.servers]]
           address = "<private-ip-server-1>:<private-port-server-1>"
         [[tcp.services.my-service.loadBalancer.servers]]
           address = "<private-ip-server-2>:<private-port-server-2>"
    ```

??? example "Declaring a UDP Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    udp:
      services:
        my-service:
          loadBalancer:
            servers:
            - address: "<private-ip-server-1>:<private-port-server-1>"
            - address: "<private-ip-server-2>:<private-port-server-2>"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [udp.services]
      [udp.services.my-service.loadBalancer]
         [[udp.services.my-service.loadBalancer.servers]]
           address = "<private-ip-server-1>:<private-port-server-1>"
         [[udp.services.my-service.loadBalancer.servers]]
           address = "<private-ip-server-2>:<private-port-server-2>"
    ```

## Configuring HTTP Services

### Servers Load Balancer

The load balancers are able to load balance the requests between multiple instances of your programs.

Each service has a load-balancer, even if there is only one server to forward traffic to.

??? example "Declaring a Service with Two Servers (with Load Balancing) -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
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

#### Servers

Servers declare a single instance of your program.

The `url` option point to a specific instance.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            servers:
              - url: "http://private-ip-server-1/"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service.loadBalancer]
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
    ```

The `preservePath` option allows to preserve the URL path.

!!! info "Health Check"

    When a [health check](#health-check) is configured for the server, the path is not preserved.

??? example "A Service with One Server and PreservePath -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        my-service:
          loadBalancer:
            servers:
              - url: "http://private-ip-server-1/base"
                preservePath: true
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service.loadBalancer]
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/base"
          preservePath = true
    ```

#### Load Balancing Strategy

The `strategy` option allows to choose the load balancing algorithm.

Two load balancing algorithms are supported:

- Weighed round-robin (wrr)
- Power of two choices (p2c)

##### WRR

Weighed round-robin is the default strategy (and does not need to be specified).

The `weight` option allows for weighted load balancing on the servers.

??? example "A Service with Two Servers with Weight -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
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

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.my-service.loadBalancer]
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-1/"
          weight = 2
        [[http.services.my-service.loadBalancer.servers]]
          url = "http://private-ip-server-2/"
          weight = 1
    ```

##### P2C

Power of two choices algorithm is a load balancing strategy that selects two servers at random and chooses the one with the least number of active requests.

??? example "P2C Load Balancing -- Using the [File Provider](../../providers/file.md)"

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

#### Sticky sessions

When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.
On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.

!!! info "Stickiness on multiple levels"

    When chaining or mixing load-balancers (e.g. a load-balancer of servers is one of the "children" of a load-balancer of services), for stickiness to work all the way, the option needs to be specified at all required levels. Which means the client needs to send a cookie with as many key/value pairs as there are sticky levels.

!!! info "Stickiness & Unhealthy Servers"

    If the server specified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).

!!! info "Cookie Name"

    The default cookie name is an abbreviation of a sha1 (ex: `_1d52e`).

!!! info "MaxAge"

    By default, the affinity cookie will never expire as the `MaxAge` option is set to zero.

    This option indicates the number of seconds until the cookie expires.  
    When set to a negative number, the cookie expires immediately.
    
!!! info "Secure & HTTPOnly & SameSite flags"

    By default, the affinity cookie is created without those flags.
    One however can change that through configuration.

    `SameSite` can be `none`, `lax`, `strict` or empty.

!!! info "Domain"

    The Domain attribute of a cookie specifies the domain for which the cookie is valid. 
    
    By setting the Domain attribute, the cookie can be shared across subdomains (for example, a cookie set for example.com would be accessible to www.example.com, api.example.com, etc.). This is particularly useful in cases where sticky sessions span multiple subdomains, ensuring that the session is maintained even when the client interacts with different parts of the infrastructure.

??? example "Adding Stickiness -- Using the [File Provider](../../providers/file.md)"

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

??? example "Adding Stickiness with custom Options -- Using the [File Provider](../../providers/file.md)"

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

??? example "Setting Stickiness on all the required levels -- Using the [File Provider](../../providers/file.md)"

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

#### Health Check

Configure health check to remove unhealthy servers from the load balancing rotation.
Traefik will consider HTTP(s) servers healthy as long as they return a status code to the health check request (carried out every `interval`) between `2XX` and `3XX`, or matching the configured status.
For gRPC servers, Traefik will consider them healthy as long as they return `SERVING` to [gRPC health check v1](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) requests.

To propagate status changes (e.g. all servers of this service are down) upwards, HealthCheck must also be enabled on the parent(s) of this service.

Below are the available options for the health check mechanism:

- `path` (required), defines the server URL path for the health check endpoint .
- `scheme` (optional), replaces the server URL `scheme` for the health check endpoint.
- `mode` (default: http), if defined to `grpc`, will use the gRPC health check protocol to probe the server.
- `hostname` (optional), sets the value of `hostname` in the `Host` header of the health check request.
- `port` (optional), replaces the server URL `port` for the health check endpoint.
- `interval` (default: 30s), defines the frequency of the health check calls for healthy targets.
- `unhealthyInterval` (default: 30s), defines the frequency of the health check calls for unhealthy targets.  When not defined, it defaults to the `interval` value.
- `timeout` (default: 5s), defines the maximum duration Traefik will wait for a health check request before considering the server unhealthy.
- `headers` (optional), defines custom headers to be sent to the health check endpoint.
- `followRedirects` (default: true), defines whether redirects should be followed during the health check calls.
- `method` (default: GET), defines the HTTP method that will be used while connecting to the endpoint.
- `status` (optional), defines the expected HTTP status code of the response to the health check request.

!!! info "Interval & Timeout Format"

    Interval, UnhealthyInterval and Timeout are to be given in a format understood by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

!!! info "Recovering Servers"

    Traefik keeps monitoring the health of unhealthy servers.
    If a server has recovered (returning `2xx` -> `3xx` responses again), it will be added back to the load balancer rotation pool.

!!! warning "Health check with Kubernetes"

    Kubernetes has an health check mechanism to remove unhealthy pods from Kubernetes services (cf [readiness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes)).
    As unhealthy pods have no Kubernetes endpoints, Traefik will not forward traffic to them.
    Therefore, Traefik health check is not available for `kubernetesCRD` and `kubernetesIngress` providers.

??? example "Custom Interval & Timeout -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              interval: "10s"
              timeout: "3s"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          interval = "10s"
          timeout = "3s"
    ```

??? example "Custom Port -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              port: 8080
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          port = 8080
    ```

??? example "Custom Scheme -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service-1:
          loadBalancer:
            healthCheck:
              path: /health
              scheme: http
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"
          scheme = "http"
    ```

??? example "Additional HTTP Headers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
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

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.healthCheck]
          path = "/health"

          [http.services.Service-1.loadBalancer.healthCheck.headers]
            My-Custom-Header = "foo"
            My-Header = "bar"
    ```

#### Pass Host Header

The `passHostHeader` allows to forward client Host header to server.

By default, `passHostHeader` is true.

??? example "Don't forward the host header -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service01:
          loadBalancer:
            passHostHeader: false
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service01]
        [http.services.Service01.loadBalancer]
          passHostHeader = false
    ```

#### ServersTransport

`serversTransport` allows to reference an [HTTP ServersTransport](./index.md#serverstransport_1) configuration for the communication between Traefik and your servers.

??? example "Specify an HTTP transport -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service01:
          loadBalancer:
            serversTransport: mytransport
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service01]
        [http.services.Service01.loadBalancer]
          serversTransport = "mytransport"
    ```

!!! info Default Servers Transport
    If no serversTransport is specified, the `default@internal` will be used.
    The `default@internal` serversTransport is created from the [static configuration](../overview.md#http-servers-transports).

#### Response Forwarding

This section is about configuring how Traefik forwards the response from the backend server to the client.

Below are the available options for the Response Forwarding mechanism:

- `FlushInterval` specifies the interval in between flushes to the client while copying the response body.
  It is a duration in milliseconds, defaulting to 100.
  A negative value means to flush immediately after each write to the client.
  The FlushInterval is ignored when ReverseProxy recognizes a response as a streaming response;
  for such responses, writes are flushed to the client immediately.

??? example "Using a custom FlushInterval -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      services:
        Service-1:
          loadBalancer:
            responseForwarding:
              flushInterval: 1s
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.services]
      [http.services.Service-1]
        [http.services.Service-1.loadBalancer.responseForwarding]
          flushInterval = "1s"
    ```

### ServersTransport

ServersTransport allows to configure the transport between Traefik and your HTTP servers.

#### `serverName`

_Optional_

`serverName` configure the server name that will be used for SNI.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      serverName: "myhost"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  serverName = "myhost"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  serverName: "test"
```

#### `certificates`

_Optional_

`certificates` is the list of certificates (as file paths, or data bytes)
that will be set as client certificates for mTLS.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      certificates:
        - certFile: foo.crt
          keyFile: bar.crt
```

```toml tab="File (TOML)"
## Dynamic configuration
[[http.serversTransports.mytransport.certificates]]
  certFile = "foo.crt"
  keyFile = "bar.crt"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  certificatesSecrets:
    - mycert

---
apiVersion: v1
kind: Secret
metadata:
  name: mycert

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

#### `insecureSkipVerify`

_Optional_

`insecureSkipVerify` controls whether the server's certificate chain and host name is verified.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  insecureSkipVerify = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  insecureSkipVerify: true
```

#### `rootCAs`

_Optional_

`rootCAs` defines the set of root certificate authorities (as file paths, or data bytes) to use when verifying server certificates.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      rootCAs:
        - foo.crt
        - bar.crt
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  rootCAs = ["foo.crt", "bar.crt"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  rootCAsSecrets:
    - myca
---
apiVersion: v1
kind: Secret
metadata:
  name: myca

data:
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
```

#### `maxIdleConnsPerHost`

_Optional, Default=2_

If non-zero, `maxIdleConnsPerHost` controls the maximum idle (keep-alive) connections to keep per-host.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      maxIdleConnsPerHost: 7
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  maxIdleConnsPerHost = 7
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  maxIdleConnsPerHost: 7
```

#### `disableHTTP2`

_Optional, Default=false_

`disableHTTP2` disables HTTP/2 for connections with servers.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      disableHTTP2: true
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  disableHTTP2 = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  disableHTTP2: true
```

#### `peerCertURI`

_Optional, Default=""_

`peerCertURI` defines the URI used to match against SAN URIs during the server's certificate verification.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      peerCertURI: foobar
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  peerCertURI = "foobar"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  peerCertURI: foobar
```

#### `spiffe`

Please note that [SPIFFE](../../https/spiffe.md) must be enabled in the static configuration
before using it to secure the connection between Traefik and the backends.

##### `spiffe.ids`

_Optional_

`ids` defines the allowed SPIFFE IDs. 
This takes precedence over the SPIFFE TrustDomain.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      spiffe:
        ids:
          - spiffe://trust-domain/id1
          - spiffe://trust-domain/id2
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.spiffe]
  ids = ["spiffe://trust-domain/id1", "spiffe://trust-domain/id2"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
    spiffe:
      ids:
        - spiffe://trust-domain/id1
        - spiffe://trust-domain/id2
```

##### `spiffe.trustDomain`

_Optional_

`trustDomain` defines the allowed SPIFFE trust domain.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
        spiffe:
          trustDomain: spiffe://trust-domain
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.spiffe]
  trustDomain = "spiffe://trust-domain"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
    spiffe:
      trustDomain: "spiffe://trust-domain"
```

#### `forwardingTimeouts`

`forwardingTimeouts` are the timeouts applied when forwarding requests to the servers.

##### `forwardingTimeouts.dialTimeout`

_Optional, Default=30s_

`dialTimeout` is the maximum duration allowed for a connection to a backend server to be established.
Zero means no timeout.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      forwardingTimeouts:
        dialTimeout: "1s"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.forwardingTimeouts]
  dialTimeout = "1s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
    forwardingTimeouts:
      dialTimeout: "1s"
```

##### `forwardingTimeouts.responseHeaderTimeout`

_Optional, Default=0s_

`responseHeaderTimeout`, if non-zero, specifies the amount of time to wait for a server's response headers
after fully writing the request (including its body, if any).
This time does not include the time to read the response body.
Zero means no timeout.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      forwardingTimeouts:
        responseHeaderTimeout: "1s"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.forwardingTimeouts]
  responseHeaderTimeout = "1s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  forwardingTimeouts:
    responseHeaderTimeout: "1s"
```

##### `forwardingTimeouts.idleConnTimeout`

_Optional, Default=90s_

`idleConnTimeout` is the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
Zero means no limit.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      forwardingTimeouts:
        idleConnTimeout: "1s"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.forwardingTimeouts]
  idleConnTimeout = "1s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  forwardingTimeouts:
    idleConnTimeout: "1s"
```

##### `forwardingTimeouts.readIdleTimeout`

_Optional, Default=0s_

`readIdleTimeout` is the timeout after which a health check using ping frame will be carried out
if no frame is received on the HTTP/2 connection.
Note that a ping response will be considered a received frame,
so if there is no other traffic on the connection,
the health check will be performed every `readIdleTimeout` interval.
If zero, no health check is performed.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      forwardingTimeouts:
        readIdleTimeout: "1s"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.forwardingTimeouts]
  readIdleTimeout = "1s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  forwardingTimeouts:
    readIdleTimeout: "1s"
```

##### `forwardingTimeouts.pingTimeout`

_Optional, Default=15s_

`pingTimeout` is the timeout after which the HTTP/2 connection will be closed
if a response to ping is not received.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      forwardingTimeouts:
        pingTimeout: "1s"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport.forwardingTimeouts]
  pingTimeout = "1s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  forwardingTimeouts:
    pingTimeout: "1s"
```

### Weighted Round Robin (service)

The WRR is able to load balance the requests between multiple services based on weights.

This strategy is only available to load balance between [services](./index.md) and not between [servers](./index.md#servers).

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../providers/file.md) or [IngressRoute](../../providers/kubernetes-crd.md) providers.

```yaml tab="YAML"
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

```toml tab="TOML"
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

#### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. whenever
one of its children is reported as down, this service becomes aware of it, and
takes it into account (i.e. it ignores the down child) when running the
load-balancing algorithm. In addition, if the parent of this service also has
HealthCheck enabled, this service reports to its parent any status change.

!!! info "All or nothing"

    If HealthCheck is enabled for a given service, but any of its descendants does
    not have it enabled, the creation of the service will fail.

    HealthCheck on Weighted services can be defined currently only with the [File](../../providers/file.md) provider.

```yaml tab="YAML"
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

```toml tab="TOML"
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

### Mirroring (service)

The mirroring is able to mirror requests sent to a service to other services.
Please note that by default the whole request is buffered in memory while it is being mirrored.
See the maxBodySize option in the example below for how to modify this behaviour.
You can also omit the request body by setting the mirrorBody option to `false`.

!!! warning "Default behavior of `percent`"

    When configuring a `mirror` service, if the `percent` field is not set, it defaults to `0`, meaning **no traffic will be sent to the mirror**.

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../providers/file.md) or [IngressRoute](../../providers/kubernetes-crd.md) providers.

```yaml tab="YAML"
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
        - url: "http://private-ip-server-2/"
```

```toml tab="TOML"
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

#### Health Check

HealthCheck enables automatic self-healthcheck for this service, i.e. if the
main handler of the service becomes unreachable, the information is propagated
upwards to its parent.

!!! info "All or nothing"

    If HealthCheck is enabled for a given service, but any of its descendants does
    not have it enabled, the creation of the service will fail.

    HealthCheck on Mirroring services can be defined currently only with the [File](../../providers/file.md) provider.

```yaml tab="YAML"
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

```toml tab="TOML"
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

### Failover (service)

A failover service job is to forward all requests to a fallback service when the main service becomes unreachable.

!!! info "Relation to HealthCheck"

    The failover service relies on the HealthCheck system to get notified when its main service becomes unreachable,
    which means HealthCheck needs to be enabled and functional on the main service.
    However, HealthCheck does not need to be enabled on the failover service itself for it to be functional.
    It is only required in order to propagate upwards the information when the failover itself becomes down
    (i.e. both its main and its fallback are down too).

!!! info "Supported Providers"

    This strategy can currently only be defined with the [File](../../providers/file.md) provider.

```yaml tab="YAML"
## Dynamic configuration
http:
  services:
    app:
      failover:
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
        servers:
        - url: "http://private-ip-server-2/"
```

```toml tab="TOML"
## Dynamic configuration
[http.services]
  [http.services.app]
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
      [[http.services.backup.loadBalancer.servers]]
        url = "http://private-ip-server-2/"
```

#### Health Check

HealthCheck enables automatic self-healthcheck for this service,
i.e. if the main and the fallback services become unreachable,
the information is propagated upwards to its parent.

!!! info "All or nothing"

    If HealthCheck is enabled for a given service, but any of its descendants does
    not have it enabled, the creation of the service will fail.

    HealthCheck on a Failover service can currently only be defined with the [File](../../providers/file.md) provider.

```yaml tab="YAML"
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

```toml tab="TOML"
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

## Configuring TCP Services

### General

Each of the fields of the service section represents a kind of service.
Which means, that for each specified service, one of the fields, and only one,
has to be enabled to define what kind of service is created.
Currently, the two available kinds are `LoadBalancer`, and `Weighted`.

### Servers Load Balancer

The servers load balancer is in charge of balancing the requests between the servers of the same service.

??? example "Declaring a Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
            - address: "xx.xx.xx.xx:xx"
            - address: "xx.xx.xx.xx:xx"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
        [[tcp.services.my-service.loadBalancer.servers]]
           address = "xx.xx.xx.xx:xx"
    ```

#### Servers

Servers declare a single instance of your program.

#### `address`

The `address` option (IP:Port) point to a specific instance.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
              - address: "xx.xx.xx.xx:xx"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
    ```

#### `tls`

The `tls` determines whether to use TLS when dialing with the backend.

??? example "A Service with One Server Using TLS -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            servers:
              - address: "xx.xx.xx.xx:xx"
                tls: true
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
          tls = true
    ```

#### ServersTransport

`serversTransport` allows to reference a [TCP ServersTransport](./index.md#serverstransport_3) configuration for the communication between Traefik and your servers.

??? example "Specify a TCP transport -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        Service01:
          loadBalancer:
            serversTransport: mytransport
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.Service01]
        [tcp.services.Service01.loadBalancer]
          serversTransport = "mytransport"
    ```

!!! info "Default Servers Transport"

    If no serversTransport is specified, the `default@internal` will be used.
    The `default@internal` serversTransport is created from the [static configuration](../overview.md#tcp-servers-transports).

#### PROXY Protocol

Traefik supports [PROXY Protocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2 on TCP Services.
It can be enabled by setting `proxyProtocol` on the load balancer.

Below are the available options for the PROXY protocol:

- `version` specifies the version of the protocol to be used. Either `1` or `2`.

!!! info "Version"

    Specifying a version is optional. By default the version 2 will be used.

??? example "A Service with Proxy Protocol v1 -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            proxyProtocol:
              version: 1
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [tcp.services.my-service.loadBalancer.proxyProtocol]
          version = 1
    ```

#### Termination Delay

!!! warning

    Deprecated in favor of [`serversTransport.terminationDelay`](#terminationdelay).
    Please note that if any `serversTransport` configuration on the servers load balancer is found,
    it will take precedence over the servers load balancer `terminationDelay` value,
    even if the `serversTransport.terminationDelay` is undefined.

As a proxy between a client and a server, it can happen that either side (e.g. client side) decides to terminate its writing capability on the connection (i.e. issuance of a FIN packet).
The proxy needs to propagate that intent to the other side, and so when that happens, it also does the same on its connection with the other side (e.g. backend side).

However, if for some reason (bad implementation, or malicious intent) the other side does not eventually do the same as well,
the connection would stay half-open, which would lock resources for however long.

To that end, as soon as the proxy enters this termination sequence, it sets a deadline on fully terminating the connections on both sides.

The termination delay controls that deadline.
It is a duration in milliseconds, defaulting to 100.
A negative value means an infinite deadline (i.e. the connection is never fully terminated by the proxy itself).

??? example "A Service with a termination delay -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      services:
        my-service:
          loadBalancer:
            terminationDelay: 200
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.services]
      [tcp.services.my-service.loadBalancer]
        [[tcp.services.my-service.loadBalancer]]
          terminationDelay = 200
    ```

### Weighted Round Robin

The Weighted Round Robin (alias `WRR`) load-balancer of services is in charge of balancing the requests between multiple services based on provided weights.

This strategy is only available to load balance between [services](./index.md) and not between [servers](./index.md#servers).

!!! info "Supported Providers"

    This strategy can be defined currently with the [File](../../providers/file.md) or [IngressRoute](../../providers/kubernetes-crd.md) providers.

```yaml tab="YAML"
## Dynamic configuration
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

```toml tab="TOML"
## Dynamic configuration
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

### ServersTransport

ServersTransport allows to configure the transport between Traefik and your TCP servers.

#### `dialTimeout`

_Optional, Default="30s"_

`dialTimeout` defines the timeout when dialing the backend TCP service. If zero, no timeout exists.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      dialTimeout: 30s
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  dialTimeout = "30s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  dialTimeout: 30s
```

#### `dialKeepAlive`

_Optional, Default="15s"_

`dialKeepAlive` defines the interval between keep-alive probes for an active network connection.
If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and
operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative,
keep-alive probes are disabled.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      dialKeepAlive: 30s
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  dialKeepAlive = "30s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  dialKeepAlive: 30s
```

#### `terminationDelay`

_Optional, Default="100ms"_

As a proxy between a client and a server, it can happen that either side (e.g. client side) decides to terminate its writing capability on the connection (i.e. issuance of a FIN packet).
The proxy needs to propagate that intent to the other side, and so when that happens, it also does the same on its connection with the other side (e.g. backend side).

However, if for some reason (bad implementation, or malicious intent) the other side does not eventually do the same as well,
the connection would stay half-open, which would lock resources for however long.

To that end, as soon as the proxy enters this termination sequence, it sets a deadline on fully terminating the connections on both sides.

The termination delay controls that deadline.
A negative value means an infinite deadline (i.e. the connection is never fully terminated by the proxy itself).

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      terminationDelay: 100ms
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  terminationDelay = "100ms"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  terminationDelay: 100ms
```

#### `tls`

`tls` defines the TLS configuration.

_Optional_

An empty `tls` section enables TLS.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls: {}
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.tls]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls: {}
```

#### `tls.serverName`

_Optional_

`tls.serverName` configure the server name that will be used for SNI.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls:
        serverName: "myhost"
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.tls]
  serverName = "myhost"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    serverName: "test"
```

#### `tls.certificates`

_Optional_

`tls.certificates` is the list of certificates (as file paths, or data bytes)
that will be set as client certificates for mTLS.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls:
        certificates:
          - certFile: foo.crt
            keyFile: bar.crt
```

```toml tab="File (TOML)"
## Dynamic configuration
[[tcp.serversTransports.mytransport.tls.certificates]]
  certFile = "foo.crt"
  keyFile = "bar.crt"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    certificatesSecrets:
      - mycert

---
apiVersion: v1
kind: Secret
metadata:
  name: mycert

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

#### `tls.insecureSkipVerify`

_Optional_

`tls.insecureSkipVerify` controls whether the server's certificate chain and host name is verified.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.tls]
  insecureSkipVerify = true
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    insecureSkipVerify: true
```

#### `tls.rootCAs`

_Optional_

`tls.rootCAs` defines the set of root certificate authorities (as file paths, or data bytes) to use when verifying server certificates.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls:
        rootCAs:
          - foo.crt
          - bar.crt
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.tls]
  rootCAs = ["foo.crt", "bar.crt"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    rootCAsSecrets:
      - myca
---
apiVersion: v1
kind: Secret
metadata:
  name: myca

data:
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
```

#### `tls.peerCertURI`

_Optional, Default=false_

`tls.peerCertURI` defines the URI used to match against SAN URIs during the server's certificate verification.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      tls:
        peerCertURI: foobar
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.tls]
  peerCertURI = "foobar"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  tls:
    peerCertURI: foobar
```

#### `spiffe`

Please note that [SPIFFE](../../https/spiffe.md) must be enabled in the static configuration
before using it to secure the connection between Traefik and the backends.

##### `spiffe.ids`

_Optional_

`ids` defines the allowed SPIFFE IDs.
This takes precedence over the SPIFFE TrustDomain.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      spiffe:
        ids:
          - spiffe://trust-domain/id1
          - spiffe://trust-domain/id2
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.spiffe]
  ids = ["spiffe://trust-domain/id1", "spiffe://trust-domain/id2"]
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
    spiffe:
      ids:
        - spiffe://trust-domain/id1
        - spiffe://trust-domain/id2
```

##### `spiffe.trustDomain`

_Optional_

`trustDomain` defines the allowed SPIFFE trust domain.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
        spiffe:
          trustDomain: spiffe://trust-domain
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport.spiffe]
  trustDomain = "spiffe://trust-domain"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
    spiffe:
      trustDomain: "spiffe://trust-domain"
```

## Configuring UDP Services

### General

Each of the fields of the service section represents a kind of service.
Which means, that for each specified service, one of the fields, and only one,
has to be enabled to define what kind of service is created.
Currently, the two available kinds are `LoadBalancer`, and `Weighted`.

### Servers Load Balancer

The servers load balancer is in charge of balancing the requests between the servers of the same service.

??? example "Declaring a Service with Two Servers -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    udp:
      services:
        my-service:
          loadBalancer:
            servers:
            - address: "xx.xx.xx.xx:xx"
            - address: "xx.xx.xx.xx:xx"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [udp.services]
      [udp.services.my-service.loadBalancer]
        [[udp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
        [[udp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
    ```

#### Servers

The Servers field defines all the servers that are part of this load-balancing group,
i.e. each address (IP:Port) on which an instance of the service's program is deployed.

??? example "A Service with One Server -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    udp:
      services:
        my-service:
          loadBalancer:
            servers:
              - address: "xx.xx.xx.xx:xx"
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [udp.services]
      [udp.services.my-service.loadBalancer]
        [[udp.services.my-service.loadBalancer.servers]]
          address = "xx.xx.xx.xx:xx"
    ```

### Weighted Round Robin

The Weighted Round Robin (alias `WRR`) load-balancer of services is in charge of balancing the requests between multiple services based on provided weights.

This strategy is only available to load balance between [services](./index.md) and not between [servers](./index.md#servers).

This strategy can only be defined with [File](../../providers/file.md).

```yaml tab="YAML"
## Dynamic configuration
udp:
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

```toml tab="TOML"
## Dynamic configuration
[udp.services]
  [udp.services.app]
    [[udp.services.app.weighted.services]]
      name = "appv1"
      weight = 3
    [[udp.services.app.weighted.services]]
      name = "appv2"
      weight = 1

  [udp.services.appv1]
    [udp.services.appv1.loadBalancer]
      [[udp.services.appv1.loadBalancer.servers]]
        address = "private-ip-server-1:8080/"

  [udp.services.appv2]
    [udp.services.appv2.loadBalancer]
      [[udp.services.appv2.loadBalancer.servers]]
        address = "private-ip-server-2:8080/"
```

{!traefik-for-business-applications.md!}
