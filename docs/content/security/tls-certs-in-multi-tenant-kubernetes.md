---
title: "TLS Certificates in Multi-Tenant Kubernetes"
description: "Isolate TLS certificates in multi-tenant Kubernetes clusters by keeping Secrets and routes in the same namespace and disabling cross-namespace look-ups in Traefik. Read the technical guidelines."
---

# TLS Certificates in Multi-Tenant Kubernetes

In a shared (multi-tenant) cluster, different teams can create `Ingress` or `IngressRoute` objects that Traefik consumes. If one tenant can reference TLS secrets owned by another, they could
serve traffic for someone else’s domain or read private keys.

The **short rule**: keep certificate material and the routes that use it inside the same namespace and stop Traefik from crossing that boundary.

## Why this matters

* The Ingress object lets you point to any Secret that has `type: kubernetes.io/tls`.  
* By default Traefik will _not_ follow cross-namespace references — `allowCrossNamespace` is `false`. 
See the [Kubernetes Ingress Provider Documentation](https://doc.traefik.io/traefik/reference/install-configuration/providers/kubernetes/kubernetes-ingress/).  
* Turning that flag on, running a single global Traefik instance, or giving tenants broad Secret rights breaks isolation.

## Recommended setup

| Goal | What to do |
|------|------------|
| **One tenant, one namespace** | Put every workload, Ingress, and TLS Secret for a tenant in its own namespace. |
| **Stop cross-namespace look-ups (default)** | Leave `allowCrossNamespace` at `false`. |
| **Scope Traefik to that tenant only** | Run a dedicated Traefik instance per namespace **or** start Traefik with:<br/>`--providers.kubernetescrd.namespaces=<tenant-ns>`<br/>`--providers.kubernetesingress.namespaces=<tenant-ns>` |
| **Let tenants issue their own certs safely** | Give each tenant a namespaced `Issuer` (cert-manager) instead of a cluster-wide `ClusterIssuer`. |
| **Lock down RBAC** | Grant tenants `get/create/update` on Secrets only in their namespace. No wildcard `*` verbs. |

---

## Example — dedicated Traefik per tenant

```yaml
# Helm values
providers:
  kubernetesCRD:
    namespaces:
      - tenant-a
  kubernetesIngress:
    namespaces:
      - tenant-a
    allowCrossNamespace: false
```
