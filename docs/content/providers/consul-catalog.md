# Traefik & Consul Catalog

A Story of Tags, Services & Instances
{: .subtitle }

![Consul Catalog](../assets/img/providers/consul.png)

Attach tags to your services and let Traefik do the rest!

## Configuration Examples

??? example "Configuring Consul Catalog & Deploying / Exposing Services"

    Enabling the consul catalog provider

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

    Attaching tags to services

    ```yaml
    - traefik.http.routers.my-router.rule=Host(`example.com`)
    ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/consul-catalog.md).

## Provider Configuration

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

_required, Default="traefik"_

```toml tab="File (TOML)"
[providers.consulCatalog]
  prefix = "test"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    prefix: test
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.prefix=test
# ...
```

The prefix for Consul Catalog tags defining traefik labels.

### `requireConsistent`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.consulCatalog]
  requireConsistent = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    requireConsistent: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.requireConsistent=true
# ...
```

Forces the read to be fully consistent.

### `stale`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.consulCatalog]
  stale = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    stale: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.stale=true
# ...
```

Use stale consistency for catalog reads.

### `cache`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.consulCatalog]
  cache = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    cache: true
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.cache=true
# ...
```

Use local agent caching for catalog reads.

### `endpoint`

Defines the Consul server endpoint.

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

Defines the address of the Consul server.

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

Defines the URI scheme for the Consul server.

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

Defines the Data center to use.
If not provided, the default agent data center is used.

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

Token is used to provide a per-request ACL token which overrides the agent's default token.

#### `endpointWaitTime`

_Optional, Default=""_

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
--providers.consulcatalog.endpoint.endpointwaittime=15s
# ...
```

WaitTime limits how long a Watch will block.
If not provided, the agent default values will be used

#### `httpAuth`

_Optional_

Used to authenticate http client with HTTP Basic Authentication.

##### `username`

_Optional_

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.httpAuth]
  username = "test"
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      httpAuth:
        username: test
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.httpauth.username=test
```

Username to use for HTTP Basic Authentication

##### `password`

_Optional_

```toml tab="File (TOML)"
[providers.consulCatalog.endpoint.httpAuth]
  password = "test"
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    endpoint:
      httpAuth:
        password: test
```

```bash tab="CLI"
--providers.consulcatalog.endpoint.httpauth.password=test
```

Password to use for HTTP Basic Authentication

#### `tls`

_Optional_

Defines TLS options for Consul server endpoint.

##### `ca`

_Optional_

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

`ca` is the path to the CA certificate used for Consul communication, defaults to the system bundle if not specified.

##### `caOptional`

_Optional_

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
--providers.consulcatalog.endpoint.tls.caoptional=true
```

Policy followed for the secured connection with TLS Client Authentication to Consul.
Requires `tls.ca` to be defined.

- `true`: VerifyClientCertIfGiven
- `false`: RequireAndVerifyClientCert
- if `tls.ca` is undefined NoClientCert

##### `cert`

_Optional_

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

`cert` is the path to the public certificate for Consul communication.
If this is set then you need to also set `key.

##### `key`

_Optional_

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

`key` is the path to the private key for Consul communication.
If this is set then you need to also set `cert`.

##### `insecureSkipVerify`

_Optional_

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
--providers.consulcatalog.endpoint.tls.insecureskipverify=true
```

If `insecureSkipVerify` is `true`, TLS for the connection to Consul server accepts any certificate presented by the server and any host name in that certificate.

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
If set to false, services that don't have a `traefik.enable=true` tag will be ignored from the resulting routing configuration.

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

For a given service if no routing rule was defined by a tag, it is defined by this defaultRule instead.
It must be a valid [Go template](https://golang.org/pkg/text/template/),
augmented with the [sprig template functions](http://masterminds.github.io/sprig/).
The service name can be accessed as the `Name` identifier,
and the template has access to all the labels (i.e. tags beginning with the `prefix`) defined on this service.

The option can be overridden on an instance basis with the `traefik.http.routers.{name-of-your-choice}.rule` tag.

### `constraints`

_Optional, Default=""_

```toml tab="File (TOML)"
[providers.consulCatalog]
  constraints = "Tag(`a.tag.name`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  consulCatalog:
    constraints: "Tag(`a.tag.name`)"
    # ...
```

```bash tab="CLI"
--providers.consulcatalog.constraints="Tag(`a.tag.name`)"
# ...
```

Constraints is an expression that Traefik matches against the service's tags to determine whether to create any route for that service.
That is to say, if none of the service's tags match the expression, no route for that service is created.
If the expression is empty, all detected services are included.

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

See also [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).
