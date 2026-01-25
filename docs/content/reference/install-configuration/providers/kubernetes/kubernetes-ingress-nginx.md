---
title: "Traefik Kubernetes Ingress NGINX Documentation"
description: "Understand the requirements, routing configuration, and how to set up the Kubernetes Ingress NGINX provider. Read the technical documentation."
---

# Traefik & Ingresses with NGINX Annotations

This provider is a Kubernetes Ingress controller that manages access to cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.
It also supports many of the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/) annotations on Ingresses, enabling teams to migrate from NGINX Ingress Controller to Traefik with minimal configuration changes.

!!! warning "NGINX Ingress Controller Retirement"

    The Kubernetes NGINX Ingress Controller project has announced its retirement in **March 2026** and will no longer receive updates or security patches.
    Traefik provides a migration path by supporting NGINX annotations, allowing you to transition your workloads without rewriting all your Ingress configurations.

    **→ See the [NGINX to Traefik Migration Guide](../../../../migrate/nginx-to-traefik.md) for step-by-step instructions.**

    For more information about the NGINX Ingress Controller retirement, see the [official Kubernetes blog announcement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement).

## Requirements

When you install Traefik without using the Helm Chart, 
ensure that you add/update the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for the Traefik Kubernetes Ingress NGINX provider. 

!!! note "Additional RBAC for Namespace Selector"

    When using the `watchNamespaceSelector` option, Traefik requires permissions to list and watch namespaces.
    These permissions are included in the RBAC configuration below.

```bash
# Install RBAC for Traefik Ingress NGINX provider:
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.6/docs/content/reference/dynamic-configuration/kubernetes-ingress-nginx-rbac.yml
```

## Ingress Discovery

This provider discovers all Ingresses in the cluster by default, which may lead to duplicated routers if you are also using the standard Kubernetes Ingress provider.

**Best Practices:**

- Use IngressClass to specify which Ingresses should be handled by this provider
- Configure `watchNamespace` to limit discovery to specific namespaces
- Use `watchNamespaceSelector` to target Ingresses based on namespace labels

## Configuration Example

You can enable the Kubernetes Ingress NGINX provider as detailed below:

```yaml tab="File (YAML)"
providers:
  kubernetesIngressNGINX:
    # Namespace discovery
    watchNamespace: "default"
    # OR use namespace selector (mutually exclusive with watchNamespace)
    # watchNamespaceSelector: "environment=production"

    # IngressClass configuration
    ingressClass: "nginx"
    controllerClass: "k8s.io/ingress-nginx"
    watchIngressWithoutClass: false
    ingressClassByName: false
```

```toml tab="File (TOML)"
[providers.kubernetesIngressNGINX]
  # Namespace discovery
  watchNamespace = "default"
  # OR use namespace selector (mutually exclusive with watchNamespace)
  # watchNamespaceSelector = "environment=production"

  # IngressClass configuration
  ingressClass = "nginx"
  controllerClass = "k8s.io/ingress-nginx"
  watchIngressWithoutClass = false
  ingressClassByName = false
```

```bash tab="CLI"
--providers.kubernetesingressnginx=true
--providers.kubernetesingressnginx.watchnamespace=default
--providers.kubernetesingressnginx.ingressclass=nginx
--providers.kubernetesingressnginx.controllerclass=k8s.io/ingress-nginx
--providers.kubernetesingressnginx.watchingresswithoutclass=false
--providers.kubernetesingressnginx.ingressclassbyname=false
```

```yaml tab="Helm Chart Values"
providers:
  kubernetesIngressNginx:
    # -- Enable Kubernetes Ingress NGINX provider
    enabled: true

    # Namespace discovery
    # -- Namespace the controller watches for updates to Kubernetes objects
    # When using rbac.namespaced, it will watch helm release namespace and namespaces listed in this array
    namespaces:
      - default
    # OR use namespace selector (mutually exclusive with namespaces)
    # namespaceSelector: "environment=production"

    # IngressClass configuration
    # -- Name of the ingress class this controller satisfies
    ingressClass: "nginx"
    # -- Ingress Class Controller value this controller satisfies
    controllerClass: "k8s.io/ingress-nginx"
    # -- Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified
    watchIngressWithoutClass: false
    # -- Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class
    ingressClassByName: false
```

This provider watches for incoming Ingress events and automatically translates NGINX annotations into Traefik's dynamic configuration, creating the corresponding routers, services, middlewares, and other components needed to route traffic to your cluster services.

## Configuration Options
<!-- markdownlint-disable MD013 -->

