---
title: "Traefik Kubernetes Routing"
description: "Reference the dynamic configuration with the Kubernetes Gateway provider in Traefik Proxy. Read the technical documentation."
---

# Kubernetes Configuration Reference

Dynamic configuration with Kubernetes Gateway provider.
{: .subtitle }

## Definitions

```yaml
--8<-- "content/reference/dynamic-configuration/gateway.networking.k8s.io_gatewayclasses.yaml"
--8<-- "content/reference/dynamic-configuration/gateway.networking.k8s.io_gateways.yaml"
--8<-- "content/reference/dynamic-configuration/gateway.networking.k8s.io_httproutes.yaml"
--8<-- "content/reference/dynamic-configuration/gateway.networking.k8s.io_tcproutes.yaml"
--8<-- "content/reference/dynamic-configuration/gateway.networking.k8s.io_tlsroutes.yaml"
```

## Resources

```yaml
--8<-- "content/reference/dynamic-configuration/kubernetes-gateway-resource.yml"
```

## RBAC

```yaml
--8<-- "content/reference/dynamic-configuration/kubernetes-gateway-rbac.yml"
```

{!traefik-for-business-applications.md!}
