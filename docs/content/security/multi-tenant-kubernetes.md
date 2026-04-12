---
title: "Traefik in Multi-Tenant Kubernetes Clusters"
description: "Traefik is not recommended for multi-tenant Kubernetes clusters due to TLS certificate management and broader isolation, traffic, and security concerns. Read the technical guidelines."
---

# Traefik in Multi-Tenant Kubernetes Clusters

Traefik is primarily designed as a cluster-wide ingress controller. For this reason, when using the Kubernetes `Ingress` or `IngressRoute` specifications, **it is not recommended to use a shared instance of Traefik in multi-tenant Kubernetes clusters**, where multiple teams or tenants share the same cluster.

The main reasons include:

* **Resource visibility and isolation**: Traefik requires cluster-level permissions and watches resources across namespaces. Misconfigurations in one tenant’s resources may affect others.
* **Shared CRDs**: Advanced configuration resources, like Middleware or TLSOptions, are cluster-scoped. Conflicting definitions can impact multiple tenants.
* **Traffic and availability risks**: Routing rules, middleware, or heavy traffic from one tenant can interfere with others, affecting reliability and performance.
* **Observability and privacy**: Logs, metrics, and traces are shared by default, which may expose sensitive information across tenants.

## TLS Certificates Management

At the core of this limitation is the TLS Store, which holds all the TLS certificates used by Traefik. 
As this Store is global in Traefik, it is shared across all namespaces, meaning any `Ingress` or `IngressRoute` in the cluster can potentially reference or affect TLS configurations intended for other tenants.

This lack of isolation poses a risk in multi-tenant environments where different teams or applications require strict boundaries between resources, especially around sensitive data like TLS certificates.

## Recommended Setup

The recommended approach for multi-tenant Kubernetes clusters is to **deploy one dedicated Traefik instance per tenant**, scoped to that tenant's namespace(s).
This ensures strict boundaries between tenants and prevents cross-tenant configuration leakage.

Each Traefik instance should be:

* Deployed in the tenant's namespace with a dedicated `ServiceAccount`
* Configured with `providers.kubernetesCRD.namespaces` and/or `providers.kubernetesIngress.namespaces` to restrict resource watch to that tenant's namespaces only
* Subject to Kubernetes RBAC that limits its access to only the resources it needs

This namespace-per-tenant topology, combined with Kubernetes `NetworkPolicy` resources, provides the strongest available isolation when running Traefik on a shared cluster.

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

## Gateway API

We also encourage users to adopt [Kubernetes Gateway API](../reference/install-configuration/providers/kubernetes/kubernetes-gateway.md), which is designed
with multi-tenancy as a first-class concern. Gateway API provides a more expressive and role-oriented model
for managing ingress, with built-in support for delegating route configuration to individual tenants while
preserving infrastructure-level control for cluster operators. This tooling handles multi-tenant scenarios
more robustly than the traditional Ingress model and is the recommended direction for new deployments.

For example, concerning TLS certificate management, the `Listener` resource allows administrators to explicitly
define which `Route` resources (e.g., `HTTPRoute`) are permitted to bind to which domain names or ports.
This enforces stricter ownership and isolation between tenants, making the Gateway API a safer and more robust
choice for multi-tenant use cases than the traditional Ingress model.

For new deployments in multi-tenant environments, adopting Gateway API is the recommended direction.