| Field                                                                                                                                                                                                                                                                                            | Description | Default | Required |
|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------|:--------|:---------|
| <a id="opt-providers-providers-ThrottleDuration" href="#opt-providers-providers-ThrottleDuration" title="#opt-providers-providers-ThrottleDuration">`providers.providers`<br/>`ThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-endpoint" href="#opt-providers-kubernetesIngressNGINX-endpoint" title="#opt-providers-kubernetesIngressNGINX-endpoint">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`endpoint`</a> | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                        | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-token" href="#opt-providers-kubernetesIngressNGINX-token" title="#opt-providers-kubernetesIngressNGINX-token">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`token`</a> | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                           | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-certAuthFilePath" href="#opt-providers-kubernetesIngressNGINX-certAuthFilePath" title="#opt-providers-kubernetesIngressNGINX-certAuthFilePath">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`certAuthFilePath`</a> | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                           | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-throttleDuration" href="#opt-providers-kubernetesIngressNGINX-throttleDuration" title="#opt-providers-kubernetesIngressNGINX-throttleDuration">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`throttleDuration`</a> | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                           | 0s      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-watchNamespace" href="#opt-providers-kubernetesIngressNGINX-watchNamespace" title="#opt-providers-kubernetesIngressNGINX-watchNamespace">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`watchNamespace`</a> | Namespace the controller watches for updates to Kubernetes objects. All namespaces are watched if this parameter is left empty.                                                                                                                                                                                                                                                      | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-watchNamespaceSelector" href="#opt-providers-kubernetesIngressNGINX-watchNamespaceSelector" title="#opt-providers-kubernetesIngressNGINX-watchNamespaceSelector">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`watchNamespaceSelector`</a> | Selector selects namespaces the controller watches for updates to Kubernetes objects.                                                                                                                                                                                                                                                                                                | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-ingressClass" href="#opt-providers-kubernetesIngressNGINX-ingressClass" title="#opt-providers-kubernetesIngressNGINX-ingressClass">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`ingressClass`</a> | Name of the ingress class this controller satisfies.                                                                                                                                                                                                                                                                                                                                 | "nginx"      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-controllerClass" href="#opt-providers-kubernetesIngressNGINX-controllerClass" title="#opt-providers-kubernetesIngressNGINX-controllerClass">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`controllerClass`</a> | Ingress Class Controller value this controller satisfies.                                                                                                                                                                                                                                                                                                                            | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-watchIngressWithoutClass" href="#opt-providers-kubernetesIngressNGINX-watchIngressWithoutClass" title="#opt-providers-kubernetesIngressNGINX-watchIngressWithoutClass">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`watchIngressWithoutClass`</a> | Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified.                                                                                                                                                                                                                                                                    | false   | No       |
| <a id="opt-providers-kubernetesIngressNGINX-ingressClassByName" href="#opt-providers-kubernetesIngressNGINX-ingressClassByName" title="#opt-providers-kubernetesIngressNGINX-ingressClassByName">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`ingressClassByName`</a> | Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class.                                                                                                                                                                                                                                                                                  | false   | No       |
| <a id="opt-providers-kubernetesIngressNGINX-publishService" href="#opt-providers-kubernetesIngressNGINX-publishService" title="#opt-providers-kubernetesIngressNGINX-publishService">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`publishService`</a> | Service fronting the Ingress controller. Takes the form `namespace/name`.                                                                                                                                                                                                                                                                                                              | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-publishStatusAddress" href="#opt-providers-kubernetesIngressNGINX-publishStatusAddress" title="#opt-providers-kubernetesIngressNGINX-publishStatusAddress">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`publishStatusAddress`</a> | Customized address (or addresses, separated by comma) to set as the load-balancer status of Ingress objects this controller satisfies.                                                                                                                                                                                                                                               | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-defaultBackendService" href="#opt-providers-kubernetesIngressNGINX-defaultBackendService" title="#opt-providers-kubernetesIngressNGINX-defaultBackendService">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`defaultBackendService`</a> | Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form 'namespace/name'.                                                                                                                                                                                                                                                                 | ""      | No       |
| <a id="opt-providers-kubernetesIngressNGINX-disableSvcExternalName" href="#opt-providers-kubernetesIngressNGINX-disableSvcExternalName" title="#opt-providers-kubernetesIngressNGINX-disableSvcExternalName">`providers.`<br/>`kubernetesIngressNGINX.`<br/>`disableSvcExternalName`</a> | Disable support for Services of type ExternalName.                                                                                                                                                                                                                                                                                                                                   | false   | No       |

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

{% include-markdown "includes/traefik-for-business-applications.md" %}
