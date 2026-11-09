---
title: "Multi-Tenancy in Kubernetes"
description: "Isolate tenants in a shared Kubernetes cluster with the Traefik Operator, which provisions a dedicated Traefik data plane for each Gateway API Gateway, and learn the limits of a single shared Traefik instance."
---

# Multi-Tenancy in Kubernetes

In a multi-tenant Kubernetes cluster, several teams or tenants share the same cluster.
How well Traefik Proxy isolates tenants from each other depends on how you deploy it:

| Deployment | Routing Configuration | Address | Multi-Tenancy |
|---|---|---|---|
| Traefik Operator with Kubernetes Gateway API | One per Gateway | One per Gateway | Recommended |
| Single Traefik instance with Kubernetes Gateway API | Shared by all Gateways | Shared by all Gateways | Not recommended |
| Single Traefik instance with Ingress or IngressRoute | Shared by all namespaces | Shared by all namespaces | Not recommended |

For multi-tenant clusters, use Kubernetes Gateway API with the [Traefik Operator](#traefik-operator), and create one Gateway per tenant.

## Gateway API

Kubernetes Gateway API treats each Gateway as an independent entry point to the cluster, with its own listeners, routes, and addresses.
A single Traefik instance doesn't keep Gateways separate.
To give each Gateway its own Traefik data plane, use the [Traefik Operator](#traefik-operator).

### Single Traefik Instance

When you run one Traefik instance with the [Kubernetes Gateway API provider](../reference/install-configuration/providers/kubernetes/kubernetes-gateway.md), that instance serves every Gateway it watches from one shared routing configuration:

* **Routes are not isolated per Gateway.** Routers built from routes attached to different Gateways share the same entry points, so a route attached to one tenant's Gateway can match requests intended for another tenant's Gateway.
* **Gateways share one address.** Every Gateway reports the address of the same Traefik instance in its `status.addresses`.
* **Gateways share one instance.** Heavy traffic or a misconfiguration on one Gateway affects every other Gateway served by that instance.

Use a single Traefik instance only when every Gateway it serves belongs to the same tenant.

### Traefik Operator

The [Traefik Operator](https://github.com/traefik/gateway-operator) watches Gateways and provisions a dedicated Traefik data plane for each one, in the Gateway's own namespace.
Each data plane is a separate Traefik instance, with its own `Deployment`, `Service`, and `ServiceAccount`, and it serves only the Gateway it was provisioned for:

* **Each Gateway has its own routing configuration.** A route attached to one tenant's Gateway can't match requests sent to another tenant's Gateway.
* **Each Gateway reports its own address.** The data plane reports the address of its own `Service` in the Gateway `status.addresses`, and honors the addresses requested in `spec.addresses`.
* **Each Gateway carries its own infrastructure metadata.** The labels and annotations set in the Gateway `spec.infrastructure` are added to the resources generated for its data plane.
* **Each data plane runs in the tenant's namespace.** Kubernetes deletes the data plane with its Gateway, and the data plane counts against the tenant's namespace quotas.

The Traefik Operator requires Traefik v3.8 or later as the data plane.
To install it, follow the Traefik Operator [Getting Started guide](https://github.com/traefik/gateway-operator/blob/main/docs/getting-started.md).
The Traefik Operator is validated against the Gateway API conformance suite. See the Traefik Operator [conformance report](https://github.com/traefik/gateway-operator/blob/main/docs/conformance.md) for the results.

Isolation applies per Gateway, so tenants that share a Gateway also share its data plane.
Create one Gateway per tenant, in the tenant's namespace.

Listeners add control inside each Gateway.
A cluster operator can use the listener `hostname` to limit the domain names a Gateway serves, and `allowedRoutes` to limit which namespaces can attach routes to it.
See [Gateway API routing configuration](../reference/routing-configuration/kubernetes/gateway-api.md) for the listener options Traefik supports.

## Ingress and IngressRoute

The Kubernetes Ingress and IngressRoute providers build one routing configuration for every namespace a Traefik instance watches, so tenants share the limits described in [Limits of a Shared Traefik Instance](#limits-of-a-shared-traefik-instance).

For multi-tenant clusters, migrate tenant routing to Kubernetes Gateway API and deploy it with the [Traefik Operator](#traefik-operator).
See the [Kubernetes Gateway API provider](../reference/install-configuration/providers/kubernetes/kubernetes-gateway.md) to configure it.

## Limits of a Shared Traefik Instance

A single Traefik instance shared by several tenants has these limits, whatever the provider:

* **Resource visibility and isolation**: Traefik requires cluster-level permissions and watches resources across namespaces. A misconfiguration in one tenant's resources can affect other tenants.
* **Traffic and availability**: Routing rules, middlewares, or heavy traffic from one tenant can affect the reliability and performance of other tenants.
* **Observability and privacy**: Logs, metrics, and traces are shared by default, which can expose sensitive information across tenants.

### TLS Store

Traefik holds all its TLS certificates in a TLS store that is global to the instance.
The TLS store is shared across all namespaces, so any Ingress, IngressRoute, or Gateway served by the instance can use or affect TLS configuration intended for another tenant.
This lack of isolation is a risk when tenants require strict boundaries around TLS certificates.
