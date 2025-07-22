---
title: "Nomad Service Discovery"
description: "Learn how to use Nomad as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Nomad Service Discovery

A Story of Tags, Services & Nomads
{: .subtitle }

![Nomad Service Discovery](../assets/img/providers/nomad.png)

Attach tags to your Nomad services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Nomad & Deploying Services"

    Enabling the nomad provider

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

    ```
    ...
    service {
      name = "myService"
      tags = [
        "traefik.http.routers.my-router.rule=Host(`example.com`)",
      ]
    }
    ...
    ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/nomad.md).

## Provider Configuration

### `refreshInterval`

_Optional, Default=15s_

Defines the polling interval.

!!! note "This option is ignored when the [watch](#watch) mode is enabled."

```yaml tab="File (YAML)"
providers:
  nomad:
    refreshInterval: 30s
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  refreshInterval = "30s"
  # ...
```

```bash tab="CLI"
--providers.nomad.refreshInterval=30s
# ...
```

### `watch`

_Optional, Default=false_

Enables the watch mode to refresh the configuration on a per-event basis.

```yaml tab="File (YAML)"
providers:
  nomad:
    watch: true
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  watch = true
  # ...
```

```bash tab="CLI"
--providers.nomad.watch
# ...
```

### `throttleDuration`

_Optional, Default=0s_

The `throttleDuration` option defines how often the provider is allowed to handle service events from Nomad.
This prevents a Nomad cluster that updates many times per second from continuously changing your Traefik configuration.

If left empty, the provider does not apply any throttling and does not drop any Nomad service events.

The value of `throttleDuration` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

!!! warning "This option is only compatible with the [watch](#watch) mode."

```yaml tab="File (YAML)"
providers:
  nomad:
    throttleDuration: 2s
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  throttleDuration = "2s"
  # ...
```

```bash tab="CLI"
--providers.nomad.throttleDuration=2s
# ...
```

### `prefix`

_required, Default="traefik"_

The prefix for Nomad service tags defining Traefik labels.

```yaml tab="File (YAML)"
providers:
  nomad:
    prefix: test
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  prefix = "test"
  # ...
```

```bash tab="CLI"
--providers.nomad.prefix=test
# ...
```

### `stale`

_Optional, Default=false_

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

### `endpoint`

Defines the Nomad server endpoint.

#### `address`

Defines the address of the Nomad server.

_Optional, Default="http://127.0.0.1:4646"_

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      address: http://127.0.0.1:4646
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  [providers.nomad.endpoint]
    address = "http://127.0.0.1:4646"
    # ...
```

```bash tab="CLI"
--providers.nomad.endpoint.address=http://127.0.0.1:4646
# ...
```

#### `token`

_Optional, Default=""_

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

#### `endpointWaitTime`

_Optional, Default=""_

Limits the duration for which a Watch can block.
If not provided, the agent default values will be used.

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      endpointWaitTime: 15s
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  [providers.nomad.endpoint]
    endpointWaitTime = "15s"
    # ...
```

```bash tab="CLI"
--providers.nomad.endpoint.endpointwaittime=15s
# ...
```

#### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to the Nomad API.

##### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Nomad,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      tls:
        ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.nomad.endpoint.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.nomad.endpoint.tls.ca=path/to/ca.crt
```

##### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to the Nomad API.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.nomad.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.nomad.endpoint.tls.cert=path/to/foo.cert
--providers.nomad.endpoint.tls.key=path/to/foo.key
```

##### `key`

_Optional_

`key` is the path to the private key used for the secure connection to the Nomad API.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      tls:
        cert: path/to/foo.cert
        key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.nomad.endpoint.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.nomad.endpoint.tls.cert=path/to/foo.cert
--providers.nomad.endpoint.tls.key=path/to/foo.key
```

##### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Nomad accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  nomad:
    endpoint:
      tls:
        insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.nomad.endpoint.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.nomad.endpoint.tls.insecureskipverify=true
```

### `exposedByDefault`

_Optional, Default=true_

Expose Nomad services by default in Traefik.
If set to `false`, services that do not have a `traefik.enable=true` tag will be ignored from the resulting routing configuration.

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  nomad:
    exposedByDefault: false
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  exposedByDefault = false
  # ...
```

```bash tab="CLI"
--providers.nomad.exposedByDefault=false
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

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

### `namespaces`

??? warning "Deprecated in favor of the [`namespaces`](#namespaces) option."

    _Optional, Default=""_

    The `namespace` option defines the namespace in which the Nomad services will be discovered.
    
    !!! warning
    
        One should only define either the `namespaces` option or the `namespace` option.

    ```yaml tab="File (YAML)"
    providers:
      nomad:
        namespace: "production"
        # ...
    ```

    ```toml tab="File (TOML)"
    [providers.nomad]
      namespace = "production"
      # ...
    ```

    ```bash tab="CLI"
    --providers.nomad.namespace=production
    # ...
    ```

### `namespaces`

_Optional, Default=""_

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

### `allowEmptyServices`

_Optional, Default: false_

If the parameter is set to `true`,
it allows the creation of an empty [servers load balancer](../routing/services/index.md#servers-load-balancer) if the targeted Nomad service has no endpoints available. This results in a `503` HTTP response instead of a `404`.

```yaml tab="File (YAML)"
providers:
  nomad:
    allowEmptyServices: true
    # ...
```

```toml tab="File (TOML)"
[providers.nomad]
  allowEmptyServices = true
  # ...
```

```bash tab="CLI"
--providers.nomad.allowEmptyServices=true
```
