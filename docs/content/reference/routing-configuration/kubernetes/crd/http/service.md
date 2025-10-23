---
title: "Kubernetes Service"
description: "A Service is a not Traefik CRD, it allows you to describe the Service option in an IngressRoute or a Traefik Service."
---

`Service` is the implementation of a [Traefik HTTP service](../../../http/load-balancing/service.md). 

There is no dedicated CRD, a `Service` is part of:

- [`IngressRoute`](./ingressroute.md)
- [`TraefikService`](./traefikservice.md)

Note that, before creating `IngressRoute` or `TraefikService` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the Traefik-specific resources.

## Configuration Example

You can declare a `Service` either as part of an `IngressRoute` or a `TraefikService` as detailed below:

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: test-name
  namespace: apps
spec:
  entryPoints:
    - web
  routes:
  - kind: Rule
    # Rule on the Host
    match: Host(`test.example.com`)
    services:
    # Target a Kubernetes Service
    - kind: Service
      name: foo
      namespace: apps
      # Customize the connection between Traefik and the backend
      passHostHeader: true
      port: 80
      responseForwarding:
        flushInterval: 1ms
      scheme: https
      sticky:
        cookie:
          httpOnly: true
          name: cookie
          secure: true
      strategy: wrr
```

```yaml tab="TraefikService"
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: wrr1
  namespace: apps

spec:
  weighted:
    services:
    # Target a Kubernetes Service
    - kind: Service
      name: foo
      namespace: apps
      # Customize the connection between Traefik and the backend
      passHostHeader: true
      port: 80
      responseForwarding:
        flushInterval: 1ms
      scheme: https
      sticky:
        cookie:
          httpOnly: true
          name: cookie
          secure: true
      strategy: wrr
