# Traefik & Rancher

A Story of Labels, Services & Containers
{: .subtitle }

![Rancher](../assets/img/providers/rancher.png)

Attach labels to your services and let Traefik do the rest!

!!! important
    This provider is specific to Rancher 1.x.
    Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
    As such, Rancher 2.x users should utilize the [Kubernetes provider](./kubernetes-crd.md) directly.

## Configuration Examples

??? example "Configuring Rancher & Deploying / Exposing Services"

    Enabling the rancher provider

    ```toml
    [Providers.Rancher]
    ```

    Attaching labels to services

    ```yaml
    labels:
      - traefik.http.services.my-service.rule=Host(`my-domain`)
    ```

## Provider Configuration Options

??? tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the configuration reference:
    
    ```toml
    --8<-- "content/providers/rancher.toml"
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
[Providers.Rancher]
defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
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

### `constraints`

_Optional, Default=""_

Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.
That is to say, if none of the container's labels match the expression, no route for the container is created.
If the expression is empty, all detected containers are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegexp("key", "value")` functions, as well as the usual boolean logic, as shown in examples below.

??? example "Constraints Expression Examples"

    ```toml
    # Includes only containers having a label with key `a.label.name` and value `foo`
    constraints = "Label(`a.label.name`, `foo`)"
    ```
    
    ```toml
    # Excludes containers having any label with key `a.label.name` and value `foo`
    constraints = "!Label(`a.label.name`, `value`)"
    ```
    
    ```toml
    # With logical AND.
    constraints = "Label(`a.label.name`, `valueA`) && Label(`another.label.name`, `valueB`)"
    ```
    
    ```toml
    # With logical OR.
    constraints = "Label(`a.label.name`, `valueA`) || Label(`another.label.name`, `valueB`)"
    ```
    
    ```toml
    # With logical AND and OR, with precedence set by parentheses.
    constraints = "Label(`a.label.name`, `valueA`) && (Label(`another.label.name`, `valueB`) || Label(`yet.another.label.name`, `valueC`))"
    ```
    
    ```toml
    # Includes only containers having a label with key `a.label.name` and a value matching the `a.+` regular expression.
    constraints = "LabelRegexp(`a.label.name`, `a.+`)"
    ```

## Routing Configuration Options

### General

Traefik creates, for each rancher service, a corresponding [service](../routing/services/index.md) and [router](../routing/routers/index.md).

The Service automatically gets a server per container in this rancher service, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the container, add labels starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.
For example, to change the rule, you could add the label `traefik.http.routers.my-container.rule=Host(my-domain)`.

Every [Router](../routing/routers/index.md) parameter can be updated this way.

### Services

To update the configuration of the Service automatically attached to the container, add labels starting with `traefik.http.services.{name-of-your-choice}.`,
followed by the option you want to change. For example, to change the passhostheader behavior,
you'd add the label `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

Every [Service](../routing/services/index.md) parameter can be updated this way.

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.
For example, to declare a middleware [`redirectscheme`](../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    labels:
     - traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
     - traefik.http.routers.my-container.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

More information about available middlewares in the dedicated [middlewares section](../middlewares/overview.md).

### Specific Options

#### `traefik.enable`

You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### Port Lookup

Traefik is now capable of detecting the port to use, by following the default rancher flow.
That means, if you just expose lets say port :1337 on the rancher ui, traefik will pick up this port and use it.
