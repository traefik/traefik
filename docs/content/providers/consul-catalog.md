---
title: "Consul Catalog Configuration Discovery"
description: "Learn how to use Consul Catalog as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Consul Catalog

A Story of Tags, Services & Instances
{: .subtitle }

![Consul Catalog](../assets/img/providers/consul.png)

Attach tags to your services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Consul Catalog & Deploying / Exposing Services"

    Enabling the consul catalog provider

    ```yaml tab="File (YAML)"
    providers:
      consulCatalog: {}
    ```

    ```toml tab="File (TOML)"
    [providers.consulCatalog]
    ```

    ```bash tab="CLI"
    --providers.consulcatalog=true
    ```

    Attaching tags to services

    ```yaml
    - traefik.http.routers.my-router.rule=Host(`example.com`)
    ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/consul-catalog.md).

## Provider Configuration

### `refreshInterval`

_Optional, Default=15s_

Defines the polling interval.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    refreshInterval: 30s
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  refreshInterval = "30s"
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.refreshInterval=30s
# ...
```

### `prefix`

_required, Default="traefik"_

The prefix for Consul Catalog tags defining Traefik labels.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    prefix: test
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  prefix = "test"
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.prefix=test
# ...
```

### `requireConsistent`

_Optional, Default=false_

Forces the read to be fully consistent.

!!! note ""

    It is more expensive due to an extra round-trip but prevents ever performing a stale read.

    For more information, see the consul [documentation on consistency](https://www.consul.io/api-docs/features/consistency).

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    requireConsistent: true
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  requireConsistent = true
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.requireConsistent=true
# ...
```

### `stale`

_Optional, Default=false_

Use stale consistency for catalog reads.

!!! note ""

    This makes reads very fast and scalable at the cost of a higher likelihood of stale values.

    For more information, see the consul [documentation on consistency](https://www.consul.io/api-docs/features/consistency).

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    stale: true
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  stale = true
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.stale=true
# ...
```

### `cache`

_Optional, Default=false_

Use local agent caching for catalog reads.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    cache: true
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  cache = true
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.cache=true
# ...
```

### `endpoint`

Defines the Consul server endpoint.

#### `address`

Defines the address of the Consul server.

_Optional, Default="127.0.0.1:8500"_

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      address: 127.0.0.1:8500
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    address = "127.0.0.1:8500"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.address=127.0.0.1:8500
# ...
```

#### `scheme`

_Optional, Default=""_

Defines the URI scheme for the Consul server.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      scheme: https
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    scheme = "https"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.scheme=https
# ...
```

#### `datacenter`

_Optional, Default=""_

Defines the datacenter to use.
If not provided in Traefik, Consul uses the default agent datacenter.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      datacenter: test
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    datacenter = "test"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.datacenter=test
# ...
```

#### `token`

_Optional, Default=""_

Token is used to provide a per-request ACL token which overwrites the agent's default token.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      token: test
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    token = "test"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.token=test
# ...
```

#### `endpointWaitTime`

_Optional, Default=""_

Limits the duration for which a Watch can block.
If not provided, the agent default values will be used.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      endpointWaitTime: 15s
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  [providers.consulCatalog.endpoint]
    endpointWaitTime = "15s"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.endpointwaittime=15s
# ...
```

#### `httpAuth`

_Optional_

Used to authenticate the HTTP client using HTTP Basic Authentication.

##### `username`

_Optional, Default=""_

Username to use for HTTP Basic Authentication.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      httpAuth:
        username: test
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.httpAuth]
  username = "test"
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.httpauth.username=test
```

##### `password`

_Optional, Default=""_

Password to use for HTTP Basic Authentication.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      httpAuth:
        password: test
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.httpAuth]
  password = "test"
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.httpauth.password=test
```

#### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to Consul Catalog.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Consul Catalog,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to Consul Catalog.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.cert=path/to/foo.cert
--providers.consulcatalog.endpoint.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to Consul Catalog.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.cert=path/to/foo.cert
--providers.consulcatalog.endpoint.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Consul accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.tls.insecureskipverify=true
```

### `exposedByDefault`

_Optional, Default=true_

Expose Consul Catalog services by default in Traefik.
If set to `false`, services that don't have a `traefik.enable=true` tag will be ignored from the resulting routing configuration.

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    exposedByDefault: false
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  exposedByDefault = false
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.exposedByDefault=false
# ...
```

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

The default host rule for all services.

For a given service, if no routing rule was defined by a tag, it is defined by this `defaultRule` instead.
The `defaultRule` must be set to a valid [Go template](https://pkg.go.dev/text/template/),
and can include [sprig template functions](https://masterminds.github.io/sprig/).
The service name can be accessed with the `Name` identifier,
and the template has access to all the labels (i.e. tags beginning with the `prefix`) defined on this service.

The option can be overridden on an instance basis with the `traefik.http.routers.{name-of-your-choice}.rule` tag.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.defaultRule=Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)
# ...
```

