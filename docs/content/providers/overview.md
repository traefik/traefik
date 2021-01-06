# Overview

Traefik's Many Friends
{: .subtitle }

![Providers](../assets/img/providers.png)

Configuration discovery in Traefik is achieved through _Providers_.

The _providers_ are existing infrastructure components, whether orchestrators, container engines, cloud providers, or key-value stores. 
The idea is that Traefik will query the providers' API in order to find relevant information about routing,
and each time Traefik detects a change, it dynamically updates the routes.

Deploy and forget is Traefik's credo.

## Orchestrators

Even if each provider is different, we can categorize them in four groups:

- Label based (each deployed container has a set of labels attached to it)
- Key-Value based (each deployed container updates a key-value store with relevant information)
- Annotation based (a separate object, with annotations, defines the characteristics of the container)
- File based (the good old configuration file)

## Provider Namespace

When you declare certain objects, in Traefik dynamic configuration,
such as middleware, service, TLS options or servers transport, they live in its provider's namespace.
For example, if you declare a middleware using a Docker label, under the hoods, it will reside in the docker provider namespace.

If you use multiple providers and wish to reference such an object declared in another provider 
(aka referencing a cross-provider object, e.g. middleware), then you'll have to append the `@` separator, 
followed by the provider name to the object name.

```text
<resource-name>@<provider-name>
```

!!! important "Kubernetes Namespace"

    As Kubernetes also has its own notion of namespace,
    one should not confuse the "provider namespace" with the "kubernetes namespace" of a resource when in the context of a cross-provider usage.  
    In this case, since the definition of a traefik dynamic configuration object is not in kubernetes,
    specifying a "kubernetes namespace" when referring to the resource does not make any sense,
    and therefore this specification would be ignored even if present.  
    On the other hand, if you, say, declare a middleware as a Custom Resource in Kubernetes and use the non-crd Ingress objects,
    you'll have to add the Kubernetes namespace of the middleware to the annotation like this `<middleware-namespace>-<middleware-name>@kubernetescrd`.

!!! abstract "Referencing a Traefik dynamic configuration object from Another Provider"

    Declaring the add-foo-prefix in the file provider.

    ```toml tab="File (TOML)"
    [http.middlewares]
      [http.middlewares.add-foo-prefix.addPrefix]
        prefix = "/foo"
    ```
    
    ```yaml tab="File (YAML)"
    http:
      middlewares:
        add-foo-prefix:
          addPrefix:
            prefix: "/foo"
    ```

    Using the add-foo-prefix middleware from other providers:

    ```yaml tab="Docker"
    your-container: #
      image: your-docker-image

      labels:
        # Attach add-foo-prefix@file middleware (declared in file)
        - "traefik.http.routers.my-container.middlewares=add-foo-prefix@file"
    ```

    ```yaml tab="Kubernetes Ingress Route"
    apiVersion: traefik.containo.us/v1alpha1
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
    
    ```yaml tab="Kubernetes Ingress"
    apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: stripprefix
      namespace: appspace
    spec:
      stripPrefix:
        prefixes:
          - /stripit
    
    ---
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: ingress
      namespace: appspace
      annotations:
        # referencing a middleware from Kubernetes CRD provider: 
        # <middleware-namespace>-<middleware-name>@kubernetescrd
        "traefik.ingress.kubernetes.io/router.middlewares": appspace-stripprefix@kubernetescrd
    spec:
      # ... regular ingress definition
    ```

## Supported Providers 

Below is the list of the currently supported providers in Traefik. 

| Provider                              | Type         | Configuration Type         |
|---------------------------------------|--------------|----------------------------|
| [Docker](./docker.md)                 | Orchestrator | Label                      |
| [Kubernetes](./kubernetes-crd.md)     | Orchestrator | Custom Resource or Ingress |
| [Consul Catalog](./consul-catalog.md) | Orchestrator | Label                      |
| [ECS](./ecs.md)                       | Orchestrator | Label                      |
| [Marathon](./marathon.md)             | Orchestrator | Label                      |
| [Rancher](./rancher.md)               | Orchestrator | Label                      |
| [File](./file.md)                     | Manual       | TOML/YAML format           |
| [Consul](./consul.md)                 | KV           | KV                         |
| [Etcd](./etcd.md)                     | KV           | KV                         |
| [ZooKeeper](./zookeeper.md)           | KV           | KV                         |
| [Redis](./redis.md)                   | KV           | KV                         |
| [HTTP](./http.md)                     | Manual       | JSON format                |

!!! info "More Providers"

    The current version of Traefik doesn't support (yet) every provider.
    See the [previous version (v1.7)](https://doc.traefik.io/traefik/v1.7/) for more providers.

### Configuration reload frequency

In some cases, some providers might undergo a sudden burst of changes,
which would generate a lot of configuration change events.
If Traefik took them all into account,
that would trigger a lot more configuration reloads than what is necessary,
or even useful.

In order to mitigate that, the `providers.providersThrottleDuration` option can be set.
It is the duration that Traefik waits for, after a configuration reload,
before taking into account any new configuration refresh event.
If any event arrives during that duration, only the most recent one is taken into account,
and all the previous others are dropped.

This option cannot be set per provider,
but the throttling algorithm applies independently to each of them.
It defaults to 2 seconds.

```toml tab="File (TOML)"
[providers]
  providers.providersThrottleDuration = 10s
```

```yaml tab="File (YAML)"
providers:
  providersThrottleDuration: 10s
```

```bash tab="CLI"
--providers.providersThrottleDuration=10s
```

<!--
TODO (document TCP VS HTTP dynamic configuration)
-->

## Restrict the Scope of Service Discovery

By default Traefik will create routes for all detected containers.

If you want to limit the scope of Traefik's service discovery,
i.e. disallow route creation for some containers,
you can do so in two different ways:
either with the generic configuration option `exposedByDefault`,
or with a finer granularity mechanism based on constraints.

### `exposedByDefault` and `traefik.enable`

List of providers that support that feature:

- [Docker](./docker.md#exposedbydefault)
- [Consul Catalog](./consul-catalog.md#exposedbydefault)
- [Rancher](./rancher.md#exposedbydefault)
- [Marathon](./marathon.md#exposedbydefault)

### Constraints

List of providers that support constraints:

- [Docker](./docker.md#constraints)
- [Consul Catalog](./consul-catalog.md#constraints)
- [Rancher](./rancher.md#constraints)
- [Marathon](./marathon.md#constraints)
- [Kubernetes CRD](./kubernetes-crd.md#labelselector)
- [Kubernetes Ingress](./kubernetes-ingress.md#labelselector)
