# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

<!--
 TODO (Link "Kubernetes Ingress controller" to ./kubernetes-ingress.md)
 -->

The Traefik Kubernetes provider used to be a Kubernetes Ingress controller in the strict sense of the term; that is to say,
it would manage access to a cluster services by supporting the [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) specification.

However, as the community expressed the need to benefit from Traefik features without resorting to (lots of) annotations,
we ended up writing a [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (alias CRD in the following) for an IngressRoute type, defined below, in order to provide a better way to configure access to a Kubernetes cluster.

## Provider Configuration

### `endpoint`

_Optional, Default=empty_

The Kubernetes server endpoint as URL.

When deployed into Kubernetes, Traefik will read the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token will be looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are provided mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik will try to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

```toml tab="File"
[Providers.KubernetesCRD]
  endpoint = "http://localhost:8080"
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.endpoint="http://localhost:8080"
```

### `token`

_Optional, Default=empty_

Bearer token used for the Kubernetes client configuration.

```toml tab="File"
[Providers.KubernetesCRD]
  token = "mytoken"
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.token="mytoken"
```

### `certAuthFilePath`

_Optional, Default=empty_

Path to the certificate authority file.
Used for the Kubernetes client configuration.

```toml tab="File"
[Providers.KubernetesCRD]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.certauthfilepath="/my/ca.crt"
```

### `namespaces`

_Optional, Default: all namespaces (empty array)_

Array of namespaces to watch.

```toml tab="File"
[Providers.KubernetesCRD]
  namespaces = ["default", "production"]
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.namespaces="default,production"
```

### `labelselector`

_Optional,Default: empty (process all Ingresses)_

By default, Traefik processes all Ingress objects in the configured namespaces.
A label selector can be defined to filter on specific Ingress objects only.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

```toml tab="File"
[Providers.KubernetesCRD]
  labelselector = "A and not B"
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.labelselector="A and not B"
```

### `ingressClass`

_Optional, Default: empty_

Value of `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

If the parameter is non-empty, only Ingresses containing an annotation with the same value are processed.
Otherwise, Ingresses missing the annotation, having an empty value, or the value `traefik` are processed.

```toml tab="File"
[Providers.KubernetesCRD]
  ingressClass = "traefik-internal"
  # ...
```

```txt tab="CLI"
--providers.kubernetescrd
--providers.kubernetescrd.ingressclass="traefik-internal"
```

## Resource Configuration

If you're in a hurry, maybe you'd rather go through the [dynamic](../reference/dynamic-configuration/kubernetes-crd.md) configuration reference.

### Traefik IngressRoute definition

```yaml
--8<-- "content/providers/crd_ingress_route.yml"
```

That `IngressRoute` kind can then be used to define an `IngressRoute` object, such as in:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutefoo

spec:
  entryPoints:
    - web
  routes:
  # Match is the rule corresponding to an underlying router.
  # Later on, match could be the simple form of a path prefix, e.g. just "/bar",
  # but for now we only support a traefik style matching rule.
  - match: Host(`foo.com`) && PathPrefix(`/bar`)
    # kind could eventually be one of "Rule", "Path", "Host", "Method", "Header",
    # "Parameter", etc, to support simpler forms of rule matching, but for now we
    # only support "Rule".
    kind: Rule
    # Priority disambiguates rules of the same length, for route matching.
    priority: 12
    services:
    - name: whoami
      port: 80

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: ingressroutetcpfoo.crd

spec:
  entryPoints:
    - footcp
  routes:
  # Match is the rule corresponding to an underlying router.
  - match: HostSNI(`*`)
    services:
    - name: whoamitcp
      port: 8080
```

### Middleware

Additionally, to allow for the use of middlewares in an `IngressRoute`, we defined the CRD below for the `Middleware` kind.

```yaml
--8<-- "content/providers/crd_middlewares.yml"
```

Once the `Middleware` kind has been registered with the Kubernetes cluster, it can then be used in `IngressRoute` definitions, such as:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: stripprefix

spec:
  stripPrefix:
    prefixes:
      - /stripit

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutebar

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`bar.com`) && PathPrefix(`/stripit`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: stripprefix
```

More information about available middlewares in the dedicated [middlewares section](../middlewares/overview.md).

### Traefik TLS Option Definition

Additionally, to allow for the use of tls options in an IngressRoute, we defined the CRD below for the TLSOption kind.
More information about TLS Options is available in the dedicated [TLS Configuration Options](../../https/tls/#tls-options).

```yaml
--8<-- "content/providers/crd_tls_option.yml"
```

Once the TLSOption kind has been registered with the Kubernetes cluster or defined in the File Provider, it can then be used in IngressRoute definitions, such as:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: TLSOption
metadata:
  name: mytlsoption
  namespace: default

spec:
  minversion: VersionTLS12

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutebar

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`bar.com`) && PathPrefix(`/stripit`)
    kind: Rule
    services:
    - name: whoami
      port: 80
  tls:
    options: 
      name: mytlsoption
      namespace: default
```

!!! note "TLS Option reference and namespace"
    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the IngressRoute.

### TLS

To allow for TLS, we made use of the `Secret` kind, as it was already defined, and it can be directly used in an `IngressRoute`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: supersecret

data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=

---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutetls

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`foo.com`) && PathPrefix(`/bar`)
    kind: Rule
    services:
    - name: whoami
      port: 443
  tls:
    secretName: supersecret
```

## Further

Also see the [full example](../user-guides/crd-acme/index.md) with Let's Encrypt.
