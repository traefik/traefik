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

??? info "`traefik.ingress.kubernetes.io/router.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.entrypoints: ep1,ep2
    ```

??? info "`traefik.ingress.kubernetes.io/router.middlewares`"

    See [middlewares overview](../http/middlewares/overview.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.middlewares: auth@file,default-prefix@kubernetescrd
    ```

??? info "`traefik.ingress.kubernetes.io/router.priority`"

    See [priority](../http/router/rules-and-priority.md#priority-calculation) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.priority: "42"
    ```

??? info "`traefik.ingress.kubernetes.io/router.rulesyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    See [rule syntax](../http/router/rules-and-priority.md#rulesyntax) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.rulesyntax: "v2"
    ```

??? info "`traefik.ingress.kubernetes.io/router.pathmatcher`"

    Overrides the default router rule type used for a path.
    Only path-related matcher name should be specified: `Path`, `PathPrefix` or `PathRegexp`.

    Default `PathPrefix`

    ```yaml
    traefik.ingress.kubernetes.io/router.pathmatcher: Path
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls`"

    <!-- See [tls](../http/tls/tls-certificates.md) for more information. -->

    ```yaml
    traefik.ingress.kubernetes.io/router.tls: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.certresolver`"

    <!-- See [certResolver](.) for more information. -->

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.certresolver: myresolver
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.domains.n.main`"

    <!-- See [domains](.) for more information. -->

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.domains.0.main: example.org
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.domains.n.sans`"

    <!-- See [domains](.) for more information. -->

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.domains.0.sans: test.example.org,dev.example.org
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.options`"

    See [options](../kubernetes/crd/http/tlsoption.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.options: foobar@file
    ```

??? info "`traefik.ingress.kubernetes.io/router.observability.accesslogs`"

    See [here](../http/router/observability.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.observability.accesslogs: true
    ```

??? info "`traefik.ingress.kubernetes.io/router.observability.metrics`"

    See [here](../http/router/observability.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.observability.metrics: true
    ```

??? info "`traefik.ingress.kubernetes.io/router.observability.tracing`"

    See [here](../http/router/observability.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.observability.tracing: true
    ```

### On Service

??? info "`traefik.ingress.kubernetes.io/service.nativelb`"

    Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.
    The Kubernetes Service itself does load-balance to the pods.
    Please note that, by default, Traefik reuses the established connections to the backends for performance purposes. This can prevent the requests load balancing between the replicas from behaving as one would expect when the option is set.
    By default, NativeLB is false.

    ```yaml
    traefik.ingress.kubernetes.io/service.nativelb: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.nodeportlb`"

    Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is NodePort.
    It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.
    By default, NodePortLB is false.

    ```yaml
    traefik.ingress.kubernetes.io/service.nodeportlb: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.serversscheme`"

    Overrides the default scheme.

    ```yaml
    traefik.ingress.kubernetes.io/service.serversscheme: h2c
    ```

??? info "`traefik.ingress.kubernetes.io/service.serverstransport`"

    See [ServersTransport](../kubernetes/crd/http/serverstransport.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.serverstransport: foobar@file
    ```

??? info "`traefik.ingress.kubernetes.io/service.passhostheader`"

    ```yaml
    traefik.ingress.kubernetes.io/service.passhostheader: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.name`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.name: foobar
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.secure`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.secure: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.samesite`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.samesite: "none"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.httponly`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.httponly: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.maxage`"

    See [sticky sessions](../kubernetes/crd/http/traefikservice.md#stickiness-on-multiple-levels) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.maxage: 42
    ```

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
              image: traefik:v3.5
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
    See the [`insecureSkipVerify` TLSOption](../kubernetes/crd/http/tlsoption.md) setting for more details.

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

{!traefik-for-business-applications.md!}
