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
| <a id="opt-services" href="#opt-services" title="#opt-services">`services`</a> | List of any combination of TraefikService and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br />. Exhaustive list of option in the [`Service`](./service.md#configuration-options) documentation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| <a id="opt-servicesm-weight" href="#opt-servicesm-weight" title="#opt-servicesm-weight">`services[m].weight`</a> | Service weight.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | ""                                                                   | No       |
| <a id="opt-sticky-cookie-name" href="#opt-sticky-cookie-name" title="#opt-sticky-cookie-name">`sticky.`<br />`cookie.name`</a> | Name of the cookie used for the stickiness at the WRR service level.<br />When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.<br />On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.<br />If the server pecified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels) | Abbreviation of a sha1<br />(ex: `_1d52e`).                          | No       |
| <a id="opt-sticky-cookie-httpOnly" href="#opt-sticky-cookie-httpOnly" title="#opt-sticky-cookie-httpOnly">`sticky.`<br />`cookie.httpOnly`</a> | Allow the cookie used for the stickiness at the WRR service level to be accessed by client-side APIs, such as JavaScript.<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| <a id="opt-sticky-cookie-secure" href="#opt-sticky-cookie-secure" title="#opt-sticky-cookie-secure">`sticky.`<br />`cookie.secure`</a> | Allow the cookie used for the stickiness at the WRR service level to be only transmitted over an encrypted connection (i.e. HTTPS).<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                                                                                                                | false                                                                | No       |
| <a id="opt-sticky-cookie-sameSite" href="#opt-sticky-cookie-sameSite" title="#opt-sticky-cookie-sameSite">`sticky.`<br />`cookie.sameSite`</a> | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy for the cookie used for the stickiness at the WRR service level.<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />More information about WRR stickiness [here](#stickiness-on-multiple-levels)                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| <a id="opt-sticky-cookie-maxAge" href="#opt-sticky-cookie-maxAge" title="#opt-sticky-cookie-maxAge">`sticky.`<br />`cookie.maxAge`</a> | Number of seconds until the cookie used for the stickiness at the WRR service level expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.                                                                                                                                                                                                                                                                                                                                                                                                                                 | 0                                                                    | No       |

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
    mirrorBody: true                # Set to false by default
    maxBodySize: 1M
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

#### Main Service Options

The main service properties are set as the option root level.

The main service provides the same options as a [`Service`](./service.md).

The exhaustive list of the service options is described in the [`Service`](./service.md#configuration-options) documentation.
The mirror main service dedicated option are described below.

| Field                                                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Default                                                              | Required |
|:--------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| <a id="opt-mirrorBody" href="#opt-mirrorBody" title="#opt-mirrorBody">`mirrorBody`</a> | Defines whether the request body should be mirrored.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | true                                                                 | No       |
| <a id="opt-maxBodySize" href="#opt-maxBodySize" title="#opt-maxBodySize">`maxBodySize`</a> | Maximum size allowed for the body of the request.<br />If the body is larger, the request is not mirrored.<br />-1 means unlimited size.                                                                                                                                                                                                                                                                                                                                                                                                                                          | -1                                                                   | No       |
| <a id="opt-mirrors" href="#opt-mirrors" title="#opt-mirrors">`mirrors`</a> | List of mirrored services to target.<br /> It can be any combination of TraefikService and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br />Exhaustive list of option in the [`Service`](./service.md#configuration-options) documentation.                                                                                                                                                                                                                                                                                                                                                      |                                                                   | Yes       |

#### Mirrored Services Options

The mirrored services properties are set in the `mirrors` list.

A mirrored service provides the same options as a [`Service`](./service.md).

The exhaustive list of the service options is described in the [`Service`](./service.md#configuration-options) documentation.
The mirrorerd service dedicated option are described below.


| Field                                                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Default                                                              | Required |
|:--------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| <a id="opt-mirrorsm-percent" href="#opt-mirrorsm-percent" title="#opt-mirrorsm-percent">`mirrors[m].percent`</a> | Traffic percentage to route to the service.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | 0                                                                   | No       |
