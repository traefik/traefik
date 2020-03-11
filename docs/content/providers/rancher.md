# Traefik & Rancher

A Story of Labels, Services & Containers
{: .subtitle }

![Rancher](../assets/img/providers/rancher.png)

Attach labels to your services and let Traefik do the rest!

!!! important "This provider is specific to Rancher 1.x."
    
    Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
    As such, Rancher 2.x users should utilize the [Kubernetes provider](./kubernetes-crd.md) directly.

## Configuration Examples

??? example "Configuring Rancher & Deploying / Exposing Services"

    Enabling the rancher provider

    ```toml tab="File (TOML)"
    [providers.rancher]
    ```
    
    ```yaml tab="File (YAML)"
    providers:
      rancher: {}
    ```
    
    ```bash tab="CLI"
    --providers.rancher=true
    ```

    Attaching labels to services

    ```yaml
    labels:
      - traefik.http.services.my-service.rule=Host(`example.com`)
    ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/rancher.md).

## Provider Configuration

??? tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the configuration reference:
    
    ```toml tab="File (TOML)"
    --8<-- "content/providers/rancher.toml"
    ```
    
    ```yaml tab="File (YAML)"
    --8<-- "content/providers/rancher.yml"
    ```
    
    ```bash tab="CLI"
    --8<-- "content/providers/rancher.txt"
    ```

### `exposedByDefault`

_Optional, Default=true_

```toml tab="File (TOML)"
[providers.rancher]
  exposedByDefault = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    exposedByDefault: false
    # ...
```

```bash tab="CLI"
--providers.rancher.exposedByDefault=false
# ...
```

Expose Rancher services by default in Traefik.
If set to false, services that don't have a `traefik.enable=true` label will be ignored from the resulting routing configuration.

See also [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

```toml tab="File (TOML)"
[providers.rancher]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```bash tab="CLI"
--providers.rancher.defaultRule=Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)
# ...
```

The default host rule for all services.

For a given container if no routing rule was defined by a label, it is defined by this defaultRule instead.
It must be a valid [Go template](https://golang.org/pkg/text/template/),
augmented with the [sprig template functions](http://masterminds.github.io/sprig/).
The service name can be accessed as the `Name` identifier,
and the template has access to all the labels defined on this container.

This option can be overridden on a container basis with the `traefik.http.routers.Router1.rule` label.

### `enableServiceHealthFilter`

_Optional, Default=true_

```toml tab="File (TOML)"
[providers.rancher]
  enableServiceHealthFilter = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    enableServiceHealthFilter: false
    # ...
```

```bash tab="CLI"
--providers.rancher.enableServiceHealthFilter=false
# ...
```

Filter services with unhealthy states and inactive states.

### `refreshSeconds`

_Optional, Default=15_

```toml tab="File (TOML)"
[providers.rancher]
  refreshSeconds = 30
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    refreshSeconds: 30
    # ...
```

```bash tab="CLI"
--providers.rancher.refreshSeconds=30
# ...
```

Defines the polling interval (in seconds).

### `intervalPoll`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.rancher]
  intervalPoll = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    intervalPoll: true
    # ...
```

```bash tab="CLI"
--providers.rancher.intervalPoll=true
# ...
```

Poll the Rancher metadata service for changes every `rancher.refreshSeconds`,
which is less accurate than the default long polling technique which will provide near instantaneous updates to Traefik.

### `prefix`

_Optional, Default=/latest_

```toml tab="File (TOML)"
[providers.rancher]
  prefix = "/test"
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    prefix: "/test"
    # ...
```

```bash tab="CLI"
--providers.rancher.prefix=/test
# ...
```

Prefix used for accessing the Rancher metadata service

### `constraints`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.rancher]
  constraints = "Label(`a.label.name`,`foo`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  rancher:
    constraints: "Label(`a.label.name`,`foo`)"
    # ...
```

```bash tab="CLI"
--providers.rancher.constraints=Label(`a.label.name`,`foo`)
# ...
```

Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container.
That is to say, if none of the container's labels match the expression, no route for the container is created.
If the expression is empty, all detected containers are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegex("key", "value")` functions, as well as the usual boolean logic, as shown in examples below.

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
    constraints = "LabelRegex(`a.label.name`, `a.+`)"
    ```

See also [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).
