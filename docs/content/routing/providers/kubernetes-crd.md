# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

## Resource Configuration

If you're in a hurry, maybe you'd rather go through the [dynamic configuration](../../reference/dynamic-configuration/kubernetes-crd.md) reference.

### Traefik IngressRoute definition

```yaml
--8<-- "content/routing/providers/crd_ingress_route.yml"
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
    # (optional) Priority disambiguates rules of the same length, for route matching.
    priority: 12
    services:
    - name: whoami
      port: 80
      # (default 1) A weight used by the weighted round-robin strategy (WRR).  
      weight: 1
      # (default true) PassHostHeader controls whether to leave the request's Host
      # Header as it was before it reached the proxy, or whether to let the proxy set it
      # to the destination (backend) host.
      passHostHeader: true
      responseForwarding:
        # (default 100ms) Interval between flushes of the buffered response body to the client.
        flushInterval: 100ms

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
--8<-- "content/routing/providers/crd_middlewares.yml"
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
(in the reference to the middleware) with the [provider namespace](../../middlewares/overview.md#provider-namespace),
when the definition of the middleware is from another provider.
In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/overview.md).

### TLS Option

Additionally, to allow for the use of TLS options in an IngressRoute, we defined the CRD below for the TLSOption kind.
More information about TLS Options is available in the dedicated [TLS Configuration Options](../../../https/tls/#tls-options).

```yaml
--8<-- "content/routing/providers/crd_tls_option.yml"
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
just as in the [middleware case](../../middlewares/overview.md#provider-namespace).
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

Also see the [full example](../../user-guides/crd-acme/index.md) with Let's Encrypt.
