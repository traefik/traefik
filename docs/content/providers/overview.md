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

| Provider                        | Type         | Configuration Type |
|---------------------------------|--------------|--------------------|
| [Docker](./docker.md)           | Orchestrator | Label              |
| [File](./file.md)               | Orchestrator | Custom Annotation  |
| [Kubernetes](kubernetes-crd.md) | Orchestrator | Custom Resource    |
| [Marathon](marathon.md)         | Orchestrator | Label              |

!!! note "More Providers"

    The current version of Traefik is in development and doesn't support (yet) every provider. See the previous version (1.7) for more providers.

<!--
 TODO (document TCP VS HTTP dynamic configuration)
 -->
    
## Constraints Configuration

If you want to limit the scope of Traefik's service discovery, you can set constraints.
Doing so, Traefik will create routes for containers that match these constraints only.

??? example "Containers with the api Tag"

    ```toml
    constraints = ["tag==api"]
    ```

??? example "Containers without the api Tag"

    ```toml
    constraints = ["tag!=api"]
    ```
    
??? example "Containers with tags starting with 'us-'"

    ```toml
    constraints = ["tag==us-*"]
    ```

??? example "Multiple constraints"

    ```toml
    # Multiple constraints
    #   - "tag==" must match with at least one tag
    #   - "tag!=" must match with none of tags
    constraints = ["tag!=us-*", "tag!=asia-*"]
    ```

??? note "List of Providers that Support Constraints"

    - Docker
    - Consul K/V
    - BoltDB
    - Zookeeper
    - ECS
    - Etcd
    - Consul Catalog
    - Rancher
    - Marathon
    - Kubernetes (using a provider-specific mechanism based on label selectors)
    
!!! note

    The constraint option belongs to the provider configuration itself.
   
    ??? example "Setting the Constraint Options for Docker"
   
        ```toml
        [providers]
         [providers.docker]
            constraints = ["tag==api"]
        ```
