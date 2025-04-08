---
title: "Traefik Kubernetes Services Documentation"
description: "Learn how to configure routing and load balancing in Traefik Proxy to reach Services, which handle incoming requests. Read the technical documentation."
--- 

A `TraefikService` is a custom resource that sits on top of the Kubernetes Services. It enables advanced load-balancing features such as a [Weighted Round Robin](#weighted-round-robin) load balancing or a [Mirroring](#mirroring) between your Kubernetes Services.

Services configure how to reach the actual endpoints that will eventually handle incoming requests. In Traefik, the target service can be either a standard [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)—which exposes a pod—or a TraefikService. The latter allows you to combine advanced load-balancing options like:

- [Weighted Round Robin load balancing](#weighted-round-robin).
- [Mirroring](#mirroring). 

## Weighted Round Robin

The WRR is able to load balance the requests between multiple services based on weights. The WRR `TraefikService` allows you to load balance the traffic between Kubernetes Services and other instances of `TraefikService` (another WRR service -, or a mirroring service).

### Configuration Example

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: test-name
  namespace: apps

spec:
  entryPoints:
  - websecure
  routes:
  - match: Host(`example.com`) && PathPrefix(`/foo`)
    kind: Rule
    services:
    # Set a WRR TraefikService
    - name: wrr1
      namespace: apps
      kind: TraefikService
  tls:
    # Add a TLS certificate from a Kubernetes Secret
    secretName: supersecret
```

```yaml tab="TraefikService WRR Level#1"
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: wrr1
  namespace: apps

spec:
  weighted:
    services:
        # Kubernetes Service
      - name: svc1
        namespace: apps
        port: 80
        weight: 1
        # Second level WRR service
      - name: wrr2
        namespace: apps
        kind: TraefikService
        weight: 1
        # Mirroring service
        # The service is described in the Mirroring example
      - name: mirror1
        namespace: apps
        kind: TraefikService
        weight: 1
```

```yaml tab="TraefikService WRR Level#2"
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: wrr2
  namespace: apps

spec:
  weighted:
    services:
      # Kubernetes Service
      - name: svc2
        namespace: apps
        port: 80
        weight: 1
      # Kubernetes Service
      - name: svc3
        namespace: apps
        port: 80
        weight: 1

```

```yaml tab="Kubernetes Services"
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: apps

spec:
  ports:
  - name: http
      port: 80
  selector:
    app: traefiklabs
    task: app1
---
apiVersion: v1
kind: Service
metadata:
  name: svc2
  namespace: apps

spec:
  ports:
  - name: http
    port: 80
  selector:
    app: traefiklabs
    task: app2
---
apiVersion: v1
kind: Service
metadata:
  name: svc3
  namespace: apps

spec:
  ports:
  - name: http
    port: 80
  selector:
    app: traefiklabs
    task: app3
```

```yaml tab="Secret"
apiVersion: v1
kind: Secret
metadata:
  name: supersecret
  namespace: apps

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
```

### Configuration Options

| Field                                                          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Default                                                              | Required |
|:---------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| `services`                                                     | List of any combination of TraefikService and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br />.                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| `services[m].`<br />`kind`                                     | Kind of the service targeted.<br />Two values allowed:<br />- **Service**: Kubernetes Service<br /> - **TraefikService**: Traefik Service.                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| `services[m].`<br />`name`                                     | Service name.<br />The character `@` is not authorized.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | ""                                                                   | Yes      |
| `services[m].`<br />`namespace`                                | Service namespace.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| `services[m].`<br />`port`                                     | Service port (number or port name).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | ""                                                                   | No       |
| `services[m].`<br />`responseForwarding.`<br />`flushInterval` | Interval, in milliseconds, in between flushes to the client while copying the response body.<br />A negative value means to flush immediately after each write to the client.<br />This configuration is ignored when a response is a streaming response; for such responses, writes are flushed to the client immediately.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                          | 100ms                                                                | No       |
| `services[m].`<br />`scheme`                                   | Scheme to use for the request to the upstream Kubernetes Service.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | "http"<br />"https" if `port` is 443 or contains the string *https*. | No       |
| `services[m].`<br />`serversTransport`                         | Name of ServersTransport resource to use to configure the transport between Traefik and your servers.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                | ""                                                                   | No       |
| `services[m].`<br />`passHostHeader`                           | Forward client Host header to server.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | true                                                                 | No       |
| `services[m].`<br />`healthCheck.scheme`                       | Server URL scheme for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                        | ""                                                                   | No       |
| `services[m].`<br />`healthCheck.mode`                         | Health check mode.<br /> If defined to grpc, will use the gRPC health check protocol to probe the server.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                               | "http"                                                               | No       |
| `services[m].`<br />`healthCheck.path`                         | Server URL path for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                          | ""                                                                   | No       |
| `services[m].`<br />`healthCheck.interval`                     | Frequency of the health check calls for healthy targets.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName]`ExternalName`.                                                                                                                                                                                                                                                                                                                                                                  | "100ms"                                                              | No       |
| `services[m].`<br />`healthCheck.unhealthyInterval`            | Frequency of the health check calls for unhealthy targets.<br />When not defined, it defaults to the `interval` value.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName]`ExternalName`.                                                                                                                                                                                                                                                                                                    | "100ms"                                                              | No       |
| `services[m].`<br />`healthCheck.method`                       | HTTP method for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                              | "GET"                                                                | No       |
| `services[m].`<br />`healthCheck.status`                       | Expected HTTP status code of the response to the health check request.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type ExternalName.<br />If not set, expect a status between 200 and 399.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                              |                                                                      | No       |
| `services[m].`<br />`healthCheck.port`                         | URL port for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                                 |                                                                      | No       |
| `services[m].`<br />`healthCheck.timeout`                      | Maximum duration to wait before considering the server unhealthy.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                       | "5s"                                                                 | No       |
| `services[m].`<br />`healthCheck.hostname`                     | Value in the Host header of the health check request.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| `services[m].`<br />`healthCheck.`<br />`followRedirect`       | Follow the redirections during the healtchcheck.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                        | true                                                                 | No       |
| `services[m].`<br />`healthCheck.headers`                      | Map of header to send to the health check endpoint<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type `ExternalName`.                                                                                                                                                                                                                                                                                                                                                                                      |                                                                      | No       |
| `services[m].`<br />`sticky.`<br />`cookie.name`               | Name of the cookie used for the stickiness.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | Abbreviation of a sha1<br />(ex: `_1d52e`).                          | No       |
| `services[m].`<br />`sticky.`<br />`cookie.httpOnly`           | Allow the cookie can be accessed by client-side APIs, such as JavaScript.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | false                                                                | No       |
| `services[m].`<br />`sticky.`<br />`cookie.secure`             | Allow the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | false                                                                | No       |
| `services[m].`<br />`sticky.`<br />`cookie.sameSite`           | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy.<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                        | ""                                                                   | No       |
| `services[m].`<br />`sticky.`<br />`cookie.maxAge`             | Number of seconds until the cookie expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                  | 0                                                                    | No       |
| `services[m].`<br />`strategy`                                 | Load balancing strategy between the servers.<br />RoundRobin is the only supported value yet.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | "RoundRobin"                                                         | No       |
| `services[m].`<br />`weight`                                   | Service weight.<br />To use only to refer to WRR TraefikService                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| `services[m].`<br />`nativeLB`                                 | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                           | false                                                                | No       |
| `services[m].`<br />`nodePortLB`                               | Use the nodePort IP address when the service type is NodePort.<br />It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                            | false                                                                | No       |
| `sticky.`<br />`cookie.name`                                   | Name of the cookie used for the stickiness at the WRR service level.<br />When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.<br />On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.<br />If the server pecified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels) | Abbreviation of a sha1<br />(ex: `_1d52e`).                          | No       |
| `sticky.`<br />`cookie.httpOnly`                               | Allow the cookie used for the stickiness at the WRR service level to be accessed by client-side APIs, such as JavaScript.<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| `sticky.`<br />`cookie.secure`                                 | Allow the cookie used for the stickiness at the WRR service level to be only transmitted over an encrypted connection (i.e. HTTPS).<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                                                                                                                | false                                                                | No       |
| `sticky.`<br />`cookie.sameSite`                               | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy for the cookie used for the stickiness at the WRR service level.<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| `sticky.`<br />`cookie.maxAge`                                 | Number of seconds until the cookie used for the stickiness at the WRR service level expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.                                                                                                                                                                                                                                                                                                                                                                                                                                 | 0                                                                    | No       |

#### Stickiness on multiple levels

When chaining or mixing load-balancers (e.g. a load-balancer of servers is one of the "children" of a load-balancer of services),
for stickiness to work all the way, the option needs to be specified at all required levels.
Which means the client needs to send a cookie with as many key/value pairs as there are sticky levels.

Sticky sessions, for stickiness to work all the way, must be specified at each load-balancing level.

For instance, in the example below, there is a first level of load-balancing because there is a (Weighted Round Robin) load-balancing of the two `whoami` services,
and there is a second level because each whoami service is a `replicaset` and is thus handled as a load-balancer of servers.

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutebar
  namespace: apps

spec:
  entryPoints:
  - web
  routes:
  - match: Host(`example.com`) && PathPrefix(`/foo`)
    kind: Rule
    services:
    - name: wrr1
      namespace: apps
      kind: TraefikService
```

```yaml tab="TraefikService WRR with 2 level of stickiness"
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: wrr1
  namespace: apps
  
spec:
  weighted:
    services:
    - name: whoami1
      kind: Service
      port: 80
      weight: 1
      # Stickiness level2 (on the Kubernetes service)
      sticky:
        cookie:
        name: lvl2
    - name: whoami2
      kind: Service
      weight: 1
      port: 80
      # Stickiness level2 (on the Kubernetes service)
      sticky:
        cookie:
        name: lvl2
  # Stickiness level2 (on the WRR service)
  sticky:
    cookie:
    name: lvl1
```

In the example above, to keep a session open with the same server, the client would then need to specify the two levels within the cookie for each request, e.g. with curl:

```bash
# Assuming `10.42.0.6` is the IP address of one of the replicas (a pod then) of the `whoami1` service.
curl -H Host:example.com -b "lvl1=default-whoami1-80; lvl2=http://10.42.0.6:80" http://localhost:8000/foo
```

## Mirroring

The mirroring is able to mirror requests sent to a service to other services.

A mirroring service allows you to send the trafiic to many services together:

- The **main** service receives 100% of the traffic,
- The **mirror** services receive a percentage of the traffic.

For example, to upgrade the version of your application. You can set the service that targets current version as the **main** service, and the service of the new version a **mirror** service.
Thus you can start testing the behavior of the new version keeping the current version reachable.

The mirroring `TraefikService` allows you to reference Kubernetes Services and other instances of `TraefikService` (another WRR service -, or a mirroring service).

Please note that by default the whole request is buffered in memory while it is being mirrored.
See the `maxBodySize` option in the example below for how to modify this behavior.

### Configuration Example

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutebar
  namespace: default

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`example.com`) && PathPrefix(`/foo`)
    kind: Rule
    services:
    - name: mirror1
      namespace: default
      kind: TraefikService
```

```yaml tab="Mirroring from a Kubernetes Service"
# Mirroring from a k8s Service
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: mirror1
  namespace: default

spec:
  mirroring:
    name: svc1                      # svc1 receives 100% of the traffic
    port: 80
    mirrors:
      - name: svc2                  # svc2 receives a copy of 20% of this traffic
        port: 80
        percent: 20
      - name: svc3                  # svc3 receives a copy of 15% of this traffic
        kind: TraefikService
        percent: 15
```

```yaml tab="Mirroring from a TraefikService (WRR)"
# Mirroring from a Traefik Service
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: mirror1
  namespace: default

spec:
  mirroring:
    name: wrr1                      # wrr1 receives 100% of the traffic
    kind: TraefikService
    mirrors:
      - name: svc2                  # svc2 receives a copy of 20% of this traffic
        port: 80
        percent: 20
      - name: svc3                  # svc3 receives a copy of 10% of this traffic
        kind: TraefikService
        percent: 10
```

```yaml tab="Kubernetes Services"
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: default

spec:
  ports:
    - name: http
      port: 80
  selector:
    app: traefiklabs
    task: app1
---
apiVersion: v1
kind: Service
metadata:
  name: svc2
  namespace: default

spec:
  ports:
    - name: http
      port: 80
  selector:
    app: traefiklabs
    task: app2
```

### Configuration Options

!!!note "Main and mirrored services"

    The main service properties are set as the option root level.

    The mirrored services properties are set in the `mirrors` list.

| Field                                                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Default                                                              | Required |
|:--------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| `kind`                                                        | Kind of the main service.<br />Two values allowed:<br />- **Service**: Kubernetes Service<br />- **TraefikService**: Traefik Service.<br />More information [here](#services)                                                                                                                                                                                                                                                                                                                                                                                                     | ""                                                                   | No       |
| `name`                                                        | Main service name.<br />The character `@` is not authorized.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | ""                                                                   | Yes      |
| `namespace`                                                   | Main service namespace.<br />More information [here](#services).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | ""                                                                   | No       |
| `port`                                                        | Main service port (number or port name).<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| `responseForwarding.`<br />`flushInterval`                    | Interval, in milliseconds, in between flushes to the client while copying the response body.<br />A negative value means to flush immediately after each write to the client.<br />This configuration is ignored when a response is a streaming response; for such responses, writes are flushed to the client immediately.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                   | 100ms                                                                | No       |
| `scheme`                                                      | Scheme to use for the request to the upstream Kubernetes Service.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                             | "http"<br />"https" if `port` is 443 or contains the string *https*. | No       |
| `serversTransport`                                            | Name of ServersTransport resource to use to configure the transport between Traefik and the main service's servers.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| `passHostHeader`                                              | Forward client Host header to main service's server.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                          | true                                                                 | No       |
| `healthCheck.scheme`                                          | Server URL scheme for the health check endpoint.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| `healthCheck.mode`                                            | Health check mode.<br /> If defined to grpc, will use the gRPC health check protocol to probe the server.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                             | "http"                                                               | No       |
| `healthCheck.path`                                            | Server URL path for the health check endpoint.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                        | ""                                                                   | No       |
| `healthCheck.interval`                                        | Frequency of the health check calls for healthy targets.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                              | "100ms"                                                              | No       |
| `healthCheck.unhealthyInterval`                               | Frequency of the health check calls for unhealthy targets.<br />When not defined, it defaults to the `interval` value.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                | "100ms"                                                              | No       |
| `healthCheck.method`                                          | HTTP method for the health check endpoint.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                            | "GET"                                                                | No       |
| `healthCheck.status`                                          | Expected HTTP status code of the response to the health check request.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type ExternalName.<br />If not set, expect a status between 200 and 399.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                       |                                                                      | No       |
| `healthCheck.port`                                            | URL port for the health check endpoint.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                               |                                                                      | No       |
| `healthCheck.timeout`                                         | Maximum duration to wait before considering the server unhealthy.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                     | "5s"                                                                 | No       |
| `healthCheck.hostname`                                        | Value in the Host header of the health check request.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                 | ""                                                                   | No       |
| `healthCheck.`<br />`followRedirect`                          | Follow the redirections during the healtchcheck.<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                      | true                                                                 | No       |
| `healthCheck.headers`                                         | Map of header to send to the health check endpoint<br />Evaluated only if the kind of the main service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                    |                                                                      | No       |
| `sticky.`<br />`cookie.name`                                  | Name of the cookie used for the stickiness on the main service.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                               | Abbreviation of a sha1<br />(ex: `_1d52e`).                          | No       |
| `sticky.`<br />`cookie.httpOnly`                              | Allow the cookie can be accessed by client-side APIs, such as JavaScript.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                     | false                                                                | No       |
| `sticky.`<br />`cookie.secure`                                | Allow the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                           | false                                                                | No       |
| `sticky.`<br />`cookie.sameSite`                              | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy.<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                 | ""                                                                   | No       |
| `sticky.`<br />`cookie.maxAge`                                | Number of seconds until the cookie expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                           | 0                                                                    | No       |
| `strategy`                                                    | Load balancing strategy between the main service's servers.<br />RoundRobin is the only supported value yet.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                  | "RoundRobin"                                                         | No       |
| `weight`                                                      | Service weight.<br />To use only to refer to WRR TraefikService                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| `nativeLB`                                                    | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                    | false                                                                | No       |
| `nodePortLB`                                                  | Use the nodePort IP address when the service type is NodePort.<br />It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br />Evaluated only if the kind of the main service is **Service**.                                                                                                                                                                                                                                                                                                     | false                                                                | No       |
| `maxBodySize`                                                 | Maximum size allowed for the body of the request.<br />If the body is larger, the request is not mirrored.<br />-1 means unlimited size.                                                                                                                                                                                                                                                                                                                                                                                                                                          | -1                                                                   | No       |
| `mirrors`                                                     | List of mirrored services to target.<br /> It can be any combination of TraefikService and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br />More information [here](#services).                                                                                                                                                                                                                                                                                                                                                      |                                                                      | No       |
| `mirrors[m].`<br />`kind`                                     | Kind of the mirrored service targeted.<br />Two values allowed:<br />- **Service**: Kubernetes Service<br />- **TraefikService**: Traefik Service.<br />More information [here](#services)                                                                                                                                                                                                                                                                                                                                                                                        | ""                                                                   | No       |
| `mirrors[m].`<br />`name`                                     | Mirrored service name.<br />The character `@` is not authorized.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | ""                                                                   | Yes      |
| `mirrors[m].`<br />`namespace`                                | Mirrored service namespace.<br />More information [here](#services).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | ""                                                                   | No       |
| `mirrors[m].`<br />`port`                                     | Mirrored service port (number or port name).<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                              | ""                                                                   | No       |
| `mirrors[m].`<br />`percent`                                  | Part of the traffic to mirror in percent (from 0 to 100)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | 0                                                                    | No       |
| `mirrors[m].`<br />`responseForwarding.`<br />`flushInterval` | Interval, in milliseconds, in between flushes to the client while copying the response body.<br />A negative value means to flush immediately after each write to the client.<br />This configuration is ignored when a response is a streaming response; for such responses, writes are flushed to the client immediately.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                               | 100ms                                                                | No       |
| `mirrors[m].`<br />`scheme`                                   | Scheme to use for the request to the mirrored service.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                    | "http"<br />"https" if `port` is 443 or contains the string *https*. | No       |
| `mirrors[m].`<br />`serversTransport`                         | Name of ServersTransport resource to use to configure the transport between Traefik and the mirrored service servers.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                     | ""                                                                   | No       |
| `mirrors[m].`<br />`passHostHeader`                           | Forward client Host header to the mirrored service servers.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                               | true                                                                 | No       |
| `mirrors[m].`<br />`healthCheck.scheme`                       | Server URL scheme for the health check endpoint.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                  | ""                                                                   | No       |
| `mirrors[m].`<br />`healthCheck.mode`                         | Health check mode.<br /> If defined to grpc, will use the gRPC health check protocol to probe the server.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                         | "http"                                                               | No       |
| `mirrors[m].`<br />`healthCheck.path`                         | Server URL path for the health check endpoint.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                    | ""                                                                   | No       |
| `mirrors[m].`<br />`healthCheck.interval`                     | Frequency of the health check calls.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                              | "100ms"                                                              | No       |
| `mirrors[m].`<br />`healthCheck.unhealthyInterval`            | Frequency of the health check calls for unhealthy targets.<br />When not defined, it defaults to the `interval` value.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                            | "100ms"                                                              | No       |
| `mirrors[m].`<br />`healthCheck.method`                       | HTTP method for the health check endpoint.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                        | "GET"                                                                | No       |
| `mirrors[m].`<br />`healthCheck.status`                       | Expected HTTP status code of the response to the health check request.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type ExternalName.<br />If not set, expect a status between 200 and 399.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                   |                                                                      | No       |
| `mirrors[m].`<br />`healthCheck.port`                         | URL port for the health check endpoint.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                           |                                                                      | No       |
| `mirrors[m].`<br />`healthCheck.timeout`                      | Maximum duration to wait before considering the server unhealthy.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                 | "5s"                                                                 | No       |
| `mirrors[m].`<br />`healthCheck.hostname`                     | Value in the Host header of the health check request.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                             | ""                                                                   | No       |
| `mirrors[m].`<br />`healthCheck.`<br />`followRedirect`       | Follow the redirections during the healtchcheck.<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                  | true                                                                 | No       |
| `mirrors[m].`<br />`healthCheck.headers`                      | Map of header to send to the health check endpoint<br />Evaluated only if the kind of the mirrored service is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#services).                                                                                                                                                                                                                                                                                                                |                                                                      | No       |
| `mirrors[m].`<br />`sticky.`<br />`cookie.name`               | Name of the cookie used for the stickiness.<br />When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.<br />On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.<br />If the server pecified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).<br />Evaluated only if the kind of the mirrored service is **Service**. | ""                                                                   | No       |
| `mirrors[m].`<br />`sticky.`<br />`cookie.httpOnly`           | Allow the cookie can be accessed by client-side APIs, such as JavaScript.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                 | false                                                                | No       |
| `mirrors[m].`<br />`sticky.`<br />`cookie.secure`             | Allow the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                       | false                                                                | No       |
| `mirrors[m].`<br />`sticky.`<br />`cookie.sameSite`           | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy.<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                             | ""                                                                   | No       |
| `mirrors[m].`<br />`sticky.`<br />`cookie.maxAge`             | Number of seconds until the cookie expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                       | 0                                                                    | No       |
| `mirrors[m].`<br />`strategy`                                 | Load balancing strategy between the servers.<br />RoundRobin is the only supported value yet.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                             | "RoundRobin"                                                         | No       |
| `mirrors[m].`<br />`weight`                                   | Service weight.<br />To use only to refer to WRR TraefikService                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| `mirrors[m].`<br />`nativeLB`                                 | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                | false                                                                | No       |
| `mirrors[m].`<br />`nodePortLB`                               | Use the nodePort IP address when the service type is NodePort.<br />It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br />Evaluated only if the kind of the mirrored service is **Service**.                                                                                                                                                                                                                                                                                                 | false                                                                | No       |
| `mirrorBody`                                                  | Defines whether the request body should be mirrored.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | true                                                                 | No       |
