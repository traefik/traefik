---
title: "Nomad Service Discovery"
description: "Learn how to use Nomad as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Nomad Service Discovery

## Configuration Example

You can enable the Nomad provider with as detailed below:

```yaml tab="File (YAML)"
providers:
  nomad: {}
```

```toml tab="File (TOML)"
[providers.nomad]
```

```bash tab="CLI"
--providers.nomad=true
```

Attaching tags to services:

```json
...
service {
  name = "myService"
  tags = [
    "traefik.http.routers.my-router.rule=Host(`example.com`)",
  ]
}
...
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.nomad.namespaces` | Defines the namespaces in which the nomad services will be discovered.|  ""     | No   |
| `providers.nomad.refreshInterval` | Defines the polling interval. This option is ignored when the `watch` option is enabled |  15s     | No   |
| `providers.nomad.watch` | Enables the watch mode to refresh the configuration on a per-event basis. |  false     | No   |
| `providers.nomad.throttleDuration` | Defines how often the provider is allowed to handle service events from Nomad. This option is only compatible when the `watch` option is enabled |  0s     | No   |
| `providers.nomad.defaultRule` | The Default Host rule for all services. See [here](#defaultrule) for more information |   ```"Host(`{{ normalize .Name }}`)"```   | No   |
| `providers.nomad.constraints` | Defines an expression that Traefik matches against the container labels to determine whether to create any route for that container. See [here](#constraints) for more information.  |  ""   | No   |
| `providers.nomad.exposedByDefault` | Expose Nomad services by default in Traefik. If set to `false`, services that do not have a `traefik.enable=true` tag will be ignored from the resulting routing configuration. See [here](../overview.md#restrict-the-scope-of-service-discovery) for additional information |  true    | No   |
| `providers.nomad.allowEmptyServices` |  Instructs the provider to create any [servers load balancer](../../../../routing/services/index.md#servers-load-balancer) defined for Docker containers regardless of the [healthiness](https://docs.docker.com/engine/reference/builder/#healthcheck) of the corresponding containers. |  false   | No   |
| `providers.nomad.prefix` | Defines the prefix for Nomad service tags defining Traefik labels. | `traefik`     | yes   |
| `providers.nomad.stale` | Instructs Traefik to use stale consistency for Nomad service API reads. See [here](#stale) for more information | false   | No   |
| `providers.nomad.endpoint.address` | Defines the Address of the Nomad server. | `http://127.0.0.1:4646`  | No   |
| `providers.nomad.endpoint.token` | Defines a per-request ACL token if Nomad ACLs are enabled. See [here](#token) for more information | ""  | No   |
| `providers.nomad.endpoint.endpointWaitTime` | Defines a duration for which a `watch` can block. If not provided, the agent default values will be used. | ""  | No   |
| `providers.nomad.endpoint.tls` | Defines the TLS configuration used for the secure connection to the Nomad APi.  |  -   | No   |
| `providers.nomad.endpoint.tls.ca` | Defines the path to the certificate authority used for the secure connection to the Nomad API, it defaults to the system bundle.  |   ""  | No   |
| `providers.nomad.endpoint.tls.cert` | Defines the path to the public certificate used for the secure connection to the Nomad API. When using this option, setting the `key` option is required. | '"  | Yes   |
| `providers.nomad.endpoint.tls.key` | Defines the path to the private key used for the secure connection to the Nomad API. When using this option, setting the `cert` option is required. |  ""   | Yes   |
| `providers.nomad.endpoint.tls.insecureSkipVerify` | Instructs the provider to accept any certificate presented by Nomad when establishing a TLS connection, regardless of the hostnames the certificate covers. | false   | No   |

### `namespaces`

The `namespaces` option defines the namespaces in which the nomad services will be discovered.
When using the `namespaces` option, the discovered object names will be suffixed as shown below:

```text
<resource-name>@nomad-<namespace>
```

!!! warning
  
    One should only define either the `namespaces` option or the `namespace` option.

```yaml tab="File (YAML)"
providers:
  nomad:
    namespaces:
      - "ns1"
      - "ns2"
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  namespaces = ["ns1", "ns2"]
  # ...
```

```bash tab="CLI"
--providers.nomad.namespaces=ns1,ns2
# ...
```

### `stale`

Use stale consistency for Nomad service API reads.

!!! note ""

    This makes reads very fast and scalable at the cost of a higher likelihood of stale values.

    For more information, see the Nomad [documentation on consistency](https://www.nomadproject.io/api-docs#consistency-modes).

```yaml tab="File (YAML)"
providers:
  nomad:
    stale: true
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  stale = true
  # ...
```

```bash tab="CLI"
--providers.nomad.stale=true
# ...
```

### `token`

Token is used to provide a per-request ACL token, if Nomad ACLs are enabled.
The appropriate ACL privilege for this token is 'read-job', as outlined in the [Nomad documentation on ACL](https://developer.hashicorp.com/nomad/tutorials/access-control/access-control-policies).

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      token: test
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  [providers.nomad.endpoint]
    token = "test"
    # ...
```

```bash tab="CLI"
--providers.nomad.endpoint.token=test
# ...
```

### `defaultRule`

The default host rule for all services.

For a given service, if no routing rule was defined by a tag, it is defined by this `defaultRule` instead.
The `defaultRule` must be set to a valid [Go template](https://pkg.go.dev/text/template/),
and can include [sprig template functions](https://masterminds.github.io/sprig/).
The service name can be accessed with the `Name` identifier,
and the template has access to all the labels (i.e. tags beginning with the `prefix`) defined on this service.

The option can be overridden on an instance basis with the `traefik.http.routers.{name-of-your-choice}.rule` tag.

```yaml tab="File (YAML)"
providers:
  nomad:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```bash tab="CLI"
--providers.nomad.defaultRule='Host(`{{ .Name }}.{{ index .Labels "customLabel"}}`)'
# ...
```

??? info "Default rule and Traefik service"

    The exposure of the Traefik container, combined with the default rule mechanism,
    can lead to create a router targeting itself in a loop.
    In this case, to prevent an infinite loop,
    Traefik adds an internal middleware to refuse the request if it comes from the same router.

### `constraints`

The `constraints` option can be set to an expression that Traefik matches against the service tags to determine whether
to create any route for that service. If none of the service tags match the expression, no route for that service is
created. If the expression is empty, all detected services are included.

The expression syntax is based on the ```Tag(`tag`)```, and ```TagRegex(`tag`)``` functions,
as well as the usual boolean logic, as shown in examples below.

!!! tip "Constraints key limitations"

    Note that `traefik.*` is a reserved label namespace for configuration and can not be used as a key for custom constraints.

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
  nomad:
    constraints: "Tag(`a.tag.name`)"
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  constraints = "Tag(`a.tag.name`)"
  # ...
```

```bash tab="CLI"
--providers.nomad.constraints="Tag(`a.tag.name`)"
# ...
```

For additional information, refer to [Restrict the Scope of Service Discovery](../overview.md#restrict-the-scope-of-service-discovery).

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/nomad.md).
