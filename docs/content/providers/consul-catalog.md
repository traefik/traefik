# Traefik & Consul Catalog

A Story of Labels, Services & Containers
{: .subtitle }

![Consul Catalog](../assets/img/providers/consul.png)

Attach labels to your services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Consul Catalog & Deploying / Exposing Services"

    Enabling the consulcatalog provider

    ```toml tab="File (TOML)"
    [providers.consulcatalog]
    ```
    
    ```yaml tab="File (YAML)"
    providers:
      consulcatalog: {}
    ```
    
    ```bash tab="CLI"
    --providers.consulcatalog=true
    ```

    Attaching labels to services

    ```yaml
    labels:
      - traefik.http.services.my-service.rule=Host(`mydomain.com`)
    ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/consul-catalog.md).

## Provider Configuration

??? tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the configuration reference:
    
    ```toml tab="File (TOML)"
    --8<-- "content/providers/consul-catalog.toml"
    ```
    
    ```yaml tab="File (YAML)"
    --8<-- "content/providers/consul-catalog.yml"
    ```
    
    ```bash tab="CLI"
    --8<-- "content/providers/consul-catalog.txt"
    ```

### `exposedByDefault`

_Optional, Default=true_

```toml tab="File (TOML)"
[providers.consulcatalog]
  exposedByDefault = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    exposedByDefault: false
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.exposedByDefault=false
# ...
```

Expose Consul Catalog services by default in Traefik.
If set to false, services that don't have a `traefik.enable=true` label will be ignored from the resulting routing configuration.

See also [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

```toml tab="File (TOML)"
[providers.consulcatalog]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.defaultRule="Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
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
[providers.consulcatalog]
  enableServiceHealthFilter = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    enableServiceHealthFilter: false
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.enableServiceHealthFilter=false
# ...
```

Filter services with unhealthy states and inactive states.

### `refreshSeconds`

_Optional, Default=15_

```toml tab="File (TOML)"
[providers.consulcatalog]
  refreshSeconds = 30
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    refreshSeconds: 30
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.refreshSeconds=30
# ...
```

Defines the polling interval (in seconds).

### `intervalPoll`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.consulcatalog]
  intervalPoll = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    intervalPoll: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.intervalPoll=true
# ...
```

Poll the Consul Catalog metadata service for changes every `consulcatalog.refreshSeconds`,
which is less accurate than the default long polling technique which will provide near instantaneous updates to Traefik.

### `prefix`

_Optional, Default=/latest_

```toml tab="File (TOML)"
[providers.consulcatalog]
  prefix = "/test"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    prefix: "/test"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.prefix="/test"
# ...
```

Prefix used for accessing the Consul Catalog service

### `constraints`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulcatalog]
  constraints = "Label(`a.label.name`, `foo`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulcatalog:
    constraints: "Label(`a.label.name`, `foo`)"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.constraints="Label(`a.label.name`, `foo`)"
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
