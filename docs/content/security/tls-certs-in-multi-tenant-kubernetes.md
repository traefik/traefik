---
title: "TLS Certificates in Multi-Tenant Kubernetes"
description: "Isolate TLS certificates in multi-tenant Kubernetes clusters by keeping Secrets and routes in the same namespace and disabling cross-namespace look-ups in Traefik. Read the technical guidelines."
---

# TLS Certificates in Multi-Tenant Kubernetes

In a shared (multi-tenant) cluster, different teams can create `Ingress` or `IngressRoute` objects that Traefik consumes. If one tenant references TLS secrets, every other tenant can serve the certificate stored in the secret.
serve traffic for someone else's domain or read private keys.

The **short rule**: keep certificate material and the routes that use it inside the same namespace, stop Traefik from crossing that boundary, do not attach wildcard certificates to `Ingress` or `IngressRoute` but directly to Traefik using a [`TLSStore`](../routing/providers/kubernetes-crd.md#kind-tlsstore) object.

## Why this matters

* The Ingress object lets you point to any Secret that has `type: kubernetes.io/tls`.
* By default Traefik will _not_ follow cross-namespace references — `allowCrossNamespace` is `false`. 
See the [Kubernetes Ingress Provider Documentation](https://doc.traefik.io/traefik/reference/install-configuration/providers/kubernetes/kubernetes-ingress/).  
* Turning that flag on, running a single global Traefik instance, or giving tenants broad Secret rights breaks isolation.

* Even if the certificates are attached to `Ingress` or `IngressRoute`, Traefik stores them all together.
* When a HTTPS request reaches Traefik, the TLS handshake is the very first operation Traefik does. At this moment of the reverse proxy pipeline, Traefik does not know the `Ingress` or `IngressRoute` to use for the routing. For this reason, Traefik can serve a certificate contained in a different namespace than the `Ingress` or `IngressRoute` to expose.
* Using TLS certificates that only 

## Recommended setup

| Goal | What to do |
|------|------------|
| **One tenant, one namespace** | Put every workload, Ingress, and TLS Secret for a tenant in its own namespace. |
| **Stop cross-namespace look-ups (default)** | Leave `allowCrossNamespace` at `false`. |
| **Scope Traefik to that tenant only** | Run a dedicated Traefik instance per namespace **or** start Traefik with:<br/>`--providers.kubernetescrd.namespaces=<tenant-ns>`<br/>`--providers.kubernetesingress.namespaces=<tenant-ns>` |
| **Let tenants issue their own certs safely** | Give each tenant a namespaced `Issuer` (cert-manager) instead of a cluster-wide `ClusterIssuer`. |
| **Lock down RBAC** | Grant tenants `get/create/update` on Secrets only in their namespace. No wildcard `*` verbs. |
| **One certificate, one domain** | Limit the domain checked by a certificate to the domain matched by the `Ingress` or `IngressRoute` that references it. |
| **Attach wildcard certificates to TLSStore** | Store the wildcard certificates in the same namespace as Traefik and attach them to [`TLSStore`](../routing/providers/kubernetes-crd.md#kind-tlsstore) objects to allow serving them with no tenant issue. |

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

## Example - Override the default TLSStore with a wildcard TLS certificate

```yaml
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  namespace: traefik
spec:
  - certificates:
      secretName: my-wildcard-tls
```

!!! important "Default TLS Store"

    Traefik currently only uses the [TLS Store](../../https/tls.md#certificates-stores) named **"default"**.
    This _default_ `TLSStore` should be in a namespace discoverable by Traefik. Since it is used by default on [`IngressRoute`](#kind-ingressroute) and [`IngressRouteTCP`](#kind-ingressroutetcp) objects, there never is a need to actually reference it.
    This means that you cannot have two stores that are named default in different Kubernetes namespaces.
    As a consequence, with respect to TLS stores, the only change that makes sense (and only if needed) is to configure the default TLSStore.
    