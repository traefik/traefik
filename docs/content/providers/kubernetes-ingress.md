# Traefik & Kubernetes

The Kubernetes Ingress Controller.
{: .subtitle }

The Traefik Kubernetes Ingress provider is a Kubernetes Ingress controller; that is to say,
it manages access to a cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.

## Routing Configuration

See the dedicated section in [routing](../routing/providers/kubernetes-ingress.md).

## Enabling and Using the Provider

As usual, the provider is enabled through the static configuration:

```toml tab="File (TOML)"
[providers.kubernetesIngress]
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress: {}
```

```bash tab="CLI"
--providers.kubernetesingress=true
```

The provider then watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn will create the resulting routers, services, handlers, etc.

```yaml tab="File (YAML)"
kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: "foo"
  namespace: production

spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
          - path: /foo
            backend:
              serviceName: service1
              servicePort: 80
```

## LetsEncrypt Support with the Ingress Provider

By design, Traefik is a stateless application,
meaning that it only derives its configuration from the environment it runs in,
without additional configuration.
For this reason, users can run multiple instances of Traefik at the same time to achieve HA,
as is a common pattern in the kubernetes ecosystem.

When using a single instance of Traefik with LetsEncrypt, no issues should be encountered,
however this could be a single point of failure.
Unfortunately, it is not possible to run multiple instances of Traefik 2.0 with LetsEncrypt enabled,
because there is no way to ensure that the correct instance of Traefik will receive the challenge request, and subsequent responses.
Previous versions of Traefik used a [KV store](https://doc.traefik.io/traefik/v1.7/configuration/acme/#storage) to attempt to achieve this,
but due to sub-optimal performance was dropped as a feature in 2.0.

If you require LetsEncrypt with HA in a kubernetes environment,
we recommend using [Traefik Enterprise](https://traefik.io/traefik-enterprise/) where distributed LetsEncrypt is a supported feature.

If you are wanting to continue to run Traefik Community Edition,
LetsEncrypt HA can be achieved by using a Certificate Controller such as [Cert-Manager](https://docs.cert-manager.io/en/latest/index.html).
When using Cert-Manager to manage certificates,
it will create secrets in your namespaces that can be referenced as TLS secrets in your [ingress objects](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls).

## Provider Configuration

### `endpoint`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  endpoint = "http://localhost:8080"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    endpoint: "http://localhost:8080"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.endpoint=http://localhost:8080
```

The Kubernetes server endpoint as URL, which is only used when the behavior based on environment variables described below does not apply.

When deployed into Kubernetes, Traefik reads the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
They are both provided automatically as mounts in the pod where Traefik is deployed.

When the environment variables are not found, Traefik tries to connect to the Kubernetes API server with an external-cluster client.
In which case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication
and authorization of the associated kubeconfig.

### `token`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  token = "mytoken"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    token: "mytoken"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.token=mytoken
```

Bearer token used for the Kubernetes client configuration.

### `certAuthFilePath`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.certauthfilepath=/my/ca.crt
```

Path to the certificate authority file.
Used for the Kubernetes client configuration.

### `disablePassHostHeaders`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  disablePassHostHeaders = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    disablePassHostHeaders: true
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.disablepasshostheaders=true
```

Whether to disable PassHost Headers.

### `namespaces`

_Optional, Default: all namespaces (empty array)_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  namespaces = ["default", "production"]
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    namespaces:
      - "default"
      - "production"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.namespaces=default,production
```

Array of namespaces to watch.

### `labelSelector`

_Optional,Default: empty (process all Ingresses)_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  labelSelector = "A and not B"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    labelselector: "A and not B"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.labelselector="A and not B"
```

By default, Traefik processes all Ingress objects in the configured namespaces.
A label selector can be defined to filter on specific Ingress objects only.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

### `ingressClass`

_Optional, Default: empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  ingressClass = "traefik-internal"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressClass: "traefik-internal"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressclass=traefik-internal
```

Value of `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

If the parameter is non-empty, only Ingresses containing an annotation with the same value are processed.
Otherwise, Ingresses missing the annotation, having an empty value, or with the value `traefik` are processed.

!!! info "Kubernetes 1.18+"

    If the Kubernetes cluster version is 1.18+,
    the new `IngressClass` resource can be leveraged to identify Ingress objects that should be processed.
    In that case, Traefik will look for an `IngressClass` in the cluster with the controller value equal to *traefik.io/ingress-controller*. 
    
    Please see [this article](https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/) for more information or the example below.

    ```yaml tab="IngressClass"
    apiVersion: networking.k8s.io/v1beta1
    kind: IngressClass
    metadata: 
      name: traefik-lb
    spec: 
      controller: traefik.io/ingress-controller
    ```
  
    ```yaml tab="Ingress"
    apiVersion: "networking.k8s.io/v1beta1"
    kind: "Ingress"
    metadata:
      name: "example-ingress"
    spec:
      ingressClassName: "traefik-lb"
      rules:
      - host: "*.example.com"
        http:
          paths:
          - path: "/example"
            backend:
              serviceName: "example-service"
              servicePort: 80
    ```

### `ingressEndpoint`

#### `hostname`

_Optional, Default: empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  hostname = "example.net"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      hostname: "example.net"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.hostname=example.net
```

Hostname used for Kubernetes Ingress endpoints.

#### `ip`

_Optional, Default: empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  ip = "1.2.3.4"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      ip: "1.2.3.4"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.ip=1.2.3.4
```

IP used for Kubernetes Ingress endpoints.

#### `publishedService`

_Optional, Default: empty_

```toml tab="File (TOML)"
[providers.kubernetesIngress.ingressEndpoint]
  publishedService = "namespace/foo-service"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    ingressEndpoint:
      publishedService: "namespace/foo-service"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.ingressendpoint.publishedservice=namespace/foo-service
```

Published Kubernetes Service to copy status from.
Format: `namespace/servicename`.

### `throttleDuration`

_Optional, Default: 0 (no throttling)_

```toml tab="File (TOML)"
[providers.kubernetesIngress]
  throttleDuration = "10s"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesIngress:
    throttleDuration: "10s"
    # ...
```

```bash tab="CLI"
--providers.kubernetesingress.throttleDuration=10s
```

### Further

If one wants to know more about the various aspects of the Ingress spec that Traefik supports,
many examples of Ingresses definitions are located in the tests [data](https://github.com/traefik/traefik/tree/v2.3/pkg/provider/kubernetes/ingress/fixtures) of the Traefik repository.
