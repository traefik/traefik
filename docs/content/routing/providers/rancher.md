# Traefik & Rancher

A Story of Labels, Services & Containers
{: .subtitle }

![Rancher](../../assets/img/providers/rancher.png)

Attach labels to your services and let Traefik do the rest!

!!! important "This provider is specific to Rancher 1.x."
    
    Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
    As such, Rancher 2.x users should utilize the [Kubernetes provider](./kubernetes-crd.md) directly.

## Routing Configuration

!!! info "Labels"
    
    - Labels are case insensitive.
    - The complete list of labels can be found [the reference page](../../reference/dynamic-configuration/rancher.md)

### General

Traefik creates, for each rancher service, a corresponding [service](../services/index.md) and [router](../routers/index.md).

The Service automatically gets a server per container in this rancher service, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the container, add labels starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the label ```traefik.http.routers.my-container.rule=Host(`mydomain.com`)```.

Every [Router](../routers/index.md) parameter can be updated this way.

### Services

To update the configuration of the Service automatically attached to the container, add labels starting with `traefik.http.services.{name-of-your-choice}.`, followed by the option you want to change.
For example, to change the passhostheader behavior,
you'd add the label `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

Every [Service](../services/index.md) parameter can be updated this way.

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    labels:
     - traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
     - traefik.http.routers.my-container.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/overview.md).

### Specific Provider Options

#### `traefik.enable`

You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### Port Lookup

Traefik is now capable of detecting the port to use, by following the default rancher flow.
That means, if you just expose lets say port :1337 on the rancher ui, traefik will pick up this port and use it.
