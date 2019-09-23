# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

Traefik used to support Kubernetes only through the [Kubernetes Ingress provider](./kubernetes-ingress.md), which is a Kubernetes Ingress controller in the strict sense of the term.

However, as the community expressed the need to benefit from Traefik features without resorting to (lots of) annotations,
we ended up writing a [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (alias CRD in the following) for an IngressRoute type, defined below, in order to provide a better way to configure access to a Kubernetes cluster.

## Resource Configuration

See the dedicated section in [routing](../routing/providers/kubernetes-crd.md).

## Provider Configuration

### `endpoint`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  endpoint = "http://localhost:8080"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    endpoint = "http://localhost:8080"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.endpoint="http://localhost:8080"
```

The Kubernetes server endpoint as URL.

When deployed into Kubernetes, Traefik will read the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token will be looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are provided mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik will try to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

### `token`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  token = "mytoken"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    token = "mytoken"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.token="mytoken"
```

Bearer token used for the Kubernetes client configuration.

### `certAuthFilePath`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.certauthfilepath="/my/ca.crt"
```

Path to the certificate authority file.
Used for the Kubernetes client configuration.

### `namespaces`

_Optional, Default: all namespaces (empty array)_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  namespaces = ["default", "production"]
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    namespaces:
    - "default"
    - "production"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.namespaces="default,production"
```

Array of namespaces to watch.

### `labelselector`

_Optional,Default: empty (process all Ingresses)_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  labelselector = "A and not B"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    labelselector: "A and not B"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.labelselector="A and not B"
```

By default, Traefik processes all Ingress objects in the configured namespaces.
A label selector can be defined to filter on specific Ingress objects only.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

### `ingressClass`

_Optional, Default: empty_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  ingressClass = "traefik-internal"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    ingressClass: "traefik-internal"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.ingressclass="traefik-internal"
```

Value of `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

If the parameter is non-empty, only Ingresses containing an annotation with the same value are processed.
Otherwise, Ingresses missing the annotation, having an empty value, or the value `traefik` are processed.

### `throttleDuration`

_Optional, Default: 0 (no throttling)_

```toml tab="File (TOML)"
[providers.kubernetesCRD]
  throttleDuration = "10s"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesCRD:
    throttleDuration: "10s"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.throttleDuration="10s"
```

## Further

Also see the [full example](../user-guides/crd-acme/index.md) with Let's Encrypt.
