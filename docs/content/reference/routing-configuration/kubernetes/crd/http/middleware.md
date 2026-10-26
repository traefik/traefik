---
title: "Traefik Kubernetes Middleware Documentation"
description: "Learn how to configure a Traefik Proxy Kubernetes Middleware to reach Services, which handle incoming requests. Read the technical documentation."
---

`Middleware` is the CRD implementation of a [Traefik middleware](../../../http/middlewares/overview.md).

Before creating `Middleware` objects, you need to apply the [Traefik Kubernetes CRDs](../../../../install-configuration/providers/kubernetes/kubernetes-crd.md#requirements) to your Kubernetes cluster.

This registers the `Middleware` kind and other Traefik-specific resources.

!!! tip "Cross-provider namespace"
    As Kubernetes also has its own notion of namespace, one should not confuse the Kubernetes namespace of a resource (in the reference to the middleware) with the [provider namespace](../../../../install-configuration/providers/overview.md#provider-namespace), when the definition of the middleware comes from another provider. In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored. Additionally, when you want to reference a Middleware from the CRD Provider, you have to append the namespace of the resource in the resource-name as Traefik appends the namespace internally automatically.

!!! note "Cross-Namespace References"
    In the example below, the middleware is defined in the `foo` namespace while being referenced from an IngressRoute in another namespace. To enable such cross-namespace references, the `allowCrossNamespace` option must be enabled in the Traefik [Kubernetes CRD provider](../../../../install-configuration/providers/kubernetes/kubernetes-crd.md#configuration-options) configuration. If you prefer to avoid this requirement, you can define and reference the Middleware within the same namespace.

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

For more information about the available middlewares, navigate to the dedicated [middlewares overview section](../../../http/middlewares/overview.md).

## Referencing a Secret or a ConfigMap in a plugin configuration

A value inside `spec.plugin` can be a reference to a Kubernetes object instead of a literal.
Traefik resolves it while building the dynamic configuration, at any depth, in strings, lists and maps.

| Reference | Reads |
|---|---|
| <a id="opt-urnk8ssecretnamekey" href="#opt-urnk8ssecretnamekey" title="#opt-urnk8ssecretnamekey">`urn:k8s:secret:[name]:[key]`</a> | the key of that `Secret` |
| <a id="opt-urnk8sconfigmapnamekey" href="#opt-urnk8sconfigmapnamekey" title="#opt-urnk8sconfigmapnamekey">`urn:k8s:configmap:[name]:[key]`</a> | the key of that `ConfigMap`, from `data` or `binaryData` |

Both are read from the namespace of the `Middleware` itself.
A reference has no namespace, so a `Middleware` cannot read a `Secret` or a `ConfigMap` from another namespace.

Use a `Secret` for a credential, and a `ConfigMap` for a value that is not one and is too long to
sit in the manifest, such as a block of text the plugin hands to its clients.

The two references are not shown the same way in the API, the dashboard and the support dump:

- A `Secret` reference stays in the routing configuration, and Traefik reads its value only when it builds the middleware, so the value is never shown.
- A `ConfigMap` reference is replaced with its value in the routing configuration, so the value is shown.

!!! warning "Unresolved references"
    If the referenced object does not exist, does not carry the key, or if the reference is malformed,
    the whole `Middleware` is invalid and is left out of the routing configuration. Routers referencing it are not served.

```yaml tab="Middleware"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: my-plugin
  namespace: foo

spec:
  plugin:
    myplugin:
      token: urn:k8s:secret:my-plugin-credentials:token
      instructions: urn:k8s:configmap:my-plugin-instructions:text
```

```yaml tab="Secret"
apiVersion: v1
kind: Secret
metadata:
  name: my-plugin-credentials
  namespace: foo

data:
  token: dG9rZW4=
```

```yaml tab="ConfigMap"
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-plugin-instructions
  namespace: foo

data:
  text: |-
    Any text the plugin needs and a manifest should not carry.
```
