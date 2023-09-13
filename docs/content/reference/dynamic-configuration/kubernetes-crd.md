---
title: "Traefik Kubernetes CRD Documentation"
description: "Learn about the definitions, resources, and RBAC of dynamic configuration with Kubernetes CRD in Traefik Proxy. Read the technical documentation."
---

# Kubernetes Configuration Reference

Dynamic configuration with Kubernetes Custom Resource
{: .subtitle }

!!! warning "Deprecated apiextensions.k8s.io/v1beta1 CRD"

    The `apiextensions.k8s.io/v1beta1` CustomResourceDefinition is deprecated in Kubernetes `v1.16+` and will be removed in `v1.22+`.

    For Kubernetes `v1.16+`, please use the Traefik `apiextensions.k8s.io/v1` CRDs instead.

## Definitions

```yaml tab="apiextensions.k8s.io/v1 (Kubernetes v1.16+)"
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml"
```

```yaml tab="apiextensions.k8s.io/v1beta1 (Deprecated)"
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition-v1beta1.yml"
```

## Resources

```yaml
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-resource.yml"
```

## RBAC

```yaml
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-rbac.yml"
```

{!traefik-for-business-applications.md!}
