---
title: "Kubernetes IngressRouteTCP"
description: "An IngressRouteTCP is a Traefik CRD is in charge of connecting incoming TCP connections to the Services that can handle them."
---

`IngressRouteTCP` is the CRD implementation of a [Traefik TCP router](../../../tcp/router/rules-and-priority.md).

Before creating `IngressRouteTCP` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `IngressRouteTCP` kind and other Traefik-specific resources.

!!! note "General"
    If both HTTP routers and TCP routers are connected to the same EntryPoint, the TCP routers will apply before the HTTP routers. If no matching route is found for the TCP routers, then the HTTP routers will take over.

## Configuration Example

You can declare an `IngressRouteTCP` as detailed below:

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: ingressroutetcpfoo
  namespace: apps

spec:
  entryPoints:
    - footcp
  routes:
  - match: HostSNI(`*`)
    priority: 10
    middlewares:
    - name: middleware1
      namespace: default
    services:
    - name: foo
      port: 8080
      weight: 10
      proxyProtocol:
        version: 1
      serversTransport: transport
      nativeLB: true
      nodePortLB: true
      tls: false

  tls:
    secretName: supersecret
    options:
      name: opt
      namespace: default
    certResolver: foo
    domains:
    - main: example.net
      sans:                       
      - a.example.net
      - b.example.net
    passthrough: false
```

## Configuration Options

| Field                                |  Description                    | Default                                   | Required |
|-------------------------------------|-----------------------------|-------------------------------------------|-----------------------|
| `entryPoints`                       | List of entrypoints names. | | No |
| `routes`                            | List of routes. | | Yes |
| `routes[n].match`                   | Defines the [rule](../../../tcp/router/rules-and-priority.md#rules) of the underlying router. | | Yes |
| `routes[n].priority`                | Defines the [priority](../../../tcp/router/rules-and-priority.md#priority) to disambiguate rules of the same length, for route matching. | | No |
| `routes[n].middlewares[n].name`               | Defines the [MiddlewareTCP](./middlewaretcp.md) name. | | Yes |
| `routes[n].middlewares[n].namespace`          | Defines the [MiddlewareTCP](./middlewaretcp.md) namespace. | ""| No|
| `routes[n].services`                | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions. | | No |
|  `routes[n].services[n].name`                  | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). | | Yes |
| `routes[n].services[n].port`                  | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.| | Yes |
| `routes[n].services[n].weight`                | Defines the weight to apply to the server load balancing. | 1 | No |
| `routes[n].services[n].proxyProtocol`         | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) configuration. |  | No |
| `routes[n].services[n].proxyProtocol.version` | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) version. |  | No |
| `routes[n].services[n].serversTransport`      | Defines the [ServersTransportTCP](./serverstransporttcp.md).<br />The `ServersTransport` namespace is assumed to be the [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace. |  | No |
| `routes[n].services[n].nativeLB`              | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. See [here](#nativelb) for more information. | false | No |
| `routes[n].services[n].nodePortLB`            | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is `NodePort`. It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes. | false | No |
| `tls`                               | Defines [TLS](../../../../install-configuration/tls/certificate-resolvers/overview.md) certificate configuration. |  | No |
| `tls.secretName`                    | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace). | "" | No |
| `tls.options`                       | Defines the reference to a [TLSOption](../http/tlsoption.md). | "" | No |
| `tls.options.name`                  | Defines the [TLSOption](../http/tlsoption.md) name. | "" | No |
| `tls.options.namespace`             | Defines the [TLSOption](../http/tlsoption.md) namespace. | "" | No |
| `tls.certResolver`                  | Defines the reference to a [CertResolver](../../../../install-configuration/tls/certificate-resolvers/overview.md). | "" | No |
| `tls.domains`                       | List of domains. | "" | No |
| `tls.domains[n].main`               | Defines the main domain name. | "" | No |
| `tls.domains[n].sans`               | List of SANs (alternative domains).  | "" | No |
| `tls.passthrough`                   | If `true`, delegates the TLS termination to the backend. | false | No |

### ExternalName Service

Traefik connect to a backend with a domain and a port. However, Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) can be defined without any port. Accordingly, Traefik supports defining a port in two ways:

- only on `IngressRouteTCP` service
- on both sides, you'll be warned if the ports don't match, and the `IngressRouteTCP` service port is used

Thus, in case of two sides port definition, Traefik expects a match between ports.

=== "Ports defined on Resource"

    ```yaml tab="IngressRouteTCP"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
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

    ```yaml tab="IngressRouteTCP"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
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

    ```yaml tab="IngressRouteTCP"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
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

### NativeLB

To avoid creating the server load-balancer with the pods IPs and use Kubernetes Service `clusterIP` directly, one should set the `NativeLB` option to true. By default, `NativeLB` is false.

```yaml tab="IngressRouteTCP"
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: test.route
  namespace: default
spec:
  entryPoints:
    - foo
  routes:
  - match: HostSNI(`*`)
    services:
    - name: svc
      port: 80
      # Here, nativeLB instructs to build the servers load balancer with the Kubernetes Service clusterIP only.
      nativeLB: true
```

```yaml tab="Service"
apiVersion: v1
kind: Service
metadata:
  name: svc
  namespace: default
spec:
  type: ClusterIP
  ...
```
