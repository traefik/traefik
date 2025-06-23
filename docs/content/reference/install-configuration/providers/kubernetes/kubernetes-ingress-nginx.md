---
title: "Traefik Kubernetes Ingress NGINX Documentation"
description: "Understand the requirements, routing configuration, and how to set up the Kubernetes Ingress NGINX provider. Read the technical documentation."
---

# Traefik & Ingresses with NGINX Annotations 

The experimental Traefik Kubernetes Ingress NGINX provider is a Kubernetes Ingress controller; i.e,
it manages access to cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.
It also supports some of the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/) annotations on ingresses to customize their behavior.

!!! warning "Ingress Discovery"

    The Kubernetes Ingress NGINX provider is discovering by default all Ingresses in the cluster,
    which may lead to duplicated routers if you are also using the Kubernetes Ingress provider.
    We recommend to use IngressClass for the Ingresses you want to be handled by this provider,
    or to use the `watchNamespace` or `watchNamespaceSelector` options to limit the discovery of Ingresses to a specific namespace or set of namespaces.

## Configuration Example

As this provider is an experimental feature, it needs to be enabled in the experimental and in the provider sections of the configuration.
You can enable the Kubernetes Ingress NGINX provider as detailed below:

```yaml tab="File (YAML)"
experimental:
  kubernetesIngressNGINX: true

providers:
  kubernetesIngressNGINX: {}
```

```toml tab="File (TOML)"
[experimental.kubernetesIngressNGINX]

[providers.kubernetesIngressNGINX]
```

```bash tab="CLI"
--experimental.kubernetesingressnginx=true
--providers.kubernetesingressnginx=true
```

The provider then watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn creates the resulting routers, services, handlers, etc.

## Configuration Options
<!-- markdownlint-disable MD013 -->

| Field                                                       | Description                                                                                                                                                                                                                                                                                                                                                                          | Default | Required |
|:------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `providers.providersThrottleDuration`                       | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s      | No       |
| `providers.kubernetesIngressNGINX.endpoint`                 | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                        | ""      | No       |
| `providers.kubernetesIngressNGINX.token`                    | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                           | ""      | No       |
| `providers.kubernetesIngressNGINX.certAuthFilePath`         | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                           | ""      | No       |
| `providers.kubernetesIngressNGINX.throttleDuration`         | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                           | 0s      | No       |
| `providers.kubernetesIngressNGINX.watchNamespace`           | Namespace the controller watches for updates to Kubernetes objects. All namespaces are watched if this parameter is left empty.                                                                                                                                                                                                                                                      | ""      | No       |
| `providers.kubernetesIngressNGINX.watchNamespaceSelector`   | Selector selects namespaces the controller watches for updates to Kubernetes objects.                                                                                                                                                                                                                                                                                                | ""      | No       |
| `providers.kubernetesIngressNGINX.ingressClass`             | Name of the ingress class this controller satisfies.                                                                                                                                                                                                                                                                                                                                 | ""      | No       |
| `providers.kubernetesIngressNGINX.controllerClass`          | Ingress Class Controller value this controller satisfies.                                                                                                                                                                                                                                                                                                                            | ""      | No       |
| `providers.kubernetesIngressNGINX.watchIngressWithoutClass` | Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified.                                                                                                                                                                                                                                                                    | false   | No       |
| `providers.kubernetesIngressNGINX.ingressClassByName`       | Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class.                                                                                                                                                                                                                                                                                  | false   | No       |
| `providers.kubernetesIngressNGINX.publishService`           | Service fronting the Ingress controller. Takes the form namespace/name.                                                                                                                                                                                                                                                                                                              | ""      | No       |
| `providers.kubernetesIngressNGINX.publishStatusAddress`     | Customized address (or addresses, separated by comma) to set as the load-balancer status of Ingress objects this controller satisfies.                                                                                                                                                                                                                                               | ""      | No       |
| `providers.kubernetesIngressNGINX.defaultBackendService`    | Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form 'namespace/name'.                                                                                                                                                                                                                                                                 | ""      | No       |
| `providers.kubernetesIngressNGINX.disableSvcExternalName`   | Disable support for Services of type ExternalName.                                                                                                                                                                                                                                                                                                                                   | false   | No       |

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
  kubernetesIngressNGINX:
    endpoint: "http://localhost:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngressNGINX]
  endpoint = "http://localhost:8080"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingressnginx.endpoint=http://localhost:8080
```

## Routing Configuration

See the dedicated section in [routing](../../../routing-configuration/kubernetes/ingress-nginx.md).

{!traefik-for-business-applications.md!}