### `connectAware`

_Optional, Default=false_

Enable Consul Connect support.
If set to `true`, Traefik will be enabled to communicate with Connect services.

```toml tab="File (TOML)"
[providers.consulCatalog]
  connectAware = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    connectAware: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.connectAware=true
# ...
```

### `connectByDefault`

_Optional, Default=false_

Consider every service as Connect capable by default.
If set to `true`, Traefik will consider every Consul Catalog service to be Connect capable by default.
The option can be overridden on an instance basis with the `traefik.consulcatalog.connect` tag.

```toml tab="File (TOML)"
[providers.consulCatalog]
  connectByDefault = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    connectByDefault: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.connectByDefault=true
# ...
```

### `serviceName`

_Optional, Default="traefik"_

Name of the Traefik service in Consul Catalog.

```toml tab="File (TOML)"
[providers.consulCatalog]
  serviceName = "test"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    serviceName: test
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.serviceName=test
# ...
```

### `constraints`

_Optional, Default=""_

The `constraints` option can be set to an expression that Traefik matches against the service tags to determine whether
to create any route for that service. If none of the service tags match the expression, no route for that service is
created. If the expression is empty, all detected services are included.

The expression syntax is based on the ```Tag(`tag`)```, and ```TagRegex(`tag`)``` functions,
as well as the usual boolean logic, as shown in examples below.

??? example "Constraints Expression Examples"

    ```toml
    # Includes only services having the tag `a.tag.name=foo`
    constraints = "Tag(`a.tag.name=foo`)"
    ```

    ```toml
    # Excludes services having any tag `a.tag.name=foo`
    constraints = "!Tag(`a.tag.name=foo`)"
    ```

    ```toml
    # With logical AND.
    constraints = "Tag(`a.tag.name`) && Tag(`another.tag.name`)"
    ```

    ```toml
    # With logical OR.
    constraints = "Tag(`a.tag.name`) || Tag(`another.tag.name`)"
    ```

    ```toml
    # With logical AND and OR, with precedence set by parentheses.
    constraints = "Tag(`a.tag.name`) && (Tag(`another.tag.name`) || Tag(`yet.another.tag.name`))"
    ```

    ```toml
    # Includes only services having a tag matching the `a\.tag\.t.+` regular expression.
    constraints = "TagRegex(`a\.tag\.t.+`)"
    ```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    constraints: "Tag(`a.tag.name`)"
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  constraints = "Tag(`a.tag.name`)"
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.constraints="Tag(`a.tag.name`)"
# ...
```

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

### `namespace`

??? warning "Deprecated in favor of the [`namespaces`](#namespaces) option."

    _Optional, Default=""_
    
    The `namespace` option defines the namespace in which the consul catalog services will be discovered.
    
    !!! warning
    
        The namespace option only works with [Consul Enterprise](https://www.consul.io/docs/enterprise),
        which provides the [Namespaces](https://www.consul.io/docs/enterprise/namespaces) feature.
    
    !!! warning
    
        One should only define either the `namespaces` option or the `namespace` option.
    
    ```yaml tab="File (YAML)"
    providers:
      consulCatalog:
        namespace: "production" 
        # ...
    ```
    
    ```toml tab="File (TOML)"
    [providers.consulCatalog]
      namespace = "production"
      # ...
    ```
    
    ```bash tab="CLI"
    --providers.consulcatalog.namespace=production
    # ...
    ```

### `namespaces`

_Optional, Default=""_

The `namespaces` option defines the namespaces in which the consul catalog services will be discovered.
When using the `namespaces` option, the discovered configuration object names will be suffixed as shown below:

```text
<resource-name>@consulcatalog-<namespace>
```

!!! warning

    The namespaces option only works with [Consul Enterprise](https://www.consul.io/docs/enterprise),
    which provides the [Namespaces](https://www.consul.io/docs/enterprise/namespaces) feature.

!!! warning

    One should only define either the `namespaces` option or the `namespace` option.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    namespaces: 
      - "ns1"
      - "ns2"
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  namespaces = ["ns1", "ns2"]
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.namespaces=ns1,ns2
# ...
```

### `watch`

_Optional, Default=false_

When set to `true`, watches for Consul changes ([Consul watches checks](https://www.consul.io/docs/dynamic-app-config/watches#checks)).

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    watch: true
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  watch = true
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.watch=true
# ...
```
