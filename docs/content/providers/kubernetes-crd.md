# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

Traefik used to support Kubernetes only through the [Kubernetes Ingress provider](./kubernetes-ingress.md), which is a Kubernetes Ingress controller in the strict sense of the term.

However, as the community expressed the need to benefit from Traefik features without resorting to (lots of) annotations,
we ended up writing a [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) (alias CRD in the following) for an IngressRoute type, defined below, in order to provide a better way to configure access to a Kubernetes cluster.

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
  namespace: foo

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
      namespace: foo
```

!!! important "Cross-provider namespace"

	As Kubernetes also has its own notion of namespace, one should not confuse the kubernetes namespace of a resource
(in the reference to the middleware) with the [provider namespace](../middlewares/overview.md#provider-namespace),
when the definition of the middleware is from another provider.
In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored.

More information about available middlewares in the dedicated [middlewares section](../middlewares/overview.md).

### TLS Option

Additionally, to allow for the use of TLS options in an IngressRoute, we defined the CRD below for the TLSOption kind.
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
  minVersion: VersionTLS12

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

!!! important "References and namespaces"

    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the IngressRoute.

	Additionally, when the definition of the TLS option is from another provider,
the cross-provider syntax (`middlewarename@provider`) should be used to refer to the TLS option,
just as in the [middleware case](../middlewares/overview.md#provider-namespace).
Specifying a namespace attribute in this case would not make any sense, and will be ignored.

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
