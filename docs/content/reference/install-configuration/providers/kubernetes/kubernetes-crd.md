---
title: 'Kubernetes Custom Resources'
description: 'Configure the Kubernetes CRD provider that allows managing Traefik custom resources.'
---

Traefik provides some Kubernetes Custom Resources, such as `IngressRoute`, `Middleware`, etc.

When using KubernetesCRD as a provider,
Traefik uses [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to retrieve its routing configuration.
Traefik Custom Resource Definitions are [listed below](#list-of-resources).

When Traefik is installed using the Helm Chart, by default, the provider `kubernetesCRD` is enabled.

## Requirements

When you install Traefik without using the Helm Chart, or when you are upgrading the stack using Helm, ensure that you satisfy the following requirements:

- Add/update **all** the Traefik resources definitions
- Add/update the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for the Traefik custom resources

```bash
# Install Traefik Resource Definitions:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.5/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml

# Install RBAC for Traefik:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.5/docs/content/reference/dynamic-configuration/kubernetes-crd-rbac.yml
```

## Configuration Example

You can enable the `kubernetesCRD` provider as detailed below:

```yaml tab="File (YAML)"
providers:
  kubernetesCRD: {}
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
```

```bash tab="CLI"
--providers.kubernetescrd=true
```

```yaml tab="Helm Chart Values"
## Values file
providers:
  kubernetesCRD:
    enabled: true
```

## Configuration Options

| Field | Description                                               | Default | Required |
|:------|:----------------------------------------------------------|:--------|:---------|
| <a id="providers-providersThrottleDuration" href="#providers-providersThrottleDuration" title="#providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s      | No |
| <a id="providers-kubernetesCRD-endpoint" href="#providers-kubernetesCRD-endpoint" title="#providers-kubernetesCRD-endpoint">`providers.kubernetesCRD.endpoint`</a> | Server endpoint URL.<br />More information [here](#endpoint). | ""      | No |
| <a id="providers-kubernetesCRD-token" href="#providers-kubernetesCRD-token" title="#providers-kubernetesCRD-token">`providers.kubernetesCRD.token`</a> | Bearer token used for the Kubernetes client configuration. | ""      | No |
| <a id="providers-kubernetesCRD-certAuthFilePath" href="#providers-kubernetesCRD-certAuthFilePath" title="#providers-kubernetesCRD-certAuthFilePath">`providers.kubernetesCRD.certAuthFilePath`</a> | Path to the certificate authority file.<br />Used for the Kubernetes client configuration. | ""      | No |
| <a id="providers-kubernetesCRD-namespaces" href="#providers-kubernetesCRD-namespaces" title="#providers-kubernetesCRD-namespaces">`providers.kubernetesCRD.namespaces`</a> | Array of namespaces to watch.<br />If left empty, watch all namespaces. | []      | No |
| <a id="providers-kubernetesCRD-labelselector" href="#providers-kubernetesCRD-labelselector" title="#providers-kubernetesCRD-labelselector">`providers.kubernetesCRD.labelselector`</a> | Allow filtering on specific resource objects only using label selectors.<br />Only to Traefik [Custom Resources](#list-of-resources) (they all must match the filter).<br />No effect on Kubernetes `Secrets`, `EndpointSlices` and `Services`.<br />See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details. | ""      | No |
| <a id="providers-kubernetesCRD-ingressClass" href="#providers-kubernetesCRD-ingressClass" title="#providers-kubernetesCRD-ingressClass">`providers.kubernetesCRD.ingressClass`</a> | Value of `kubernetes.io/ingress.class` annotation that identifies resource objects to be processed.<br />If empty, resources missing the annotation, having an empty value, or the value `traefik` are processed. | ""      | No |
| <a id="providers-kubernetesCRD-throttleDuration" href="#providers-kubernetesCRD-throttleDuration" title="#providers-kubernetesCRD-throttleDuration">`providers.kubernetesCRD.throttleDuration`</a> | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught. | 0s      | No |
| <a id="providers-kubernetesCRD-allowEmptyServices" href="#providers-kubernetesCRD-allowEmptyServices" title="#providers-kubernetesCRD-allowEmptyServices">`providers.kubernetesCRD.allowEmptyServices`</a> | Allows creating a route to reach a service that has no endpoint available.<br />It allows Traefik to handle the requests and responses targeting this service (applying middleware or observability operations) before returning a `503` HTTP Status.  | false   | No |
| <a id="providers-kubernetesCRD-allowCrossNamespace" href="#providers-kubernetesCRD-allowCrossNamespace" title="#providers-kubernetesCRD-allowCrossNamespace">`providers.kubernetesCRD.allowCrossNamespace`</a> | Allows the `IngressRoutes` to reference resources in namespaces other than theirs. | false   | No |
| <a id="providers-kubernetesCRD-allowExternalNameServices" href="#providers-kubernetesCRD-allowExternalNameServices" title="#providers-kubernetesCRD-allowExternalNameServices">`providers.kubernetesCRD.allowExternalNameServices`</a> | Allows the `IngressRoutes` to reference ExternalName services. | false   | No |
| <a id="providers-kubernetesCRD-nativeLBByDefault" href="#providers-kubernetesCRD-nativeLBByDefault" title="#providers-kubernetesCRD-nativeLBByDefault">`providers.kubernetesCRD.nativeLBByDefault`</a> | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik for every `IngressRoute` by default.<br />It can br overridden in the [`ServerTransport`](../../../../routing/services/index.md#serverstransport). | false   | No |
| <a id="providers-kubernetesCRD-disableClusterScopeResources" href="#providers-kubernetesCRD-disableClusterScopeResources" title="#providers-kubernetesCRD-disableClusterScopeResources">`providers.kubernetesCRD.disableClusterScopeResources`</a> | Prevent from discovering cluster scope resources (`IngressClass` and `Nodes`).<br />By doing so, it alleviates the requirement of giving Traefik the rights to look up for cluster resources.<br />Furthermore, Traefik will not handle IngressRoutes with IngressClass references, therefore such Ingresses will be ignored (please note that annotations are not affected by this option).<br />This will also prevent from using the `NodePortLB` options on services. | false   | No |

### endpoint

The Kubernetes server endpoint URL.

When deployed into Kubernetes, Traefik reads the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik tries to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    endpoint: "http://localhost:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  endpoint = "http://localhost:8080"
  # ...
```

```bash tab="CLI"
--providers.kubernetesCRD.endpoint=http://localhost:8080
```

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kubernetes-crd.md).

## List of Resources

<!-- markdownlint-disable MD013 -->

| Resource  | Purpose    |
|--------------------------------------------------|--------------------------------------------------------------------|
| <a id="IngressRoute" href="#IngressRoute" title="#IngressRoute">[IngressRoute](../../../../routing/providers/kubernetes-crd.md#kind-ingressroute)</a> | HTTP Routing     |
| <a id="Middleware" href="#Middleware" title="#Middleware">[Middleware](../../../../middlewares/http/overview.md)</a> | Tweaks the HTTP requests before they are sent to your service  |
| <a id="TraefikService" href="#TraefikService" title="#TraefikService">[TraefikService](../../../../routing/providers/kubernetes-crd.md#kind-traefikservice)</a> | Abstraction for HTTP loadbalancing/mirroring  | 
| <a id="TLSOptions" href="#TLSOptions" title="#TLSOptions">[TLSOptions](../../../../routing/providers/kubernetes-crd.md#kind-tlsoption)</a> | Allows configuring some parameters of the TLS connection  |
| <a id="TLSStores" href="#TLSStores" title="#TLSStores">[TLSStores](../../../../routing/providers/kubernetes-crd.md#kind-tlsstore)</a> | Allows configuring the default TLS store    |  
| <a id="ServersTransport" href="#ServersTransport" title="#ServersTransport">[ServersTransport](../../../../routing/providers/kubernetes-crd.md#kind-serverstransport)</a> | Allows configuring the transport between Traefik and the backends | 
| <a id="IngressRouteTCP" href="#IngressRouteTCP" title="#IngressRouteTCP">[IngressRouteTCP](../../../../routing/providers/kubernetes-crd.md#kind-ingressroutetcp)</a> | TCP Routing  | 
| <a id="MiddlewareTCP" href="#MiddlewareTCP" title="#MiddlewareTCP">[MiddlewareTCP](../../../../routing/providers/kubernetes-crd.md#kind-middlewaretcp)</a> | Tweaks the TCP requests before they are sent to your service       |
| <a id="ServersTransportTCP" href="#ServersTransportTCP" title="#ServersTransportTCP">[ServersTransportTCP](../../../../routing/providers/kubernetes-crd.md#kind-serverstransporttc)</a> | Allows configuring the transport between Traefik and the backends |
| <a id="IngressRouteUDP" href="#IngressRouteUDP" title="#IngressRouteUDP">[IngressRouteUDP](../../../../routing/providers/kubernetes-crd.md#kind-ingressrouteudp)</a> | UDP Routing       |

## Particularities

- The usage of `name` **and** `namespace` to refer to another Kubernetes resource.
- The usage of [secret](https://kubernetes.io/docs/concepts/configuration/secret/) for sensitive data (TLS certificates and credentials).

## Full Example

For additional information, refer to the [full example](../../../../user-guides/crd-acme/index.md) with Let's Encrypt.

{!traefik-for-business-applications.md!}
