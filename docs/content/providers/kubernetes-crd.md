---
title: "Kubernetes IngressRoute & Traefik CRD"
description: "The Traefik team developed a Custom Resource Definition (CRD) for an IngressRoute type, to provide a better way to configure access to a Kubernetes cluster."
---

# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

In early versions, Traefik supported Kubernetes only through the [Kubernetes Ingress provider](./kubernetes-ingress.md), which is a Kubernetes Ingress controller in the strict sense of the term.

However, as the community expressed the need to benefit from Traefik features without resorting to (lots of) annotations,
the Traefik engineering team developed a [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
(CRD) for an IngressRoute type, defined below, in order to provide a better way to configure access to a Kubernetes cluster.

## Configuration Requirements

!!! tip "All Steps for a Successful Deployment"

    * Add/update **all** the Traefik resources [definitions](../reference/dynamic-configuration/kubernetes-crd.md#definitions)
    * Add/update the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for the Traefik custom resources
    * Use [Helm Chart](../getting-started/install-traefik.md#use-the-helm-chart) or use a custom Traefik Deployment
        * Enable the kubernetesCRD provider
        * Apply the needed kubernetesCRD provider [configuration](#provider-configuration)
    * Add all necessary Traefik custom [resources](../reference/dynamic-configuration/kubernetes-crd.md#resources)

!!! warning "Deprecated apiextensions.k8s.io/v1beta1 CRD"

    The `apiextensions.k8s.io/v1beta1` CustomResourceDefinition is deprecated in Kubernetes `v1.16+` and will be removed in `v1.22+`.
    
    For Kubernetes `v1.16+`, please use the Traefik `apiextensions.k8s.io/v1` CRDs instead.

!!! example "Installing Resource Definition and RBAC"

    ```bash
    # Install Traefik Resource Definitions:
    kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.11/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml
    
    # Install RBAC for Traefik:
    kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.11/docs/content/reference/dynamic-configuration/kubernetes-crd-rbac.yml
    ```

## Resource Configuration

When using KubernetesCRD as a provider,
Traefik uses [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to retrieve its routing configuration.
Traefik Custom Resource Definitions are a Kubernetes implementation of the Traefik concepts. The main particularities are:

* The usage of `name` **and** `namespace` to refer to another Kubernetes resource.
* The usage of [secret](https://kubernetes.io/docs/concepts/configuration/secret/) for sensitive data (TLS certificates and credentials).
* The structure of the configuration.
* The requirement to declare all the [definitions](../reference/dynamic-configuration/kubernetes-crd.md#definitions).

The Traefik CRDs are building blocks that you can assemble according to your needs.
See the list of CRDs in the dedicated [routing section](../routing/providers/kubernetes-crd.md).

## LetsEncrypt Support with the Custom Resource Definition Provider

By design, Traefik is a stateless application, meaning that it only derives its configuration from the environment it runs in, without additional configuration.
For this reason, users can run multiple instances of Traefik at the same time to achieve HA, as is a common pattern in the kubernetes ecosystem.

When using a single instance of Traefik with Let's Encrypt, you should encounter no issues. However, this could be a single point of failure.
Unfortunately, it is not possible to run multiple instances of Traefik Proxy 2.0 with Let's Encrypt enabled, because there is no way to ensure that the correct instance of Traefik will receive the challenge request and subsequent responses.
Previous versions of Traefik used a [KV store](https://doc.traefik.io/traefik/v1.7/configuration/acme/#storage) to attempt to achieve this, but due to sub-optimal performance that feature was dropped in 2.0.

If you need Let's Encrypt with HA in a Kubernetes environment, we recommend using [Traefik Enterprise](https://traefik.io/traefik-enterprise/), which includes distributed Let's Encrypt as a supported feature.

If you want to keep using Traefik Proxy, high availability for Let's Encrypt can be achieved by using a Certificate Controller such as [Cert-Manager](https://cert-manager.io/docs/).
When using Cert-Manager to manage certificates, it creates secrets in your namespaces that can be referenced as TLS secrets in your [ingress objects](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls).
When using the Traefik Kubernetes CRD Provider, unfortunately Cert-Manager cannot yet interface directly with the CRDs.
A workaround is to enable the [Kubernetes Ingress provider](./kubernetes-ingress.md) to allow Cert-Manager to create ingress objects to complete the challenges.
Please note that this still requires manual intervention to create the certificates through Cert-Manager, but once the certificates are created, Cert-Manager keeps them renewed.

## Provider Configuration

### `endpoint`

_Optional, Default=""_

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
--providers.kubernetescrd.endpoint=http://localhost:8080
```

### `token`

_Optional, Default=""_

Bearer token used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    token: "mytoken"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  token = "mytoken"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.token=mytoken
```

### `certAuthFilePath`

_Optional, Default=""_

Path to the certificate authority file.
Used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.certauthfilepath=/my/ca.crt
```

### `namespaces`

_Optional, Default: []_

Array of namespaces to watch.
If left empty, Traefik watches all namespaces.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    namespaces:
    - "default"
    - "production"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  namespaces = ["default", "production"]
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.namespaces=default,production
```

### `labelselector`

_Optional, Default: ""_

A label selector can be defined to filter on specific resource objects only,
this applies only to Traefik [Custom Resources](../routing/providers/kubernetes-crd.md#custom-resource-definition-crd)
and has no effect on Kubernetes `Secrets`, `Endpoints` and `Services`.
If left empty, Traefik processes all resource objects in the configured namespaces.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

!!! warning

    Because the label selector is applied to all Traefik Custom Resources, they all must match the filter.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    labelSelector: "app=traefik"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  labelSelector = "app=traefik"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.labelselector="app=traefik"
```

### `ingressClass`

_Optional, Default: ""_

Value of `kubernetes.io/ingress.class` annotation that identifies resource objects to be processed.

If the parameter is set, only resources containing an annotation with the same value are processed.
Otherwise, resources missing the annotation, having an empty value, or the value `traefik` are processed.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    ingressClass: "traefik-internal"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  ingressClass = "traefik-internal"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.ingressclass=traefik-internal
```

### `throttleDuration`

_Optional, Default: 0_

The `throttleDuration` option defines how often the provider is allowed to handle events from Kubernetes. This prevents
a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.

If left empty, the provider does not apply any throttling and does not drop any Kubernetes events.

The value of `throttleDuration` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    throttleDuration: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  throttleDuration = "10s"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.throttleDuration=10s
```

### `allowEmptyServices`

_Optional, Default: false_

If the parameter is set to `true`,
it allows the creation of an empty [servers load balancer](../routing/services/index.md#servers-load-balancer) if the targeted Kubernetes service has no endpoints available.
With IngressRoute resources,
this results in `503` HTTP responses instead of `404` ones.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    allowEmptyServices: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  allowEmptyServices = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesCRD.allowEmptyServices=true
```

### `allowCrossNamespace`

_Optional, Default: false_

If the parameter is set to `true`,
IngressRoute are able to reference resources in namespaces other than theirs.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    allowCrossNamespace: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  allowCrossNamespace = true
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.allowCrossNamespace=true
```

### `allowExternalNameServices`

_Optional, Default: false_

If the parameter is set to `true`, IngressRoutes are able to reference ExternalName services.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    allowExternalNameServices: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  allowExternalNameServices = true
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.allowexternalnameservices=true
```

### `crossProviderNamespaces`

_Optional, Default: nil_

`crossProviderNamespaces` is the list of namespaces from which IngressRoute, IngressRouteTCP, IngressRouteUDP, and TraefikService,
are allowed to declare cross-provider references to Traefik resources (Services, Middlewares, TLSOptions, ServersTransports, ...).

| Value      | Behavior                                                                                           |
|------------|----------------------------------------------------------------------------------------------------|
| not set    | Resources may declare cross-provider references from any namespace (default, backward compatible). |
| `[]`       | Every resource declaring a cross-provider reference is rejected.                                   |
| `["ns-a"]` | Only resources in the listed namespaces may declare cross-provider references.                     |

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    crossProviderNamespaces:
      - ns-a
      - ns-b
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  crossProviderNamespaces = ["ns-a", "ns-b"]
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.crossProviderNamespaces=ns-a,ns-b
```

### `defaultTLSResourcesNamespace`

_Optional, Default: ""_

`defaultTLSResourcesNamespace` restricts the namespace in which the default [TLSOption](../https/tls.md#tls-options)
and the default [TLSStore](../https/tls.md#certificates-stores) can be defined.

The TLSOption and the TLSStore named `default` are cluster-wide, whatever the namespace they are defined in:
the former holds the TLS enforcement policy of every router that does not reference a TLSOption explicitly,
the latter holds the default certificate served by every entry point.
This option allows the cluster operator to reserve their definition to a namespace they control.

| Value    | Behavior                                                                                                                         |
|----------|------------------------------------------------------------------------------------------------------------------------------------|
| not set  | A TLSOption or a TLSStore named `default` is taken into account whatever its namespace (default, backward compatible).            |
| `"ns-a"` | Only the resources named `default` in the `ns-a` namespace are taken into account, the others are ignored.                        |

!!! warning "Ignored resources"

    A TLSOption or a TLSStore named `default` defined outside of the configured namespace is ignored,
    and cannot be referenced under its namespaced name either.
    For a TLSStore, this also applies to the certificates it defines.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    defaultTLSResourcesNamespace: traefik
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  defaultTLSResourcesNamespace = "traefik"
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.defaultTLSResourcesNamespace=traefik
```

### `safeNaming`

_Optional, Default: false_

By default, the Kubernetes CRD provider generates the names of the routers, middlewares and services it builds by
joining the namespace and the name of the object they come from, which can produce the same name for two distinct
objects, one silently replacing the other.

`safeNaming` enables collision-safe naming instead: generated names are derived from the identity of the object
they come from, and Kubernetes Services referenced from several parents (a route with several services, or a
Weighted/Mirroring TraefikService) are scoped to their parent instead of being shared by identity.

| Value    | Behavior                                                                                     |
|----------|-----------------------------------------------------------------------------------------------|
| not set  | Current naming is used (default, backward compatible), and a warning is logged on startup.    |
| `true`   | Collision-safe naming is used.                                                                 |
| `false`  | Current naming is used, and the startup warning is silenced.                                  |

!!! warning "Startup warning"

    When `safeNaming` is left unset, a warning is logged on startup, since the current naming scheme is collision-prone.
    It is recommended to explicitly set this option, to `true` on new setups, or to `false` to keep the current
    behavior and silence the warning.

When `safeNaming` is enabled, the generated names are no longer normalized: their components are joined with a
`_` separator, which cannot appear in a Kubernetes namespace or name, and the ones generated for a route are
derived from the route index instead of its rule. For example, for a Kubernetes Service named `whoami` and an
IngressRoute named `test.route`, both in the `default` namespace:

```text
default-whoami-80                          ->    default_whoami_80
default-test-route-6b204d94623b3df4370c    ->    default_test.route_0
```

The services generated for the Kubernetes Services referenced by a `TraefikService` (weighted or mirroring),
or by a route with several services, are named after the parent declaring the reference,
followed by the index of the reference, and the namespace, the name and the port of the referenced Kubernetes Service:

```text
default-whoami-80    ->    default_wrr1_wrr_1_default_whoami_80
```

Each of these references carries its own options (`serversTransport`, `scheme`, `sticky`, `healthCheck`, ...),
which were not part of the generated name before: two references to the same Kubernetes Service with different
options were collapsed into a single service, and the last one built silently won. With `safeNaming` enabled,
they are distinct services, which also means that the servers of a Kubernetes Service referenced from several
parents are health checked once per reference, instead of once for all of them.

!!! warning "Observability"

    These names are user-visible: they appear in the dashboard and API, in the access logs `RouterName` and `ServiceName` fields,
    and in the `router` and `service` labels of the metrics.
    Dashboards, alerting rules, and log queries that match on Kubernetes CRD router, middleware or service names must be updated accordingly
    when `safeNaming` is enabled.

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    safeNaming: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  safeNaming = true
  # ...
```

```bash tab="CLI"
--providers.kubernetescrd.safeNaming=true
```

## Full Example

For additional information, refer to the [full example](../user-guides/crd-acme/index.md) with Let's Encrypt.

{% include-markdown "includes/traefik-for-business-applications.md" %}
