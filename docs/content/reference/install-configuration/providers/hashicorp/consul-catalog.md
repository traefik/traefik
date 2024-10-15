---
title: "Traefik Consul Catalog Documentation"
description: "Learn how to use Consul Catalog as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Consul Catalog

Attach tags to your services and let Traefik do the rest!

## Configuration Example

Enabling the consul catalog provider

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint: 
      address: "127.0.0.1:8500"
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

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `consulCatalog.refreshInterval` | Defines the polling interval.|  15s    | No   |
| `consulCatalog.prefix` | Defines the prefix for Consul Catalog tags defining Traefik labels.|  traefik    | yes   |
| `consulCatalog.requireConsistent` | Forces the read to be fully consistent. See [here](#requireconsistent) for more information.|  false    | yes   |
| `consulCatalog.exposedByDefault` | Expose Consul Catalog services by default in Traefik. If set to `false`, services that do not have a `traefik.enable=true` tag will be ignored from the resulting routing configuration. See [here](../overview.md#restrict-the-scope-of-service-discovery)| true | no |
| `consulCatalog.defaultRule` | The Default Host rule for all services. See [here](#defaultrule) for more information |  `Host(`{{ normalize .Name }}`)`   | No   |
| `consulCatalog.connectAware` | Enable Consul Connect support. If set to `true`, Traefik will be enabled to communicate with Connect services.   | false   | No |
| `consulCatalog.connectByDefault` | Consider every service as Connect capable by default. If set to true, Traefik will consider every Consul Catalog service to be Connect capable by default. The option can be overridden on an instance basis with the traefik.consulcatalog.connect tag. | false   | No |
| `consulCatalog.serviceName` | Defines the name of the Traefik service in Consul Catalog. | "traefik"   | No |
| `consulCatalog.constraints` | Defines an expression that Traefik matches against the container labels to determine whether to create any route for that container. See [here](#constraints) for more information. | ""   | No |
| `consulCatalog.namespaces` | Defines the namespaces to query. See [here](#namespaces) for more information |  ""     | no   |
| `consulCatalog.stale` | Instruct Traefik to use stale consistency for catalog reads. |  false    | no   |
| `consulCatalog.cache` | Instruct Traefik to use local agent caching for catalog reads. |  false    | no   |
| `consulCatalog.cache` | Instruct Traefik to use local agent caching for catalog reads. |  false    | no   |
| `consulCatalog.endpoint` | Defines the Consul server endpoint. |  N/A    | yes   |
| `consulCatalog.endpoint.address` | Defines the address of the Consul server. |  127.0.0.1:8500    | no   |
| `consulCatalog.endpoint.scheme` | Defines the URI scheme for the Consul server. |  ""   | no   |
| `consulCatalog.endpoint.datacenter` | Defines the datacenter to use. If not provided in Traefik, Consul uses the default agent datacenter. |  ""   | no   |
| `consulCatalog.endpoint.token` |  Defines a per-request ACL token which overwrites the agent's default token. |  ""    | no   |
| `consulCatalog.endpoint.endpointWaitTime` |  Defines a duration for which a `watch` can block. If not provided, the agent default values will be used. |  ""    | no   |
| `consulCatalog.endpoint.httpAuth` | Defines authentication settings for the HTTP client using HTTP Basic Authentication. |  N/A    | no   |
| `consulCatalog.endpoint.httpAuth.username` | Defines the username to use for HTTP Basic Authentication. |  ""    | no   |
| `consulCatalog.endpoint.httpAuth.password` | Defines the password to use for HTTP Basic Authentication. |  ""    | no   |
| `consulCatalog.tls` | Defines the TLS configuration used for the secure connection to Consul Calatog  |  N/A   | No   |
| `consulCatalog.tls.ca` | Defines the path to the certificate authority used for the secure connection to Consul Calatog, it defaults to the system bundle.  |  N/A   | No   |
| `consulCatalog.tls.cert` | Defines the path to the public certificate used for the secure connection to Consul Calatog. When using this option, setting the `key` option is required. |  N/A   | Yes   |
| `consulCatalog.tls.key` | Defines the path to the private key used for the secure connection to Consul Catalog. When using this option, setting the `cert` option is required. |  N/A   | Yes   |
| `consulCatalog.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by Consul Catalog when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

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

??? info "Default rule and Traefik service"

    The exposure of the Traefik container, combined with the default rule mechanism,
    can lead to create a router targeting itself in a loop.
    In this case, to prevent an infinite loop,
    Traefik adds an internal middleware to refuse the request if it comes from the same router.

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

For additional information, refer to [Restrict the Scope of Service Discovery](../overview.md#restrict-the-scope-of-service-discovery).

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

### `strictChecks`

_Optional, Default="passing,warning"_

Define which [Consul Service health checks](https://developer.hashicorp.com/consul/docs/services/usage/checks#define-initial-health-check-status) are allowed to take on traffic.

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    strictChecks: 
      - "passing"
      - "warning"
    # ...
```

```toml tab="File (TOML)"
[providers.consulCatalog]
  strictChecks = ["passing", "warning"]
  # ...
```

```bash tab="CLI"
--providers.consulcatalog.strictChecks=passing,warning
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

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/consul-catalog.md).
