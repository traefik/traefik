# Traefik & Rancher

A Story of Labels, Services & Container
{: .subtitle }

![Rancher](../assets/img/providers/rancher.png)

Attach labels to your services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Rancher & Deploying / Exposing Services"

    Enabling the rancher provider

    ```toml
    [provider.rancher]
    ```

    Attaching labels to services

    ```yaml
    labels:
      - traefik.http.services.my-service.rule=Host(my-domain)
    ```

## Provider Configuration Options

!!! tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the configuration reference:
    
    ```toml
    ################################################################
    # Rancher Provider
    ################################################################
    
    # Enable Rancher Provider.
    [rancher]
    
    # Expose Rancher services by default in Traefik.
    #
    # Optional
    #
    ExposedByDefault = "true"
    
    # Enable watch Rancher changes.
    #
    # Optional
    #
    watch = true
    
    # Filter services with unhealthy states and inactive states.
    #
    # Optional
    #
    EnableServiceHealthFilter = true
    
    # Defines the polling interval (in seconds).
    #
    # Optional
    #
    RefreshSeconds = true
    
    # Poll the Rancher metadata service for changes every `rancher.refreshSeconds`, which is less accurate
    #
    # Optional
    #
    IntervalPoll = false
    
    # Prefix used for accessing the Rancher metadata service
    #
    # Optional
    #
    Prefix = 15
    ```

### `ExposedByDefault`

_Optional, Default=true_

Expose Rancher services by default in Traefik.
If set to false, services that don't have a `traefik.enable=true` label will be ignored from the resulting routing configuration.

### `DefaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

The default host rule for all services.

For a given container if no routing rule was defined by a label, it is defined by this defaultRule instead.
It must be a valid [Go template](https://golang.org/pkg/text/template/),
augmented with the [sprig template functions](http://masterminds.github.io/sprig/).
The service name can be accessed as the `Name` identifier,
and the template has access to all the labels defined on this container.

```toml tab="File"
[rancher]
defaultRule = ""
# ...
```

```txt tab="CLI"
--providers.rancher
--providers.rancher.defaultRule="Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
```

This option can be overridden on a container basis with the `traefik.http.routers.Router1.rule` label.

### `EnableServiceHealthFilter`

_Optional, Default=true_

Filter services with unhealthy states and inactive states.

### `RefreshSeconds`

_Optional, Default=15_

Defines the polling interval (in seconds).

### `IntervalPoll`

_Optional, Default=false_

Poll the Rancher metadata service for changes every `rancher.refreshSeconds`,
which is less accurate than the default long polling technique which will provide near instantaneous updates to Traefik.

### `Prefix`

_Optional, Default=/latest_

Prefix used for accessing the Rancher metadata service

### General

Traefik creates, for each rancher service, a corresponding [service](../routing/services/index.md) and [router](../routing/routers/index.md).

The Service automatically gets a server per container in this rancher service, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the container, add labels starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.
For example, to change the rule, you could add the label `traefik.http.routers.my-container.rule=Host(my-domain)`.

Every [Router](../routing/routers/index.md) parameter can be updated this way.

### Services

To update the configuration of the Service automatically attached to the container, add labels starting with `traefik.http.services.{name-of-your-choice}.`,
followed by the option you want to change. For example, to change the load balancer method,
you'd add the label `traefik.http.services.{name-of-your-choice}.loadbalancer.method=drr`.

Every [Service](../routing/services/index.md) parameter can be updated this way.

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.
For example, to declare a middleware [`schemeredirect`](../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.schemeredirect.scheme: https`.

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    labels:
     - traefik.http.middlewares.my-redirect.schemeredirect.scheme=https
     - traefik.http.routers.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

More information about available middlewares in the dedicated [middlewares section](../middlewares/overview.md).

### Specific Options

#### `traefik.enable`

You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### `traefik.tags`

Sets the tags for [constraints filtering](./overview.md#constraints-configuration).

#### Port Lookup

Traefik is now capable of detecting the port to use, by following the default rancher flow.
That means, if you just expose lets say port :1337 on the rancher ui, traefik will pick up this port and use it.
