---
title: "Traefik Kubernetes Middleware Documentation"
description: "Learn how to configure a Traefik Proxy Kubernetes Middleware to reach Services, which handle incoming requests. Read the technical documentation."
---

Middleware is the CRD implementation of a [Traefik middleware](../../http/middlewares/overview.md).

!!! tip "Cross-provider namespace"
    As Kubernetes also has its own notion of namespace, one should not confuse the kubernetes namespace of a resource (in the reference to the middleware) with the [provider namespace](../../../install-configuration/providers/overview.md#provider-namespace), when the definition of the middleware comes from another provider. In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored. Additionally, when you want to reference a Middleware from the CRD Provider, you have to append the namespace of the resource in the resource-name as Traefik appends the namespace internally automatically.

## Configuration Example

```yaml tab="Middleware"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: stripprefix
  namespace: foo

spec:
  stripPrefix:
    prefixes:
      - /stripit
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutebar

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`example.com`) && PathPrefix(`/stripit`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: stripprefix
      namespace: foo
```

More information about available middlewares in the dedicated [middlewares section](../../http/middlewares/overview.md).