```

## Configuration Options

| Field                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Default                                                              | Required |
|:---------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| <a id="opt-kind" href="#opt-kind" title="#opt-kind">`kind`</a> | Kind of the service targeted.<br />Two values allowed:<br />- **Service**: Kubernetes Service<br /> **TraefikService**: Traefik Service.<br />More information [here](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                  | "Service"                                                            | No       |
| <a id="opt-name" href="#opt-name" title="#opt-name">`name`</a> | Service name.<br />The character `@` is not authorized. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                            |                                                                      | Yes      |
| <a id="opt-namespace" href="#opt-namespace" title="#opt-namespace">`namespace`</a> | Service namespace.<br />Can be empty if the service belongs to the same namespace as the IngressRoute. <br />More information [here](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                                                   |                                                                      | No       |
| <a id="opt-port" href="#opt-port" title="#opt-port">`port`</a> | Service port (number or port name).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |                                                                      | No       |
| <a id="opt-responseForwarding-flushInterval" href="#opt-responseForwarding-flushInterval" title="#opt-responseForwarding-flushInterval">`responseForwarding.`<br />`flushInterval`</a> | Interval, in milliseconds, in between flushes to the client while copying the response body.<br />A negative value means to flush immediately after each write to the client.<br />This configuration is ignored when a response is a streaming response; for such responses, writes are flushed to the client immediately.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                    | 100ms                                                                | No       |
| <a id="opt-scheme" href="#opt-scheme" title="#opt-scheme">`scheme`</a> | Scheme to use for the request to the upstream Kubernetes Service.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                              | "http"<br />"https" if `port` is 443 or contains the string *https*. | No       |
| <a id="opt-serversTransport" href="#opt-serversTransport" title="#opt-serversTransport">`serversTransport`</a> | Name of ServersTransport resource to use to configure the transport between Traefik and your servers.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                          | ""                                                                   | No       |
| <a id="opt-passHostHeader" href="#opt-passHostHeader" title="#opt-passHostHeader">`passHostHeader`</a> | Forward client Host header to server.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | true                                                                 | No       |
| <a id="opt-healthCheck-scheme" href="#opt-healthCheck-scheme" title="#opt-healthCheck-scheme">`healthCheck.scheme`</a> | Server URL scheme for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| <a id="opt-healthCheck-mode" href="#opt-healthCheck-mode" title="#opt-healthCheck-mode">`healthCheck.mode`</a> | Health check mode.<br /> If defined to grpc, will use the gRPC health check protocol to probe the server.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                  | "http"                                                               | No       |
| <a id="opt-healthCheck-path" href="#opt-healthCheck-path" title="#opt-healthCheck-path">`healthCheck.path`</a> | Server URL path for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                             | ""                                                                   | No       |
| <a id="opt-healthCheck-interval" href="#opt-healthCheck-interval" title="#opt-healthCheck-interval">`healthCheck.interval`</a> | Frequency of the health check calls for healthy targets.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                   | "100ms"                                                              | No       |
| <a id="opt-healthCheck-unhealthyInterval" href="#opt-healthCheck-unhealthyInterval" title="#opt-healthCheck-unhealthyInterval">`healthCheck.unhealthyInterval`</a> | Frequency of the health check calls for unhealthy targets.<br />When not defined, it defaults to the `interval` value.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                     | "100ms"                                                              | No       |
| <a id="opt-healthCheck-method" href="#opt-healthCheck-method" title="#opt-healthCheck-method">`healthCheck.method`</a> | HTTP method for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                 | "GET"                                                                | No       |
| <a id="opt-healthCheck-status" href="#opt-healthCheck-status" title="#opt-healthCheck-status">`healthCheck.status`</a> | Expected HTTP status code of the response to the health check request.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type ExternalName.<br />If not set, expect a status between 200 and 399.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                        |                                                                      | No       |
| <a id="opt-healthCheck-port" href="#opt-healthCheck-port" title="#opt-healthCheck-port">`healthCheck.port`</a> | URL port for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                    |                                                                      | No       |
| <a id="opt-healthCheck-timeout" href="#opt-healthCheck-timeout" title="#opt-healthCheck-timeout">`healthCheck.timeout`</a> | Maximum duration to wait before considering the server unhealthy.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                          | "5s"                                                                 | No       |
| <a id="opt-healthCheck-hostname" href="#opt-healthCheck-hostname" title="#opt-healthCheck-hostname">`healthCheck.hostname`</a> | Value in the Host header of the health check request.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| <a id="opt-healthCheck-followRedirect" href="#opt-healthCheck-followRedirect" title="#opt-healthCheck-followRedirect">`healthCheck.`<br />`followRedirect`</a> | Follow the redirections during the healtchcheck.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                           | true                                                                 | No       |
| <a id="opt-healthCheck-headers" href="#opt-healthCheck-headers" title="#opt-healthCheck-headers">`healthCheck.headers`</a> | Map of header to send to the health check endpoint<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service)).                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| <a id="opt-sticky-cookie-name" href="#opt-sticky-cookie-name" title="#opt-sticky-cookie-name">`sticky.`<br />`cookie.name`</a> | Name of the cookie used for the stickiness.<br />When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.<br />On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.<br />If the server pecified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).<br />Evaluated only if the kind is **Service**.      | ""                                                                   | No       |
| <a id="opt-sticky-cookie-httpOnly" href="#opt-sticky-cookie-httpOnly" title="#opt-sticky-cookie-httpOnly">`sticky.`<br />`cookie.httpOnly`</a> | Allow the cookie can be accessed by client-side APIs, such as JavaScript.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                      | false                                                                | No       |
| <a id="opt-sticky-cookie-secure" href="#opt-sticky-cookie-secure" title="#opt-sticky-cookie-secure">`sticky.`<br />`cookie.secure`</a> | Allow the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                            | false                                                                | No       |
| <a id="opt-sticky-cookie-sameSite" href="#opt-sticky-cookie-sameSite" title="#opt-sticky-cookie-sameSite">`sticky.`<br />`cookie.sameSite`</a> | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| <a id="opt-sticky-cookie-maxAge" href="#opt-sticky-cookie-maxAge" title="#opt-sticky-cookie-maxAge">`sticky.`<br />`cookie.maxAge`</a> | Number of seconds until the cookie expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                            | 0                                                                    | No       |
| <a id="opt-strategy" href="#opt-strategy" title="#opt-strategy">`strategy`</a> | Strategy defines the load balancing strategy between the servers.<br />Supported values are: wrr (Weighed round-robin), p2c (Power of two choices), hrw (Highest Random Weight), and leasttime (Least-Time).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                   | "RoundRobin"                                                         | No       |
| <a id="opt-nativeLB" href="#opt-nativeLB" title="#opt-nativeLB">`nativeLB`</a> | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik.<br /> Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                    | false                                                                | No       |
| <a id="opt-nodePortLB" href="#opt-nodePortLB" title="#opt-nodePortLB">`nodePortLB`</a> | Use the nodePort IP address when the service type is NodePort.<br />It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                      | false                                                                | No       |


### ExternalName Service

Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port. Accordingly, Traefik supports defining a port in two ways:

- only on `IngressRoute` service
- on both sides, you'll be warned if the ports don't match, and the `IngressRoute` service port is used

Thus, in case of two sides port definition, Traefik expects a match between ports.

=== "Ports defined on Resource"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: apps

    spec:
      entryPoints:
        - foo
      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc
          port: 80
    ```

    ```yaml tab="Service ExternalName"
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: apps

    spec:
      externalName: external.domain
      type: ExternalName
    ```

