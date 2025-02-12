---
title: "IngressRouteUDP"
description: "Understand the routing configuration for the Kubernetes IngressRouteUDP & Traefik CRD"
---

`IngressRouteUDP` is the CRD implementation of a [Traefik UDP router](../../udp/router/rules-priority.md).

Before creating `IngressRouteUDP` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `IngressRouteUDP` kind and other Traefik-specific resources.

## Configuration Examples

```yaml tab="IngressRouteUDP"
apiVersion: traefik.io/v1alpha1
kind: IngressRouteUDP
metadata:
  name: ingressrouteudpfoo
  namespace: apps
spec:
  entryPoints:
    - fooudp  # The entry point where Traefik listens for incoming traffic
  routes:
  - services:
    - name: foo # The name of the Kubernetes Service to route to
      port: 8080
      weight: 10
      nativeLB: true # Enables native load balancing between pods
      nodePortLB: true
```

## Configuration Options

| Field  |  Description | Default  | Required |
|------------------------------------|-----------------------------|-------------------------------------------|-----------------------|
|   `entryPoints`                     | List of entrypoints names  | | No |
|   ` routes `                        | List of routes  | | Yes |
| `routes[n].services`                | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions. See [here](#externalname-service) for `ExternalName Service` setup. | | No |
| `services[n].name`                  | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) | "" | Yes |
| `routes[n].services[n].port`        | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.| "" | Yes |
| `routes[n].services[n].weight`      | Defines the weight to apply to the server load balancing | "" | No |
| `routes[n].services[n].nativeLB`    | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. | false | No |
| `routes[n].services[n].nodePortLB`  | Controls, when creating the load-balancer, whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is NodePort. It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes. See [here](#nativelb) for more information. | false | No  |

### ExternalName Service

Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port. Accordingly, Traefik supports defining a port in two ways:

- only on IngressRouteUDP service
- on both sides, you'll be warned if the ports don't match, and the IngressRouteUDP service port is used

Thus, in case of two sides port definition, Traefik expects a match between port.
    
```yaml tab="IngressRouteUDP"
---
apiVersion: traefik.io/v1alpha1
kind: IngressRouteUDP
metadata:
  name: test.route
  namespace: default

spec:
  entryPoints:
    - foo

  routes:
  - services:
    - name: external-svc
      port: 80

---
apiVersion: v1
kind: Service
metadata:
  name: external-svc
  namespace: default
spec:
  externalName: external.domain
  type: ExternalName
```

```yaml tab="ExternalName Service"
---
apiVersion: traefik.io/v1alpha1
kind: IngressRouteUDP
metadata:
  name: test.route
  namespace: default

spec:
  entryPoints:
    - foo

  routes:
  - services:
    - name: external-svc

---
apiVersion: v1
kind: Service
metadata:
  name: external-svc
  namespace: default
spec:
  externalName: external.domain
  type: ExternalName
  ports:
    - port: 80
```

```yaml tab="Both sides"
---
apiVersion: traefik.io/v1alpha1
kind: IngressRouteUDP
metadata:
  name: test.route
  namespace: default

spec:
  entryPoints:
    - foo

  routes:
  - services:
    - name: external-svc
      port: 80

---
apiVersion: v1
kind: Service
metadata:
  name: external-svc
  namespace: default
spec:
  externalName: external.domain
  type: ExternalName
  ports:
    - port: 80
```

### NativeLB

To avoid creating the server load-balancer with the pods IPs and use Kubernetes Service `clusterIP` directly, one should set the TCP service `NativeLB` option to true. By default, `NativeLB` is false.

```yaml tab="IngressRouteUDP"
apiVersion: traefik.io/v1alpha1
kind: IngressRouteUDP
metadata:
  name: test.route
  namespace: default
spec:
  entryPoints:
    - foo
routes:
- services:
  - name: svc
    port: 80
    # Here, nativeLB instructs to build the servers load balancer with the Kubernetes Service clusterIP only.
    nativeLB: true
```

```yaml tab="Service"
apiVersion: v1
kind: Service
metadata:
  name: svc
  namespace: default
spec:
  type: ClusterIP
  ...
```
