---
title: "Kubernetes IngressRouteTCP"
description: "An IngressRouteTCP is a Traefik CRD is in charge of connecting incoming TCP connections to the Services that can handle them."
---

`IngressRouteTCP` is the CRD implementation of a [Traefik TCP router](../../../tcp/routing/rules-and-priority.md).

Before creating `IngressRouteTCP` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `IngressRouteTCP` kind and other Traefik-specific resources.

!!! note "General"
    If both HTTP routers and TCP routers are connected to the same EntryPoint, the TCP routers will apply before the HTTP routers. If no matching route is found for the TCP routers, then the HTTP routers will take over.

## Configuration Example

You can declare an `IngressRouteTCP` as detailed below:

```yaml tab="IngressRouteTCP"
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
      serversTransport: transport
      nativeLB: true
      nodePortLB: true

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

| Field                                | Description                                                                                                                                                                                                                                                                                  | Default                                   | Required |
|-------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------|-----------------------|
| <a id="opt-entryPoints" href="#opt-entryPoints" title="#opt-entryPoints">`entryPoints`</a> | List of entrypoints names.                                                                                                                                                                                                                                                                   | | No |
| <a id="opt-routes" href="#opt-routes" title="#opt-routes">`routes`</a> | List of routes.                                                                                                                                                                                                                                                                              | | Yes |
| <a id="opt-routesn-match" href="#opt-routesn-match" title="#opt-routesn-match">`routes[n].match`</a> | Defines the [rule](../../../tcp/routing/rules-and-priority.md#rules) of the underlying router.                                                                                                                                                                                               | | Yes |
| <a id="opt-routesn-priority" href="#opt-routesn-priority" title="#opt-routesn-priority">`routes[n].priority`</a> | Defines the [priority](../../../tcp/routing/rules-and-priority.md#priority-calculation) to disambiguate rules of the same length, for route matching.                                                                                                                                        | | No |
| <a id="opt-routesn-middlewaresn-name" href="#opt-routesn-middlewaresn-name" title="#opt-routesn-middlewaresn-name">`routes[n].middlewares[n].name`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) name.                                                                                                                                                                                                                                        | | Yes |
| <a id="opt-routesn-middlewaresn-namespace" href="#opt-routesn-middlewaresn-namespace" title="#opt-routesn-middlewaresn-namespace">`routes[n].middlewares[n].namespace`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) namespace.                                                                                                                                                                                                                                   | ""| No|
| <a id="opt-routesn-services" href="#opt-routesn-services" title="#opt-routesn-services">`routes[n].services`</a> | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions.                                                                                                                                                                                  | | No |
| <a id="opt-routesn-servicesn-name" href="#opt-routesn-servicesn-name" title="#opt-routesn-servicesn-name">`routes[n].services[n].name`</a> | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/).                                                                                                                                                                                | | Yes |
| <a id="opt-routesn-servicesn-port" href="#opt-routesn-servicesn-port" title="#opt-routesn-servicesn-port">`routes[n].services[n].port`</a> | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.                                                                                                                                       | | Yes |
| <a id="opt-routesn-servicesn-weight" href="#opt-routesn-servicesn-weight" title="#opt-routesn-servicesn-weight">`routes[n].services[n].weight`</a> | Defines the weight to apply to the server load balancing.                                                                                                                                                                                                                                    | 1 | No |
| <a id="opt-routesn-servicesn-proxyProtocol" href="#opt-routesn-servicesn-proxyProtocol" title="#opt-routesn-servicesn-proxyProtocol">`routes[n].services[n].proxyProtocol`</a> | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) configuration.                                                                                                                                                               |  | No |
| <a id="opt-routesn-servicesn-proxyProtocol-version" href="#opt-routesn-servicesn-proxyProtocol-version" title="#opt-routesn-servicesn-proxyProtocol-version">`routes[n].services[n].proxyProtocol.version`</a> | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) version.                                                                                                                                                                     |  | No |
| <a id="opt-routesn-servicesn-serversTransport" href="#opt-routesn-servicesn-serversTransport" title="#opt-routesn-servicesn-serversTransport">`routes[n].services[n].serversTransport`</a> | Defines the [ServersTransportTCP](./serverstransporttcp.md).<br />The `ServersTransport` namespace is assumed to be the [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace.                                                                    |  | No |
| <a id="opt-routesn-servicesn-nativeLB" href="#opt-routesn-servicesn-nativeLB" title="#opt-routesn-servicesn-nativeLB">`routes[n].services[n].nativeLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. See [here](#nativelb) for more information.                                                                                         | false | No |
| <a id="opt-routesn-servicesn-nodePortLB" href="#opt-routesn-servicesn-nodePortLB" title="#opt-routesn-servicesn-nodePortLB">`routes[n].services[n].nodePortLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is `NodePort`. It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes. | false | No |
| <a id="opt-tls" href="#opt-tls" title="#opt-tls">`tls`</a> | Defines [TLS](../../../../install-configuration/tls/certificate-resolvers/overview.md) certificate configuration.                                                                                                                                                                            |  | No |
| <a id="opt-tls-secretName" href="#opt-tls-secretName" title="#opt-tls-secretName">`tls.secretName`</a> | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace).                                                                                                                                        | "" | No |
| <a id="opt-tls-options" href="#opt-tls-options" title="#opt-tls-options">`tls.options`</a> | Defines the reference to a [TLSOption](tlsoption.md).                                                                                                                                                                                                                                        | "" | No |
| <a id="opt-tls-options-name" href="#opt-tls-options-name" title="#opt-tls-options-name">`tls.options.name`</a> | Defines the [TLSOption](tlsoption.md) name.                                                                                                                                                                                                                                                  | "" | No |
| <a id="opt-tls-options-namespace" href="#opt-tls-options-namespace" title="#opt-tls-options-namespace">`tls.options.namespace`</a> | Defines the [TLSOption](tlsoption.md) namespace.                                                                                                                                                                                                                                             | "" | No |
| <a id="opt-tls-certResolver" href="#opt-tls-certResolver" title="#opt-tls-certResolver">`tls.certResolver`</a> | Defines the reference to a [CertResolver](../../../../install-configuration/tls/certificate-resolvers/overview.md).                                                                                                                                                                          | "" | No |
| <a id="opt-tls-domains" href="#opt-tls-domains" title="#opt-tls-domains">`tls.domains`</a> | List of domains.                                                                                                                                                                                                                                                                             | "" | No |
| <a id="opt-tls-domainsn-main" href="#opt-tls-domainsn-main" title="#opt-tls-domainsn-main">`tls.domains[n].main`</a> | Defines the main domain name.                                                                                                                                                                                                                                                                | "" | No |
| <a id="opt-tls-domainsn-sans" href="#opt-tls-domainsn-sans" title="#opt-tls-domainsn-sans">`tls.domains[n].sans`</a> | List of SANs (alternative domains).                                                                                                                                                                                                                                                          | "" | No |
| <a id="opt-tls-passthrough" href="#opt-tls-passthrough" title="#opt-tls-passthrough">`tls.passthrough`</a> | If `true`, delegates the TLS termination to the backend.                                                                                                                                                                                                                                     | false | No |

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