=== "Port defined on the Service"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: apps

    spec:
      entryPoints:
        - foo
      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc
    ```

    ```yaml tab="Service ExternalName"
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: apps

    spec:
      externalName: external.domain
      type: ExternalName
      ports:
        - port: 80
    ```

=== "Port defined on both sides"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: apps

    spec:
      entryPoints:
        - foo
      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc
          port: 80
    ```

    ```yaml tab="Service ExternalName"
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: apps

    spec:
      externalName: external.domain
      type: ExternalName
      ports:
        - port: 80
    ```

### Port Definition

Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port. Accordingly, Traefik supports defining a port in two ways:

- only on `IngressRoute` service
- on both sides, you'll be warned if the ports don't match, and the `IngressRoute` service port is used

Thus, in case of two sides port definition, Traefik expects a match between ports.

??? example   

    ```yaml tab="IngressRoute"
    ---
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: default

    spec:
      entryPoints:
        - foo

      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc
          port: 80

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: default
    spec:
      externalName: external.domain
      type: ExternalName
    ```

    ```yaml tab="ExternalName Service"
    ---
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: default

    spec:
      entryPoints:
        - foo

      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: default
    spec:
      externalName: external.domain
      type: ExternalName
      ports:
        - port: 80
    ```

    ```yaml tab="Both sides"
    ---
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: default

    spec:
      entryPoints:
        - foo

      routes:
      - match: Host(`example.net`)
        kind: Rule
        services:
        - name: external-svc
          port: 80

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: external-svc
      namespace: default
    spec:
      externalName: external.domain
      type: ExternalName
      ports:
        - port: 80
    ```

### Load Balancing

You can declare and use Kubernetes Service load balancing as detailed below:

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
    - name: svc1
      namespace: default
    - name: svc2
      namespace: default
```

```yaml tab="K8s Service"
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

!!! important "Kubernetes Service Native Load-Balancing"

    To avoid creating the server load-balancer with the pod IPs and use Kubernetes Service clusterIP directly,
    one should set the service `NativeLB` option to true.
    Please note that, by default, Traefik reuses the established connections to the backends for performance purposes. This can prevent the requests load balancing between the replicas from behaving as one would expect when the option is set.
    By default, `NativeLB` is false.

    ??? example "Example"

        ```yaml
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRoute
        metadata:
          name: test.route
          namespace: default

        spec:
          entryPoints:
            - foo

          routes:
          - match: Host(`example.net`)
            kind: Rule
            services:
            - name: svc
              port: 80
              # Here, nativeLB instructs to build the server load-balancer with the Kubernetes Service clusterIP only.
              nativeLB: true

        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: svc
          namespace: default
        spec:
          type: ClusterIP
          ...
        ```

### Configuring Backend Protocol

There are 3 ways to configure the backend protocol for communication between Traefik and your pods:

- Setting the scheme explicitly (http/https/h2c)
- Configuring the name of the kubernetes service port to start with https (https)
- Setting the kubernetes service port to use port 443 (https)

If you do not configure the above, Traefik will assume an http connection.
