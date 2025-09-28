---
title: "Traefik Providers Documentation"
description: "Learn about Providers in Traefik Proxy."
---

# Overview

_Providers_ are infrastructure components, whether orchestrators, container engines, cloud providers, or key-value stores.
The idea is that Traefik queries the provider APIs in order to find relevant information about routing,
and when Traefik detects a change, it dynamically updates the routes.

## Provider Categories

While each provider is different, you can think of each as belonging to one of four categories:

- Label-based: each deployed container has a set of labels attached to it
- Key-Value-based: each deployed container updates a key-value store with relevant information
- Annotation-based: a separate object, with annotations, defines the characteristics of the container
- File-based: uses files to define configuration

## Provider Namespace

When you declare certain objects in the Traefik dynamic configuration,
such as middleware, services, TLS options or server transports, they reside in their provider's namespace.
For example, if you declare a middleware using a Docker label, it resides in the Docker provider namespace.

If you use multiple providers and wish to reference such an object declared in another provider
(e.g. referencing a cross-provider object like middleware), then the object name should be suffixed by the `@`
separator, and the provider name.

For the list of the providers names, see the [supported providers](#supported-providers) table below.

```text
<resource-name>@<provider-name>
```

!!! important "Kubernetes Namespace vs Traefik Namespace"

    As Kubernetes also has its own notion of namespace,
    one should not confuse the _provider namespace_ with the _Kubernetes Namespace_ of a resource when in the context of cross-provider usage.

    In this case, since the definition of a Traefik dynamic configuration object is not in Kubernetes,
    specifying a Kubernetes Namespace when referring to the resource does not make any sense.

    On the other hand, if you were to declare a middleware as a Custom Resource in Kubernetes and use the non-CRD Ingress objects,
    you would have to add the Kubernetes Namespace of the middleware to the annotation like this `<middleware-namespace>-<middleware-name>@kubernetescrd`.

## Supported Providers

Below is the list of the currently supported providers in Traefik.

| Provider                                          | Type         | Configuration Type   | Provider Name       |
|--------------------------------------------------------------|--------------|----------------------|---------------------|
| <a id="Docker" href="#Docker" title="#Docker">[Docker](./docker.md)</a> | Orchestrator | Label                | `docker`            |
| <a id="Docker-Swarm" href="#Docker-Swarm" title="#Docker-Swarm">[Docker Swarm](./swarm.md)</a> | Orchestrator | Label                | `swarm`             |
| <a id="Kubernetes-IngressRoute" href="#Kubernetes-IngressRoute" title="#Kubernetes-IngressRoute">[Kubernetes IngressRoute](./kubernetes/kubernetes-crd.md)</a> | Orchestrator | Custom Resource      | `kubernetescrd`     |
| <a id="Kubernetes-Ingress" href="#Kubernetes-Ingress" title="#Kubernetes-Ingress">[Kubernetes Ingress](./kubernetes/kubernetes-ingress.md)</a> | Orchestrator | Ingress              | `kubernetes`        |
| <a id="Kubernetes-Gateway-API" href="#Kubernetes-Gateway-API" title="#Kubernetes-Gateway-API">[Kubernetes Gateway API](./kubernetes/kubernetes-gateway.md)</a> | Orchestrator | Gateway API Resource | `kubernetesgateway` |
| <a id="Consul-Catalog" href="#Consul-Catalog" title="#Consul-Catalog">[Consul Catalog](./hashicorp/consul-catalog.md)</a> | Orchestrator | Label                | `consulcatalog`     |
| <a id="Nomad" href="#Nomad" title="#Nomad">[Nomad](./hashicorp/nomad.md)</a> | Orchestrator | Label                | `nomad`             |
| <a id="ECS" href="#ECS" title="#ECS">[ECS](./others/ecs.md)</a> | Orchestrator | Label                | `ecs`               |
| <a id="File" href="#File" title="#File">[File](./others/file.md)</a> | Manual       | YAML/TOML format     | `file`              |
| <a id="Consul" href="#Consul" title="#Consul">[Consul](./hashicorp/consul.md)</a> | KV           | KV                   | `consul`            |
| <a id="Etcd" href="#Etcd" title="#Etcd">[Etcd](./kv/etcd.md)</a> | KV           | KV                   | `etcd`              |
| <a id="ZooKeeper" href="#ZooKeeper" title="#ZooKeeper">[ZooKeeper](./kv/zk.md)</a> | KV           | KV                   | `zookeeper`         |
| <a id="Redis" href="#Redis" title="#Redis">[Redis](./kv/redis.md)</a> | KV           | KV                   | `redis`             |
| <a id="HTTP" href="#HTTP" title="#HTTP">[HTTP](./others/http.md)</a> | Manual       | JSON/YAML format          | `http`              |

!!! info "More Providers"

    The current version of Traefik does not yet support every provider that Traefik v2.11 did.
    See the [previous version (v2.11)](https://doc.traefik.io/traefik/v2.11/) for more information.

## Referencing a Traefik Dynamic Configuration Object from Another Provider

Declaring the add-foo-prefix in the file provider.

```yaml tab="File (YAML)"
http:
  middlewares:
    add-foo-prefix:
      addPrefix:
        prefix: "/foo"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.add-foo-prefix.addPrefix]
    prefix = "/foo"
```

Using the add-foo-prefix middleware from other providers:

```yaml tab="Docker & Swarm"
your-container:
  image: your-docker-image

  labels:
    # Attach add-foo-prefix@file middleware (declared in file)
    - "traefik.http.routers.my-container.middlewares=add-foo-prefix@file"
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: ingressroutestripprefix

spec:
  entryPoints:
    - web
  routes:
    - match: Host(`example.com`)
      kind: Rule
      services:
        - name: whoami
          port: 80
      middlewares:
        - name: add-foo-prefix@file
        # namespace: bar
        # A namespace specification such as above is ignored
        # when the cross-provider syntax is used.
```

```yaml tab="Ingress"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress
  namespace: appspace
  annotations:
    "traefik.ingress.kubernetes.io/router.middlewares": add-foo-prefix@file
spec:
```

By default, Traefik creates routes for all detected containers.

If you want to limit the scope of the Traefik service discovery,
i.e. disallow route creation for some containers,
you can do so in two different ways:

- the generic configuration option `exposedByDefault`,
- a finer granularity mechanism based on constraints.

### `exposedByDefault` and `traefik.enable`

List of providers that support these features:

- [Docker](./docker.md#configuration-options)
- [ECS](./others/ecs.md#configuration-options)
- [Consul Catalog](./hashicorp/consul-catalog.md#configuration-options)
- [Nomad](./hashicorp/nomad.md#configuration-options)

### Constraints

List of providers that support constraints:

- [Docker](./docker.md#constraints)
- [ECS](./others/ecs.md#constraints)
- [Consul Catalog](./hashicorp/consul-catalog.md#constraints)
- [Nomad](./hashicorp/nomad.md#constraints)

{!traefik-for-business-applications.md!}
