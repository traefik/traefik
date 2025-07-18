---
title: "Traefik Kubernetes CRD Documentation"
description: "Learn about the definitions, resources, and RBAC of dynamic configuration with Kubernetes CRD in Traefik Proxy. Read the technical documentation."
---

# Kubernetes Configuration Reference

Dynamic configuration with Kubernetes Custom Resource
{: .subtitle }

## Definitions

```yaml tab="apiextensions.k8s.io/v1 (Kubernetes v1.16+)"
--8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml"
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
