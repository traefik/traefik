---
title: "Traefik Kubernetes Ingress Documentation"
description: "Understand the requirements, routing configuration, and how to set up Traefik Proxy as your Kubernetes Ingress Controller. Read the technical documentation."
---

# Traefik & Kubernetes

The Kubernetes Ingress Controller.
{: .subtitle }

The Traefik Kubernetes Ingress provider is a Kubernetes Ingress controller; that is to say,
it manages access to cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.

## Requirements

{!kubernetes-requirements.md!}

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kubernetes-ingress.md).

## Enabling and Using the Provider

You can enable the provider in the static configuration:

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

The provider then watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn creates the resulting routers, services, handlers, etc.

```yaml tab="Ingress"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: foo
  namespace: production

spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            pathType: Exact
            backend:
              service:
                name:  service1
                port:
                  number: 80
          - path: /foo
            pathType: Exact
            backend:
              service:
                name:  service1
                port:
                  number: 80
```

## LetsEncrypt Support with the Ingress Provider

By design, Traefik is a stateless application,
meaning that it only derives its configuration from the environment it runs in,
without additional configuration.
For this reason, users can run multiple instances of Traefik at the same time to achieve HA,
as is a common pattern in the kubernetes ecosystem.

When using a single instance of Traefik Proxy with Let's Encrypt, you should encounter no issues.
However, this could be a single point of failure.
Unfortunately, it is not possible to run multiple instances of Traefik 2.0 with Let's Encrypt enabled,
because there is no way to ensure that the correct instance of Traefik receives the challenge request, and subsequent responses.
Early versions (v1.x) of Traefik used a [KV store](https://doc.traefik.io/traefik/v1.7/configuration/acme/#storage) to attempt to achieve this,
but due to sub-optimal performance that feature was dropped in 2.0.

If you need Let's Encrypt with high availability in a Kubernetes environment,
we recommend using [Traefik Enterprise](https://traefik.io/traefik-enterprise/) which includes distributed Let's Encrypt as a supported feature.

If you want to keep using Traefik Proxy,
LetsEncrypt HA can be achieved by using a Certificate Controller such as [Cert-Manager](https://cert-manager.io/docs/).
When using Cert-Manager to manage certificates,
it creates secrets in your namespaces that can be referenced as TLS secrets in your [ingress objects](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls).

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

### `token`

_Optional, Default=""_

Bearer token used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    token: "mytoken"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  token = "mytoken"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.token=mytoken
```

### `certAuthFilePath`

_Optional, Default=""_

Path to the certificate authority file.
Used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.certauthfilepath=/my/ca.crt
```

### `namespaces`

_Optional, Default: []_

Array of namespaces to watch.
If left empty, Traefik watches all namespaces.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    namespaces:
      - "default"
      - "production"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  namespaces = ["default", "production"]
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.namespaces=default,production
```

### `labelSelector`

_Optional, Default: ""_

A label selector can be defined to filter on specific Ingress objects only.
If left empty, Traefik processes all Ingress objects in the configured namespaces.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    labelSelector: "app=traefik"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  labelSelector = "app=traefik"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.labelselector="app=traefik"
```

### `ingressClass`

_Optional, Default: ""_

Value of `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

If the parameter is set, only Ingresses containing an annotation with the same value are processed.
Otherwise, Ingresses missing the annotation, having an empty value, or the value `traefik` are processed.

??? info "Example"

    ```yaml tab="IngressClass"
    apiVersion: networking.k8s.io/v1
    kind: IngressClass
    metadata:
      name: traefik-lb
    spec:
      controller: traefik.io/ingress-controller
    ```

    ```yaml tab="Ingress"
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: example-ingress
    spec:
      ingressClassName: traefik-lb
      rules:
      - host: "*.example.com"
        http:
          paths:
          - path: /example
            pathType: Exact
            backend:
              service:
                name: example-service
                port:
                    number: 80
    ```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressClass: "traefik-internal"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  ingressClass = "traefik-internal"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressclass=traefik-internal
```

### `disableIngressClassLookup`

_Optional, Default: false_

??? warning "Deprecated"

    The Kubernetes Ingress provider option `disableIngressClassLookup` has been deprecated in v3.1, and will be removed in the next major version.
	Please use the `disableClusterScopeResources` option instead.

If the parameter is set to `true`,
Traefik will not discover IngressClasses in the cluster.
By doing so, it alleviates the requirement of giving Traefik the rights to look IngressClasses up.
Furthermore, when this option is set to `true`,
Traefik is not able to handle Ingresses with IngressClass references,
therefore such Ingresses will be ignored.
Please note that annotations are not affected by this option.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    disableIngressClassLookup: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  disableIngressClassLookup = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.disableingressclasslookup=true
```

### `disableClusterScopeResources`

_Optional, Default: false_

When this parameter is set to `true`,
Traefik will not discover cluster scope resources (`IngressClass` and `Nodes`).
By doing so, it alleviates the requirement of giving Traefik the rights to look up for cluster resources.
Furthermore, Traefik will not handle Ingresses with IngressClass references, therefore such Ingresses will be ignored (please note that annotations are not affected by this option).
This will also prevent from using the `NodePortLB` options on services.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    disableClusterScopeResources: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  disableClusterScopeResources = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.disableClusterScopeResources=true
```

### `ingressEndpoint`

#### `hostname`

_Optional, Default: ""_

Hostname used for Kubernetes Ingress endpoints.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      hostname: "example.net"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  hostname = "example.net"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.hostname=example.net
```

#### `ip`

_Optional, Default: ""_

This IP will get copied to Ingress `status.loadbalancer.ip`, and currently only supports one IP value (IPv4 or IPv6).

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      ip: "1.2.3.4"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  ip = "1.2.3.4"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.ip=1.2.3.4
```

#### `publishedService`

_Optional, Default: ""_

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

### `throttleDuration`

_Optional, Default: 0_

The `throttleDuration` option defines how often the provider is allowed to handle events from Kubernetes. This prevents
a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.

If left empty, the provider does not apply any throttling and does not drop any Kubernetes events.

The value of `throttleDuration` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    throttleDuration: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  throttleDuration = "10s"
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.throttleDuration=10s
```

### `allowEmptyServices`

_Optional, Default: false_

If the parameter is set to `true`,
it allows the creation of an empty [servers load balancer](../routing/services/index.md#servers-load-balancer) if the targeted Kubernetes service has no endpoints available.
This results in `503` HTTP responses instead of `404` ones.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    allowEmptyServices: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  allowEmptyServices = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.allowEmptyServices=true
```

### `allowExternalNameServices`

_Optional, Default: false_

If the parameter is set to `true`,
Ingresses are able to reference ExternalName services.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    allowExternalNameServices: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  allowExternalNameServices = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.allowexternalnameservices=true
```

### `nativeLBByDefault`

_Optional, Default: false_

Defines whether to use Native Kubernetes load-balancing mode by default.
For more information, please check out the `traefik.ingress.kubernetes.io/service.nativelb` [service annotation documentation](../routing/providers/kubernetes-ingress.md#on-service).

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    nativeLBByDefault: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  nativeLBByDefault = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.nativeLBByDefault=true
```

### `strictPrefixMatching`

_Optional, Default: false_

Make prefix matching strictly comply with the Kubernetes Ingress specification (path-element-wise matching instead of character-by-character string matching). For example, a PathPrefix of `/foo` will match `/foo`, `/foo/`, and `/foo/bar` but not `/foobar`.

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    strictPrefixMatching: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  strictPrefixMatching = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesingress.strictPrefixMatching=true
```

### Further

To learn more about the various aspects of the Ingress specification that Traefik supports,
many examples of Ingresses definitions are located in the test [examples](https://github.com/traefik/traefik/tree/v3.5/pkg/provider/kubernetes/ingress/fixtures) of the Traefik repository.

{!traefik-for-business-applications.md!}
