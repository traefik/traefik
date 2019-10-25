# Traefik & Consul Catalog

A Story of Labels, Services & Containers
{: .subtitle }

![Consul Catalog](../assets/img/providers/consul.png)

Attach labels to your services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Consul Catalog & Deploying / Exposing Services"

    Enabling the consulcatalog provider

    ```toml tab="File (TOML)"
    [providers.consulCatalog]
    ```
    
    ```yaml tab="File (YAML)"
    providers:
      consulCatalog: {}
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

### `exposedByDefault`

_Optional, Default=true_

```toml tab="File (TOML)"
[providers.consulCatalog]
  exposedByDefault = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
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
[providers.consulCatalog]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
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

### `refreshInterval`

_Optional, Default=15s_

```toml tab="File (TOML)"
[providers.consulCatalog]
  refreshInterval = "30s"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    refreshInterval: 30s
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.refreshInterval=30s
# ...
```

Defines the polling interval.

### `prefix`

_Optional, Default=/latest_

```toml tab="File (TOML)"
[providers.consulCatalog]
  prefix = "/test"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    prefix: "/test"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.prefix="/test"
# ...
```

Prefix used for accessing the Consul service metadata.

### `constraints`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulCatalog]
  constraints = "Label(`a.label.name`, `foo`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
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

### `endpoint`

Defines Consul Catalog Provider endpoint.

#### `address`

_Optional, Default="http://127.0.0.1:8500"_

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    address = "http://127.0.0.1:8500"
    # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      address: http://127.0.0.1:8500
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.address=http://127.0.0.1:8500
# ...
```

Defines the consul address endpoint.

#### `scheme`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    scheme = "https"
    # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      scheme: https
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.scheme=https
# ...
```

Defines the scheme used.

#### `datacenter`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    datacenter = "test"
    # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      datacenter: test
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.datacenter=test
# ...
```

Defines the DC.

#### `token`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    token = "test"
    # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      token: test
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.token=test
# ...
```

Defines the token.

#### `endpointWaitTime`

_Optional, Default=""_

If not provided, the agent default values will be used

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    endpointWaitTime = "15s"
    # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      endpointWaitTime: 15s
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.endpointWaitTime=15s
# ...
```

Defines the endpoint wait time.

#### `tls`

_Optional_

Defines TLS options for Consul Catalog Provider endpoint.

##### `ca`

Certificate Authority used for the secured connection to Consul Catalog.

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  ca = "path/to/ca.crt"
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        ca: path/to/ca.crt
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.ca=path/to/ca.crt
```

##### `caOptional`

Policy followed for the secured connection with TLS Client Authentication to Consul Catalog.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  caOptional = true
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        caOptional: true
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.caOptional=true
```

##### `cert`

Public certificate used for the secured connection to Consul Catalog.

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.cert=path/to/foo.cert
--providers.consulcatalog.endpoint.tls.key=path/to/foo.key
```

##### `key`

Private certificate used for the secured connection to Consul Catalog.

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.cert=path/to/foo.cert
--providers.consulcatalog.endpoint.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

If `insecureSkipVerify` is `true`, TLS for the connection to Consul Catalog accepts any certificate presented by the server and any host name in that certificate.

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  insecureSkipVerify = true
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        insecureSkipVerify: true
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.insecureSkipVerify=true
```
