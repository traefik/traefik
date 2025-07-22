---
title: "Traefik Kubernetes Ingress Documentation"
description: "Understand the requirements, routing configuration, and how to set up Traefik Proxy as your Kubernetes Ingress Controller. Read the technical documentation."
---

# Traefik & Kubernetes 

The Traefik Kubernetes Ingress provider is a Kubernetes Ingress controller; i.e,
it manages access to cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.

??? warning "Ingress Backend Resource not supported"

    Referencing backend service endpoints using [`spec.rules.http.paths.backend.resource`](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/ingress-v1/#IngressBackend) is not supported.
    Use `spec.rules.http.paths.backend.service` instead.

## Configuration Example

You can enable the `kubernetesIngress` provider as detailed below:

```yaml tab="File (YAML)"
providers:
  kubernetesIngress: {}
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
```

```bash tab="CLI"
--providers.kubernetesingress=true
```

```yaml tab="Helm Chart Values"
## Values file
providers:
  kubernetesIngress:
    enabled: true
```

The provider then watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn creates the resulting routers, services, handlers, etc.

## Configuration Options
<!-- markdownlint-disable MD013 -->

| Field                                                                  | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Default | Required |
|:-----------------------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `providers.providersThrottleDuration`                                  | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.**                                                                                   | 2s      | No       |
| `providers.kubernetesIngress.endpoint`                                 | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                                                                                                          | ""      | No       |
| `providers.kubernetesIngress.token`                                    | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                                                                                                             | ""      | No       |
| `providers.kubernetesIngress.certAuthFilePath`                         | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                                                                             | ""      | No       |
| `providers.kubernetesCRD.namespaces`                                   | Array of namespaces to watch.<br />If left empty, watch all namespaces.                                                                                                                                                                                                                                                                                                                                                                                                |         | No       |
| `providers.kubernetesIngress.labelselector`                            | Allow filtering on Ingress objects using label selectors.<br />No effect on Kubernetes `Secrets`, `EndpointSlices` and `Services`.<br />See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.                                                                                                                                                                                                  | ""      | No       |
| `providers.kubernetesIngress.ingressClass`                             | The `IngressClass` resource name or the `kubernetes.io/ingress.class` annotation value that identifies resource objects to be processed.<br />If empty, resources missing the annotation, having an empty value, or the value `traefik` are processed.                                                                                                                                                                                                                 | ""      | No       |
| `providers.kubernetesIngress.disableIngressClassLookup`                | Prevent to discover IngressClasses in the cluster.<br />It alleviates the requirement of giving Traefik the rights to look IngressClasses up.<br />Ignore Ingresses with IngressClass.<br />Annotations are not affected by this option.                                                                                                                                                                                                                               | false   | No       |
| `providers.kubernetesIngress.`<br />`ingressEndpoint.hostname`         | Hostname used for Kubernetes Ingress endpoints.                                                                                                                                                                                                                                                                                                                                                                                                                        | ""      | No       |
| `providers.kubernetesIngress.`<br />`ingressEndpoint.ip`               | This IP will get copied to the Ingress `status.loadbalancer.ip`, and currently only supports one IP value (IPv4 or IPv6).                                                                                                                                                                                                                                                                                                                                              | ""      | No       |
| `providers.kubernetesIngress.`<br />`ingressEndpoint.publishedService` | The Kubernetes service to copy status from.<br />More information [here](#ingressendpointpublishedservice).                                                                                                                                                                                                                                                                                                                                                            | ""      | No       |
| `providers.kubernetesIngress.throttleDuration`                         | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                                                                                                             | 0s      | No       |
| `providers.kubernetesIngress.allowEmptyServices`                       | Allows creating a route to reach a service that has no endpoint available.<br />It allows Traefik to handle the requests and responses targeting this service (applying middleware or observability operations) before returning a `503` HTTP Status.                                                                                                                                                                                                                  | false   | No       |
| `providers.kubernetesIngress.allowCrossNamespace`                      | Allows the `Ingress` to reference resources in namespaces other than theirs.                                                                                                                                                                                                                                                                                                                                                                                           | false   | No       |
| `providers.kubernetesIngress.allowExternalNameServices`                | Allows the `Ingress` to reference ExternalName services.                                                                                                                                                                                                                                                                                                                                                                                                               | false   | No       |
| `providers.kubernetesIngress.nativeLBByDefault`                        | Allow using the Kubernetes Service load balancing between the pods instead of the one provided by Traefik for every `Ingress` by default.<br />It can br overridden in the [`ServerTransport`](../../../../routing/services/index.md#serverstransport).                                                                                                                                                                                                                | false   | No       |
| `providers.kubernetesIngress.disableClusterScopeResources`             | Prevent from discovering cluster scope resources (`IngressClass` and `Nodes`).<br />By doing so, it alleviates the requirement of giving Traefik the rights to look up for cluster resources.<br />Furthermore, Traefik  will not handle Ingresses with IngressClass references, therefore such Ingresses will be ignored (please note that annotations are not affected by this option).<br />This will also prevent from using the `NodePortLB` options on services. | false   | No       |
| `providers.kubernetesIngress.strictPrefixMatching`                     | Make prefix matching strictly comply with the Kubernetes Ingress specification (path-element-wise matching instead of character-by-character string matching). For example, a PathPrefix of `/foo` will match `/foo`, `/foo/`, and `/foo/bar` but not `/foobar`.                                                                                                                                                                                                       | false   | No       |

<!-- markdownlint-enable MD013 -->

### `endpoint`

The Kubernetes server endpoint URL.

When deployed into Kubernetes, Traefik reads the environment variables `KUBERNETES_SERVICE_HOST`
and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token`
and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside
a cluster.

When the environment variables are not found, Traefik tries to connect to the 
Kubernetes API server with an external-cluster client.

In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes
cluster using the granted authentication and authorization of the associated kubeconfig.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    endpoint: "http://localhost:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  endpoint = "http://localhost:8080"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.endpoint=http://localhost:8080
```

###  `ingressEndpoint.publishedService`

Format: `namespace/servicename`.

The Kubernetes service to copy status from,
depending on the service type:

- **ClusterIP:** The ExternalIPs of the service will be propagated to the ingress status.
- **NodePort:** The ExternalIP addresses of the nodes in the cluster will be propagated to the ingress status.
- **LoadBalancer:** The IPs from the service's `loadBalancer.status` field (which contains the endpoints provided by the load balancer) will be propagated to the ingress status.

When using third-party tools such as External-DNS, this option enables the copying of external service IPs to the ingress resources.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      publishedService: "namespace/foo-service"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  publishedService = "namespace/foo-service"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.publishedservice=namespace/foo-service
```


## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kubernetes-ingress.md).

## Further

To learn more about the various aspects of the Ingress specification that 
Traefik supports,
many examples of Ingresses definitions are located in the test 
[examples](https://github.com/traefik/traefik/tree/v3.1/pkg/provider/kubernetes/ingress/fixtures) 
of the Traefik repository.

{!traefik-for-business-applications.md!}
