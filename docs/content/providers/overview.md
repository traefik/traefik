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

## Supported Providers 

Below is the list of the currently supported providers in Traefik. 

| Provider                          | Type         | Configuration Type |
|-----------------------------------|--------------|--------------------|
| [Docker](./docker.md)             | Orchestrator | Label              |
| [Kubernetes](./kubernetes-crd.md) | Orchestrator | Custom Resource    |
| [Marathon](./marathon.md)         | Orchestrator | Label              |
| [Rancher](./rancher.md)           | Orchestrator | Label              |
| [File](./file.md)                 | Manual       | TOML format        |

!!! note "More Providers"

    The current version of Traefik is in development and doesn't support (yet) every provider.
    See the previous version (1.7) for more providers.

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
- [Rancher](./rancher.md#exposedbydefault)
- [Marathon](./marathon.md#exposedbydefault)

### Constraints

List of providers that support constraints:

- [Docker](./docker.md#constraints)
- [Rancher](./rancher.md#constraints)
- [Marathon](./marathon.md#constraints)
- [Kubernetes CRD](./kubernetes-crd.md#labelselector)
