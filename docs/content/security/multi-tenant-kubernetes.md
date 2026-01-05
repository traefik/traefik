---
title: "Traefik in Multi-Tenant Kubernetes Clusterss"
description: "Traefik is not recommended for multi-tenant Kubernetes clusters due to TLS certificate management and broader isolation, traffic, and security concerns. Read the technical guidelines."
---

# Traefik in Multi-Tenant Kubernetes Clusters

Traefik is primarily designed as a cluster-wide ingress controller. For this reason, when using the Kubernetes `Ingress` or `IngressRoute` specifications, **it is not recommended to use Traefik in multi-tenant Kubernetes clusters**, where multiple teams or tenants share the same cluster.

The main reasons include:

* **Resource visibility and isolation**: Traefik requires cluster-level permissions and watches resources across namespaces. Misconfigurations in one tenant’s resources may affect others.
* **Shared CRDs**: Advanced configuration resources, like Middleware or TLSOptions, are cluster-scoped. Conflicting definitions can impact multiple tenants.
* **Traffic and availability risks**: Routing rules, middleware, or heavy traffic from one tenant can interfere with others, affecting reliability and performance.
* **Observability and privacy**: Logs, metrics, and traces are shared by default, which may expose sensitive information across tenants.

## TLS Certificates Management

At the core of this limitation is the TLS Store, which holds all the TLS certificates used by Traefik. 
As this Store is global in Traefik, it is shared across all namespaces, meaning any `Ingress` or `IngressRoute` in the cluster can potentially reference or affect TLS configurations intended for other tenants.

This lack of isolation poses a risk in multi-tenant environments where different teams or applications require strict boundaries between resources, especially around sensitive data like TLS certificates.

In contrast, the [Kubernetes Gateway API](../providers/kubernetes-gateway.md) provides better primitives for secure multi-tenancy. 
Specifically, the `Listener` resource in the Gateway API allows administrators to explicitly define which Route resources (e.g., `HTTPRoute`) are permitted to bind to which domain names or ports. 
This capability enforces stricter ownership and isolation, making it a safer choice for multi-tenant use cases.

## Recommended setup

When strict boundaries are required between resources and teams, we recommend using one Traefik instance per tenant.

In Kubernetes one way to isolate a tenant is to restrict it to a namespace.
In that case, the namespace options from the Kubernetes [CRD](../providers/kubernetes-crd.md#namespaces) and [Ingress](../providers/kubernetes-ingress.md#namespaces) providers can be leveraged.  

!!! tip "Dedicate one Traefik instance per tenant using the Helm Chart" 

    ```yaml
    providers:
      kubernetesCRD:
        namespaces:
          - tenant
      kubernetesIngress:
        namespaces:
          - tenant
    ```
