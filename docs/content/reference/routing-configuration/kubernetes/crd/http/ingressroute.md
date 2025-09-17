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
| <a id="entryPoints" href="#entryPoints" title="#entryPoints">`entryPoints`</a> | List of [entry points](../../../../install-configuration/entrypoints.md) names.<br />If not specified, HTTP routers will accept requests from all EntryPoints in the list of default EntryPoints.                                                                                                                                                                                                                                                                                                                                                                                                              |                                                                      | No       |
| <a id="routes" href="#routes" title="#routes">`routes`</a> | List of routes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |                                                                      | Yes      |
| <a id="routesn-kind" href="#routesn-kind" title="#routesn-kind">`routes[n].kind`</a> | Kind of router matching, only `Rule` is allowed yet.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | "Rule"                                                               | No       |
| <a id="routesn-match" href="#routesn-match" title="#routesn-match">`routes[n].match`</a> | Defines the [rule](../../../http/router/rules-and-priority.md#rules) corresponding to an underlying router.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                      | Yes      |
| <a id="routesn-priority" href="#routesn-priority" title="#routesn-priority">`routes[n].priority`</a> | Defines the [priority](../../../http/router/rules-and-priority.md#priority-calculation) to disambiguate rules of the same length, for route matching.<br />If not set, the priority is directly equal to the length of the rule, and so the longest length has the highest priority.<br />A value of `0` for the priority is ignored, the default rules length sorting is used.                                                                                                                                                                                                                                | 0                                                                    | No       |
| <a id="routesn-middlewares" href="#routesn-middlewares" title="#routesn-middlewares">`routes[n].middlewares`</a> | List of middlewares to attach to the IngressRoute. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | ""                                                                   | No       |
| <a id="routesn-middlewaresm-name" href="#routesn-middlewaresm-name" title="#routesn-middlewaresm-name">`routes[n].`<br />`middlewares[m].`<br />`name`</a> | Middleware name.<br />The character `@` is not authorized. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |                                                                      | Yes      |
| <a id="routesn-middlewaresm-namespace" href="#routesn-middlewaresm-namespace" title="#routesn-middlewaresm-namespace">`routes[n].`<br />`middlewares[m].`<br />`namespace`</a> | Middleware namespace.<br />Can be empty if the middleware belongs to the same namespace as the IngressRoute. <br />More information [here](#middleware).                                                                                                                                                                                                                                                                                                                                                                                                                                                       |                                                                      | No       |
| <a id="routesn-observability-accesslogs" href="#routesn-observability-accesslogs" title="#routesn-observability-accesslogs">`routes[n].`<br />`observability.`<br />`accesslogs`</a> | Defines whether the route will produce [access-logs](../../../../install-configuration/observability/logs-and-accesslogs.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| <a id="routesn-observability-metrics" href="#routesn-observability-metrics" title="#routesn-observability-metrics">`routes[n].`<br />`observability.`<br />`metrics`</a> | Defines whether the route will produce [metrics](../../../../install-configuration/observability/metrics.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                                          | false                                                                | No       |
| <a id="routesn-observability-tracing" href="#routesn-observability-tracing" title="#routesn-observability-tracing">`routes[n].`<br />`observability.`<br />`tracing`</a> | Defines whether the route will produce [traces](../../../../install-configuration/observability/tracing.md). See [here](../../../http/router/observability.md) for more information.                                                                                                                                                                                                                                                                                                                                                                                                                           | false                                                                | No       |
| <a id="tls" href="#tls" title="#tls">`tls`</a> | TLS configuration.<br />Can be an empty value(`{}`):<br />A self signed is generated in such a case<br />(or the [default certificate](tlsstore.md) is used if it is defined.)                                                                                                                                                                                                                                                                                                                                                                                                                                 |                                                                      | No       |
| <a id="routesn-services" href="#routesn-services" title="#routesn-services">`routes[n].`<br />`services`</a> | List of any combination of [TraefikService](./traefikservice.md) and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). <br /> Exhaustive list of option in the [`Service`](./service.md#configuration-options) documentation.                                                                                                                                                                                                                                                                                                                                                                                                                    |                                                                      | No       |
| <a id="tls-secretName" href="#tls-secretName" title="#tls-secretName">`tls.secretName`</a> | [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the same namesapce as the `IngressRoute`)                                                                                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| <a id="tls-options-name" href="#tls-options-name" title="#tls-options-name">`tls.`<br />`options.name`</a> | Name of the [`TLSOption`](tlsoption.md) to use.<br />More information [here](#tls-options).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | ""                                                                   | No       |
| <a id="tls-options-namespace" href="#tls-options-namespace" title="#tls-options-namespace">`tls.`<br />`options.namespace`</a> | Namespace of the [`TLSOption`](tlsoption.md) to use.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                   | No       |
| <a id="tls-certResolver" href="#tls-certResolver" title="#tls-certResolver">`tls.certResolver`</a> | Name of the [Certificate Resolver](../../../../install-configuration/tls/certificate-resolvers/overview.md) to use to generate automatic TLS certificates.                                                                                                                                                                                                                                                                                                                                                                                                                                                     | ""                                                                   | No       |
| <a id="tls-domains" href="#tls-domains" title="#tls-domains">`tls.domains`</a> | List of domains to serve using the certificates generates (one `tls.domain`= one certificate).<br />More information in the [dedicated section](../../../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition).                                                                                                                                                                                                                                                                                                                                                                        |                                                                      | No       |
| <a id="tls-domainsn-main" href="#tls-domainsn-main" title="#tls-domainsn-main">`tls.`<br />`domains[n].main`</a> | Main domain name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | ""                                                                   | Yes      |
| <a id="tls-domainsn-sans" href="#tls-domainsn-sans" title="#tls-domainsn-sans">`tls.`<br />`domains[n].sans`</a> | List of alternative domains (SANs)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |                                                                      | No       |


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
