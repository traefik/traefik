---
title: "TLS Certificates in Multi‑Tenant Kubernetes"
description: "Isolate TLS certificates in multi‑tenant clusters by keeping Secrets and routes in the same namespace and disabling cross‑namespace look‑ups in Traefik. Read the technical guidelines."
---

# TLS Certificates in Multi‑Tenant Kubernetes

In a shared cluster, different teams can create `Ingress` or `IngressRoute` objects that Traefik consumes.

Traefik does not support multi-tenancy when using the Kubernetes `Ingress` or `IngressRoute` specifications due to the way TLS certificate management is handled.

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
