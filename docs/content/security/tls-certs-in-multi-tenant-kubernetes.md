---
title: "TLS Certificates in Multi‑Tenant Kubernetes"
description: "Isolate TLS certificates in multi‑tenant clusters by keeping Secrets and routes in the same namespace and disabling cross‑namespace look‑ups in Traefik. Read the technical guidelines."
---

# TLS Certificates in Multi‑Tenant Kubernetes

In a shared cluster, different teams can create `Ingress` or `IngressRoute` objects that Traefik consumes.  
If one team can reference a TLS `Secret` from another namespace, it can serve traffic for someone else’s domain or read private keys.

**Rule of thumb:**  
Keep each certificate and the routes that use it in the same namespace, prevent Traefik from crossing that boundary, and store wildcard certificates in a [`TLSStore`](../routing/providers/kubernetes-crd.md#kind-tlsstore) that lives with Traefik, not with tenants.

## Why this matters

* An `Ingress` can reference any `Secret` of type `kubernetes.io/tls`.  
* By default, Traefik blocks cross‑namespace references (`allowCrossNamespace` is `false`).  
  See the [Kubernetes Ingress provider docs](https://doc.traefik.io/traefik/reference/install-configuration/providers/kubernetes/kubernetes-ingress/).  
* Turning that flag on, running one global Traefik instance, or giving tenants wide Secret rights breaks isolation.
* Traefik stores all loaded certs together. During the TLS handshake it does not yet know which `Ingress` or `IngressRoute` will handle the request, so it might serve a cert from the wrong namespace.

---

## Recommended setup

| Goal | What to do |
|------|------------|
| **One tenant, one namespace** | Put every workload, `Ingress`, and TLS `Secret` for a tenant in its own namespace. |
| **Stop cross‑namespace look‑ups (default)** | Leave `allowCrossNamespace` at `false`. |
| **Scope Traefik to one tenant** | Run a dedicated Traefik per namespace **or** start Traefik with:<br/>`--providers.kubernetescrd.namespaces=<tenant-ns>`<br/>`--providers.kubernetesingress.namespaces=<tenant-ns>` |
| **Tenant‑managed certs** | Give each tenant a namespaced cert‑manager `Issuer` instead of a cluster‑wide `ClusterIssuer`. |
| **Restrict RBAC** | Allow tenants to work with Secrets only in their namespace. No wildcard `*` verbs. |
| **One certificate, one domain** | Match each cert’s SANs to the domains in the referencing `Ingress` or `IngressRoute`. |
| **Use wildcard certs safely** | Store wildcard certs in Traefik’s namespace and attach them through a `TLSStore`. |

---

## Example — dedicated Traefik per tenant

```yaml
providers:
  kubernetesCRD:
    namespaces:
      - tenant-a
  kubernetesIngress:
    namespaces:
      - tenant-a
    allowCrossNamespace: false
```

## Example — override the default TLSStore with a wildcard certificate

```yaml
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  namespace: traefik
spec:
  certificates:
    secretName: my-wildcard-tls
```

!!! important "Default TLS Store"

    Traefik uses only one [TLS Store](../../https/tls.md#certificates-stores) named **"default"**.
    Place it in a namespace that Traefik can watch.
    Because Traefik picks it up automatically, you never need to reference it in an [`IngressRoute`](../routing/providers/kubernetes-crd.md#kind-ingressroute) or [`IngressRouteTCP`](../routing/providers/kubernetes-crd.md#kind-ingressroutetcp) objects.
    You cannot have two stores named **default** in different namespaces.
    