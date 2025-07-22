---
title: "Kubernetes IngressRoute"
description: "An IngressRoute is a Traefik CRD is in charge of connecting incoming requests to the Services that can handle them in HTTP."
---

`IngressRoute` is the CRD implementation of a [Traefik HTTP router](../../../http/router/rules-and-priority.md).

Before creating `IngressRoute` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `IngressRoute` kind and other Traefik-specific resources.

## Configuration Example

You can declare an `IngressRoute` as detailed below:

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
    # Attach a middleware
    middlewares:
    - name: middleware1
      namespace: apps
    # Enable Router observability
    observability:
      accessLogs: true
      metrics: true
      tracing: true
    # Set a pirority
    priority: 10
    services:
    # Target a Kubernetes Support
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
      strategy: RoundRobin
      weight: 10
  tls:
    # Generate a TLS certificate using a certificate resolver
    certResolver: foo
    domains:
    - main: example.net
      sans:
      - a.example.net
      - b.example.net
    # Customize the TLS options
    options:
      name: opt
      namespace: apps
    # Add a TLS certificate from a Kubernetes Secret
    secretName: supersecret
```

## Configuration Options

| Field                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Default                                                              | Required |
|:---------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|:---------|
| `entryPoints`                                                                    | List of [entry points](../../../../install-configuration/entrypoints.md) names.<br />If not specified, HTTP routers will accept requests from all EntryPoints in the list of default EntryPoints.                                                                                                                                                                                                                                                                                                                                                                                                              |                                                                      | No       |
| `routes`                                                                         | List of routes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |                                                                      | Yes      |
| `routes[n].kind`                                                                 | Kind of router matching, only `Rule` is allowed yet.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | "Rule"                                                               | No       |
| `routes[n].match`                                                                | Defines the [rule](../../../http/router/rules-and-priority.md#rules) corresponding to an underlying router.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                      | Yes      |
| `routes[n].priority`                                                             | Defines the [priority](../../../http/router/rules-and-priority.md#priority-calculation) to disambiguate rules of the same length, for route matching.<br />If not set, the priority is directly equal to the length of the rule, and so the longest length has the highest priority.<br />A value of `0` for the priority is ignored, the default rules length sorting is used.                                                                                                                                                                                                                                | 0                                                                    | No       |
| `routes[n].middlewares`                                                          | List of middlewares to attach to the IngressRoute. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | ""                                                                   | No       |
| `routes[n].`<br />`middlewares[m].`<br />`name`                                  | Middleware name.<br />The character `@` is not authorized. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |                                                                      | Yes      |
| `routes[n].`<br />`middlewares[m].`<br />`namespace`                             | Middleware namespace.<br />Can be empty if the middleware belongs to the same namespace as the IngressRoute. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                       |                                                                      | No       |
| `routes[n].`<br />`observability.`<br />`accesslogs`                             | Defines whether the route will produce [access-logs](../../../../install-configuration/observability/logs-and-accesslogs.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| `routes[n].`<br />`observability.`<br />`metrics`                                | Defines whether the route will produce [metrics](../../../../install-configuration/observability/metrics.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| `routes[n].`<br />`observability.`<br />`tracing`                                | Defines whether the route will produce [traces](../../../../install-configuration/observability/tracing.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                                           | false                                                                | No       |
| `routes[n].`<br />`services`                                                     | List of any combination of TraefikService and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br />More information [here](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`kind`                                     | Kind of the service targeted.<br />Two values allowed:<br />- **Service**: Kubernetes Service<br /> **TraefikService**: Traefik Service.<br />More information [here](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                                                                  | "Service"                                                            | No       |
| `routes[n].`<br />`services[m].`<br />`name`                                     | Service name.<br />The character `@` is not authorized. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |                                                                      | Yes      |
| `routes[n].`<br />`services[m].`<br />`namespace`                                | Service namespace.<br />Can be empty if the service belongs to the same namespace as the IngressRoute. <br />More information [here](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                                                                                                   |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`port`                                     | Service port (number or port name).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`responseForwarding.`<br />`flushInterval` | Interval, in milliseconds, in between flushes to the client while copying the response body.<br />A negative value means to flush immediately after each write to the client.<br />This configuration is ignored when a response is a streaming response; for such responses, writes are flushed to the client immediately.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                    | 100ms                                                                | No       |
| `routes[n].`<br />`services[m].`<br />`scheme`                                   | Scheme to use for the request to the upstream Kubernetes Service.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | "http"<br />"https" if `port` is 443 or contains the string *https*. | No       |
| `routes[n].`<br />`services[m].`<br />`serversTransport`                         | Name of ServersTransport resource to use to configure the transport between Traefik and your servers.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                          | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`passHostHeader`                           | Forward client Host header to server.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | true                                                                 | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.scheme`                       | Server URL scheme for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.mode`                         | Health check mode.<br /> If defined to grpc, will use the gRPC health check protocol to probe the server.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                  | "http"                                                               | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.path`                         | Server URL path for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                             | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.interval`                     | Frequency of the health check calls for healthy targets.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                   | "100ms"                                                              | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.unhealthyInterval`            | Frequency of the health check calls for unhealthy targets.<br />When not defined, it defaults to the `interval` value.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                     | "100ms"                                                              | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.method`                       | HTTP method for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                 | "GET"                                                                | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.status`                       | Expected HTTP status code of the response to the health check request.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type ExternalName.<br />If not set, expect a status between 200 and 399.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.port`                         | URL port for the health check endpoint.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                                    |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.timeout`                      | Maximum duration to wait before considering the server unhealthy.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                          | "5s"                                                                 | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.hostname`                     | Value in the Host header of the health check request.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                      | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.`<br />`followRedirect`       | Follow the redirections during the healtchcheck.<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service).                                                                                                                                                                                                                                                                                                                                                           | true                                                                 | No       |
| `routes[n].`<br />`services[m].`<br />`healthCheck.headers`                      | Map of header to send to the health check endpoint<br />Evaluated only if the kind is **Service**.<br />Only for [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) of type [ExternalName](#externalname-service)).                                                                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| `routes[n].`<br />`services[m].`<br />`sticky.`<br />`cookie.name`               | Name of the cookie used for the stickiness.<br />When sticky sessions are enabled, a `Set-Cookie` header is set on the initial response to let the client know which server handles the first response.<br />On subsequent requests, to keep the session alive with the same server, the client should send the cookie with the value set.<br />If the server pecified in the cookie becomes unhealthy, the request will be forwarded to a new server (and the cookie will keep track of the new server).<br />Evaluated only if the kind is **Service**.                                                      | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`sticky.`<br />`cookie.httpOnly`           | Allow the cookie can be accessed by client-side APIs, such as JavaScript.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | false                                                                | No       |
| `routes[n].`<br />`services[m].`<br />`sticky.`<br />`cookie.secure`             | Allow the cookie can only be transmitted over an encrypted connection (i.e. HTTPS).<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | false                                                                | No       |
| `routes[n].`<br />`services[m].`<br />`sticky.`<br />`cookie.sameSite`           | [SameSite](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite) policy<br />Allowed values:<br />-`none`<br />-`lax`<br />`strict`<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                   | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`sticky.`<br />`cookie.maxAge`             | Number of seconds until the cookie expires.<br />Negative number, the cookie expires immediately.<br />0, the cookie never expires.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                            | 0                                                                    | No       |
| `routes[n].`<br />`services[m].`<br />`strategy`                                 | Load balancing strategy between the servers.<br />RoundRobin is the only supported value yet.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | "RoundRobin"                                                         | No       |
| `routes[n].`<br />`services[m].`<br />`weight`                                   | Service weight.<br />To use only to refer to WRR TraefikService                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | ""                                                                   | No       |
| `routes[n].`<br />`services[m].`<br />`nativeLB`                                 | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik.<br /> Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                                                                                                                    | false                                                                | No       |
| `routes[n].`<br />`services[m].`<br />`nodePortLB`                               | Use the nodePort IP address when the service type is NodePort.<br />It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br />Evaluated only if the kind is **Service**.                                                                                                                                                                                                                                                                                                                                                      | false                                                                | No       |
| `tls`                                                                            | TLS configuration.<br />Can be an empty value(`{}`):<br />A self signed is generated in such a case<br />(or the [default certificate](tlsstore.md) is used if it is defined.)                                                                                                                                                                                                                                                                                                                                                                                                                                 |                                                                      | No       |
| `tls.secretName`                                                                 | [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the same namesapce as the `IngressRoute`)                                                                                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| `tls.`<br />`options.name`                                                       | Name of the [`TLSOption`](tlsoption.md) to use.<br />More information [here](#tls-options).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | ""                                                                   | No       |
| `tls.`<br />`options.namespace`                                                  | Namespace of the [`TLSOption`](tlsoption.md) to use.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| `tls.certResolver`                                                               | Name of the [Certificate Resolver](../../../../install-configuration/tls/certificate-resolvers/overview.md) to use to generate automatic TLS certificates.                                                                                                                                                                                                                                                                                                                                                                                                                                                     | ""                                                                   | No       |
| `tls.domains`                                                                    | List of domains to serve using the certificates generates (one `tls.domain`= one certificate).<br />More information in the [dedicated section](../../../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition).                                                                                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| `tls.`<br />`domains[n].main`                                                    | Main domain name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | ""                                                                   | Yes      |
| `tls.`<br />`domains[n].sans`                                                    | List of alternative domains (SANs)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |                                                                      | No       |

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

### Middleware

- You can attach a list of [middlewares](../../../http/middlewares/overview.md) 
to each HTTP router.
- The middlewares will take effect only if the rule matches, and before forwarding
the request to the service.
- Middlewares are applied in the same order as their declaration in **router**.
- In Kubernetes, the option `middleware` allow you to attach a middleware using its
name and namespace (the namespace can be omitted when the Middleware is in the 
same namespace as the IngressRoute)

??? example "IngressRoute attached to a few middlewares"

    ```yaml 
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: my-app
      namespace: apps

    spec:
      entryPoints:
        - websecure
      routes:
      - match: Host(`example.com`)
        kind: Rule
        middlewares:
        # same namespace as the IngressRoute
        - name: middleware01
        # default namespace
        - name: middleware02
          namespace: apps
        # Other namespace
        - name: middleware03
          namespace: other-ns
        services:
        - name: whoami
          port: 80
    ```

??? abstract "routes.services.kind"

    As the field `name` can reference different types of objects, use the field `kind` to avoid any ambiguity.
    The field `kind` allows the following values:

    - `Service` (default value): to reference a [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/)
    - `TraefikService`: to reference an object [`TraefikService`](../http/traefikservice.md)

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

### TLS Options

The `options` field enables fine-grained control of the TLS parameters.
It refers to a [TLSOption](./tlsoption.md) and will be applied only if a `Host` 
rule is defined.

#### Server Name Association

A TLS options reference is always mapped to the host name found in the `Host` 
part of the rule, but neither to a router nor a router rule.
There could also be several `Host` parts in a rule.
In such a case the TLS options reference would be mapped to as many host names.

A TLS option is picked from the mapping mentioned above and based on the server 
name provided during the TLS handshake, 
and it all happens before routing actually occurs.

In the case of domain fronting,
if the TLS options associated with the Host Header and the SNI are different then
Traefik will respond with a status code `421`.

#### Conflicting TLS Options

Since a TLS options reference is mapped to a host name, if a configuration introduces
a situation where the same host name (from a `Host` rule) gets matched with two 
TLS options references, a conflict occurs, such as in the example below.

??? example

    ```yaml tab="IngressRoute01"
      apiVersion: traefik.io/v1alpha1
      kind: IngressRoute
      metadata:
        name: IngressRoute01
        namespace: apps

      spec:
        entryPoints:
          - foo
        routes:
        - match: Host(`example.net`)
          kind: Rule
        tls:
          options: foo
          ...

    ```

    ```yaml tab="IngressRoute02"
      apiVersion: traefik.io/v1alpha1
      kind: IngressRoute
      metadata:
        name: IngressRoute02
        namespace: apps

      spec:
        entryPoints:
          - foo
        routes:
        - match: Host(`example.net`)
          kind: Rule
        tls:
          options: bar
        ...
    ```

If that happens, both mappings are discarded, and the host name 
(`example.net` in the example) for these routers gets associated with
 the default TLS options instead.

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
