# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

Traefik used to support Kubernetes only through the [Kubernetes Ingress provider](./kubernetes-ingress.md), which is a Kubernetes Ingress controller in the strict sense of the term.

However, as the community expressed the need to benefit from Traefik features without resorting to (lots of) annotations,
we ended up writing a [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (alias CRD in the following) for an IngressRoute type, defined below, in order to provide a better way to configure access to a Kubernetes cluster.

## Configuration Requirements

!!! tip "All Steps for a Successful Deployment"

    * Add/update **all** the Traefik resources [definitions](../reference/dynamic-configuration/kubernetes-crd.md#definitions)
    * Add/update the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for the Traefik custom resources
    * Use [Helm Chart](../getting-started/install-traefik.md#use-the-helm-chart) or use a custom Traefik Deployment 
        * Enable the kubernetesCRD provider
        * Apply the needed kubernetesCRD provider [configuration](#provider-configuration)
    * Add all needed traefik custom [resources](../reference/dynamic-configuration/kubernetes-crd.md#resources)
 
??? example "Initializing Resource Definition and RBAC"

    ```yaml tab="Traefik Resource Definition"
    # All resources definition must be declared
    --8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition.yml"
    ```

    ```yaml tab="RBAC for Traefik CRD"
    --8<-- "content/reference/dynamic-configuration/kubernetes-crd-rbac.yml"
    ```

## Resource Configuration

When using KubernetesCRD as a provider,
Traefik uses [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to retrieve its routing configuration.
Traefik Custom Resource Definitions are a Kubernetes implementation of the Traefik concepts. The main particularities are:

* The usage of `name` **and** `namespace` to refer to another Kubernetes resource.
* The usage of [secret](https://kubernetes.io/docs/concepts/configuration/secret/) for sensible data like:
    * TLS certificate.
    * Authentication data.
* The structure of the configuration.
* The obligation to declare all the [definitions](../reference/dynamic-configuration/kubernetes-crd.md#definitions).

The Traefik CRD are building blocks which you can assemble according to your needs.
See the list of CRDs in the dedicated [routing section](../routing/providers/kubernetes-crd.md).

## LetsEncrypt Support with the Custom Resource Definition Provider

By design, Traefik is a stateless application, meaning that it only derives its configuration from the environment it runs in, without additional configuration.
For this reason, users can run multiple instances of Traefik at the same time to achieve HA, as is a common pattern in the kubernetes ecosystem.

When using a single instance of Traefik with LetsEncrypt, no issues should be encountered, however this could be a single point of failure.
Unfortunately, it is not possible to run multiple instances of Traefik 2.0 with LetsEncrypt enabled, because there is no way to ensure that the correct instance of Traefik will receive the challenge request, and subsequent responses.
Previous versions of Traefik used a [KV store](https://doc.traefik.io/traefik/v1.7/configuration/acme/#storage) to attempt to achieve this, but due to sub-optimal performance was dropped as a feature in 2.0.

If you require LetsEncrypt with HA in a kubernetes environment, we recommend using [Traefik Enterprise](https://traefik.io/traefik-enterprise/) where distributed LetsEncrypt is a supported feature.

If you are wanting to continue to run Traefik Community Edition, LetsEncrypt HA can be achieved by using a Certificate Controller such as [Cert-Manager](https://docs.cert-manager.io/en/latest/index.html).
When using Cert-Manager to manage certificates, it will create secrets in your namespaces that can be referenced as TLS secrets in your [ingress objects](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls).
When using the Traefik Kubernetes CRD Provider, unfortunately Cert-Manager cannot interface directly with the CRDs _yet_, but this is being worked on by our team.
A workaround is to enable the [Kubernetes Ingress provider](./kubernetes-ingress.md) to allow Cert-Manager to create ingress objects to complete the challenges.
Please note that this still requires manual intervention to create the certificates through Cert-Manager, but once created, Cert-Manager will keep the certificate renewed.

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
    endpoint: "http://localhost:8080"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.endpoint=http://localhost:8080
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
    token: "mytoken"
    # ...
```

```bash tab="CLI"
--providers.kubernetescrd.token=mytoken
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
--providers.kubernetescrd.certauthfilepath=/my/ca.crt
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
--providers.kubernetescrd.namespaces=default,production
```

Array of namespaces to watch.

### `labelselector`

_Optional,Default: empty (process all resources)_

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

By default, Traefik processes all resource objects in the configured namespaces.
A label selector can be defined to filter on specific resource objects only.

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
--providers.kubernetescrd.ingressclass=traefik-internal
```

Value of `kubernetes.io/ingress.class` annotation that identifies resource objects to be processed.

If the parameter is non-empty, only resources containing an annotation with the same value are processed.
Otherwise, resources missing the annotation, having an empty value, or the value `traefik` are processed.

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
--providers.kubernetescrd.throttleDuration=10s
```

## Further

Also see the [full example](../user-guides/crd-acme/index.md) with Let's Encrypt.
