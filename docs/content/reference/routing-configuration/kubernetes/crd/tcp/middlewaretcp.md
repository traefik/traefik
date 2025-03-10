---
title: "Kubernetes MiddlewareTCP"
description: "Learn how to configure a Traefik Proxy Kubernetes Middleware to reach TCP Services, which handle incoming requests. Read the technical documentation."
---

`MiddlewareTCP` is the CRD implementation of a [Traefik TCP middleware](../../../tcp/middlewares/overview.md).

Before creating `MiddlewareTCP` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `MiddlewareTCP` kind and other Traefik-specific resources.

!!! tip "Cross-provider namespace"
    As Kubernetes also has its own notion of namespace, one should not confuse the kubernetes namespace of a resource (in the reference to the middleware) with the [provider namespace](../../../../install-configuration/providers/overview.md#provider-namespace), when the definition of the middleware comes from another provider. In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored. Additionally, when you want to reference a Middleware from the CRD Provider, you have to append the namespace of the resource in the resource-name as Traefik appends the namespace internally automatically.

## Configuration Example

```yaml tab="MiddlewareTCP"
apiVersion: traefik.io/v1alpha1
kind: MiddlewareTCP
metadata:
  name: ipallowlist
spec:
  ipAllowList:
    sourceRange:
      - 127.0.0.1/32
      - 192.168.1.7
```

```yaml tab="IngressRouteTCP"
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: ingressroutebar

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`example.com`) && PathPrefix(`/allowlist`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: ipallowlist
      namespace: foo
```

More information about available TCP middlewares in the dedicated [middlewares section](../../../tcp/middlewares/overview.md).
