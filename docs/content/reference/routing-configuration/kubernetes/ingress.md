---
title: "Kubernetes Ingress Routing Configuration"
description: "Understand the routing configuration for the Kubernetes Ingress Controller and Traefik Proxy. Read the technical documentation."
---

# Traefik & Kubernetes with Ingress

## Routing Configuration

The Kubernetes Ingress provider watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn will create the resulting routers, services, handlers, etc.

## Configuration Example

```yaml tab="Ingress"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myingress
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: web

spec:
  rules:
    - host: example.com
      http:
        paths:
          - path: /bar
            pathType: Exact
            backend:
              service:
                name:  whoami
                port:
                  number: 80
          - path: /foo
            pathType: Exact
            backend:
              service:
                name:  whoami
                port:
                  number: 80
```

## Annotations

!!! warning "Referencing resources in annotations"
    
    In an annotation, when referencing a resource defined by another provider,
    the [provider namespace syntax](../../install-configuration/providers/overview.md#provider-namespace) must be used.

### On Ingress

| Annotation | Description | Value |
|------------|-------------|-------|
| <a id="opt-traefik-ingress-kubernetes-iorouter-entrypoints" href="#opt-traefik-ingress-kubernetes-iorouter-entrypoints" title="#opt-traefik-ingress-kubernetes-iorouter-entrypoints">`traefik.ingress.kubernetes.io/router.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-middlewares" href="#opt-traefik-ingress-kubernetes-iorouter-middlewares" title="#opt-traefik-ingress-kubernetes-iorouter-middlewares">`traefik.ingress.kubernetes.io/router.middlewares`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth@file,default-prefix@kubernetescrd` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-priority" href="#opt-traefik-ingress-kubernetes-iorouter-priority" title="#opt-traefik-ingress-kubernetes-iorouter-priority">`traefik.ingress.kubernetes.io/router.priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `"42"` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-rulesyntax" href="#opt-traefik-ingress-kubernetes-iorouter-rulesyntax" title="#opt-traefik-ingress-kubernetes-iorouter-rulesyntax">`traefik.ingress.kubernetes.io/router.rulesyntax`</a> | See [rule syntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>**Deprecated:** RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `"v2"` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-pathmatcher" href="#opt-traefik-ingress-kubernetes-iorouter-pathmatcher" title="#opt-traefik-ingress-kubernetes-iorouter-pathmatcher">`traefik.ingress.kubernetes.io/router.pathmatcher`</a> | Overrides the default router rule type used for a path.<br/>Only path-related matcher name should be specified: `Path`, `PathPrefix` or `PathRegexp`.<br/>Default: `PathPrefix` | `Path` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-tls" href="#opt-traefik-ingress-kubernetes-iorouter-tls" title="#opt-traefik-ingress-kubernetes-iorouter-tls">`traefik.ingress.kubernetes.io/router.tls`</a> | Enables TLS for the router. | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-tls-certresolver" href="#opt-traefik-ingress-kubernetes-iorouter-tls-certresolver" title="#opt-traefik-ingress-kubernetes-iorouter-tls-certresolver">`traefik.ingress.kubernetes.io/router.tls.certresolver`</a> | Specifies the certificate resolver to use for TLS certificates. | `myresolver` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-main" href="#opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-main" title="#opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-main">`traefik.ingress.kubernetes.io/router.tls.domains.n.main`</a> | Defines the main domain for TLS certificate (where n is the domain index). | `example.org` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-sans" href="#opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-sans" title="#opt-traefik-ingress-kubernetes-iorouter-tls-domains-n-sans">`traefik.ingress.kubernetes.io/router.tls.domains.n.sans`</a> | Defines the Subject Alternative Names (SANs) for TLS certificate (where n is the domain index). | `test.example.org,dev.example.org` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-tls-options" href="#opt-traefik-ingress-kubernetes-iorouter-tls-options" title="#opt-traefik-ingress-kubernetes-iorouter-tls-options">`traefik.ingress.kubernetes.io/router.tls.options`</a> | See [TLS options](../kubernetes/crd/tls/tlsoption.md) for more information. | `foobar@file` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-observability-accesslogs" href="#opt-traefik-ingress-kubernetes-iorouter-observability-accesslogs" title="#opt-traefik-ingress-kubernetes-iorouter-observability-accesslogs">`traefik.ingress.kubernetes.io/router.observability.accesslogs`</a> | Controls whether the router produces access logs.<br/>See [observability](../http/routing/observability.md) for more information. | `true` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-observability-metrics" href="#opt-traefik-ingress-kubernetes-iorouter-observability-metrics" title="#opt-traefik-ingress-kubernetes-iorouter-observability-metrics">`traefik.ingress.kubernetes.io/router.observability.metrics`</a> | Controls whether the router produces metrics.<br/>See [observability](../http/routing/observability.md) for more information. | `true` |
| <a id="opt-traefik-ingress-kubernetes-iorouter-observability-tracing" href="#opt-traefik-ingress-kubernetes-iorouter-observability-tracing" title="#opt-traefik-ingress-kubernetes-iorouter-observability-tracing">`traefik.ingress.kubernetes.io/router.observability.tracing`</a> | Controls whether the router produces traces.<br/>See [observability](../http/routing/observability.md) for more information. | `true` |

### On Service

| Annotation | Description | Value |
|------------|-------------|-------|
| <a id="opt-traefik-ingress-kubernetes-ioservice-nativelb" href="#opt-traefik-ingress-kubernetes-ioservice-nativelb" title="#opt-traefik-ingress-kubernetes-ioservice-nativelb">`traefik.ingress.kubernetes.io/service.nativelb`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.<br/>The Kubernetes Service itself does load-balance to the pods.<br/>Please note that, by default, Traefik reuses the established connections to the backends for performance purposes. This can prevent the requests load balancing between the replicas from behaving as one would expect when the option is set.<br/>Default: `false` | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-nodeportlb" href="#opt-traefik-ingress-kubernetes-ioservice-nodeportlb" title="#opt-traefik-ingress-kubernetes-ioservice-nodeportlb">`traefik.ingress.kubernetes.io/service.nodeportlb`</a> | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is NodePort.<br/>It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.<br/>Default: `false` | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-serversscheme" href="#opt-traefik-ingress-kubernetes-ioservice-serversscheme" title="#opt-traefik-ingress-kubernetes-ioservice-serversscheme">`traefik.ingress.kubernetes.io/service.serversscheme`</a> | Overrides the default scheme. | `h2c` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-serverstransport" href="#opt-traefik-ingress-kubernetes-ioservice-serverstransport" title="#opt-traefik-ingress-kubernetes-ioservice-serverstransport">`traefik.ingress.kubernetes.io/service.serverstransport`</a> | See [ServersTransport](../kubernetes/crd/http/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-passhostheader" href="#opt-traefik-ingress-kubernetes-ioservice-passhostheader" title="#opt-traefik-ingress-kubernetes-ioservice-passhostheader">`traefik.ingress.kubernetes.io/service.passhostheader`</a> | Controls whether to forward the Host header to the backend. | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie">`traefik.ingress.kubernetes.io/service.sticky.cookie`</a> | Enables sticky sessions using cookies.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-name" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-name" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-name">`traefik.ingress.kubernetes.io/service.sticky.cookie.name`</a> | Defines the cookie name for sticky sessions.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `foobar` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-secure" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-secure" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-secure">`traefik.ingress.kubernetes.io/service.sticky.cookie.secure`</a> | Sets the Secure flag on the sticky session cookie.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-samesite" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-samesite" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-samesite">`traefik.ingress.kubernetes.io/service.sticky.cookie.samesite`</a> | Sets the SameSite attribute on the sticky session cookie.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `"none"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-domain" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-domain" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-domain">`traefik.ingress.kubernetes.io/service.sticky.cookie.domain`</a> | Sets the Domain attribute on the sticky session cookie, defining the host to which the cookie will be sent.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `"foo.com"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-httponly" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-httponly" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-httponly">`traefik.ingress.kubernetes.io/service.sticky.cookie.httponly`</a> | Sets the HttpOnly flag on the sticky session cookie.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `"true"` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-maxage" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-maxage" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-maxage">`traefik.ingress.kubernetes.io/service.sticky.cookie.maxage`</a> | Sets the Max-Age attribute (in seconds) on the sticky session cookie.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `42` |
| <a id="opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-path" href="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-path" title="#opt-traefik-ingress-kubernetes-ioservice-sticky-cookie-path">`traefik.ingress.kubernetes.io/service.sticky.cookie.path`</a> | Sets the Path attribute on the sticky session cookie, defining the path that must exist in the requested URL.<br/>See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information. | `/foobar` |

## TLS

### Enabling TLS via HTTP Options on Entrypoint

TLS can be enabled through the [HTTP options](../../install-configuration/entrypoints.md) of an Entrypoint:

```bash tab="CLI"
# Static configuration
--entryPoints.websecure.address=:443
--entryPoints.websecure.http.tls
```

```yaml tab="File (YAML)"
# Static configuration
entryPoints:
  websecure:
    address: ':443'
    http:
      tls: {}
```

```toml tab="File (TOML)"
# Static configuration
[entryPoints.websecure]
  address = ":443"

    [entryPoints.websecure.http.tls]
```

This way, any Ingress attached to this Entrypoint will have TLS termination by default.

??? example "Configuring Kubernetes Ingress Controller with TLS on Entrypoint"

    ```yaml tab="RBAC"
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: traefik-ingress-controller
    rules:
      - apiGroups:
          - ""
        resources:
          - services
          - secrets
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - discovery.k8s.io
        resources:
          - endpointslices
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - extensions
          - networking.k8s.io
        resources:
          - ingresses
          - ingressclasses
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - extensions
          - networking.k8s.io
        resources:
          - ingresses/status
        verbs:
          - update

    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: traefik-ingress-controller
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: traefik-ingress-controller
    subjects:
      - kind: ServiceAccount
        name: traefik-ingress-controller
        namespace: default
    ```

    ```yaml tab="Ingress"
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: myingress
      annotations:
        traefik.ingress.kubernetes.io/router.entrypoints: websecure

    spec:
      rules:
        - host: example.com
          http:
            paths:
              - path: /bar
                pathType: Exact
                backend:
                  service:
                    name:  whoami
                    port:
                      number: 80
              - path: /foo
                pathType: Exact
                backend:
                  service:
                    name:  whoami
                    port:
                      number: 80
    ```

    ```yaml tab="Traefik"
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: traefik-ingress-controller

    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: traefik
      labels:
        app: traefik

    spec:
      replicas: 1
      selector:
        matchLabels:
          app: traefik
      template:
        metadata:
          labels:
            app: traefik
        spec:
          serviceAccountName: traefik-ingress-controller
          containers:
            - name: traefik
              image: traefik:v3.6
              args:
                - --entryPoints.websecure.address=:443
                - --entryPoints.websecure.http.tls
                - --providers.kubernetesingress
              ports:
                - name: websecure
                  containerPort: 443

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: traefik
    spec:
      type: LoadBalancer
      selector:
        app: traefik
      ports:
        - protocol: TCP
          port: 443
          name: websecure
          targetPort: 443
    ```

    ```yaml tab="Whoami"
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: whoami
      labels:
        app: traefiklabs
        name: whoami

    spec:
      replicas: 2
      selector:
        matchLabels:
          app: traefiklabs
          task: whoami
      template:
        metadata:
          labels:
            app: traefiklabs
            task: whoami
        spec:
          containers:
            - name: whoami
              image: traefik/whoami
              ports:
                - containerPort: 80

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: whoami

    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: whoami
    ```

### Enabling TLS via Annotations

To enable TLS on the underlying router created from an Ingress, one should configure it through annotations:

```yaml
traefik.ingress.kubernetes.io/router.tls: "true"
```

For more options, please refer to the available [annotations](#on-ingress).

??? example "Configuring Kubernetes Ingress Controller with TLS"

    ```yaml tab="Ingress"
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: myingress
      annotations:
        traefik.ingress.kubernetes.io/router.entrypoints: websecure
        traefik.ingress.kubernetes.io/router.tls: "true"

    spec:
      rules:
        - host: example.com
          http:
            paths:
              - path: /bar
                pathType: Exact
                backend:
                  service:
                    name:  whoami
                    port:
                      number: 80
              - path: /foo
                pathType: Exact
                backend:
                  service:
                    name:  whoami
                    port:
                      number: 80
    ```

### Certificates Management

??? example "Using a secret"

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
      # Only selects which certificate(s) should be loaded from the secret, in order to terminate TLS.
      # Doesn't enable TLS for that ingress (hence for the underlying router).
      # Please see the TLS annotations on ingress made for that purpose.
      tls:
      - secretName: supersecret
    ```

    ```yaml tab="Secret"
    apiVersion: v1
    kind: Secret
    metadata:
      name: supersecret

    data:
      tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
      tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
    ```

TLS certificates can be managed in Secrets objects.

!!! info

    Only TLS certificates provided by users can be stored in Kubernetes Secrets.
    [Let's Encrypt](../../install-configuration/tls/certificate-resolvers/acme.md) certificates cannot be managed in Kubernetes Secrets yet.

### Communication Between Traefik and Pods

!!! info "Routing directly to [Kubernetes services](https://kubernetes.io/docs/concepts/services-networking/service/ "Link to Kubernetes service docs")"

    To route directly to the Kubernetes service,
    one can use the `traefik.ingress.kubernetes.io/service.nativelb` annotation on the Kubernetes service.
    It controls, when creating the load-balancer,
    whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.

    One alternative is to use an `ExternalName` service to forward requests to the Kubernetes service through DNS.
    To do so, one must allow external name services.

Traefik automatically requests endpoint information based on the service provided in the ingress spec.
Although Traefik will connect directly to the endpoints (pods),
it still checks the service port to see if TLS communication is required.

There are 3 ways to configure Traefik to use HTTPS to communicate with pods:

1. If the service port defined in the ingress spec is `443` (note that you can still use `targetPort` to use a different port on your pod).
1. If the service port defined in the ingress spec has a name that starts with `https` (such as `https-api`, `https-web` or just `https`).
1. If the service spec includes the annotation `traefik.ingress.kubernetes.io/service.serversscheme: https`.

If either of those configuration options exist, then the backend communication protocol is assumed to be TLS,
and will connect via TLS automatically.

!!! info

    Please note that by enabling TLS communication between traefik and your pods,
    you will have to have trusted certificates that have the proper trust chain and IP subject name.
    If this is not an option, you may need to skip TLS certificate verification.
    See the [`insecureSkipVerify` TLSOption](../kubernetes/crd/tls/tlsoption.md) setting for more details.

## Global Default Backend Ingresses

Ingresses can be created that look like the following:

```yaml tab="Ingress"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
 name: cheese

spec:
  defaultBackend:
    service:
      name: stilton
      port:
        number: 80
```

This ingress follows the Global Default Backend property of ingresses.
This will allow users to create a "default router" that will match all unmatched requests.

!!! info

    Due to Traefik's use of priorities, you may have to set this ingress priority lower than other ingresses in your environment,
    to avoid this global ingress from satisfying requests that could match other ingresses.

    To do this, use the `traefik.ingress.kubernetes.io/router.priority` annotation (as seen in [Annotations on Ingress](#on-ingress)) on your ingresses accordingly.

{% include-markdown "includes/traefik-for-business-applications.md" %}
