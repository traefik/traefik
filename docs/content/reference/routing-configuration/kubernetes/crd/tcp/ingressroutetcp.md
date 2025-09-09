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

<<<<<<< HEAD
| Field                                         | Description                                                                                                                                                                                                                                                                                            | Default | Required |
|-----------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="entryPoints" href="#entryPoints" title="#entryPoints">`entryPoints`</a> | List of entrypoints names.                                                                                                                                                                                                                                                                             |         | No       |
| <a id="routes" href="#routes" title="#routes">`routes`</a> | List of routes.                                                                                                                                                                                                                                                                                        |         | Yes      |
| <a id="routesn-match" href="#routesn-match" title="#routesn-match">`routes[n].match`</a> | Defines the [rule](../../../tcp/router/rules-and-priority.md#rules) of the underlying router.                                                                                                                                                                                                          |         | Yes      |
| <a id="routesn-priority" href="#routesn-priority" title="#routesn-priority">`routes[n].priority`</a> | Defines the [priority](../../../tcp/router/rules-and-priority.md#priority) to disambiguate rules of the same length, for route matching.                                                                                                                                                               |         | No       |
| <a id="routesn-middlewaresn-name" href="#routesn-middlewaresn-name" title="#routesn-middlewaresn-name">`routes[n].middlewares[n].name`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) name.                                                                                                                                                                                                                                                  |         | Yes      |
| <a id="routesn-middlewaresn-namespace" href="#routesn-middlewaresn-namespace" title="#routesn-middlewaresn-namespace">`routes[n].middlewares[n].namespace`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) namespace.                                                                                                                                                                                                                                             | ""      | No       |
| <a id="routesn-services" href="#routesn-services" title="#routesn-services">`routes[n].services`</a> | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions.                                                                                                                                                                                            |         | No       |
| <a id="routesn-servicesn-name" href="#routesn-servicesn-name" title="#routesn-servicesn-name">`routes[n].services[n].name`</a> | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/).                                                                                                                                                                                          |         | Yes      |
| <a id="routesn-servicesn-port" href="#routesn-servicesn-port" title="#routesn-servicesn-port">`routes[n].services[n].port`</a> | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.                                                                                                                                                 |         | Yes      |
| <a id="routesn-servicesn-weight" href="#routesn-servicesn-weight" title="#routesn-servicesn-weight">`routes[n].services[n].weight`</a> | Defines the weight to apply to the server load balancing.                                                                                                                                                                                                                                              | 1       | No       |
| <a id="routesn-servicesn-serversTransport" href="#routesn-servicesn-serversTransport" title="#routesn-servicesn-serversTransport">`routes[n].services[n].serversTransport`</a> | Defines the [ServersTransportTCP](./serverstransporttcp.md).<br />The `ServersTransport` namespace is assumed to be the [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace.                                                                              |         | No       |
| <a id="routesn-servicesn-nativeLB" href="#routesn-servicesn-nativeLB" title="#routesn-servicesn-nativeLB">`routes[n].services[n].nativeLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. See [here](#nativelb) for more information.                                                                                                   | false   | No       |
| <a id="routesn-servicesn-nodePortLB" href="#routesn-servicesn-nodePortLB" title="#routesn-servicesn-nodePortLB">`routes[n].services[n].nodePortLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is `NodePort`. It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes. | false   | No       |
| <a id="tls" href="#tls" title="#tls">`tls`</a> | Defines [TLS](../../../../install-configuration/tls/certificate-resolvers/overview.md) certificate configuration.                                                                                                                                                                                      |         | No       |
| <a id="tls-secretName" href="#tls-secretName" title="#tls-secretName">`tls.secretName`</a> | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace).                                                                                                                                                  | ""      | No       |
| <a id="tls-options" href="#tls-options" title="#tls-options">`tls.options`</a> | Defines the reference to a [TLSOption](../http/tlsoption.md).                                                                                                                                                                                                                                          | ""      | No       |
| <a id="tls-options-name" href="#tls-options-name" title="#tls-options-name">`tls.options.name`</a> | Defines the [TLSOption](../http/tlsoption.md) name.                                                                                                                                                                                                                                                    | ""      | No       |
| <a id="tls-options-namespace" href="#tls-options-namespace" title="#tls-options-namespace">`tls.options.namespace`</a> | Defines the [TLSOption](../http/tlsoption.md) namespace.                                                                                                                                                                                                                                               | ""      | No       |
| <a id="tls-certResolver" href="#tls-certResolver" title="#tls-certResolver">`tls.certResolver`</a> | Defines the reference to a [CertResolver](../../../../install-configuration/tls/certificate-resolvers/overview.md).                                                                                                                                                                                    | ""      | No       |
| <a id="tls-domains" href="#tls-domains" title="#tls-domains">`tls.domains`</a> | List of domains.                                                                                                                                                                                                                                                                                       | ""      | No       |
| <a id="tls-domainsn-main" href="#tls-domainsn-main" title="#tls-domainsn-main">`tls.domains[n].main`</a> | Defines the main domain name.                                                                                                                                                                                                                                                                          | ""      | No       |
| <a id="tls-domainsn-sans" href="#tls-domainsn-sans" title="#tls-domainsn-sans">`tls.domains[n].sans`</a> | List of SANs (alternative domains).                                                                                                                                                                                                                                                                    | ""      | No       |
| <a id="tls-passthrough" href="#tls-passthrough" title="#tls-passthrough">`tls.passthrough`</a> | If `true`, delegates the TLS termination to the backend.                                                                                                                                                                                                                                               | false   | No       |
=======
| Field                                |  Description                    | Default                                   | Required |
|-------------------------------------|-----------------------------|-------------------------------------------|-----------------------|
| <a id="entryPoints" href="#entryPoints" title="#entryPoints">`entryPoints`</a> | List of entrypoints names. | | No |
| <a id="routes" href="#routes" title="#routes">`routes`</a> | List of routes. | | Yes |
| <a id="routesn-match" href="#routesn-match" title="#routesn-match">`routes[n].match`</a> | Defines the [rule](../../../tcp/router/rules-and-priority.md#rules) of the underlying router. | | Yes |
| <a id="routesn-priority" href="#routesn-priority" title="#routesn-priority">`routes[n].priority`</a> | Defines the [priority](../../../tcp/router/rules-and-priority.md#priority) to disambiguate rules of the same length, for route matching. | | No |
| <a id="routesn-middlewaresn-name" href="#routesn-middlewaresn-name" title="#routesn-middlewaresn-name">`routes[n].middlewares[n].name`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) name. | | Yes |
| <a id="routesn-middlewaresn-namespace" href="#routesn-middlewaresn-namespace" title="#routesn-middlewaresn-namespace">`routes[n].middlewares[n].namespace`</a> | Defines the [MiddlewareTCP](./middlewaretcp.md) namespace. | ""| No|
| <a id="routesn-services" href="#routesn-services" title="#routesn-services">`routes[n].services`</a> | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions. | | No |
| <a id="routesn-servicesn-name" href="#routesn-servicesn-name" title="#routesn-servicesn-name">`routes[n].services[n].name`</a> | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). | | Yes |
| <a id="routesn-servicesn-port" href="#routesn-servicesn-port" title="#routesn-servicesn-port">`routes[n].services[n].port`</a> | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.| | Yes |
| <a id="routesn-servicesn-weight" href="#routesn-servicesn-weight" title="#routesn-servicesn-weight">`routes[n].services[n].weight`</a> | Defines the weight to apply to the server load balancing. | 1 | No |
| <a id="routesn-servicesn-proxyProtocol" href="#routesn-servicesn-proxyProtocol" title="#routesn-servicesn-proxyProtocol">`routes[n].services[n].proxyProtocol`</a> | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) configuration. |  | No |
| <a id="routesn-servicesn-proxyProtocol-version" href="#routesn-servicesn-proxyProtocol-version" title="#routesn-servicesn-proxyProtocol-version">`routes[n].services[n].proxyProtocol.version`</a> | Defines the [PROXY protocol](../../../../install-configuration/entrypoints.md#proxyprotocol-and-load-balancers) version. |  | No |
| <a id="routesn-servicesn-serversTransport" href="#routesn-servicesn-serversTransport" title="#routesn-servicesn-serversTransport">`routes[n].services[n].serversTransport`</a> | Defines the [ServersTransportTCP](./serverstransporttcp.md).<br />The `ServersTransport` namespace is assumed to be the [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace. |  | No |
| <a id="routesn-servicesn-nativeLB" href="#routesn-servicesn-nativeLB" title="#routesn-servicesn-nativeLB">`routes[n].services[n].nativeLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. See [here](#nativelb) for more information. | false | No |
| <a id="routesn-servicesn-nodePortLB" href="#routesn-servicesn-nodePortLB" title="#routesn-servicesn-nodePortLB">`routes[n].services[n].nodePortLB`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is `NodePort`. It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes. | false | No |
| <a id="tls" href="#tls" title="#tls">`tls`</a> | Defines [TLS](../../../../install-configuration/tls/certificate-resolvers/overview.md) certificate configuration. |  | No |
| <a id="tls-secretName" href="#tls-secretName" title="#tls-secretName">`tls.secretName`</a> | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace). | "" | No |
| <a id="tls-options" href="#tls-options" title="#tls-options">`tls.options`</a> | Defines the reference to a [TLSOption](../http/tlsoption.md). | "" | No |
| <a id="tls-options-name" href="#tls-options-name" title="#tls-options-name">`tls.options.name`</a> | Defines the [TLSOption](../http/tlsoption.md) name. | "" | No |
| <a id="tls-options-namespace" href="#tls-options-namespace" title="#tls-options-namespace">`tls.options.namespace`</a> | Defines the [TLSOption](../http/tlsoption.md) namespace. | "" | No |
| <a id="tls-certResolver" href="#tls-certResolver" title="#tls-certResolver">`tls.certResolver`</a> | Defines the reference to a [CertResolver](../../../../install-configuration/tls/certificate-resolvers/overview.md). | "" | No |
| <a id="tls-domains" href="#tls-domains" title="#tls-domains">`tls.domains`</a> | List of domains. | "" | No |
| <a id="tls-domainsn-main" href="#tls-domainsn-main" title="#tls-domainsn-main">`tls.domains[n].main`</a> | Defines the main domain name. | "" | No |
| <a id="tls-domainsn-sans" href="#tls-domainsn-sans" title="#tls-domainsn-sans">`tls.domains[n].sans`</a> | List of SANs (alternative domains).  | "" | No |
| <a id="tls-passthrough" href="#tls-passthrough" title="#tls-passthrough">`tls.passthrough`</a> | If `true`, delegates the TLS termination to the backend. | false | No |
>>>>>>> 9c932124f (Add anchors in reference tables)

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
